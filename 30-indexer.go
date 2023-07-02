package main

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func LabelIndexFunc(obj interface{}) ([]string, error) {
	metadata, err := meta.Accessor(obj)
	if err != nil {
		return []string{""}, fmt.Errorf("object has no meta: %v", err)
	}

	var indexKeys []string
	for key, value := range metadata.GetLabels() {
		indexKeys = append(indexKeys, key, fmt.Sprintf("%s=%s", key, value))
	}
	return indexKeys, nil

}

func main() {
	indexers := cache.Indexers{
		cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
		"labelindex":         LabelIndexFunc,
	}
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, indexers)

	pods := []*v1.Pod{
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod1",
				Namespace: "namespace1",
				Labels:    map[string]string{"label1": "pod1", "label2": "pod1"},
			},
		},
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				Namespace: "namespace2",
				Labels:    map[string]string{"label1": "pod2"},
			},
		},
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod3",
				Namespace: "namespace3",
				Labels:    map[string]string{"label1": "pod3", "label2": "pod3"},
			},
		},
	}

	for _, pod := range pods {
		_ = indexer.Add(pod)
	}

	keys, _ := indexer.IndexKeys(cache.NamespaceIndex, "namespace1")
	fmt.Printf("get key by namespace 'namespace1': %v\n", keys)

	keys, _ = indexer.IndexKeys("labelindex", "label1")
	fmt.Printf("get key by label 'label1': %v\n", keys)

	keys, _ = indexer.IndexKeys("labelindex", "label2")
	fmt.Printf("get key by label 'label2': %v\n", keys)

	keys, _ = indexer.IndexKeys("labelindex", "label1=pod2")
	fmt.Printf("get key by label 'label1=pod2': %v\n", keys)
}
