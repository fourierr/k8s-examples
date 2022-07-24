package application

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"

	"github.com/oam-dev/kubevela/apis/core.oam.dev/common"
	"github.com/oam-dev/kubevela/apis/core.oam.dev/v1beta1"
	"github.com/oam-dev/kubevela/pkg/oam/testutil"
	"github.com/oam-dev/kubevela/pkg/oam/util"
)

// TODO: Refactor the tests to not copy and paste duplicated code 10 times
var _ = Describe("Test Application Controller", func() {

	cd := &v1beta1.ComponentDefinition{}
	cDDefJson, _ := yaml.YAMLToJSON([]byte(componentDefYaml))

	td := &v1beta1.TraitDefinition{}
	tDDefJson, _ := yaml.YAMLToJSON([]byte(traitDefYaml))

	ctx := context.TODO()
	appWithControlPlaneOnly := &v1beta1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "core.oam.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "app-controlplaneonly",
		},
		Spec: v1beta1.ApplicationSpec{
			Components: []common.ApplicationComponent{
				{
					Name:       "app-controlplaneonly-component",
					Type:       "worker",
					Properties: &runtime.RawExtension{Raw: []byte("{\"cmd\":[\"sleep\",\"1000\"],\"image\":\"busybox\"}")},
				},
			},
		},
	}
	appWithControlPlaneOnly.Spec.Components[0].Traits = []common.ApplicationTrait{
		{
			Type:       "hubcpuscaler",
			Properties: &runtime.RawExtension{Raw: []byte("{\"min\": 1,\"max\": 10,\"cpuPercent\": 60}")},
		},
	}

	BeforeEach(func() {
		Expect(json.Unmarshal(cDDefJson, cd)).Should(BeNil())
		Expect(k8sClient.Create(ctx, cd.DeepCopy())).Should(SatisfyAny(BeNil(), &util.AlreadyExistMatcher{}))

		Expect(json.Unmarshal(tDDefJson, td)).Should(BeNil())
		Expect(k8sClient.Create(ctx, td.DeepCopy())).Should(SatisfyAny(BeNil(), &util.AlreadyExistMatcher{}))

		importHubCpuScaler := &v1beta1.TraitDefinition{}
		hubCpuScalerJson, hubCpuScalerErr := yaml.YAMLToJSON([]byte(hubCpuScalerYaml))
		Expect(hubCpuScalerErr).ShouldNot(HaveOccurred())
		Expect(json.Unmarshal(hubCpuScalerJson, importHubCpuScaler)).Should(BeNil())
		Expect(k8sClient.Create(ctx, importHubCpuScaler.DeepCopy())).Should(SatisfyAny(BeNil(), &util.AlreadyExistMatcher{}))
	})
	It("test application with controlPlaneOnly trait ", func() {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "vela-test-with-controlplaneonly",
			},
		}
		Expect(k8sClient.Create(ctx, ns)).Should(BeNil())

		appWithControlPlaneOnly.SetNamespace(ns.Name)
		app := appWithControlPlaneOnly.DeepCopy()
		Expect(k8sClient.Create(ctx, app)).Should(BeNil())

		appKey := client.ObjectKey{
			Name:      app.Name,
			Namespace: app.Namespace,
		}
		testutil.ReconcileOnceAfterFinalizer(reconciler, reconcile.Request{NamespacedName: appKey})

		By("Check App running successfully")
		curApp := &v1beta1.Application{}
		Expect(k8sClient.Get(ctx, appKey, curApp)).Should(BeNil())
		Expect(curApp.Status.Phase).Should(Equal(common.ApplicationRunning))

		appRevision := &v1beta1.ApplicationRevision{}
		Expect(k8sClient.Get(ctx, client.ObjectKey{
			Namespace: app.Namespace,
			Name:      curApp.Status.LatestRevision.Name,
		}, appRevision)).Should(BeNil())

		By("Check affiliated resource tracker is created")
		expectRTName := fmt.Sprintf("%s-%s", appRevision.GetName(), appRevision.GetNamespace())
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Name: expectRTName}, &v1beta1.ResourceTracker{})
		}, 10*time.Second, 500*time.Millisecond).Should(Succeed())

		By("Check AppRevision Created with the expected workload spec")
		appRev := &v1beta1.ApplicationRevision{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Name: app.Name + "-v1", Namespace: app.GetNamespace()}, appRev)
		}, 10*time.Second, 500*time.Millisecond).Should(Succeed())

		By("Check secret Created with the expected trait-storage spec")
		hpa := &autoscalingv1.HorizontalPodAutoscaler{}
		Expect(k8sClient.Get(ctx, client.ObjectKey{
			Namespace: app.GetNamespace(),
			Name:      app.Spec.Components[0].Name,
		}, hpa)).Should(BeNil())

		Expect(k8sClient.Delete(ctx, hpa)).Should(BeNil())
		Expect(k8sClient.Delete(ctx, app)).Should(BeNil())
	})

})

