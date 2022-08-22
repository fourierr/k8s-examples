package main

import (
	"context"
	"cuelang.org/go/cue"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	v1beta12 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	compbasemetrics "k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog/v2"
	"net/http"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
	"sync"
	"time"
)

var registerMetrics sync.Once

var metrics = []compbasemetrics.Registerable{
	ocmProxiedRequestsByResourceTotal,
	ocmProxiedRequestsByClusterTotal,
	ocmProxiedClusterEscalationRequestDurationHistogram,
}

const (
	namespace = "ocm"
	subsystem = "proxy"
)

// labels
const (
	proxiedResource = "resource"
	proxiedVerb     = "verb"
	proxiedCluster  = "cluster"
	success         = "success"
	code            = "code"
)

var (
	requestDurationSecondsBuckets = []float64{0, 0.005, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30}
)

var (
	ocmProxiedRequestsByResourceTotal = compbasemetrics.NewCounterVec(
		&compbasemetrics.CounterOpts{
			Namespace:      namespace,
			Subsystem:      subsystem,
			Name:           "proxied_resource_requests_by_resource_total",
			Help:           "Number of requests proxied requests",
			StabilityLevel: compbasemetrics.ALPHA,
		},
		[]string{proxiedResource, proxiedVerb, code},
	)
	ocmProxiedRequestsByClusterTotal = compbasemetrics.NewCounterVec(
		&compbasemetrics.CounterOpts{
			Namespace:      namespace,
			Subsystem:      subsystem,
			Name:           "proxied_requests_by_cluster_total",
			Help:           "Number of requests proxied requests",
			StabilityLevel: compbasemetrics.ALPHA,
		},
		[]string{proxiedCluster, code},
	)
	ocmProxiedClusterEscalationRequestDurationHistogram = compbasemetrics.NewHistogramVec(
		&compbasemetrics.HistogramOpts{
			Namespace:      namespace,
			Subsystem:      subsystem,
			Name:           "cluster_escalation_access_review_duration_seconds",
			Help:           "Cluster escalation access review time cost",
			Buckets:        requestDurationSecondsBuckets,
			StabilityLevel: compbasemetrics.ALPHA,
		},
		[]string{success},
	)
)

func RecordProxiedRequestsByResource(resource string, verb string, code int) {
	ocmProxiedRequestsByResourceTotal.
		WithLabelValues(resource, verb, strconv.Itoa(code)).
		Inc()
}

func RecordProxiedRequestsByCluster(cluster string, code int) {
	ocmProxiedRequestsByClusterTotal.
		WithLabelValues(cluster, strconv.Itoa(code)).
		Inc()
}

const json2 = `{"envs":[{"JAVA-OPTS":"xmx"}, {"CMB_LOGGING": "xms"}],"age":47}`

