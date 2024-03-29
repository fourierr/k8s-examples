package main

import (
	"context"
	"cuelang.org/go/cue"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/dustin/go-humanize"
	carbon "github.com/golang-module/carbon/v2"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	v1beta12 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/flowcontrol/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog/v2"
	"math"
	"net/http"
	"sigs.k8s.io/yaml"
	"strings"
	"sync"
	"time"
)

var registerMetrics sync.Once

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

const json2 = `{"envs":[{"JAVA-OPTS":"xmx"}, {"CMB_LOGGING": "xms"}],"age":47}`

var err error

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

func humanateBytes(s uint64, base float64, sizes []string) string {
	if s < 10 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(logn(float64(s), base))
	suffix := sizes[int(e)]
	val := math.Floor(float64(s)/math.Pow(base, e)*10+0.5) / 10
	f := "%.0f %s"
	if val < 10 {
		f = "%.1f %s"
	}

	return fmt.Sprintf(f, val, suffix)
}

var testSlice []int

func main() {
	for i := 0; i < 1000; i++ {
		tmp := i
		go func() {
			testSlice = append(testSlice, tmp)
		}()
	}
	time.Sleep(30 * time.Second)
	for _, val := range testSlice {
		//fmt.Printf("idx:%d val:%d\n", idx, val)
		fmt.Printf(" val:%d\n", val)
	}
	str := "512Mi"
	u, err := humanize.ParseBytes(str)
	if err != nil {
		return
	}
	s1 := float32(u) / 1024 / 1024 / 1024
	fmt.Println(s1)
	s2 := humanize.IBytes(u)
	fmt.Println(s2)
	s3 := humanateBytes(u, 1024, []string{"GiB"})
	fmt.Println(s3)

	carbon.Now().Between(carbon.Now(), carbon.Now())
	var prints []func()
	for _, v := range []int{1, 2, 3} {
		//v := v
		prints = append(prints, func() { fmt.Println(v) })
	}
	for _, print := range prints {
		print()
	}

	parseBytes1, err := humanize.ParseBytes("79G")
	parseBytes2, err := humanize.ParseBytes("78G")
	if parseBytes1 > parseBytes2 {
		fmt.Println("111")
	}

	time.Now()

	// MB  MiB 转 B
	bytes := humanize.Bytes(82854982)
	iBytes := humanize.IBytes(82854982)
	parseBytes, err := humanize.ParseBytes("79G")
	if err != nil {
		return
	}
	iBytes = humanize.IBytes(parseBytes)

	parseBytesB, err := humanize.ParseBytes("79MiB")
	fmt.Println(bytes)
	fmt.Println(iBytes)
	fmt.Println(parseBytes)
	fmt.Println(parseBytesB)
	parseQuantity, err := resource.ParseQuantity("79Mi")
	if err != nil {
		return
	}
	s := parseQuantity.Value()
	fmt.Println(s)

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
	_ = v1alpha1.FlowSchema{}
	_ = v1alpha1.PriorityLevelConfiguration{}
	_ = v1beta12.ValidatingWebhookConfiguration{}
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
	clusterYaml = ""

	memoryYaml = ""
)
