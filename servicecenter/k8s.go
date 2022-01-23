package servicecenter

import (
	"context"
	"fmt"
	"log"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ ServiceCenter = (*k8sServiceCenter)(nil)

func newK8sServiceCenter() ServiceCenter {
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	return &k8sServiceCenter{clientSet: clientset}
}

type k8sServiceCenter struct {
	clientSet *kubernetes.Clientset
}

func (k *k8sServiceCenter) Register(param RegisterParam) (bool, error) {
	return true, nil
}

func (k *k8sServiceCenter) GetService(name string) (Service, error) {
	api := k.clientSet.CoreV1()

	labelSelector := v1.LabelSelector{MatchLabels: map[string]string{"app": name}}
	listOptions := v1.ListOptions{LabelSelector: labels.Set(labelSelector.MatchLabels).String()}

	serviceList, err := api.Services("default").List(context.Background(), listOptions)
	if err != nil {
		return Service{}, err
	}
	if len(serviceList.Items) == 0 || len(serviceList.Items[0].Spec.Ports) == 0 {
		return Service{}, fmt.Errorf("service with label: 'app: %s' not found", name)
	}
	podList, err := api.Pods("default").List(context.Background(), listOptions)
	if err != nil {
		return Service{}, err
	}

	var hosts []Instance
	for _, p := range podList.Items {
		hosts = append(hosts, Instance{Ip: p.Status.PodIP, Port: uint64(serviceList.Items[0].Spec.Ports[0].Port), Healthy: true})
	}

	return Service{Name: name, Hosts: hosts}, nil
}
