package main

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	util_runtime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/flowcontrol"
	"time"
)

/*
	使用Informer实例的Lister()方法，List/Get Kubernetes中的Object时，
	Informer不会去请求Kubernetes API，而是直接查找缓存在本地内存中的数据(这份数据由Informer自己维护)。
	通过这种方式，Informer既可以更快地返回结果，又能减少对Kubernetes API的直接调用。
*/

func main() {
	// 1、生成一个kubeConfig
	cfg := &rest.Config{
		Host:            "regionEndPoint",
		BearerToken:     "regionToken",
		Timeout:         30 * time.Minute,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
		RateLimiter:     flowcontrol.NewTokenBucketRateLimiter(500, 500),
	}
	// 2、生成一个clientSet
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}

	stopCh := make(chan struct{})
	defer close(stopCh)
	// 3、new一个factory, informer watch apiserver,每隔 10 分钟 resync 一次(list)
	sharedInformerFactory := informers.NewSharedInformerFactory(clientSet, time.Minute*10)
	// 4、初始化 informer
	deploymentInformer := sharedInformerFactory.Apps().V1().Deployments().Informer()
	// 5、informer增加索引
	indexFunc := func(obj interface{}) ([]string, error) {
		return []string{obj.(*appsv1.Deployment).Spec.Template.Spec.Containers[0].Image}, nil
	}
	if err := deploymentInformer.AddIndexers(map[string]cache.IndexFunc{"spec.template.spec.containers[0].image": indexFunc}); err != nil {
		panic(err.Error())
	}
	defer util_runtime.HandleCrash()
	// 6、启动 informer，list & watch
	go sharedInformerFactory.Start(stopCh)
	// 7、等待从apiserver同步资源，即 list
	if !cache.WaitForCacheSync(stopCh, deploymentInformer.HasSynced) {
		util_runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	// 8、informer注册事件回调
	deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: func(oldObj, newObj interface{}) {},
		DeleteFunc: func(obj interface{}) {},
	})
	// 9、从LocalStore中获取Lister
	deploymentLister := sharedInformerFactory.Apps().V1().Deployments().Lister()

	// 10、从 lister 中获取 items
	deployments, err := deploymentLister.Deployments("namespace").List(labels.Everything())
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(deployments)
	<-stopCh
}

func onAdd(obj interface{}) {
	deployment := obj.(*appsv1.Deployment)
	fmt.Println("add a deployment:", deployment.Name)
}
