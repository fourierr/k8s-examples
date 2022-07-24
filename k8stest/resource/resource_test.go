package resource

import (
	"context"
	"encoding/json"
	"github.com/oam-dev/kubevela/apis/core.oam.dev/common"
	"github.com/oam-dev/kubevela/apis/core.oam.dev/v1alpha1"
	"github.com/oam-dev/kubevela/apis/core.oam.dev/v1alpha2"
	"github.com/oam-dev/kubevela/apis/core.oam.dev/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	genericclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/yaml"
	"testing"
	"time"
)

var (
	ctx context.Context
	// Scheme用于提供GVK与对应Go types的映射关系，每一种Controller都需要一个scheme。
	testScheme = runtime.NewScheme()
	testEnv    *envtest.Environment
	cfg        *rest.Config
	k8sClient  client.Client
)

const (
	componentDefYaml = `
apiVersion: core.oam.dev/v1beta1
kind: ComponentDefinition
metadata:
  name: worker
  namespace: vela-system
  annotations:
    definition.oam.dev/description: "Long-running scalable backend worker without network endpoint"
spec:
  workload:
    definition:
      apiVersion: apps/v1
      kind: Deployment
  extension:
    template: |
      output: {
          apiVersion: "apps/v1"
          kind:       "Deployment"
          metadata: {
              annotations: {
                  if context["config"] != _|_ {
                      for _, v in context.config {
                          "\(v.name)" : v.value
                      }
                  }
              }
          }
          spec: {
              selector: matchLabels: {
                  "app.oam.dev/component": context.name
              }
              template: {
                  metadata: labels: {
                      "app.oam.dev/component": context.name
                  }

                  spec: {
                      containers: [{
                          name:  context.name
                          image: parameter.image

                          if parameter["cmd"] != _|_ {
                              command: parameter.cmd
                          }
                      }]
                  }
              }

              selector:
                  matchLabels:
                      "app.oam.dev/component": context.name
          }
      }

      parameter: {
          // +usage=Which image would you like to use for your service
          // +short=i
          image: string

          cmd?: [...string]
      }
`
)

func init() {
	_ = clientgoscheme.AddToScheme(testScheme)
	_ = v1.AddToScheme(testScheme)
}

// Fake Client
func TestCreatePodWithNs(t *testing.T) {
	ctx = context.TODO()
	cases := []struct {
		name string
		ns   string
	}{
		{
			name: "test-pod01",
			ns:   "test-ns02",
		},
		{
			name: "test-pod02",
			ns:   "test-ns02",
		},
	}

	api := &API{
		FakeClient: fake.NewSimpleClientset(),
	}
	for _, c := range cases {
		// create the postfixed namespace
		err := api.CreatePodWithNs(c.name, c.ns)
		if err != nil {
			t.Fatal(err.Error())
		}

		if p, err := api.FakeClient.CoreV1().Pods(c.ns).Get(ctx, c.name, metav1.GetOptions{}); nil != err {
			t.Errorf("get pod  err %v", err)
		} else if p.Name != c.name {
			t.Errorf("pod name err")
		}
	}
}

// Fake Client
func TestCreatePodEvent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the fake client.
	client := fake.NewSimpleClientset()

	c := Cache{Pods: make(map[string]*v1.Pod)}
	// Create an informer that writes added pods to a channel.
	pods := make(chan *v1.Pod, 1)
	informers := informers.NewSharedInformerFactory(client, 0)
	podInformer := informers.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			//t.Logf("pod added: %s/%s", pod.Namespace, pod.Name)
			if err := c.AddPodToCache(pod); nil != err {
				t.Errorf("add pod err %v", err)
			}
			pods <- pod
		},
	})

	// Make sure informers are running.
	informers.Start(ctx.Done())

	// This is not required in tests, but it serves as a proof-of-concept by
	// ensuring that the informer goroutine have warmed up and called List before
	// we send any events to it.
	cache.WaitForCacheSync(ctx.Done(), podInformer.HasSynced)

	// Inject an event into the fake client.
	p := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-pod03"}}
	_, err := client.CoreV1().Pods("test-ns03").Create(ctx, p, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error injecting pod add: %v", err)
	}

	select {
	case pod := <-pods:
		if _, ok := c.Pods[pod.Name]; !ok {
			t.Errorf("no pod after add event")
		}
	case <-time.After(wait.ForeverTestTimeout):
		t.Error("Informer did not get the added pod")
	}
}

