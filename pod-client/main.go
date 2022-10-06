package main

import (
	"context"
	"fmt"
	corev1alpha1 "github.com/tamalsaha/duckdemo/apis/core/v1alpha1"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2/klogr"
	"kmodules.xyz/client-go/client/duck"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func NewClient() (client.Client, error) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = corev1alpha1.AddToScheme(scheme)

	ctrl.SetLogger(klogr.New())
	cfg := ctrl.GetConfigOrDie()
	cfg.QPS = 100
	cfg.Burst = 100

	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		return nil, err
	}

	return client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: mapper,
		//Opts: client.WarningHandlerOptions{
		//	SuppressWarnings:   false,
		//	AllowDuplicateLogs: false,
		//},
	})
}

func main() {
	if err := useKubebuilderClient(); err != nil {
		panic(err)
	}
}

func useKubebuilderClient() error {
	fmt.Println("Using kubebuilder client")
	kc, err := NewClient()
	if err != nil {
		return err
	}

	cc, err := duck.NewClient().
		ForDuckType(&corev1alpha1.MyPod{}).
		WithUnderlyingType(apps.SchemeGroupVersion.WithKind("Deployment")).
		Build(kc)
	if err != nil {
		return err
	}

	var appobj corev1alpha1.MyPod
	err = cc.Get(context.TODO(), client.ObjectKey{
		Namespace: "kube-system",
		Name:      "coredns",
	}, &appobj)
	if err != nil {
		return err
	}
	fmt.Println(appobj.UID)

	// var applist apps.DeploymentList
	var applist corev1alpha1.MyPodList
	err = cc.List(context.TODO(), &applist)
	if err != nil {
		return err
	}
	fmt.Println()
	for _, db := range applist.Items {
		fmt.Println(client.ObjectKeyFromObject(&db))
	}
	return nil
}
