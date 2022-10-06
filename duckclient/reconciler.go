package duckclient

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type DuckReconciler interface {
	SetClient(c client.Client)
	SetLister(l Lister)
	// Reconcile performs a full reconciliation for the object referred to by the Request.
	// The Controller will requeue the Request to be processed again if an error is non-nil or
	// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
	Reconcile(context.Context, reconcile.Request) (reconcile.Result, error)
}

type ReconcilerBuilder func() DuckReconciler
