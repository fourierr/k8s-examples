package main

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	pkgClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {

	var cfg *rest.Config
	// todo 初始化cfg
	cache, err := cache.New(cfg, cache.Options{})
	if err != nil {
		panic(err.Error())
	}
	client, err := pkgClient.New(cfg, pkgClient.Options{})
	if err != nil {
		panic(err.Error())
	}
	// todo 使用Expect(err).NotTo(HaveOccurred())
	dReader, err := pkgClient.NewDelegatingClient(pkgClient.NewDelegatingClientInput{
		CacheReader:       cache,
		Client:            client,
		CacheUnstructured: false,
	})
	if err != nil {
		panic(err.Error())
	}
	var deploymentList appsv1.DeploymentList
	if err := dReader.List(context.TODO(), &deploymentList, &pkgClient.ListOptions{}); err != nil {
		panic(err.Error())
	}
}
