package main

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	crdv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	util_runtime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	crtl_cache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"time"
)

// 测试client-go的几个客户端
// client-go的rest client、clientset、dynamic client都没有informer
// ctrl client 走informer
func main() {

	// 获取config
	config, err := clientcmd.BuildConfigFromFlags("", "/mnt/f/vela/config")
	//config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err)
	}

	// 1、client-go rest client, 没有使用informer, 直接访问kube-apiserver
	httpClient, err := rest.HTTPClientFor(config)
	if err != nil {
		return
	}
	configShadow := *config
	configShadow.GroupVersion = &appsv1.SchemeGroupVersion
	configShadow.APIPath = "/apis"
	configShadow.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	configShadow.UserAgent = rest.DefaultKubernetesUserAgent()
	restClient, err := rest.RESTClientForConfigAndClient(&configShadow, httpClient)
	if err != nil {
		return
	}
	restDeploy := new(appsv1.Deployment)
	getOpts := new(metav1.GetOptions)
	err = restClient.Get().
		Namespace("vela-system").
		Resource("deployments").
		Name("kubevela-vela-core").
		VersionedParams(getOpts, scheme.ParameterCodec).Do(context.TODO()).Into(restDeploy)
	if err != nil {
		return
	}
	fmt.Println(restDeploy)

	// 2、client-go clientset, 没有使用informer, 直接访问kube-apiserver
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	clientsetDeploy := new(appsv1.Deployment)
	clientsetDeploy, err = clientset.AppsV1().Deployments("vela-system").
		Get(context.TODO(), "kubevela-vela-core", metav1.GetOptions{})
	if err != nil {
		return
	}
	fmt.Println(clientsetDeploy)

	// 3、client-go dynamic client, 没有使用informer, 直接访问kube-apiserver
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return
	}
	unDeploy := new(unstructured.Unstructured)
	unDeploy, err = dynamicClient.Resource(schema.GroupVersionResource{
		Group:    appsv1.GroupName,
		Version:  appsv1.SchemeGroupVersion.Version,
		Resource: "deployments"}).Namespace("vela-system").Get(context.TODO(), "kubevela-vela-core", metav1.GetOptions{})
	if err != nil {
		return
	}
	fmt.Println(unDeploy)

	// ctrl/pkg/client, 包含 typedClient, unstructuredClient, metadataClient
	ctrlClient, err := client.New(config, client.Options{})
	if err != nil {
		return
	}

	// 4、ctrl typedClient 没有使用informer
	ctrlDeploy := new(appsv1.Deployment)
	err = ctrlClient.Get(context.TODO(), client.ObjectKey{Namespace: "vela-system", Name: "kubevela-vela-core"}, ctrlDeploy)
	if err != nil {
		return
	}
	fmt.Println(ctrlDeploy)

	// 5、ctrl unstructuredClient 没有使用informer
	unDeployy := new(unstructured.Unstructured)
	unDeployy.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   appsv1.GroupName,
		Version: appsv1.SchemeGroupVersion.Version,
		Kind:    "Deployment"})
	err = ctrlClient.Get(context.TODO(), client.ObjectKey{Namespace: "vela-system", Name: "kubevela-vela-core"}, unDeployy)
	if err != nil {
		return
	}
	fmt.Println(unDeployy)

	//	6、ctrl metadataClient 没有informer, 只会返回metadata部分
	pom := new(metav1.PartialObjectMetadata)
	pom.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   appsv1.GroupName,
		Version: appsv1.SchemeGroupVersion.Version,
		Kind:    "Deployment"})
	err = ctrlClient.Get(context.TODO(), client.ObjectKey{Namespace: "vela-system", Name: "kubevela-vela-core"}, pom)
	if err != nil {
		return
	}

	// 7、NamespacedClient
	//client.NewNamespacedClient()

	// 8、NewDelegatingClient

	// Scheme defines methods for serializing and deserializing API objects, a type
	// registry for converting group, version, and kind information to and from Go
	// schemas, and mappings between Go schemas of different versions. A scheme is the
	// foundation for a versioned API and versioned configuration over time.
	k8sScheme := runtime.NewScheme()
	_ = scheme.AddToScheme(k8sScheme)
	_ = crdv1.AddToScheme(k8sScheme)
	// MapperProvider provides the rest mapper used to map go types to Kubernetes APIs
	mapper, err := apiutil.NewDiscoveryRESTMapper(config)
	if err != nil {
		return
	}
	fiveMinute := 5 * time.Minute
	informerCache, err := crtl_cache.New(config, crtl_cache.Options{Scheme: k8sScheme, Mapper: mapper, Namespace: "vela-system", Resync: &fiveMinute})
	if err != nil {
		return
	}

	uncachedObjects := []client.Object{&corev1.ConfigMap{}, &corev1.Service{}}
	delegatingClient, err := client.NewDelegatingClient(client.NewDelegatingClientInput{
		CacheReader:     informerCache,
		Client:          ctrlClient,
		UncachedObjects: uncachedObjects,
	})
	if err != nil {
		return
	}
	// 启动 informer，list & watch
	defer util_runtime.HandleCrash()
	informerCacheCtx, informerCacheCancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		defer util_runtime.HandleCrash()
		go informerCache.Start(informerCacheCtx)
	}(informerCacheCtx)
label:
	waitForCacheSync := informerCache.WaitForCacheSync(context.TODO())
	if waitForCacheSync {
		delegatingDeploy := new(appsv1.Deployment)
		err = delegatingClient.Get(context.TODO(), client.ObjectKey{Namespace: "vela-system", Name: "kubevela-vela-core"}, delegatingDeploy)
		if err != nil {
			return
		}
		informerCacheCancel()
	} else {
		goto label
	}
}
