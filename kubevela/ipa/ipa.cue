ipodscaler: {
	type: "trait"
	annotations: {}
	labels: {
		"ui-hidden": "true"
	}
	description: "Automatically scale the component based on CPU usage."
	attributes: appliesToWorkloads: ["deployments.apps"]
}

template: {

	hpaMetrics: *[
			for m in parameter.intelligentPodAutoscaler.horizationPodAutoscaler.metrics {
			{
				  type: m.type
    			resource: {
    				name: m.resource.name
    				target: {
    					type: m.resource.target.type
    					averageUtilization: m.resource.target.averageUtilization
    				}
    			}
			}
		},
	] | []

	vpaMetrics: *[
			for m in parameter.intelligentPodAutoscaler.verticalPodAutoscaler.metrics {
			{
				  type: m.type
    			resource: {
    				name: m.resource.name
    				target: {
    					type: m.resource.target.type
    					averageUtilization: m.resource.target.averageUtilization
    				}
    			}
			}
		},
	] | []

	hpaCrons: *[
			for c in parameter.intelligentPodAutoscaler.horizationPodAutoscaler.crons {
			{
    			name: c.name
    			timezone: c.timezone
    			description: c.description
					start: c.start
					end: c.end
					targetReplicas: c.targetReplicas
			}
		},
	] | []

	vpaCrons: *[
			for c in parameter.intelligentPodAutoscaler.verticalPodAutoscaler.crons {
			{
    			name: c.name
    			timezone: c.timezone
    			description: c.description
					start: c.start
					end: c.end
					targetReplicas: c.targetReplicas
			}
		},
	] | []

	outputs: {
		for sTR in parameter.intelligentPodAutoscaler.scaleTargetRef {
			"\(sTR.name)-\(sTR.cluster)": {
				apiVersion: "autoscaling.oam.cmb/v1alpha1"
				kind:       "IntelligentPodAutoscaler"
				metadata: name: sTR.name+"-"+sTR.cluster
					spec: scaleTargetRef: {
						name:       context.name
						apiVersion: sTR.apiVersion
						kind:       sTR.kind
						cluster:    sTR.cluster
						namespace:  sTR.namespace
					}
					upDelay: sTR.upDelay
					downDelay: sTR.downDelay
					scaleStrategy: sTR.scaleStrategy
					horizationPodAutoscaler: {
						minReplicas: sTR.horizationPodAutoscaler.minReplicas
    		    maxReplicas: sTR.horizationPodAutoscaler.minReplicas
    		    metrics: hpaMetrics
    		    crons: hpaCrons
					}
					verticalPodAutoscaler: {
						minCPU: sTR.horizationPodAutoscaler.minCPU
    		    maxCPU: sTR.horizationPodAutoscaler.maxCPU
    		    minMemory: sTR.horizationPodAutoscaler.minMemory
    		    maxMemory: sTR.horizationPodAutoscaler.maxMemory
    		    metrics: vpaMetrics
    		    crons: vpaCrons
					}
			}
		}
	}

	parameter: {
		intelligentPodAutoscaler?: {
			name:      string
			scaleTargetRef: [...{
						apiVersion: *"apps/v1" | string
					  kind:       *"Deployment" | string
						cluster:      string
						namespace:    string
					}]
			upDelay: *15 | int32
			downDelay: *15 | int32
			scaleStrategy: *"DryRun" | "Auto"
			horizationPodAutoscaler?: {
				minReplicas: *1 | int32
    		maxReplicas: *10 | int32
    		metrics?: #Metrics
    		crons?: #Crons
			}
			verticalPodAutoscaler?: {
				minCPU: *1 | int32
    		maxCPU: *4 | int32
    		minMemory: *1 | int32
    		maxMemory: *4 | int32
    		metrics?: #Metrics
    		crons?: #Crons
			}
			prediction?: #Prediction
		}
	}

	#Metrics: [...{
			type: string
			resource: {
				name: string
				target: {
					type: string
					averageUtilization: int32
				}
			}
	}]

	#Crons: [...{
		name: string
		timezone: *"Local" | string
		description: string
		start: string
		end: string
		targetReplicas: int32
	}]

	#Prediction: {
    predictionWindowSeconds: *3600 | int32
    predictionAlgorithm?: {
    	algorithmType: string
      dsp?: {
      	sampleInterval: int32
        historyLength: int32
      }
    }
	}

}
