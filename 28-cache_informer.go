package main

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {
	//cfg := &rest.Config{
	//	Host:            "regionEndPoint",
	//	BearerToken:     "regionToken",
	//	Timeout:         30 * time.Minute,
	//	TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	//	RateLimiter:     flowcontrol.NewTokenBucketRateLimiter(500, 500),
	//}
	cfg := config.GetConfigOrDie()
	ctx := context.Background()
	var opts ctrlcache.Options
	ls := labels.SelectorFromSet(map[string]string{"aaa": "bbbb"})
	fd := fields.OneTermEqualSelector("metadata.namespace", "bbb")
	opts.SelectorsByObject[&appsv1.Deployment{}] = ctrlcache.ObjectSelector{Label: ls, Field: fd}
	informer, err := ctrlcache.New(cfg, opts)
	if err != nil {
		panic(err)
	}
	if err := informer.Start(ctx); err != nil {
		panic(err)
	}
	if hasSyncd := informer.WaitForCacheSync(ctx); !hasSyncd {
		panic("informer for application hasn't synced")
	}

}
