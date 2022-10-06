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
	"github.com/tamalsaha/duckdemo/duckclient"
	apps "k8s.io/api/apps/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/tamalsaha/duckdemo/apis/core/v1alpha1"
)

// MyPodReconciler reconciles a MyPod object
type MyPodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	lister duckclient.Lister
}

var _ duckclient.DuckReconciler = &MyPodReconciler{}

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
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

func (r *MyPodReconciler) SetClient(c client.Client) {
	r.Client = c
}

func (r *MyPodReconciler) SetScheme(s *runtime.Scheme) {
	r.Scheme = s
}

func (r *MyPodReconciler) SetLister(l duckclient.Lister) {
	r.lister = l
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyPodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return duckclient.ControllerManagedBy(mgr).
		For(&corev1alpha1.MyPod{}).
		WithUnderlyingTypes(
			apps.SchemeGroupVersion.WithKind("Deployment"),
			apps.SchemeGroupVersion.WithKind("StatefulSet"),
			apps.SchemeGroupVersion.WithKind("DaemonSet"),
		).
		Complete(func() duckclient.DuckReconciler {
			return new(MyPodReconciler)
		})
}
