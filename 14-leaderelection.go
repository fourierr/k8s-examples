package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"os"
	"os/signal"
	"syscall"
	"time"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
	ctrlleader "sigs.k8s.io/controller-runtime/pkg/leaderelection"
)

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func main() {
	klog.InitFlags(nil)

	var kubeConfig string
	var leaseLockName string
	var leaseLockNamespace string
	// 初始化客户端的部分
	flag.StringVar(&kubeConfig, "kube-config", "", "absolute path to the kubeConfig file")
	flag.StringVar(&leaseLockName, "lease-lock-name", "", "the lease lock resource name")
	flag.StringVar(&leaseLockNamespace, "lease-lock-namespace", "", "the lease lock resource namespace")
	flag.Parse()

	if leaseLockName == "" {
		klog.Fatal("unable to get lease lock resource name (missing lease-lock-name flag).")
	}
	if leaseLockNamespace == "" {
		klog.Fatal("unable to get lease lock resource namespace (missing lease-lock-namespace flag).")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	//config, err := config2.GetConfig()
	if err != nil {
		klog.Fatal(err)
	}
	client := clientset.NewForConfigOrDie(config)

	// ctrl 的 leaderelection
	options := ctrlleader.Options{}
	fmt.Println(options)

	run := func(ctx context.Context) {
		// 实现的业务逻辑，这里仅仅为实验，就直接打印了
		klog.Info("Controller loop...")

		for {
			fmt.Println("I am leader, I was working.")
			time.Sleep(time.Second * 5)
		}
	}

	// use a Go context so we can tell the leaderelection code when we
	// want to step down
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听系统中断
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		klog.Info("Received termination, signaling shutdown")
		cancel()
	}()
	id, err := os.Hostname()
	if err != nil {
		return
	}
	id = id + "_" + uuid.New().String()
	fmt.Printf("id: %s", id)
	fmt.Println()

	// 创建一个资源锁, 这里使用了一个lease锁，注释中说愿意为集群中存在lease的监听较少
	// configmapsleases能够兼容更多的k8s版本，老版本的仅支持ConfigMap锁
	lock, err := resourcelock.New("configmapsleases", leaseLockNamespace, leaseLockName,
		client.CoreV1(), client.CoordinationV1(), resourcelock.ResourceLockConfig{Identity: id})
	if err != nil {
		return
	}
	// 开启一个选举的循环
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   15 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				// 当选举为leader后所运行的业务逻辑
				run(ctx)
			},
			OnStoppedLeading: func() {
				// we can do cleanup here
				klog.Infof("leader lost: %s", id)
				os.Exit(0)
			},
			OnNewLeader: func(identity string) { // 申请一个选举时的动作
				if identity == id {
					return
				}
				klog.Infof("new leader elected: %s", identity)
			},
		},
	})
}
