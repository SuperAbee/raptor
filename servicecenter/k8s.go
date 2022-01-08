package servicecenter

import (
	"context"
	"log"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ ServiceCenter = (*k8sServiceCenter)(nil)

func newK8sServiceCenter() ServiceCenter {
	return &k8sServiceCenter{}
}

type k8sServiceCenter struct {
}

func (n *k8sServiceCenter) Get(name string) Service {
	return Service{}
}

func K8SNamingTest() {
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	api := clientset.CoreV1()

	podList, err := api.Pods("default").List(context.Background(), v1.ListOptions{})

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", podList)
}
