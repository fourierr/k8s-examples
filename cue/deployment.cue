package cue

// 导出渲染结果为 YAML 格式: cue export deployment.cue -e template --out yaml > deployment.yaml
parameter: {
    name: string
    image: string
}

template: {
    apiVersion: "apps/v1"
    kind:       "Deployment"
    spec: {
        selector: matchLabels: {
            "app.oam.dev/component": parameter.name
        }
        template: {
            metadata: labels: {
                "app.oam.dev/component": parameter.name
            }
            spec: {
                containers: [{
                    name:  parameter.name
                    image: parameter.image
                }]
            }}}
}

parameter:{
   name: "mytest"
   image: "nginx:v1"
}
