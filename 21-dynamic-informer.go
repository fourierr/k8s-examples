package main

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

func main() {

	config := ctrl.GetConfigOrDie()
	timeout := time.Duration(60 * time.Second)
	stopCh := make(chan struct{})
	defer close(stopCh)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	resource := schema.GroupVersionResource{Group: "core.oam.dev", Version: "v1beta1", Resource: "applications"}
	informerReceiveObjectCh := make(chan *unstructured.Unstructured, 1)

	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 5)
	informerListerForGvr := informerFactory.ForResource(resource)
	// informer注册事件回调
	informerListerForGvr.Informer().AddEventHandler(handleEvent(informerReceiveObjectCh))
	// 5、informer增加索引
	//indexFunc := func(obj interface{}) ([]string, error) {
	//	return []string{obj.(*unstructured.Unstructured).GetNamespace()}, nil
	//}
	err = informerListerForGvr.Informer().AddIndexers(cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	if err != nil {
		panic(err)
	}
	go informerFactory.Start(stopCh)
	if synced := informerFactory.WaitForCacheSync(stopCh); !synced[resource] {
		panic("informer for application hasn't synced")
	}

	object, err := informerListerForGvr.Lister().ByNamespace("fourier").Get("fourierapp02")
	if err != nil {
		panic(err)
	}
	fmt.Println(object)
	// 处理事件回调通过的资源
	select {
	case objFromInformer := <-informerReceiveObjectCh:
		if objFromInformer.GetName() == "fourierapp02" {
			fmt.Println(objFromInformer.Object)
		}
	case <-ctx.Done():
		fmt.Println("informer haven't received an object, waited ")
	}
}

func handleEvent(rcvCh chan<- *unstructured.Unstructured) *cache.ResourceEventHandlerFuncs {
	return &cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			rcvCh <- obj.(*unstructured.Unstructured)
		},
		DeleteFunc: func(obj interface{}) {
			rcvCh <- obj.(*unstructured.Unstructured)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if !reflect.DeepEqual(oldObj, newObj) {
				rcvCh <- newObj.(*unstructured.Unstructured)
			}
		},
	}
}