const (
	scopeDefYaml = `apiVersion: core.oam.dev/v1beta1
kind: ScopeDefinition
metadata:
  name: healthscopes.core.oam.dev
  namespace: vela-system
spec:
  workloadRefsPath: spec.workloadRefs
  allowComponentOverlap: true
  definitionRef:
    name: healthscopes.core.oam.dev`

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

	unhealthyComponentDefYaml = `
apiVersion: core.oam.dev/v1beta1
kind: ComponentDefinition
metadata:
  name: unhealthy-worker
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
  status:
    healthPolicy: |-
      isHealth: false
`

	wDImportYaml = `
apiVersion: core.oam.dev/v1beta1
kind: WorkloadDefinition
metadata:
  name: worker-import
  namespace: vela-system
  annotations:
    definition.oam.dev/description: "Long-running scalable backend worker without network endpoint"
spec:
  definitionRef:
    name: deployments.apps
  extension:
    template: |
      import (
          "k8s.io/apps/v1"
          appsv1 "kube/apps/v1"
      )
      output: v1.#Deployment & appsv1.#Deployment & {
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

	tdImportedYaml = `apiVersion: core.oam.dev/v1alpha2
kind: TraitDefinition
metadata:
  name: ingress-import
  namespace: vela-system
spec:
  appliesToWorkloads:
    - "*"
  schematic:
    cue:
      template: |
        import (
        	kubev1 "k8s.io/core/v1"
        	network "k8s.io/networking/v1beta1"
        )

        parameter: {
        	domain: string
        	http: [string]: int
        }

        outputs: {
        service: kubev1.#Service
        ingress: network.#Ingress
        }

        // trait template can have multiple outputs in one trait
        outputs: service: {
        	metadata:
        		name: context.name
        	spec: {
        		selector:
        			"app.oam.dev/component": context.name
        		ports: [
        			for k, v in parameter.http {
        				port:       v
        				targetPort: v
        			},
        		]
        	}
        }

        outputs: ingress: {
        	metadata:
        		name: context.name
        	spec: {
        		rules: [{
        			host: parameter.domain
        			http: {
        				paths: [
        					for k, v in parameter.http {
        						path: k
        						backend: {
        							serviceName: context.name
        							servicePort: v
        						}
        					},
        				]
        			}
        		}]
        	}
        }`

	webComponentDefYaml = `apiVersion: core.oam.dev/v1alpha2
kind: ComponentDefinition
metadata:
  name: webserver
  namespace: vela-system
  annotations:
    definition.oam.dev/description: "webserver was composed by deployment and service"
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

      					if parameter["env"] != _|_ {
      						env: parameter.env
      					}

      					if context["config"] != _|_ {
      						env: context.config
      					}

      					ports: [{
      						containerPort: parameter.port
      					}]

      					if parameter["cpu"] != _|_ {
      						resources: {
      							limits:
      								cpu: parameter.cpu
      							requests:
      								cpu: parameter.cpu
      						}
      					}
      				}]
      		}
      		}
      	}
      }
      // workload can have extra object composition by using 'outputs' keyword
      outputs: service: {
      	apiVersion: "v1"
      	kind:       "Service"
      	spec: {
      		selector: {
      			"app.oam.dev/component": context.name
      		}
      		ports: [
      			{
      				port:       parameter.port
      				targetPort: parameter.port
      			},
      		]
      	}
      }
      parameter: {
      	image: string
      	cmd?: [...string]
      	port: *80 | int
      	env?: [...{
      		name:   string
      		value?: string
      		valueFrom?: {
      			secretKeyRef: {
      				name: string
      				key:  string
      			}
      		}
      	}]
      	cpu?: string
      }

