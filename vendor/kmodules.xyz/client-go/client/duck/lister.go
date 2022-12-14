/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package duck

import (
	"context"
	"strings"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// Lister knows how to list Kubernetes objects.
type Lister interface {
	// List retrieves list of objects for a given namespace and list options. On a
	// successful call, Items field in the list will be populated with the
	// result returned from the server.
	List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error

	Client(obj client.Object) (client.Client, error)
}

type ListerImpl struct {
	c       client.Client // reader?
	duckObj Object
	duckGVK schema.GroupVersionKind
	rawObjs []client.Object
}

var _ Lister = &ListerImpl{}

type ListerBuilder struct {
	cc *ListerImpl
}

func NewLister() *ListerBuilder {
	return &ListerBuilder{
		cc: new(ListerImpl),
	}
}

func (b *ListerBuilder) ForDuckType(obj Object) *ListerBuilder {
	b.cc.duckObj = obj
	return b
}

func (b *ListerBuilder) WithUnderlyingTypes(objs client.Object, rest ...client.Object) *ListerBuilder {
	b.cc.rawObjs = make([]client.Object, 0, len(rest)+1)
	b.cc.rawObjs = append(b.cc.rawObjs, objs)
	b.cc.rawObjs = append(b.cc.rawObjs, rest...)
	return b
}

func (b *ListerBuilder) Build(c client.Client) (Lister, error) {
	b.cc.c = c
	gvk, err := apiutil.GVKForObject(b.cc.duckObj, c.Scheme())
	if err != nil {
		return nil, err
	}
	b.cc.duckGVK = gvk
	return b.cc, nil
}

// Scheme returns the scheme this client is using.
func (d *ListerImpl) Scheme() *runtime.Scheme {
	return d.c.Scheme()
}

// RESTMapper returns the rest this client is using.
func (d *ListerImpl) RESTMapper() apimeta.RESTMapper {
	return d.c.RESTMapper()
}

func (d *ListerImpl) Client(obj client.Object) (client.Client, error) {
	return NewClient().
		ForDuckType(d.duckObj).
		WithUnderlyingType(obj).
		Build(d.c)
}

func (d *ListerImpl) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	gvk, err := apiutil.GVKForObject(list, d.c.Scheme())
	if err != nil {
		return err
	}
	if strings.HasSuffix(gvk.Kind, listType) && apimeta.IsListType(list) {
		gvk.Kind = gvk.Kind[:len(gvk.Kind)-4]
	}

	if gvk != d.duckGVK {
		return d.c.List(ctx, list, opts...)
	}

	var items []runtime.Object
	for _, rawObj := range d.rawObjs {
		_, isUnstructured := rawObj.(*unstructured.Unstructured)

		rawGVK := rawObj.GetObjectKind().GroupVersionKind()
		listGVK := rawGVK
		listGVK.Kind += listType

		var llo client.ObjectList
		if isUnstructured {
			var ul unstructured.UnstructuredList
			llo.GetObjectKind().SetGroupVersionKind(listGVK)
			err := d.c.List(ctx, &ul, opts...)
			if err != nil {
				return err
			}
			llo = &ul
		} else {
			ll, err := d.c.Scheme().New(listGVK)
			if err != nil {
				return err
			}
			llo = ll.(client.ObjectList)
			err = d.c.List(ctx, llo, opts...)
			if err != nil {
				return err
			}
		}

		list.SetResourceVersion(llo.GetResourceVersion())
		// list.SetContinue(llo.GetContinue())
		// list.SetSelfLink(llo.GetSelfLink())
		// list.SetRemainingItemCount(llo.GetRemainingItemCount())

		err = apimeta.EachListItem(llo, func(object runtime.Object) error {
			d2, err := d.c.Scheme().New(d.duckGVK)
			if err != nil {
				return err
			}
			dd := d2.(Object)
			err = dd.Duckify(object)
			if err != nil {
				return err
			}
			items = append(items, d2)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return apimeta.SetList(list, items)
}
