package configcenter

import (
	"context"
	"log"
	"raptor/constants"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ ConfigCenter = (*k8sConfigCenter)(nil)

func newK8sConfigCenter() ConfigCenter {
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	return &k8sConfigCenter{clientSet: clientset}
}

type k8sConfigCenter struct {
	clientSet *kubernetes.Clientset
}

func (k *k8sConfigCenter) Save(config Config) (bool, error) {
	configMap := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.ID,
			Namespace: constants.K8S_NAMESPACE,
			Labels:    map[string]string{"app": constants.K8S_APP_LABEL},
		},
		Data: map[string]string{constants.K8S_CONFIGMAP_CONTENT_KEY: config.Content},
	}
	_, err := k.clientSet.CoreV1().ConfigMaps(constants.K8S_NAMESPACE).Create(context.Background(), &configMap, metav1.CreateOptions{})
	return err == nil, err
}

func (k *k8sConfigCenter) Get(id string) (Config, error) {
	configMap, err := k.clientSet.CoreV1().ConfigMaps(constants.K8S_NAMESPACE).Get(context.Background(), id, metav1.GetOptions{})
	if err != nil {
		return Config{}, nil
	}
	return Config{ID: id, Content: configMap.Data[constants.K8S_CONFIGMAP_CONTENT_KEY]}, nil
}

func (k *k8sConfigCenter) OnChange(id string, handler func(config Config)) error {
	watcher, err := k.clientSet.CoreV1().ConfigMaps(constants.K8S_NAMESPACE).Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case e, _ := <-watcher.ResultChan():
				cm := e.Object.(*v1.ConfigMap)
				config := Config{
					ID:      cm.Name,
					Content: cm.Data[constants.K8S_CONFIGMAP_CONTENT_KEY],
				}
				handler(config)
			}
		}
	}()
	return nil
}
