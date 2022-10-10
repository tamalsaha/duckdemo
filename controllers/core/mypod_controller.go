/*
Copyright 2022.

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

package core

import (
	"context"
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	corev1alpha1 "github.com/tamalsaha/duckdemo/apis/core/v1alpha1"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"kmodules.xyz/client-go/client/duck"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// MyPodReconciler reconciles a MyPod object
type MyPodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var _ duck.Reconciler = &MyPodReconciler{}

//+kubebuilder:rbac:groups=core.duck.dev,resources=mypods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.duck.dev,resources=mypods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.duck.dev,resources=mypods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyPod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *MyPodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// TODO(user): your logic here
	var mypod corev1alpha1.MyPod
	if err := r.Get(ctx, req.NamespacedName, &mypod); err != nil {
		log.Error(err, "unable to fetch CronJob")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	//images := sets.NewString()
	//for _, c := range mypod.Spec.Template.Spec.Containers {
	//	images.Insert(c.Image)
	//}
	//for _, c := range mypod.Spec.Template.Spec.InitContainers {
	//	images.Insert(c.Image)
	//}
	//for _, c := range mypod.Spec.Template.Spec.EphemeralContainers {
	//	images.Insert(c.Image)
	//}

	// fmt.Println(images.List())

	sel, err := metav1.LabelSelectorAsSelector(mypod.Spec.Selector)
	if err != nil {
		return ctrl.Result{}, err
	}

	var pods corev1.PodList
	err = r.List(context.TODO(), &pods,
		client.InNamespace(mypod.Namespace),
		client.MatchingLabelsSelector{Selector: sel})
	if err != nil {
		return ctrl.Result{}, err
	}

	fnRef := func(c corev1.ContainerStatus) (string, error) {
		imageRef, err := name.ParseReference(c.Image)
		if err != nil {
			return "", err
		}
		imageIDRef, err := name.ParseReference(c.ImageID)
		if err != nil {
			return "", err
		}
		if imageRef.Context() != imageIDRef.Context() {
			return c.Image, nil
		}
		return c.ImageID, nil
	}

	refs := sets.NewString()
	for _, pod := range pods.Items {
		for _, c := range pod.Status.ContainerStatuses {
			ref, err := fnRef(c)
			if err != nil {
				return ctrl.Result{}, err
			}
			refs.Insert(ref)
		}
		for _, c := range pod.Status.InitContainerStatuses {
			ref, err := fnRef(c)
			if err != nil {
				return ctrl.Result{}, err
			}
			refs.Insert(ref)
		}
		for _, c := range pod.Status.EphemeralContainerStatuses {
			ref, err := fnRef(c)
			if err != nil {
				return ctrl.Result{}, err
			}
			refs.Insert(ref)
		}
	}
	fmt.Println(refs.List())

	return ctrl.Result{}, nil
}

func (r *MyPodReconciler) InjectClient(c client.Client) {
	r.Client = c
}

func (r *MyPodReconciler) InjectScheme(s *runtime.Scheme) error {
	r.Scheme = s
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyPodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return duck.ControllerManagedBy(mgr).
		For(&corev1alpha1.MyPod{}).
		WithUnderlyingTypes(
			apps.SchemeGroupVersion.WithKind("Deployment"),
			apps.SchemeGroupVersion.WithKind("StatefulSet"),
			apps.SchemeGroupVersion.WithKind("DaemonSet"),
		).
		Complete(func() duck.Reconciler {
			return new(MyPodReconciler)
		})
}
