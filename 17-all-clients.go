package main

import (
	"context"
	"fmt"
	//"github.com/oam-dev/kubevela/pkg/apiserver/infrastructure/clients"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	crdv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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
	err = ctrlClient.Get(context.TODO(), client.ObjectKeyFromObject(unDeployy), unDeployy)
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

	// 7、ctrl NamespacedClient, 没有informer, 所有请求都使用相同的ns
	namespacedDeploy := new(appsv1.Deployment)
	err = client.NewNamespacedClient(ctrlClient, "vela-system").Get(context.TODO(), types.NamespacedName{Name: "kubevela-vela-core"}, namespacedDeploy)
	if err != nil {
		return
	}
	fmt.Println(namespacedDeploy)

	// 8、ctrl NewDelegatingClient, 主动声明informer, get和list会走informer, update走k8s apiserver
	defer util_runtime.HandleCrash()
	informerCacheCtx, informerCacheCancel := context.WithCancel(context.Background())
	// GVK和资源go types的对应关系，资源的默认值函数，资源版本转换函数，资源的标签转换等等全部由 schema 维护
	k8sScheme := runtime.NewScheme()
	_ = scheme.AddToScheme(k8sScheme)
	_ = crdv1.AddToScheme(k8sScheme)
	// go types 和 k8s api的对印关系由 restmapper维护
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
	delegatingClient, delegatingClientErr := client.NewDelegatingClient(client.NewDelegatingClientInput{
		CacheReader:     informerCache,
		Client:          ctrlClient,
		UncachedObjects: uncachedObjects,
	})
	if delegatingClientErr != nil {
		return
	}

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
		fmt.Println(delegatingDeploy)
	} else {
		goto label
	}

	// 9、vela client, 可以通过clustergateway 访问到子集群
	//err = clients.SetKubeConfig(apiConfig.Config{
	//	KubeQPS:   50,
	//	KubeBurst: 300,
	//})
	//if err != nil {
	//	return
	//}
	//// 在secretMultiClusterRoundTripper.RoundTrip对req.URL.Path做了修改
	//velaClient, err := clients.GetKubeClient()
	//if err != nil {
	//	return
	//}
	//remoteCtx := context.WithValue(context.Background(), "ClusterName", "cluster01")
	//velaDeploy := new(appsv1.Deployment)
	//err = velaClient.Get(remoteCtx, client.ObjectKey{Namespace: "vela-system", Name: "kubevela-vela-core"}, velaDeploy)
	//if err != nil {
	//	return
	//}
	//fmt.Println(velaDeploy)
}