// Generic Client
func TestCreateConfigMapWithNs(t *testing.T) {
	ctx = context.TODO()
	cases := []*v1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm03",
				Namespace: "test-ns03",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm04",
				Namespace: "test-ns04",
			},
		},
	}
	UpdateCases := []*v1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm03",
				Namespace: "test-ns03",
			},
			Data: map[string]string{"03": "03"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm04",
				Namespace: "test-ns04",
			},
			Data: map[string]string{"04": "04"},
		},
	}
	api := &API{
		GenericClient: genericclient.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(cases[0], cases[1]).Build(),
	}
	for _, c := range UpdateCases {
		cm := &v1.ConfigMap{}
		if err := api.GenericClient.Get(ctx, types.NamespacedName{Name: c.Name, Namespace: c.Namespace}, cm); nil != err {
			t.Errorf("get cm  err %v", err)
		} else if cm.Name != c.Name {
			t.Errorf("cm name err")
		}
		// create the postfixed namespace
		err := api.UpdateConfigMap(c)
		if err != nil {
			t.Fatal(err.Error())
		}
		cm2 := &v1.ConfigMap{}
		if err := api.GenericClient.Get(ctx, types.NamespacedName{Name: c.Name, Namespace: c.Namespace}, cm2); nil != err {
			t.Errorf("get cm  err %v", err)
		} else if cm2.Name != c.Name {
			t.Errorf("cm name err")
		}
	}
}

// envtest
func TestCreateAppliction(t *testing.T) {
	//yamlPath := filepath.Join("testdata", "crds")
	yamlPath := "../testdata/crds"
	testEnv = &envtest.Environment{
		ControlPlaneStartTimeout: time.Minute,
		ControlPlaneStopTimeout:  time.Minute,
		UseExistingCluster:       pointer.BoolPtr(false),
		//CRDDirectoryPaths:        []string{"./testdata/crds"},
		CRDDirectoryPaths: []string{yamlPath},
	}
	var err error
	cfg, err = testEnv.Start()
	if err != nil {
		t.Errorf(err.Error())
	}
	defer testEnv.Stop()

	err = v1alpha2.SchemeBuilder.AddToScheme(testScheme)
	if err != nil {
		t.Errorf(err.Error())
	}

	err = v1alpha1.SchemeBuilder.AddToScheme(testScheme)
	if err != nil {
		t.Errorf(err.Error())
	}

	err = v1beta1.SchemeBuilder.AddToScheme(testScheme)
	if err != nil {
		t.Errorf(err.Error())
	}

	k8sClient, err = client.New(cfg, client.Options{Scheme: testScheme})
	if err != nil {
		t.Errorf(err.Error())
	}
	velaSystemNs := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "vela-system",
		},
	}
	err = k8sClient.Create(ctx, velaSystemNs)
	if err != nil {
		t.Errorf(err.Error())
	}
	cd := &v1beta1.ComponentDefinition{}
	cDDefJson, _ := yaml.YAMLToJSON([]byte(componentDefYaml))
	err = json.Unmarshal(cDDefJson, cd)
	if err != nil {
		t.Errorf(err.Error())
	}
	err = k8sClient.Create(ctx, cd.DeepCopy())
	if err != nil {
		t.Errorf(err.Error())
	}

	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "app-ns-01",
		},
	}
	err = k8sClient.Create(ctx, ns)
	if err != nil {
		t.Errorf(err.Error())
	}
	app := &v1beta1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "core.oam.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "app-test01",
		},
		Spec: v1beta1.ApplicationSpec{
			Components: []common.ApplicationComponent{
				{
					Name:       "app-test01-component",
					Type:       "worker",
					Properties: &runtime.RawExtension{Raw: []byte("{\"cmd\":[\"sleep\",\"1000\"],\"image\":\"busybox\"}")},
				},
			},
		},
	}
	app.SetNamespace(ns.Name)
	appCopy := app.DeepCopy()
	err = k8sClient.Create(ctx, appCopy)
	if err != nil {
		t.Errorf(err.Error())
	}
	curApp := &v1beta1.Application{}
	err = k8sClient.Get(ctx, client.ObjectKey{Name: app.Name, Namespace: app.Namespace}, curApp)
	if err != nil {
		t.Errorf(err.Error())
	}
}

// ginkgo & gomega
