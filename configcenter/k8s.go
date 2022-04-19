package configcenter

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/labels"
	"log"
	"raptor/constants"
	"raptor/monitor"
	"strings"
	"time"

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
	return &k8sConfigCenter{clientSet: clientset, cache: newMemoryConfigCenter()}
}

type k8sConfigCenter struct {
	clientSet *kubernetes.Clientset
	cache ConfigCenter
}

func (k *k8sConfigCenter) GetByGroup(id, group string) (Config, error) {
	start := time.Now()
	defer func() {
		monitor.ConfigCenterDurationHistogram.
			With(prometheus.Labels{"group": group}).Observe(float64(time.Now().Sub(start)))
		monitor.ConfigCenterDurationSummary.
			With(prometheus.Labels{"group": group}).Observe(float64(time.Now().Sub(start)))
	}()
	monitor.ConfigCenterQuery.With(prometheus.Labels{"group": group}).Inc()
	if r, err := k.cache.GetByGroup(id, group); err == nil && r.ID == id {
		monitor.ConfigCenterCacheHit.With(prometheus.Labels{"group": group}).Inc()
		monitor.ConfigCenterContentLengthHistogram.
			With(prometheus.Labels{"group": group}).Observe(float64(len(r.Content)))
		monitor.ConfigCenterContentLengthSummary.
			With(prometheus.Labels{"group": group}).Observe(float64(len(r.Content)))
		return r, nil
	}
	if group == "" {
		group = constants.K8S_NAMESPACE
	}
	configMap, err := k.clientSet.CoreV1().ConfigMaps(group).Get(context.Background(), id, metav1.GetOptions{})
	if err != nil {
		return Config{}, nil
	}

	monitor.ConfigCenterContentLengthHistogram.
		With(prometheus.Labels{"group": group}).Observe(float64(len(configMap.Data[constants.K8S_CONFIGMAP_CONTENT_KEY])))
	monitor.ConfigCenterContentLengthSummary.
		With(prometheus.Labels{"group": group}).Observe(float64(len(configMap.Data[constants.K8S_CONFIGMAP_CONTENT_KEY])))

	return Config{ID: id, Content: configMap.Data[constants.K8S_CONFIGMAP_CONTENT_KEY]}, nil
}

func (k *k8sConfigCenter) GetByKV(kv map[string]Search, group string) ([]Config, error) {
	start := time.Now()
	defer func() {
		monitor.ConfigCenterDurationHistogram.
			With(prometheus.Labels{"group": group}).Observe(float64(time.Now().Sub(start)))
		monitor.ConfigCenterDurationSummary.
			With(prometheus.Labels{"group": group}).Observe(float64(time.Now().Sub(start)))
	}()
	monitor.ConfigCenterQuery.With(prometheus.Labels{"group": group}).Inc()
	if r, err := k.cache.GetByKV(kv, group); err == nil && len(r) != 0 {
		monitor.ConfigCenterCacheHit.With(prometheus.Labels{"group": group}).Inc()
		for _, t := range r {
			monitor.ConfigCenterContentLengthHistogram.
				With(prometheus.Labels{"group": group}).Observe(float64(len(t.Content)))
			monitor.ConfigCenterContentLengthSummary.
				With(prometheus.Labels{"group": group}).Observe(float64(len(t.Content)))
		}
		return r, nil
	}
	if group == "" {
		group = constants.K8S_GROUP_LABEL
	}
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"group": group}}
	listOptions := metav1.ListOptions{LabelSelector: labels.Set(labelSelector.MatchLabels).String()}

	configMap, err := k.clientSet.CoreV1().ConfigMaps(constants.K8S_NAMESPACE).List(context.Background(), listOptions)
	if err != nil {
		log.Printf("GetByKV error: %v", err)
		return nil, err
	}
	var ret []Config
	for _, cm := range configMap.Items {
		match := true
		for k, v := range kv {
			s := cm.Data[constants.K8S_CONFIGMAP_CONTENT_KEY]
			if v.Exact {
				if gjson.Get(s, k).String() != v.Keyword {
					match = false
					break
				}
			} else {
				if !strings.Contains(v.Keyword, gjson.Get(s, k).String()) {
					match = false
					break
				}
			}
		}
		if match {
			ret = append(ret, Config{
				ID:      cm.Name,
				Content: cm.Data[constants.K8S_CONFIGMAP_CONTENT_KEY],
			})
		}
	}

	for _, t1 := range ret {
		monitor.ConfigCenterContentLengthHistogram.
			With(prometheus.Labels{"group": group}).Observe(float64(len(t1.Content)))
		monitor.ConfigCenterContentLengthSummary.
			With(prometheus.Labels{"group": group}).Observe(float64(len(t1.Content)))
	}

	return ret, nil
}

func (k *k8sConfigCenter) Save(config Config) (bool, error) {
	_, err := k.cache.Save(config)
	if err != nil {
		log.Println(err)
	}
	if config.Group == "" {
		config.Group = constants.K8S_GROUP_LABEL
	}
	configMap := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.ID,
			Namespace: constants.K8S_NAMESPACE,
			Labels:    map[string]string{"group": config.Group},
		},
		Data: map[string]string{constants.K8S_CONFIGMAP_CONTENT_KEY: config.Content},
	}
	_, err = k.clientSet.CoreV1().ConfigMaps(constants.K8S_NAMESPACE).Create(context.Background(), &configMap, metav1.CreateOptions{})
	return err == nil, err
}

func (k *k8sConfigCenter) Get(id string) (Config, error) {
	start := time.Now()
	defer func() {
		monitor.ConfigCenterDurationHistogram.
			With(prometheus.Labels{"group": constants.K8S_NAMESPACE}).Observe(float64(time.Now().Sub(start)))
		monitor.ConfigCenterDurationSummary.
			With(prometheus.Labels{"group": constants.K8S_NAMESPACE}).Observe(float64(time.Now().Sub(start)))
	}()
	monitor.ConfigCenterQuery.With(prometheus.Labels{"group": constants.K8S_NAMESPACE}).Inc()
	if r, err := k.cache.Get(id); err == nil && r.ID == id {
		monitor.ConfigCenterCacheHit.With(prometheus.Labels{"group": constants.K8S_NAMESPACE}).Inc()
		monitor.ConfigCenterContentLengthHistogram.
			With(prometheus.Labels{"group": constants.K8S_NAMESPACE}).Observe(float64(len(r.Content)))
		monitor.ConfigCenterContentLengthSummary.
			With(prometheus.Labels{"group": constants.K8S_NAMESPACE}).Observe(float64(len(r.Content)))
		return r, nil
	}
	configMap, err := k.clientSet.CoreV1().ConfigMaps(constants.K8S_NAMESPACE).Get(context.Background(), id, metav1.GetOptions{})
	if err != nil {
		return Config{}, nil
	}

	monitor.ConfigCenterContentLengthHistogram.
		With(prometheus.Labels{"group": constants.K8S_NAMESPACE}).Observe(float64(len(configMap.Data[constants.K8S_CONFIGMAP_CONTENT_KEY])))
	monitor.ConfigCenterContentLengthSummary.
		With(prometheus.Labels{"group": constants.K8S_NAMESPACE}).Observe(float64(len(configMap.Data[constants.K8S_CONFIGMAP_CONTENT_KEY])))

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
