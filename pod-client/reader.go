package main

import (
	"context"
	"github.com/tamalsaha/duckdemo/apis/core/v1alpha1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"strings"
)

type DuckReader struct {
	c       client.Client // reader?
	obj     v1alpha1.DuckObject
	duckGVK schema.GroupVersionKind
	rawGVK  schema.GroupVersionKind
}

var _ client.Reader = &DuckReader{}

func NewDuckReader(c client.Client, obj v1alpha1.DuckObject, rawGVK schema.GroupVersionKind) (client.Reader, error) {
	cc := &DuckReader{
		c:      c,
		obj:    obj,
		rawGVK: rawGVK,
	}
	gvk, err := apiutil.GVKForObject(obj, c.Scheme())
	if err != nil {
		return nil, err
	}
	cc.duckGVK = gvk
	return cc, nil
}

func (d DuckReader) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {

	//TODO implement me
	panic("implement me")
}

func (d DuckReader) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	gvk, err := apiutil.GVKForObject(list, d.c.Scheme())
	if err != nil {
		return err
	}
	if strings.HasSuffix(gvk.Kind, "List") && apimeta.IsListType(list) {
		gvk.Kind = gvk.Kind[:len(gvk.Kind)-4]
	}

	if gvk != d.duckGVK {
		return d.c.List(ctx, list, opts...)
	}

	listGVK := d.rawGVK
	listGVK.Kind += "List"

	ll, err := d.c.Scheme().New(listGVK)
	if err != nil {
		return err
	}
	llo := ll.(client.ObjectList)
	err = d.c.List(ctx, llo, opts...)
	if err != nil {
		return err
	}

	list.SetResourceVersion(llo.GetResourceVersion())
	list.SetContinue(llo.GetContinue())
	list.SetSelfLink(llo.GetSelfLink())
	list.SetRemainingItemCount(llo.GetRemainingItemCount())

	items := make([]runtime.Object, 0, apimeta.LenList(llo))
	err = apimeta.EachListItem(llo, func(object runtime.Object) error {
		d2, err := d.c.Scheme().New(d.duckGVK)
		if err != nil {
			return err
		}
		dd := d2.(v1alpha1.DuckType)
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
	return apimeta.SetList(list, items)
}
