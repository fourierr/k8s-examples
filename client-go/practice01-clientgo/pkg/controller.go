package pkg

import (
	"context"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"reflect"
	"time"
)

const (
	workerNum = 5
	maxRetry  = 10
)

type controller struct {
	client          kubernetes.Interface
	serviceLister   corev1lister.ServiceLister
	configMapLister corev1lister.ConfigMapLister
	queue           workqueue.RateLimitingInterface
}

func NewController(client kubernetes.Interface, serviceInformer corev1informer.ServiceInformer, configMapInformer corev1informer.ConfigMapInformer) *controller {
	c := controller{
		client:          client,
		serviceLister:   serviceInformer.Lister(),
		configMapLister: configMapInformer.Lister(),
		queue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "svc-cm controller"),
	}
	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.serviceAddFunc,
		UpdateFunc: c.serviceUpdateFunc,
	})
	configMapInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.configmapDeleteFunc,
	})
	return &c
}

// 消费者
func (c controller) Run(stopCh chan struct{}) {
	for i := 0; i < workerNum; i++ {
		// 每隔1分钟运行一次 c.worker 函数
		go wait.Until(c.worker, time.Minute, stopCh)
	}
	<-stopCh
}

func (c controller) worker() {
	for c.processNextItem() {

	}
}

func (c controller) processNextItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(item)

	key := item.(string)

	err := c.reconcileService(key)
	if err != nil {
		c.handleErr(key, err)
	}
	return true
}

func (c controller) reconcileService(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	// 检查service是否存在
	service, err := c.serviceLister.Services(namespace).Get(name)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	// 新增或删除 configmap
	_, ok := service.GetAnnotations()["service/config"]
	configMap, err := c.configMapLister.ConfigMaps(namespace).Get(name)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if ok && apierrors.IsNotFound(err) {
		// create configmap
		ig := constructConfigMap(service)
		_, err := c.client.CoreV1().ConfigMaps(namespace).Create(context.TODO(), ig, metav1.CreateOptions{})
		if err != nil {
			return err
		}

	} else if !ok && configMap != nil {
		// delete confimap
		err := c.client.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c controller) handleErr(key string, err error) {
	if c.queue.NumRequeues(key) < maxRetry {
		c.queue.AddRateLimited(key)
	}
	runtime.HandleError(err)
	c.queue.Forget(key)
}

func constructConfigMap(service *corev1.Service) *corev1.ConfigMap {
	configMap := corev1.ConfigMap{}
	//ctrl.SetControllerReference(service,configMap,corev)
	configMap.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(service, corev1.SchemeGroupVersion.WithKind("Service")),
	}
	configMap.Name = service.Name
	configMap.Namespace = service.Namespace
	configMap.Data = map[string]string{
		"a": "aa",
		"b": "bb",
	}
	return &configMap
}

// 生产者
func (c controller) enqueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
	}
	c.queue.Add(key)
}

func (c *controller) serviceAddFunc(obj interface{}) {
	logrus.Info("add svc")
	c.enqueue(obj)
}

func (c controller) serviceUpdateFunc(olbObj interface{}, newObj interface{}) {
	logrus.Info("update svc")
	if reflect.DeepEqual(olbObj, newObj) {
		return
	}
	c.enqueue(newObj)
}

func (c controller) configmapDeleteFunc(obj interface{}) {
	logrus.Info("delete cm")
	configMap := obj.(*corev1.ConfigMap)
	ownerReference := metav1.GetControllerOf(configMap)
	if ownerReference == nil {
		return
	}
	if ownerReference.Kind != "Service" {
		return
	}
	c.queue.AddRateLimited(configMap.Namespace + "/" + configMap.Name)
}
