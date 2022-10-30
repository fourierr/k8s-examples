package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {

	config := ctrl.GetConfigOrDie()
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	resource := schema.GroupVersionResource{Group: "core.oam.dev", Version: "v1beta1", Resource: "applications"}
	watcher, err := dynamicClient.Resource(resource).Namespace("fourier").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	resultCh := watcher.ResultChan()
	for {
		select {
		case event := <-resultCh:
			switch event.Type {
			case watch.Added:
				fmt.Println("add", event)
			case watch.Modified:
				fmt.Println("modify", event)
			case watch.Deleted:
				fmt.Println("delete", event)
			}
		}
	}
}
