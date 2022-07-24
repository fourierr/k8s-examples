module k8s-examples

go 1.17

require (
	cuelang.org/go v0.2.2
	github.com/google/go-github/v32 v32.1.0
	github.com/google/uuid v1.3.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/tidwall/gjson v1.12.1
	k8s.io/api v0.23.6
	k8s.io/apimachinery v0.23.6
	k8s.io/client-go v0.23.6
	sigs.k8s.io/controller-runtime v0.11.2
)

require (
	github.com/AlecAivazis/survey/v2 v2.1.1
	github.com/crossplane/crossplane-runtime v0.14.1-0.20210722005935-0b469fcc77cd
	github.com/emicklei/go-restful/v3 v3.0.0-rc2
	github.com/go-logr/logr v1.2.2
	github.com/google/go-cmp v0.5.8
	github.com/google/go-containerregistry v0.9.0
	github.com/oam-dev/kubevela v1.4.2
	github.com/oam-dev/terraform-controller v0.7.3
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.19.0
	github.com/prometheus/client_golang v1.11.0
	github.com/pyroscope-io/client v0.2.3
	github.com/robfig/cron/v3 v3.0.1
	github.com/spf13/cast v1.4.1
	github.com/tidwall/sjson v1.2.4
	github.com/vrischmann/envconfig v1.3.0
	golang.org/x/sync v0.0.0-20220513210516-0976fa681c29
	gotest.tools v2.2.0+incompatible
	k8s.io/apiextensions-apiserver v0.23.6
	k8s.io/component-base v0.23.6
	k8s.io/klog/v2 v2.60.1
	k8s.io/kubectl v0.23.6
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9
	sigs.k8s.io/yaml v1.3.0
)

replace (
	cuelang.org/go => github.com/fourierr/cue-0.2.2-fix v0.0.2
	github.com/docker/cli => github.com/docker/cli v20.10.9+incompatible
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/wercker/stern => github.com/oam-dev/stern v1.13.2
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client => sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.24
)