func main() {
	_ = v1beta12.ValidatingWebhookConfiguration{}
	// 测试 apimachinery 的周期内重试
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	wait.Until(func() {
		klog.Infof("retry")
	}, 10*time.Second, ctx.Done())
	klog.Infof("retry over")

	// 测试gjson和sjson
	array := gjson.Parse(json2).Get("envs").Array()
	for i, env := range array {
		if env.Get("JAVA-OPTS").Exists() {
			ops := env.Get("JAVA-OPTS")
			fmt.Println(ops)
			array[i].Raw = "xmxx"
		}
	}
	set, err2 := sjson.Set(json2, "envs", array)
	if err2 != nil {
		return
	}
	fmt.Println(set)

	// 为 Go 应用添加 Prometheus 自定义监控指标
	registerMetrics.Do(func() {
		for _, metric := range metrics {
			legacyregistry.MustRegister(metric)
		}
	})

	RecordProxiedRequestsByCluster("aa", 001)
	RecordProxiedRequestsByCluster("bb", 002)
	//temp := prometheus.NewGauge(prometheus.GaugeOpts{
	//	Name: "home_temperature_celsius",
	//	Help: "The current temperature in degrees Celsius.",
	//})
	//
	//// 在默认的注册表中注册该指标
	//prometheus.MustRegister(temp)
	//
	//// 设置 gauge 的值为 39
	//temp.Set(39)

	// 暴露指标
	//http.Handle("/metrics", promhttp.Handler())
	http.Handle("/metrics", legacyregistry.HandlerWithReset())
	http.ListenAndServe(":8080", nil)

	// clientset与sharedInformerFactory的使用
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	sharedInformerFactory := informers.NewSharedInformerFactoryWithOptions(clientset, 0, informers.WithNamespace("default"))

	podInformer := sharedInformerFactory.Core().V1().Pods().Informer()

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Println("add")
		},
	})

	stopCh := make(chan struct{})
	sharedInformerFactory.Start(stopCh)
	sharedInformerFactory.WaitForCacheSync(stopCh)
	<-stopCh

	// cron 定时器的使用
	attribute := cue.Attribute{}
	fmt.Println(attribute)
	c := cron.New(cron.WithSeconds())
	go func() {
		fmt.Println("end")
	}()
	go func() {
		c.AddFunc("30 * * ? * *", func() {
			fmt.Println("end")
		})
	}()

	c.Start()
	defer c.Stop()
	time.Sleep(3000 * time.Second)

	_ = v1.PodStatus{}
	_ = v1.PodSpec{}
	_ = appsv1.Deployment{}
	_ = v1.ConfigMap{}
	_ = v1beta1.CronJob{}
	color := ""
	time.Sleep(100 * time.Minute)
	prompt := &survey.Select{
		Message: "Choose a color:",
		Options: []string{"red", "blue", "green"},
	}
	survey.AskOne(prompt, &color)

	//ctx := context.Background()
	path := "spec.template.spec.containers[0].resources"
	//path := "spec.template.spec.containers[0].image"
	//path := "spec.replicas"

	//path := "spec.template[0].spec.containers[0].resources"
	//path := "spec.template.spec.containers[1].resources"

	clusterManifest := &unstructured.Unstructured{}
	clusterJson, _ := yaml.YAMLToJSON([]byte(clusterYaml))
	err = json.Unmarshal(clusterJson, clusterManifest)
	if err != nil {
		return
	}
	memoryManifest := &unstructured.Unstructured{}
	memoryJson, _ := yaml.YAMLToJSON([]byte(memoryYaml))
	err = json.Unmarshal(memoryJson, memoryManifest)
	if err != nil {
		return
	}

	prefix, i, suffix := splitPathWithIndex(context.Background(), path)
	fmt.Println(prefix, i, suffix)

	// 1、取 path 值
	value, err := fieldpath.Pave(clusterManifest.UnstructuredContent()).GetValue(path)

	//value, err := getNestedValueWithSlice(ctx, clusterManifest.DeepCopy(), path)
	if err != nil || value == nil {
		fmt.Println(err)
		return
	}
	fmt.Println(value)

	err = fieldpath.Pave(memoryManifest.UnstructuredContent()).SetValue(path, value)
	if err != nil {
		fmt.Println(err)
	}
	// 2、设置 path 值
	//un, _ := setNestedValueWithSlice(ctx, memoryManifest.DeepCopy(), path, value)
	//fmt.Println(un.UnstructuredContent())

	newValue, err := fieldpath.Pave(memoryManifest.DeepCopy().UnstructuredContent()).GetValue(path)
	//newValue, _ := getNestedValueWithSlice(ctx, un.DeepCopy(), path)
	fmt.Println(newValue)
}

func test03(string2 string) error {
	fmt.Println(string2, time.Now())
	return nil
}

// getNestedValueWithSlice get the value of unstructured.Unstructured via the path, like spec.template.spec.containers[0].image
func getNestedValueWithSlice(ctx context.Context, manifest *unstructured.Unstructured, path string) (interface{}, error) {
label:
	prefix, index, suffix := splitPathWithIndex(ctx, path)
	// 是 slice
	if index != -1 {
		dotPath := strings.Split(prefix, ".")
		nestedSlice, found, err := unstructured.NestedSlice(manifest.UnstructuredContent(), dotPath...)
		if err != nil {
			return nil, err
		}
		if !found || len(nestedSlice) <= index {
			return nil, errors.Errorf("the path(%s) was not found ", path)
		}
		manifest.Object = nestedSlice[index].(map[string]interface{})
		path = suffix
		goto label
	}
	// 不是 slice
	if index == -1 {
		dotPath := strings.Split(prefix, ".")
		nestedField, found, err := unstructured.NestedFieldCopy(manifest.UnstructuredContent(), dotPath...)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, errors.Errorf("the path(%s) was not found ", path)
		}
		return nestedField, nil
	}
	return nil, nil
}

