package v1alpha1

import (
	"fmt"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func (dst *MyPod) Duckify(srcRaw runtime.Object) error {
	switch src := srcRaw.(type) {
	case *apps.Deployment:
		dst.TypeMeta = metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: apps.SchemeGroupVersion.String(),
		}
		dst.ObjectMeta = src.ObjectMeta
		dst.Spec.Template = src.Spec.Template
		return nil
	case *apps.StatefulSet:
		dst.TypeMeta = metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: apps.SchemeGroupVersion.String(),
		}
		dst.ObjectMeta = src.ObjectMeta
		dst.Spec.Template = src.Spec.Template
		return nil
	case *apps.DaemonSet:
		dst.TypeMeta = metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: apps.SchemeGroupVersion.String(),
		}
		dst.ObjectMeta = src.ObjectMeta
		dst.Spec.Template = src.Spec.Template
		return nil
	case *unstructured.Unstructured:
		return runtime.DefaultUnstructuredConverter.FromUnstructured(src.UnstructuredContent(), dst)
	default:
		return fmt.Errorf("unknown src type %T", srcRaw)
	}
}