`
	componentDefWithHealthYaml = `
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
    healthPolicy: |
      isHealth: context.output.status.readyReplicas == context.output.status.replicas 
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
	gatewayYaml = `apiVersion: core.oam.dev/v1beta1
kind: TraitDefinition
metadata:
  annotations:
    definition.oam.dev/description: Enable public web traffic for the component, the ingress API matches K8s v1.20+.
  name: gateway
  namespace: vela-system
spec:
  appliesToWorkloads:
    - '*'
  podDisruptive: false
  schematic:
    cue:
      template: |
        // trait template can have multiple outputs in one trait
        outputs: service: {
        	apiVersion: "v1"
        	kind:       "Service"
        	metadata: name: context.name
        	spec: {
        		selector: "app.oam.dev/component": context.name
        		ports: [
        			for k, v in parameter.http {
        				port:       v
        				targetPort: v
        			},
        		]
        	}
        }
        outputs: ingress: {
        	apiVersion: "networking.k8s.io/v1"
        	kind:       "Ingress"
        	metadata: {
        		name: context.name
        		annotations: {
        			if !parameter.classInSpec {
        				"kubernetes.io/ingress.class": parameter.class
        			}
        		}
        	}
        	spec: {
        		if parameter.classInSpec {
        			ingressClassName: parameter.class
        		}
        		if parameter.secretName != _|_ {
        			tls: [{
        				hosts: [
        					parameter.domain,
        				]
        				secretName: parameter.secretName
        			}]
        		}
        		rules: [{
        			host: parameter.domain
        			http: paths: [
        				for k, v in parameter.http {
        					path:     k
        					pathType: "ImplementationSpecific"
        					backend: service: {
        						name: context.name
        						port: number: v
        					}
        				},
        			]
        		}]
        	}
        }
        parameter: {
        	// +usage=Specify the domain you want to expose
        	domain: string

        	// +usage=Specify the mapping relationship between the http path and the workload port
        	http: [string]: int

        	// +usage=Specify the class of ingress to use
        	class: *"nginx" | string

        	// +usage=Set ingress class in '.spec.ingressClassName' instead of 'kubernetes.io/ingress.class' annotation.
        	classInSpec: *false | bool

        	// +usage=Specify the secret name you want to quote.
        	secretName?: string
        }
  status:
    customStatus: |-
      let igs = context.outputs.ingress.status.loadBalancer.ingress
      if igs == _|_ {
        message: "No loadBalancer found, visiting by using 'vela port-forward " + context.appName + "'\n"
      }
      if len(igs) > 0 {
        if igs[0].ip != _|_ {
      	  message: "Visiting URL: " + context.outputs.ingress.spec.rules[0].host + ", IP: " + igs[0].ip
        }
        if igs[0].ip == _|_ {
      	  message: "Visiting URL: " + context.outputs.ingress.spec.rules[0].host
        }
      }
    healthPolicy: 'isHealth: len(context.outputs.service.spec.clusterIP) > 0'

`
	cdDefWithHealthStatusYaml = `apiVersion: core.oam.dev/v1beta1
kind: ComponentDefinition
metadata:
  name: nworker
  namespace: vela-system
  annotations:
    definition.oam.dev/description: "Describes long-running, scalable, containerized services that running at backend. They do NOT have network endpoint to receive external network traffic."
spec:
  workload:
    definition:
      apiVersion: apps/v1
      kind: Deployment
  status:
    healthPolicy: |
      isHealth: (context.output.status.readyReplicas > 0) && (context.output.status.readyReplicas == context.output.status.replicas)
    customStatus: |-
      message: "type: " + context.output.spec.template.spec.containers[0].image + ",\t enemies:" + context.outputs.gameconfig.data.enemies
  schematic:
    cue:
      template: |
        output: {
        	apiVersion: "apps/v1"
        	kind:       "Deployment"
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
        					envFrom: [{
        						configMapRef: name: context.name + "game-config"
        					}]
        					if parameter["cmd"] != _|_ {
        						command: parameter.cmd
        					}
        				}]
        			}
        		}
        	}
        }

        outputs: gameconfig: {
        	apiVersion: "v1"
        	kind:       "ConfigMap"
        	metadata: {
        		name: context.name + "game-config"
        	}
        	data: {
        		enemies: parameter.enemies
        		lives:   parameter.lives
        	}
        }

        parameter: {
        	// +usage=Which image would you like to use for your service
        	// +short=i
        	image: string
        	// +usage=Commands to run in the container
        	cmd?: [...string]
        	lives:   string
        	enemies: string
        }
`
	workloadDefYaml = `
apiVersion: core.oam.dev/v1beta1
kind: WorkloadDefinition
metadata:
  name: task
  namespace: vela-system
  annotations:
    definition.oam.dev/description: "Describes jobs that run code or a script to completion."
spec:
  definitionRef:
    name: jobs.batch
  schematic:
    cue:
      template: |
        output: {
        	apiVersion: "batch/v1"
        	kind:       "Job"
        	spec: {
        		parallelism: parameter.count
        		completions: parameter.count
        		template: spec: {
        			restartPolicy: parameter.restart
        			containers: [{
        				name:  context.name
        				image: parameter.image
        
        				if parameter["cmd"] != _|_ {
        					command: parameter.cmd
        				}
        			}]
        		}
        	}
        }
        parameter: {
        	// +usage=specify number of tasks to run in parallel
        	// +short=c
        	count: *1 | int
        
        	// +usage=Which image would you like to use for your service
        	// +short=i
        	image: string
        
        	// +usage=Define the job restart policy, the value can only be Never or OnFailure. By default, it's Never.
        	restart: *"Never" | string
        
        	// +usage=Commands to run in the container
        	cmd?: [...string]
        }
`
	traitDefYaml = `
apiVersion: core.oam.dev/v1beta1
kind: TraitDefinition
metadata:
  annotations:
    definition.oam.dev/description: "Manually scale the app"
  name: scaler
  namespace: vela-system
spec:
  appliesToWorkloads:
    - deployments.apps
  definitionRef:
    name: manualscalertraits.core.oam.dev
  workloadRefPath: spec.workloadRef
  extension:
    template: |-
      outputs: scaler: {
      	apiVersion: "core.oam.dev/v1alpha2"
      	kind:       "ManualScalerTrait"
      	spec: {
      		replicaCount: parameter.replicas
      	}
      }
      parameter: {
      	//+short=r
      	replicas: *1 | int
      }

`
	tdDefYamlWithHttp = `
apiVersion: core.oam.dev/v1beta1
kind: TraitDefinition
metadata:
  annotations:
    definition.oam.dev/description: "Manually scale the app"
  name: scaler
  namespace: vela-system
spec:
  appliesToWorkloads:
    - deployments.apps
  definitionRef:
    name: manualscalertraits.core.oam.dev
  workloadRefPath: spec.workloadRef
  extension:
    template: |-
      outputs: scaler: {
      	apiVersion: "core.oam.dev/v1alpha2"
      	kind:       "ManualScalerTrait"
      	spec: {
          replicaCount: parameter.replicas
          token: processing.output.token
      	}
      }
      parameter: {
      	//+short=r
        replicas: *1 | int
        serviceURL: *"http://127.0.0.1:8090/api/v1/token?val=test-token" | string
      }
      processing: {
        output: {
          token ?: string
        }
        http: {
          method: *"GET" | string
          url: parameter.serviceURL
          request: {
              body ?: bytes
              header: {}
              trailer: {}
          }
        }
      }
`
	tDDefWithHealthYaml = `
apiVersion: core.oam.dev/v1beta1
kind: TraitDefinition
metadata:
  annotations:
    definition.oam.dev/description: "Manually scale the app"
  name: scaler
  namespace: vela-system
spec:
  appliesToWorkloads:
    - deployments.apps
  definitionRef:
    name: manualscalertraits.core.oam.dev
  workloadRefPath: spec.workloadRef
  extension:
    healthPolicy: |
      isHealth: context.output.status.conditions[0].status == "True"
    template: |-
      outputs: scaler: {
      	apiVersion: "core.oam.dev/v1alpha2"
      	kind:       "ManualScalerTrait"
      	spec: {
      		replicaCount: parameter.replicas
      	}
      }
      parameter: {
      	//+short=r
      	replicas: *1 | int
      }
`

	tDDefWithHealthStatusYaml = `apiVersion: core.oam.dev/v1beta1
kind: TraitDefinition
metadata:
  name: ingress
  namespace: vela-system
spec:
  status:
    customStatus: |-
      message: "type: "+ context.outputs.service.spec.type +",\t clusterIP:"+ context.outputs.service.spec.clusterIP+",\t ports:"+ "\(context.outputs.service.spec.ports[0].port)"+",\t domain"+context.outputs.ingress.spec.rules[0].host
    healthPolicy: |
      isHealth: len(context.outputs.service.spec.clusterIP) > 0
  schematic:
    cue:
      template: |
        parameter: {
        	domain: string
        	http: [string]: int
        }
        // trait template can have multiple outputs in one trait
        outputs: service: {
        	apiVersion: "v1"
        	kind:       "Service"
        	spec: {
        		selector:
        			app: context.name
        		ports: [
        			for k, v in parameter.http {
        				port:       v
        				targetPort: v
        			},
        		]
        	}
        }
        outputs: ingress: {
        	apiVersion: "networking.k8s.io/v1beta1"
        	kind:       "Ingress"
        	metadata:
        		name: context.name
        	spec: {
        		rules: [{
        			host: parameter.domain
        			http: {
        				paths: [
        					for k, v in parameter.http {
        						path: k
        						backend: {
        							serviceName: context.name
        							servicePort: v
        						}
        					},
        				]
        			}
        		}]
        	}
        }
`
	workloadWithContextRevision = `
apiVersion: core.oam.dev/v1beta1
kind: ComponentDefinition
metadata:
  name: worker-revision
  namespace: vela-system
  annotations:
    definition.oam.dev/description: "Long-running scalable backend worker without network endpoint"
spec:
  workload:
    definition:
      apiVersion: apps/v1
      kind: Deployment
  extension:
    healthPolicy: |
      isHealth: context.output.status.readyReplicas == context.output.status.replicas 
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
                      "app.oam.dev/revision": context.revision
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
      }`

	shareFsTraitDefinition = `
apiVersion: core.oam.dev/v1beta1
kind: TraitDefinition
metadata:
  name: share-fs
  namespace: default
spec:
  schematic:
    cue:
      template: |
        outputs: pv: {
        	apiVersion: "v1"
        	kind:       "PersistentVolume"
        	metadata: {
        		name:      context.name
        	}
        	spec: {
        		accessModes: ["ReadWriteMany"]
        		capacity: storage: "999Gi"
        		persistentVolumeReclaimPolicy: "Retain"
        		csi: {
        			driver: "nasplugin.csi.alibabacloud.com"
        			volumeAttributes: {
        				host: nasConn.MountTargetDomain
        				path: "/"
        				vers: "3.0"
        			}
        			volumeHandle: context.name
        		}
        	}
        }
        outputs: pvc: {
        	apiVersion: "v1"
        	kind:       "PersistentVolumeClaim"
        	metadata: {
        		name:      parameter.pvcName
        	}
        	spec: {
        		accessModes: ["ReadWriteMany"]
        		resources: {
        			requests: {
        				storage: "999Gi"
        			}
        		}
        		volumeName: context.name
        	}
        }
        parameter: {
        	pvcName: string
        	// +insertSecretTo=nasConn
        	nasSecret: string
        }
        nasConn: {
        	MountTargetDomain: string
        }
`
	rolloutTraitDefinition = `
apiVersion: core.oam.dev/v1beta1
kind: TraitDefinition
metadata:
  name: rollout
  namespace: default
spec:
  manageWorkload: true
  skipRevisionAffect: true
  schematic:
    cue:
      template: |
        outputs: rollout: {
        	apiVersion: "standard.oam.dev/v1alpha1"
        	kind:       "Rollout"
        	metadata: {
        		name:  context.name
                namespace: context.namespace
        	}
        	spec: {
                   targetRevisionName: parameter.targetRevision
                   componentName: "myweb1"
                   rolloutPlan: {
                   	rolloutStrategy: "IncreaseFirst"
                    rolloutBatches:[
                    	{ replicas: 3}]    
                    targetSize: 5
                   }
        		 }
        	}

         parameter: {
             targetRevision: *context.revision|string
         }
`
	applyCompWfStepDefinition = `
apiVersion: core.oam.dev/v1beta1
kind: WorkflowStepDefinition
metadata:
  annotations:
    definition.oam.dev/description: Apply components and traits for your workflow steps
  name: apply-component
  namespace: vela-system
spec:
  schematic:
    cue:
      template: |
        import (
        	"vela/op"
        )

        // apply components and traits
        apply: op.#ApplyComponent & {
        	component: parameter.component
        }
        parameter: {
        	// +usage=Declare the name of the component
        	component: string
        }
`

	k8sObjectsComponentDefinitionYaml = `
apiVersion: core.oam.dev/v1beta1
kind: ComponentDefinition
metadata:
  annotations:
    definition.oam.dev/description: K8s-objects allow users to specify raw K8s objects in properties
  name: k8s-objects
  namespace: vela-system
spec:
  schematic:
    cue:
      template: |
        output: parameter.objects[0]
        outputs: {
          for i, v in parameter.objects {
            if i > 0 {
              "objects-\(i)": v
            }
          }
        }
        parameter: objects: [...{}]
`
	applyInParallelWorkflowDefinitionYaml = `
apiVersion: core.oam.dev/v1beta1
kind: WorkflowStepDefinition
metadata:
  name: apply-test
  namespace: vela-system
spec:
  schematic:
    cue:
      template: |
        import (
                "vela/op"
                "list"
        )

        components:      op.#LoadInOrder & {}
        targetComponent: components.value[0]
        resources:       op.#RenderComponent & {
                value: targetComponent
        }
        workload:       resources.output
        arr:            list.Range(0, parameter.parallelism, 1)
        patchWorkloads: op.#Steps & {
                for idx in arr {
                        "\(idx)": op.#PatchK8sObject & {
                                value: workload
                                patch: {
                                        // +patchStrategy=retainKeys
                                        metadata: name: "\(targetComponent.name)-\(idx)"
                                }
                        }
                }
        }
        workloads: [ for patchResult in patchWorkloads {patchResult.result}]
        apply: op.#ApplyInParallel & {
                value: workloads
        }
        parameter: parallelism: int

`

	storageYaml = `apiVersion: core.oam.dev/v1beta1
kind: TraitDefinition
metadata:
  annotations:
    definition.oam.dev/description: Add storages on K8s pod for your workload which follows the pod spec in path 'spec.template'.
  name: storage
  namespace: vela-system
spec:
  appliesToWorkloads:
    - deployments.apps
  podDisruptive: true
  schematic:
    cue:
      template: |
        pvcVolumesList: *[
        		for v in parameter.pvc {
        		{
        			name: "pvc-" + v.name
        			persistentVolumeClaim: claimName: v.name
        		}
        	},
        ] | []
        configMapVolumesList: *[
        			for v in parameter.configMap if v.mountPath != _|_ {
        		{
        			name: "configmap-" + v.name
        			configMap: {
        				defaultMode: v.defaultMode
        				name:        v.name
        				if v.items != _|_ {
        					items: v.items
        				}
        			}
        		}
        	},
        ] | []
        secretVolumesList: *[
        			for v in parameter.secret if v.mountPath != _|_ {
        		{
        			name: "secret-" + v.name
        			secret: {
        				defaultMode: v.defaultMode
        				secretName:  v.name
        				if v.items != _|_ {
        					items: v.items
        				}
        			}
        		}
        	},
        ] | []
        emptyDirVolumesList: *[
        			for v in parameter.emptyDir {
        		{
        			name: "emptydir-" + v.name
        			emptyDir: medium: v.medium
        		}
        	},
        ] | []
        pvcVolumeMountsList: *[
        			for v in parameter.pvc {
        		if v.volumeMode == "Filesystem" {
        			{
        				name:      "pvc-" + v.name
        				mountPath: v.mountPath
        			}
        		}
        	},
        ] | []
        configMapVolumeMountsList: *[
        				for v in parameter.configMap if v.mountPath != _|_ {
        		{
        			name:      "configmap-" + v.name
        			mountPath: v.mountPath
        		}
        	},
        ] | []
        configMapEnvMountsList: *[
        			for v in parameter.configMap if v.mountToEnv != _|_ {
        		{
        			name: v.mountToEnv.envName
        			valueFrom: configMapKeyRef: {
        				name: v.name
        				key:  v.mountToEnv.configMapKey
        			}
        		}
        	},
        ] | []
        configMapMountToEnvsList: *[
        			for v in parameter.configMap if v.mountToEnvs != _|_ for k in v.mountToEnvs {
        		{
        			name: k.envName
        			valueFrom: configMapKeyRef: {
        				name: v.name
        				key:  k.configMapKey
        			}
        		}
        	},
        ] | []
        secretVolumeMountsList: *[
        			for v in parameter.secret if v.mountPath != _|_ {
        		{
        			name:      "secret-" + v.name
        			mountPath: v.mountPath
        		}
        	},
        ] | []
        secretEnvMountsList: *[
        			for v in parameter.secret if v.mountToEnv != _|_ {
        		{
        			name: v.mountToEnv.envName
        			valueFrom: secretKeyRef: {
        				name: v.name
        				key:  v.mountToEnv.secretKey
        			}
        		}
        	},
        ] | []
        secretMountToEnvsList: *[
        			for v in parameter.secret if v.mountToEnvs != _|_ for k in v.mountToEnvs {
        		{
        			name: k.envName
        			valueFrom: secretKeyRef: {
        				name: v.name
        				key:  k.secretKey
        			}
        		}
        	},
        ] | []
        emptyDirVolumeMountsList: *[
        				for v in parameter.emptyDir {
        		{
        			name:      "emptydir-" + v.name
        			mountPath: v.mountPath
        		}
        	},
        ] | []
        volumeDevicesList: *[
        			for v in parameter.pvc if v.volumeMode == "Block" {
        		{
        			name:       "pvc-" + v.name
        			devicePath: v.mountPath
        		}
        	},
        ] | []
        patch: spec: template: spec: {
        	// +patchKey=name
        	volumes: pvcVolumesList + configMapVolumesList + secretVolumesList + emptyDirVolumesList

        	containers: [{
        		// +patchKey=name
        		env: configMapEnvMountsList + secretEnvMountsList + configMapMountToEnvsList + secretMountToEnvsList
        		// +patchKey=name
        		volumeDevices: volumeDevicesList
        		// +patchKey=name
        		volumeMounts: pvcVolumeMountsList + configMapVolumeMountsList + secretVolumeMountsList + emptyDirVolumeMountsList
        	},...]

        }
        outputs: {
        	for v in parameter.pvc {
        		if v.mountOnly == false {
        			"pvc-\(v.name)": {
        				apiVersion: "v1"
        				kind:       "PersistentVolumeClaim"
        				metadata: name: v.name
        				spec: {
        					accessModes: v.accessModes
        					volumeMode:  v.volumeMode
        					if v.volumeName != _|_ {
        						volumeName: v.volumeName
        					}
        					if v.storageClassName != _|_ {
        						storageClassName: v.storageClassName
        					}

        					if v.resources.requests.storage == _|_ {
        						resources: requests: storage: "8Gi"
        					}
        					if v.resources.requests.storage != _|_ {
        						resources: requests: storage: v.resources.requests.storage
        					}
        					if v.resources.limits.storage != _|_ {
        						resources: limits: storage: v.resources.limits.storage
        					}
        					if v.dataSourceRef != _|_ {
        						dataSourceRef: v.dataSourceRef
        					}
        					if v.dataSource != _|_ {
        						dataSource: v.dataSource
        					}
        					if v.selector != _|_ {
        						dataSource: v.selector
        					}
        				}
        			}
        		}
        	}

        	for v in parameter.configMap {
        		if v.mountOnly == false {
        			"configmap-\(v.name)": {
        				apiVersion: "v1"
        				kind:       "ConfigMap"
        				metadata: name: v.name
        				if v.data != _|_ {
        					data: v.data
        				}
        			}
        		}
        	}

        	for v in parameter.secret {
        		if v.mountOnly == false {
        			"secret-\(v.name)": {
        				apiVersion: "v1"
        				kind:       "Secret"
        				metadata: name: v.name
        				if v.data != _|_ {
        					data: v.data
        				}
        				if v.stringData != _|_ {
        					stringData: v.stringData
        				}
        			}
        		}
        	}

        }
        parameter: {
        	// +usage=Declare pvc type storage
        	pvc?: [...{
        		name:              string
        		mountOnly:         *false | bool
        		mountPath:         string
        		volumeMode:        *"Filesystem" | string
        		volumeName?:       string
        		accessModes:       *["ReadWriteOnce"] | [...string]
        		storageClassName?: string
        		resources?: {
        			requests: storage: =~"^([1-9][0-9]{0,63})(E|P|T|G|M|K|Ei|Pi|Ti|Gi|Mi|Ki)$"
        			limits?: storage:  =~"^([1-9][0-9]{0,63})(E|P|T|G|M|K|Ei|Pi|Ti|Gi|Mi|Ki)$"
        		}
        		dataSourceRef?: {
        			name:     string
        			kind:     string
        			apiGroup: string
        		}
        		dataSource?: {
        			name:     string
        			kind:     string
        			apiGroup: string
        		}
        		selector?: {
        			matchLabels?: [string]: string
        			matchExpressions?: {
        				key: string
        				values: [...string]
        				operator: string
        			}
        		}
        	}]

        	// +usage=Declare config map type storage
        	configMap?: [...{
        		name:      string
        		mountOnly: *false | bool
        		mountToEnv?: {
        			envName:      string
        			configMapKey: string
        		}
        		mountToEnvs?: [...{
        			envName:      string
        			configMapKey: string
        		}]
        		mountPath?:   string
        		defaultMode: *420 | int
        		readOnly:    *false | bool
        		data?: {...}
        		items?: [...{
        			key:  string
        			path: string
        			mode: *511 | int
        		}]
        	}]

        	// +usage=Declare secret type storage
        	secret?: [...{
        		name:      string
        		mountOnly: *false | bool
        		mountToEnv?: {
        			envName:   string
        			secretKey: string
        		}
        		mountToEnvs?: [...{
        			envName:   string
        			secretKey: string
        		}]
        		mountPath?:   string
        		defaultMode: *420 | int
        		readOnly:    *false | bool
        		stringData?: {...}
        		data?: {...}
        		items?: [...{
        			key:  string
        			path: string
        			mode: *511 | int
        		}]
        	}]

        	// +usage=Declare empty dir type storage
        	emptyDir?: [...{
        		name:      string
        		mountPath: string
        		medium:    *"" | "Memory"
        	}]
        }
`

	envYaml = `apiVersion: core.oam.dev/v1beta1
kind: TraitDefinition
metadata:
  annotations:
    definition.oam.dev/description: Add env on K8s pod for your workload which follows the pod spec in path 'spec.template'
  labels:
    custom.definition.oam.dev/ui-hidden: "true"
  name: env
  namespace: vela-system
spec:
  appliesToWorkloads:
    - '*'
  schematic:
    cue:
      template: |
        #PatchParams: {
        	// +usage=Specify the name of the target container, if not set, use the component name
        	containerName: *"" | string
        	// +usage=Specify if replacing the whole environment settings for the container
        	replace: *false | bool
        	// +usage=Specify the  environment variables to merge, if key already existing, override its value
        	env: [string]: string
        	// +usage=Specify which existing environment variables to unset
        	unset: *[] | [...string]
        }
        PatchContainer: {
        	_params: #PatchParams
        	name:    _params.containerName
        	_delKeys: {for k in _params.unset {"\(k)": ""}}
        	_baseContainers: context.output.spec.template.spec.containers
        	_matchContainers_: [ for _container_ in _baseContainers if _container_.name == name {_container_}]
        	_baseContainer: *_|_ | {...}
        	if len(_matchContainers_) == 0 {
        		err: "container \(name) not found"
        	}
        	if len(_matchContainers_) > 0 {
        		_baseContainer: _matchContainers_[0]
        		_baseEnv:       _baseContainer.env
        		if _baseEnv == _|_ {
        			// +patchStrategy=replace
        			env: [ for k, v in _params.env if _delKeys[k] == _|_ {
        				name:  k
        				value: v
        			}]
        		}
        		if _baseEnv != _|_ {
        			_baseEnvMap: {for envVar in _baseEnv {"\(envVar.name)": envVar.value}}
        			// +patchStrategy=replace
        			env: [ for envVar in _baseEnv if _delKeys[envVar.name] == _|_ && !_params.replace {
        				name: envVar.name
        				if _params.env[envVar.name] != _|_ {
        					value: _params.env[envVar.name]
        				}
        				if _params.env[envVar.name] == _|_ {
        					value: envVar.value
        				}
        			}] + [ for k, v in _params.env if _delKeys[k] == _|_ && (_params.replace || _baseEnvMap[k] == _|_) {
        				name:  k
        				value: v
        			}]
        		}
        	}
        }
        patch: spec: template: spec: {
        	if parameter.containers == _|_ {
        		// +patchKey=name
        		containers: [{
        			PatchContainer & {_params: {
        				if parameter.containerName == "" {
        					containerName: context.name
        				}
        				if parameter.containerName != "" {
        					containerName: parameter.containerName
        				}
        				replace: parameter.replace
        				env:     parameter.env
        				unset:   parameter.unset
        			}}
        		}]
        	}
        	if parameter.containers != _|_ {
        		// +patchKey=name
        		containers: [ for c in parameter.containers {
        			if c.containerName == "" {
        				err: "containerName must be set for containers"
        			}
        			if c.containerName != "" {
        				PatchContainer & {_params: c}
        			}
        		}]
        	}
        }
        parameter: *#PatchParams | close({
        	// +usage=Specify the environment variables for multiple containers
        	containers: [...#PatchParams]
        })
        errs: [ for c in patch.spec.template.spec.containers if c.err != _|_ {c.err}]

`

	hubCpuScalerYaml = `apiVersion: core.oam.dev/v1beta1
kind: TraitDefinition
metadata:
  annotations:
    definition.oam.dev/description: Automatically scale the component based on CPU usage.
  labels:
    custom.definition.oam.dev/ui-hidden: "true"
  name: hubcpuscaler
  namespace: vela-system
spec:
  appliesToWorkloads:
    - deployments.apps
  controlPlaneOnly: true
  schematic:
    cue:
      template: |
        outputs: hubcpuscaler: {
        	apiVersion: "autoscaling/v1"
        	kind:       "HorizontalPodAutoscaler"
        	metadata: name: context.name
        	spec: {
        		scaleTargetRef: {
        			apiVersion: parameter.targetAPIVersion
        			kind:       parameter.targetKind
        			name:       context.name
        		}
        		minReplicas:                    parameter.min
        		maxReplicas:                    parameter.max
        		targetCPUUtilizationPercentage: parameter.cpuUtil
        	}
        }
        parameter: {
        	// +usage=Specify the minimal number of replicas to which the autoscaler can scale down
        	min: *1 | int
        	// +usage=Specify the maximum number of of replicas to which the autoscaler can scale up
        	max: *10 | int
        	// +usage=Specify the average CPU utilization, for example, 50 means the CPU usage is 50%
        	cpuUtil: *50 | int
        	// +usage=Specify the apiVersion of scale target
        	targetAPIVersion: *"apps/v1" | string
        	// +usage=Specify the kind of scale target
        	targetKind: *"Deployment" | string
        }
`
	affinityYaml = `apiVersion: core.oam.dev/v1beta1
kind: TraitDefinition
metadata:
  annotations:
    definition.oam.dev/description: affinity specify affinity and tolerationon K8s pod for your workload which follows the pod spec in path 'spec.template'.
  labels:
    custom.definition.oam.dev/ui-hidden: "true"
  name: affinity
  namespace: vela-system
spec:
  appliesToWorkloads:
    - '*'
  podDisruptive: true
  schematic:
    cue:
      template: |
        patch: spec: template: spec: {
        	if parameter.podAffinity != _|_ {
        		affinity: podAffinity: {
        			if parameter.podAffinity.required != _|_ {
        				requiredDuringSchedulingIgnoredDuringExecution: [
        					for k in parameter.podAffinity.required {
        						if k.labelSelector != _|_ {
        							labelSelector: k.labelSelector
        						}
        						if k.namespace != _|_ {
        							namespace: k.namespace
        						}
        						topologyKey: k.topologyKey
        						if k.namespaceSelector != _|_ {
        							namespaceSelector: k.namespaceSelector
        						}
        					}]
        			}
        			if parameter.podAffinity.preferred != _|_ {
        				preferredDuringSchedulingIgnoredDuringExecution: [
        					for k in parameter.podAffinity.preferred {
        						weight:          k.weight
        						podAffinityTerm: k.podAffinityTerm
        					}]
        			}
        		}
        	}
        	if parameter.podAntiAffinity != _|_ {
        		affinity: podAntiAffinity: {
        			if parameter.podAntiAffinity.required != _|_ {
        				requiredDuringSchedulingIgnoredDuringExecution: [
        					for k in parameter.podAntiAffinity.required {
        						if k.labelSelector != _|_ {
        							labelSelector: k.labelSelector
        						}
        						if k.namespace != _|_ {
        							namespace: k.namespace
        						}
        						topologyKey: k.topologyKey
        						if k.namespaceSelector != _|_ {
        							namespaceSelector: k.namespaceSelector
        						}
        					}]
        			}
        			if parameter.podAntiAffinity.preferred != _|_ {
        				preferredDuringSchedulingIgnoredDuringExecution: [
        					for k in parameter.podAntiAffinity.preferred {
        						weight:          k.weight
        						podAffinityTerm: k.podAffinityTerm
        					}]
        			}
        		}
        	}
        	if parameter.nodeAffinity != _|_ {
        		affinity: nodeAffinity: {
        			if parameter.nodeAffinity.required != _|_ {
        				requiredDuringSchedulingIgnoredDuringExecution: nodeSelectorTerms: [
        					for k in parameter.nodeAffinity.required.nodeSelectorTerms {
        						if k.matchExpressions != _|_ {
        							matchExpressions: k.matchExpressions
        						}
        						if k.matchFields != _|_ {
        							matchFields: k.matchFields
        						}
        					}]
        			}
        			if parameter.nodeAffinity.preferred != _|_ {
        				preferredDuringSchedulingIgnoredDuringExecution: [
        					for k in parameter.nodeAffinity.preferred {
        						weight:     k.weight
        						preference: k.preference
        					}]
        			}
        		}
        	}
        	if parameter.tolerations != _|_ {
        		tolerations: [
        			for k in parameter.tolerations {
        				if k.key != _|_ {
        					key: k.key
        				}
        				if k.effect != _|_ {
        					effect: k.effect
        				}
        				if k.value != _|_ {
        					value: k.value
        				}
        				operator: k.operator
        				if k.tolerationSeconds != _|_ {
        					tolerationSeconds: k.tolerationSeconds
        				}
        			}]
        	}
        }
        #labelSelector: {
        	matchLabels?: [string]: string
        	matchExpressions?: [...{
        		key:      string
        		operator: *"In" | "NotIn" | "Exists" | "DoesNotExist"
        		values?: [...string]
        	}]
        }
        #podAffinityTerm: {
        	labelSelector?: #labelSelector
        	namespaces?: [...string]
        	topologyKey:        string
        	namespaceSelector?: #labelSelector
        }
        #nodeSelecor: {
        	key:      string
        	operator: *"In" | "NotIn" | "Exists" | "DoesNotExist" | "Gt" | "Lt"
        	values?: [...string]
        }
        #nodeSelectorTerm: {
        	matchExpressions?: [...#nodeSelecor]
        	matchFields?: [...#nodeSelecor]
        }
        parameter: {
        	// +usage=Specify the pod affinity scheduling rules
        	podAffinity?: {
        		// +usage=Specify the required during scheduling ignored during execution
        		required?: [...#podAffinityTerm]
        		// +usage=Specify the preferred during scheduling ignored during execution
        		preferred?: [...{
        			// +usage=Specify weight associated with matching the corresponding podAffinityTerm
        			weight: int & >=1 & <=100
        			// +usage=Specify a set of pods
        			podAffinityTerm: #podAffinityTerm
        		}]
        	}
        	// +usage=Specify the pod anti-affinity scheduling rules
        	podAntiAffinity?: {
        		// +usage=Specify the required during scheduling ignored during execution
        		required?: [...#podAffinityTerm]
        		// +usage=Specify the preferred during scheduling ignored during execution
        		preferred?: [...{
        			// +usage=Specify weight associated with matching the corresponding podAffinityTerm
        			weight: int & >=1 & <=100
        			// +usage=Specify a set of pods
        			podAffinityTerm: #podAffinityTerm
        		}]
        	}
        	// +usage=Specify the node affinity scheduling rules for the pod
        	nodeAffinity?: {
        		// +usage=Specify the required during scheduling ignored during execution
        		required?: {
        			// +usage=Specify a list of node selector
        			nodeSelectorTerms: [...#nodeSelectorTerm]
        		}
        		// +usage=Specify the preferred during scheduling ignored during execution
        		preferred?: [...{
        			// +usage=Specify weight associated with matching the corresponding nodeSelector
        			weight: int & >=1 & <=100
        			// +usage=Specify a node selector
        			preference: #nodeSelectorTerm
        		}]
        	}
        	// +usage=Specify tolerant taint
        	tolerations?: [...{
        		key?:     string
        		operator: *"Equal" | "Exists"
        		value?:   string
        		effect?:  "NoSchedule" | "PreferNoSchedule" | "NoExecute"
        		// +usage=Specify the period of time the toleration
        		tolerationSeconds?: int
        	}]
        }
`
)

func newMockHTTP() *httptest.Server {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			fmt.Printf("Expected 'GET' request, got '%s'", r.Method)
		}
		if r.URL.EscapedPath() != "/api/v1/token" {
			fmt.Printf("Expected request to '/person', got '%s'", r.URL.EscapedPath())
		}
		r.ParseForm()
		token := r.Form.Get("val")
		tokenBytes, _ := json.Marshal(map[string]interface{}{"token": token})

		w.WriteHeader(http.StatusOK)
		w.Write(tokenBytes)
	}))
	l, _ := net.Listen("tcp", "127.0.0.1:8090")
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	return ts
}
