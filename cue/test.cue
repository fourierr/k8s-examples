package cue

import (
  "vela/op"
)

"deploy-webserver": {
    annotations: {}
    attributes: {}
    description: "deploy webserver and wait till it's running."
    labels: {}
    type: "workflow-step"
}

template: {
  // 部署应用中的所有资源
  apply: op.#ApplyApplication & {}

  resource: op.#Read & {
     value: {
       kind: "Deployment"
       apiVersion: "apps/v1"
       metadata: {
         name: "webserver-demo"
         // 可以使用 context 来获取该 Application 的任意元信息
         namespace: context.namespace
       }
     }
  }

  workload: resource.value
  // 等待 webserver 的 deployment 可用
  wait: op.#ConditionalWait & {
    continue: workload.status.readyReplicas == workload.status.replicas && workload.status.observedGeneration == workload.metadata.generation
  }
}