// setNestedValueWithSlice set the value to unstructured.Unstructured via the path, like spec.template.spec.containers[0].image
func setNestedValueWithSlice(ctx context.Context, manifest *unstructured.Unstructured, path string, value interface{}) (*unstructured.Unstructured, error) {
	raw := manifest.DeepCopy()
	var pathSlice []string
	var indexSlice []int
	var valueSlice []interface{}
setLabel:
	prefix, index, suffix := splitPathWithIndex(ctx, path)
	// 是 slice
	if index != -1 {
		dotPath := strings.Split(prefix, ".")
		nestedSlice, found, err := unstructured.NestedSlice(manifest.UnstructuredContent(), dotPath...)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, errors.Errorf("the path(%s) was not found ", path)
		}
		manifest.Object = nestedSlice[index].(map[string]interface{})
		pathSlice = append(pathSlice, prefix)
		indexSlice = append(indexSlice, index)
		valueSlice = append(valueSlice, nestedSlice)
		path = suffix
		goto setLabel
	}
	// 不是 slice
	if index == -1 {
		pathSlice = append(pathSlice, prefix)
		indexSlice = append(indexSlice, index)
		valueSlice = append(valueSlice, nil)
	}

	// set area
	resUnstructured := map[string]interface{}{}
	for i := len(pathSlice) - 1; i >= 0; i-- {
		// 是 slice
		if indexSlice[i] != -1 {
			split := strings.Split(pathSlice[i], ".")
			err := unstructured.SetNestedSlice(raw.UnstructuredContent(), valueSlice[i].([]interface{}), split...)
			if err != nil {
				return nil, err
			}
			resUnstructured = raw.Object
		}
		// 不是 slice
		if indexSlice[i] == -1 {
			split := strings.Split(pathSlice[i], ".")
			if i-1 >= 0 {
				temp := valueSlice[i-1].([]interface{})[indexSlice[i-1]].(map[string]interface{})
				err := unstructured.SetNestedField(temp, value, split...)
				if err != nil {
					return nil, err
				}
				resUnstructured = temp
			} else {
				err := unstructured.SetNestedField(raw.UnstructuredContent(), value, split...)
				if err != nil {
					return nil, err
				}
				resUnstructured = raw.Object
			}
		}
	}
	toUnstructured, err := toUnstructured(resUnstructured)
	if err != nil {
		return nil, err
	}
	return toUnstructured, nil
}

func splitPathWithIndex(ctx context.Context, path string) (prefix string, index int, suffix string) {
	if strings.Contains(path, "[") {
		prefixSplit := strings.SplitN(path, "[", 2)
		prefix = prefixSplit[0]
		suffixSplit := strings.SplitN(prefixSplit[1], "].", 2)
		index = cast.ToInt(suffixSplit[0])
		suffix = suffixSplit[1]
		return prefix, index, suffix
	} else {
		return path, -1, ""
	}
}

func toUnstructured(in map[string]interface{}) (*unstructured.Unstructured, error) {
	marshal, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	un := new(unstructured.Unstructured)
	err = un.UnmarshalJSON(marshal)
	if err != nil {
		return nil, nil
	}
	return un, nil
}

