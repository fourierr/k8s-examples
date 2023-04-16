package main

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {
	cfg := config.GetConfigOrDie()
	ctx := context.Background()
	var opts ctrlcache.Options
	ls := labels.SelectorFromSet(map[string]string{"aaa": "bbbb"})
	fd := fields.OneTermEqualSelector("metadata.namespace", "bbb")
	// 按照label、field过滤部分资源进行listAndWatch
	opts.SelectorsByObject[&appsv1.Deployment{}] = ctrlcache.ObjectSelector{Label: ls, Field: fd}
	// TransformFunc 能够将对象并将其放入缓存之前调用相应的处理程序之前对其进行转换。最常见的使用模式是清除对象的某些部分，从而在给定控制器不关心它们的情况下减少内存占用。
	opts.DefaultTransform = func(i interface{}) (interface{}, error) {
		obj := i.(runtime.Object)

		accessor, err := meta.Accessor(obj)
		if err != nil {
			return i, err
		}
		accessor.SetManagedFields(nil)
		annotations := accessor.GetAnnotations()

		_, exists := annotations["applied"]
		if !exists {
			return i, nil
		}
		delete(annotations, "applied")
		accessor.SetAnnotations(annotations)
		return i, nil
	}
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
