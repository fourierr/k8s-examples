anno: {
	type: "trait"
	annotations: {}
	labels: {
		"ui-hidden": "true"
	}
	description: "Add annotations on K8s pod for your workload which follows the pod spec in path 'spec.template'."
	attributes: {
		podDisruptive: true
		appliesToWorkloads: ["*"]
	}
}
template: {
	// +patchStrategy=retainKeys
	patch: {
		metadata: {
			annotations: {
				for k, v in parameter {
					"\(k)": v
				}
				"fourier-policy-test1": context.topologyClusters
				"fourier-policy-test2": context.topologyNamespace
				"fourier-policy-test3": context.policies
			}
		}
	}
	parameter: [string]: string | null
}