const (
	clusterYaml = `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    app.alauda.io/display-name: fourier-container-040
    app.alauda.io/replicas: '1'
    deployment.kubernetes.io/revision: '31'
    io.cmb/liveness_probe_alert_level: warning
    io.cmb/readiness_probe_alert_level: warning
    owners.alauda.io/info: '[{"name":"马祥博","phone":"17854227913","employee_id":"80310624"}]'
  creationTimestamp: '2022-01-12T05:59:50Z'
  generation: 77
  labels:
    app.alauda.io/name: fourier-container-040.lt31-04-fourier
    app.cmboam.io/name: fourier-appfile-040.lt31-04
    component.cmboam.io/name: fourier-component-040.lt31-04-fourier
    workload-type: Deployment
  name: fourier-container-040
  namespace: lt31-04-fourier
  resourceVersion: '547401259'
  uid: c74afeba-18a2-412a-84b4-bd48144356e0
spec:
  progressDeadlineSeconds: 600
  replicas: 10
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app.alauda.io/name: fourier-container-040.lt31-04-fourier
      workload-type: Deployment
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      creationTimestamp: null
      labels:
        app.alauda.io/name: fourier-container-040.lt31-04-fourier
        cmb: lt31.04
        cmb-app-name: fourier-container-040
        cmb-log-tag: lt31-04-fourier-Deployment-fourier-container-040
        cmb-org: lt31.04
        cmb-service-unit-id: LT31.04...yxr-link-gray-test-three
        workload-type: Deployment
    spec:
      affinity: {}
      containers:
        - env:
            - name: CMB_LOGGING_PLATFORM_URL
              value: 'http://alpmng.redev.cmbchina.net:60000'
          image: 'csbase.registry.cmbchina.cn/console/proj_gin_test:v2_clusterYaml'
          imagePullPolicy: IfNotPresent
          lifecycle:
            preStop:
              exec:
                command:
                  - sh
                  - '-c'
                  - sleep 30
          livenessProbe:
            failureThreshold: 3
            initialDelaySeconds: 30
            periodSeconds: 30
            successThreshold: 1
            tcpSocket:
              port: 8001
            timeoutSeconds: 5
          name: fourier-container-040
          ports:
            - containerPort: 8001
              protocol: TCP
          readinessProbe:
            failureThreshold: 3
            initialDelaySeconds: 30
            periodSeconds: 30
            successThreshold: 1
            tcpSocket:
              port: 8001
            timeoutSeconds: 5
          resources:
            limits:
              cpu: '10'
              memory: 10Gi
            requests:
              cpu: 10m
              memory: 10Mi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      terminationGracePeriodSeconds: 30
status:
  availableReplicas: 1
  conditions:
    - lastTransitionTime: '2022-04-15T02:03:07Z'
      lastUpdateTime: '2022-04-15T02:03:07Z'
      message: Deployment has minimum availability.
      reason: MinimumReplicasAvailable
      status: 'True'
      type: Available
    - lastTransitionTime: '2022-01-12T07:00:52Z'
      lastUpdateTime: '2022-04-26T07:29:24Z'
      message: >-
        ReplicaSet "fourier-container-040-79b8f79fd9" has successfully
        progressed.
      reason: NewReplicaSetAvailable
      status: 'True'
      type: Progressing
  observedGeneration: 77
  readyReplicas: 1
  replicas: 1
  updatedReplicas: 1

`

	memoryYaml = `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    app.alauda.io/display-name: fourier-container-040
    app.alauda.io/replicas: '1'
    deployment.kubernetes.io/revision: '31'
    io.cmb/liveness_probe_alert_level: warning
    io.cmb/readiness_probe_alert_level: warning
    owners.alauda.io/info: '[{"name":"马祥博","phone":"17854227913","employee_id":"80310624"}]'
  creationTimestamp: '2022-01-12T05:59:50Z'
  generation: 77
  labels:
    app.alauda.io/name: fourier-container-040.lt31-04-fourier
    app.cmboam.io/name: fourier-appfile-040.lt31-04
    component.cmboam.io/name: fourier-component-040.lt31-04-fourier
    workload-type: Deployment
  name: fourier-container-040
  namespace: lt31-04-fourier
  resourceVersion: '547401259'
  uid: c74afeba-18a2-412a-84b4-bd48144356e0
spec:
  progressDeadlineSeconds: 600
  replicas: 5
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app.alauda.io/name: fourier-container-040.lt31-04-fourier
      workload-type: Deployment
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      creationTimestamp: null
      labels:
        app.alauda.io/name: fourier-container-040.lt31-04-fourier
        cmb: lt31.04
        cmb-app-name: fourier-container-040
        cmb-log-tag: lt31-04-fourier-Deployment-fourier-container-040
        cmb-org: lt31.04
        cmb-service-unit-id: LT31.04...yxr-link-gray-test-three
        workload-type: Deployment
    spec:
      affinity: {}
      containers:
        - env:
            - name: CMB_LOGGING_PLATFORM_URL
              value: 'http://alpmng.redev.cmbchina.net:60000'
          image: 'csbase.registry.cmbchina.cn/console/proj_gin_test:v2_memoryYaml'
          imagePullPolicy: IfNotPresent
          lifecycle:
            preStop:
              exec:
                command:
                  - sh
                  - '-c'
                  - sleep 30
          livenessProbe:
            failureThreshold: 3
            initialDelaySeconds: 30
            periodSeconds: 30
            successThreshold: 1
            tcpSocket:
              port: 8001
            timeoutSeconds: 5
          name: fourier-container-040
          ports:
            - containerPort: 8001
              protocol: TCP
          readinessProbe:
            failureThreshold: 3
            initialDelaySeconds: 30
            periodSeconds: 30
            successThreshold: 1
            tcpSocket:
              port: 8001
            timeoutSeconds: 5
          resources:
            limits:
              cpu: '5'
              memory: 5Gi
            requests:
              cpu: 5m
              memory: 5Mi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      terminationGracePeriodSeconds: 30
status:
  availableReplicas: 1
  conditions:
    - lastTransitionTime: '2022-04-15T02:03:07Z'
      lastUpdateTime: '2022-04-15T02:03:07Z'
      message: Deployment has minimum availability.
      reason: MinimumReplicasAvailable
      status: 'True'
      type: Available
    - lastTransitionTime: '2022-01-12T07:00:52Z'
      lastUpdateTime: '2022-04-26T07:29:24Z'
      message: >-
        ReplicaSet "fourier-container-040-79b8f79fd9" has successfully
        progressed.
      reason: NewReplicaSetAvailable
      status: 'True'
      type: Progressing
  observedGeneration: 77
  readyReplicas: 1
  replicas: 1
  updatedReplicas: 1

`
)
