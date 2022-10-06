package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type DuckType interface {
	Duckify(srcRaw runtime.Object) error
}

type DuckObject interface {
	metav1.Object
	runtime.Object
	DuckType
}
