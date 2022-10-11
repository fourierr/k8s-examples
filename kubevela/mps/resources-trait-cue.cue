resources: {
	type: "trait"
	annotations: {}
	labels: {
		"ui-hidden": "true"
	}
	description: "Add resource requests and limits, replicas, podScaler on K8s pod for your workload which follows the pod spec in path 'spec.template.'"
	attributes: {
		podDisruptive: true
		appliesToWorkloads: ["*"]
	}
}
template: {
	patch: spec: {
			if context.output.kind != "CronJob" {
				// +patchStrategy=retainKeys
				if parameter.execReplicas != _|_ {
					replicas: parameter.execReplicas
				}
				if parameter.execReplicas == _|_ {
					replicas: parameter.replicas
				}
				template: spec: containers: [...{
					resources: {
						if parameter.requests != _|_ {
							requests: {
								cpu:    parameter.requests.cpu
								memory: parameter.requests.memory
							}
						}
						if parameter.execCPU != _|_ && parameter.execMemory != _|_{
							limits: {
								cpu:    parameter.execCPU
								memory: parameter.execMemory
							}
						}
						if parameter.execCPU == _|_ || parameter.execMemory == _|_ {
							limits: {
								cpu:    parameter.limits.cpu
								memory: parameter.limits.memory
							}
						}
					}
				}]
			}
			if context.output.kind == "CronJob" {
				jobTemplate: spec: template: spec: {
					//patchKey=name
					containers: [{
						resources: {
							if parameter.requests != _|_ {
								requests: {
									cpu:    parameter.requests.cpu
									memory: parameter.requests.memory
								}
							}
							if parameter.limits != _|_ {
								limits: {
									cpu:    parameter.limits.cpu
									memory: parameter.limits.memory
								}
							}
						}
					}]
				}
			}
		}

	cronPodScalerList: *[
		for c in parameter.cronPodScaler if c.name != _|_ {
				{
					name: c.name
					if c.description != _|_ {
						description: c.description
					}
					start: c.start
					if c.end != _|_ {
						end: c.end
					}
					if c.targetReplicas != _|_ {
						targetReplicas: c.targetReplicas
					}
					if c.targetMemory != _|_ {
						targetMemory: c.targetMemory
					}
					if c.targetCPU != _|_ {
						targetCPU: c.targetCPU
					}
					if c.targetJAVAOPTS != _|_ {
						targetJAVAOPTS: c.targetJAVAOPTS
					}
				}
			},
	] | []

	triggerList: *[
		for t in parameter.metricPodScaler.triggers if t.value != _|_ {
				{
					type: t.type
					metricType: t.metricType
					value: t.value
					if t.externalKey != _|_ {
						externalKey: t.externalKey
					}
					if t.externalValue != _|_ {
						externalValue: t.externalValue
					}
				}
			},
	] | []

  metricTargetValue: *{
  	if parameter.metricPodScaler.metricTarget.metricReplicas != _|_ {
				metricReplicas: parameter.metricPodScaler.metricTarget.metricReplicas
			}
		if parameter.metricPodScaler.metricTarget.metricCPU != _|_ {
				metricCPU: parameter.metricPodScaler.metricTarget.metricCPU
			}
		if parameter.metricPodScaler.metricTarget.metricMemory != _|_ {
				metricMemory: parameter.metricPodScaler.metricTarget.metricMemory
			}
  } | {}

  podScalerConfigValue: *{
  	priorityStrategy: parameter.podScalerConfig.priorityStrategy
  	cooldownPeriod: parameter.podScalerConfig.cooldownPeriod
  	coolupPeriod: parameter.podScalerConfig.coolupPeriod
  	minReplicas: parameter.podScalerConfig.minReplicas
  	maxReplicas: parameter.podScalerConfig.maxReplicas
  	minCPU: parameter.podScalerConfig.minCPU
  	maxCPU: parameter.podScalerConfig.maxCPU
  	minMemory: parameter.podScalerConfig.minMemory
  	maxMemory: parameter.podScalerConfig.maxMemory
  } | {}

	flag: parameter.metricPodScaler != _|_ || len(parameter.cronPodScaler) != 0

	outputs: {
			if parameter.generateCR == true {
				"PodScaler": {
					apiVersion: "scaler.oam.cmb/v1alpha1"
					kind: "PodScaler"
					metadata: {
						name: context.appName + "-" + context.name
						finalizers: [ "scaler.oam.cmb/oamscaler-tracker-finalizer" ]
						labels: {
							"scaler.oam.cmb/sharding": parameter.shardingInfo
						}
					}
					spec: {
						scaleTargetRef: {
							clusters: context.topologyClusters
							namespace: context.topologyNamespace
							appName: context.appName
							compName: context.name
							serviceUnitID: context.serviceUnitID
						}
						scaleStrategy: "Auto"
						podScalerConfig: podScalerConfigValue
						if len(parameter.cronPodScaler) != 0 {
							cronPodScaler: cronPodScalerList
						}
						if parameter.metricPodScaler != _|_ {
							if parameter.metricPodScaler.metricTarget != _|_ {
								metricPodScaler: {
									metricTarget: metricTargetValue
								}
							}
							if parameter.metricPodScaler.triggers != _|_ {
								metricPodScaler: {
									triggers: triggerList
								}
							}
						}
				  }
				}
			}
	}

	parameter: {
			// +usage=Specify the number of workload
			replicas: *1 | int
			// +usage=Specify the resources in requests
			requests?: {
				// +usage=Specify the amount of cpu for requests
				cpu: *"0.5" | string
				// +usage=Specify the amount of memory for requests
				memory: *"256Mi" | string
			}
			// +usage=Specify the resources in limits
			limits?: {
				// +usage=Specify the amount of cpu for limits
				cpu: *"1" | string
				// +usage=Specify the amount of memory for limits
				memory: *"2048Mi" | string
			}
			javaOpts?: string
			execReplicas?:  int
			execCPU?:  string
			execMemory?:  string
			scaleStrategy: *"Auto" | string
			cronPodScaler?: #CronPodScaler
			metricPodScaler?: #MetricPodScaler
		  podScalerConfig: #PodScalerConfig
			// +usage=Specify the labels in the workload
			shardingInfo: *"sharding-1"| string
			generateCR: "false" | bool
	}

	#CronPodScaler: [...{
			name: string
			description?: string
			start: string
			end: string
			targetReplicas?: int32
			targetMemory?: string
			targetCPU?: string
			targetJAVAOPTS?: string
		} ]

	#MetricPodScaler: {
		metricTarget?: {
			metricReplicas?: int32
			metricCPU?: string
			metricMemory?: string
		}
		triggers: [...{
			type: string
			metricType: string
			value: int32
			externalKey?: string
			externalValue?: {
				jdkVersion: "1.8" | string
				reservedMemory: *10 | int32
				directMemory: *10 | int32
				heap: *65 | int32
				metaspace: *10 | int32
				native: *15 | int32
				stack: *10 | int32
			}
		} ]
	}

	#PodScalerConfig: {
		priorityStrategy: *"ReplicasFirst" | string
		cooldownPeriod: *900 | int32
		coolupPeriod: *300 | int32
		minReplicas: *1 | int32
		maxReplicas: *8 | int32
		minCPU: *"0.2" | string
		maxCPU: *"8" | string
		minMemory: *"0.2Gi" | string
		maxMemory: *"8Gi" | string
	}
}
