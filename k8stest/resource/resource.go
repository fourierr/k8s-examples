package resource

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//API for operate k8s
type API struct {
	FakeClient    kubernetes.Interface
	GenericClient client.Client
}

// CreatePodWithNs creates a new pod with namespace
func (a API) CreatePodWithNs(name, namespace string) error {
	ctx := context.TODO()
	p := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	_, err := a.FakeClient.CoreV1().Pods(namespace).Create(ctx, p, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// UpdateConfigMap creates a new cm with namespace
func (a API) UpdateConfigMap(cm *v1.ConfigMap) error {
	ctx := context.TODO()
	err := a.GenericClient.Update(ctx, cm)
	if err != nil {
		return err
	}
	return nil
}

//Cache for operate k8s
type Cache struct {
	Pods map[string]*v1.Pod
}

// AddPodToCache add a pod
func (c *Cache) AddPodToCache(p *v1.Pod) error {
	c.Pods[p.Name] = p
	return nil
}
