package duckclient

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/tamalsaha/duckdemo/apis/core/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	errors2 "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Builder builds a Controller.
type Builder struct {
	forInput         ForInput
	ownsInput        []OwnsInput
	watchesInput     []WatchesInput
	mgr              manager.Manager
	globalPredicates []predicate.Predicate
	ctrl             controller.Controller
	ctrlOptions      controller.Options
	name             string
}

// ControllerManagedBy returns a new controller builder that will be started by the provided Manager.
func ControllerManagedBy(m manager.Manager) *Builder {
	return &Builder{mgr: m}
}

// ForInput represents the information set by For method.
type ForInput struct {
	object  v1alpha1.DuckObject
	rawGVKs []schema.GroupVersionKind
	opts    []builder.ForOption
	err     error
}

// For defines the type of Object being *reconciled*, and configures the ControllerManagedBy to respond to create / delete /
// update events by *reconciling the object*.
// This is the equivalent of calling
// Watches(&source.Kind{Type: apiType}, &handler.EnqueueRequestForObject{}).
func (blder *Builder) For(object v1alpha1.DuckObject, opts ...builder.ForOption) *Builder {
	if blder.forInput.object != nil {
		blder.forInput.err = errors2.NewAggregate([]error{
			blder.forInput.err,
			fmt.Errorf("For(...) should only be called once, could not assign multiple objects for reconciliation"),
		})
		return blder
	}
	blder.forInput.object = object
	blder.forInput.opts = opts

	return blder
}

func (blder *Builder) WithUnderlyingTypes(rawGVK schema.GroupVersionKind, rest ...schema.GroupVersionKind) *Builder {
	if len(blder.forInput.rawGVKs) > 0 {
		blder.forInput.err = errors2.NewAggregate([]error{
			blder.forInput.err,
			fmt.Errorf("WithUnderlyingTypes(...) should only be called once"),
		})
		return blder
	}

	gvks := make([]schema.GroupVersionKind, 0, len(rest)+1)
	gvks = append(gvks, rawGVK)
	for _, gvk := range rest {
		gvks = append(gvks, gvk)
	}
	blder.forInput.rawGVKs = gvks
	return blder
}

// OwnsInput represents the information set by Owns method.
type OwnsInput struct {
	object client.Object
	opts   []builder.OwnsOption
	err    error
}

// Owns defines types of Objects being *generated* by the ControllerManagedBy, and configures the ControllerManagedBy to respond to
// create / delete / update events by *reconciling the owner object*.  This is the equivalent of calling
// Watches(&source.Kind{Type: <ForType-forInput>}, &handler.EnqueueRequestForOwner{OwnerType: apiType, IsController: true}).
func (blder *Builder) Owns(object client.Object, opts ...builder.OwnsOption) *Builder {
	input := OwnsInput{
		object: object,
		opts:   opts,
	}
	if _, ok := object.(v1alpha1.DuckObject); ok {
		input.err = fmt.Errorf("Owns(...) can't be called on duck types")
	}

	blder.ownsInput = append(blder.ownsInput, input)
	return blder
}

// WatchesInput represents the information set by Watches method.
type WatchesInput struct {
	src          source.Source
	eventhandler handler.EventHandler
	opts         []builder.WatchesOption
}

// Watches exposes the lower-level ControllerManagedBy Watches functions through the builder.  Consider using
// Owns or For instead of Watches directly.
// Specified predicates are registered only for given source.
func (blder *Builder) Watches(src source.Source, eventhandler handler.EventHandler, opts ...builder.WatchesOption) *Builder {
	input := WatchesInput{
		src:          src,
		eventhandler: eventhandler,
		opts:         opts,
	}

	blder.watchesInput = append(blder.watchesInput, input)
	return blder
}

// WithEventFilter sets the event filters, to filter which create/update/delete/generic events eventually
// trigger reconciliations.  For example, filtering on whether the resource version has changed.
// Given predicate is added for all watched objects.
// Defaults to the empty list.
func (blder *Builder) WithEventFilter(p predicate.Predicate) *Builder {
	blder.globalPredicates = append(blder.globalPredicates, p)
	return blder
}

// WithOptions overrides the controller options use in doController. Defaults to empty.
func (blder *Builder) WithOptions(options controller.Options) *Builder {
	blder.ctrlOptions = options
	return blder
}

// WithLogConstructor overrides the controller options's LogConstructor.
func (blder *Builder) WithLogConstructor(logConstructor func(*reconcile.Request) logr.Logger) *Builder {
	blder.ctrlOptions.LogConstructor = logConstructor
	return blder
}

// Named sets the name of the controller to the given name.  The name shows up
// in metrics, among other things, and thus should be a prometheus compatible name
// (underscores and alphanumeric characters only).
//
// By default, controllers are named using the lowercase version of their kind.
func (blder *Builder) Named(name string) *Builder {
	blder.name = name
	return blder
}

// Complete builds the Application Controller.
func (blder *Builder) Complete(rb ReconcilerBuilder) error {
	if rb == nil {
		return fmt.Errorf("must provide a non-nil Reconciler")
	}
	if blder.mgr == nil {
		return fmt.Errorf("must provide a non-nil Manager")
	}
	if blder.forInput.err != nil {
		return blder.forInput.err
	}
	// Checking the reconcile type exist or not
	if blder.forInput.object == nil {
		return fmt.Errorf("must provide a duck type for reconciliation")
	}
	if len(blder.forInput.rawGVKs) == 0 {
		return fmt.Errorf("must provide underlying types for reconciliation")
	}

	lister, err := NewLister().
		ForDuckType(blder.forInput.object).
		WithUnderlyingType(blder.forInput.rawGVKs[0], blder.forInput.rawGVKs[1:]...).
		Build(blder.mgr.GetClient())
	if err != nil {
		return err
	}

	for _, rawGVK := range blder.forInput.rawGVKs {
		b2 := ctrl.NewControllerManagedBy(blder.mgr)
		b2 = b2.Named(blder.name + "-" + rawGVK.String())

		ll, err := blder.mgr.GetScheme().New(rawGVK)
		if err != nil {
			return err
		}
		llo := ll.(client.Object)
		b2 = b2.For(llo, blder.forInput.opts...)

		for _, own := range blder.ownsInput {
			b2 = b2.Owns(own.object, own.opts...)
		}
		for _, w := range blder.watchesInput {
			b2 = b2.Watches(w.src, w.eventhandler, w.opts...)
		}
		for _, p := range blder.globalPredicates {
			b2.WithEventFilter(p)
		}
		b2.WithOptions(blder.ctrlOptions)
		// b2.WithLogConstructor(blder.)

		r := rb()
		r.SetLister(lister)
		cc, err := lister.Client(rawGVK)
		if err != nil {
			return err
		}
		r.SetClient(cc)
		r.SetScheme(blder.mgr.GetScheme())
		if err = b2.Complete(r); err != nil {
			return err
		}
	}
	return nil
}