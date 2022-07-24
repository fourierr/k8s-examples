package main

import (
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"proj_test/client-go/practice01/pkg"
)

func main() {
	// config
	//config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	config, err := clientcmd.BuildConfigFromFlags("", "/mnt/f/vela/config")
	if err != nil {
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalln(err)
		}
		config = inClusterConfig
	}

	// clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(err)
	}
	// informerFactory
	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientset, 0, informers.WithNamespace("default"))
	serviceInformer := informerFactory.Core().V1().Services()
	configMapInformer := informerFactory.Core().V1().ConfigMaps()

	// add event handler
	controller := pkg.NewController(clientset, serviceInformer, configMapInformer)

	// informerFactory.start
	stopCh := make(chan struct{})
	informerFactory.Start(stopCh)
	informerFactory.WaitForCacheSync(stopCh)

	// controller.start
	controller.Run(stopCh)
}
