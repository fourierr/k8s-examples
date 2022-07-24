package cue

parameter: {
    name:  string
    image: string
    env: [string]: string
}
output: {
    spec: {
        containers: [{
            name:  parameter.name
            image: parameter.image
            env: [
                for k, v in parameter.env {
                    "\(k)":  v
                },
            ]
        }]
    }
}
parameter:{
   name: "mytest"
   image: "nginx:v1"
   env: {
   	"a1": "b1"
   	"a2": "b2"
   	"a3": "b3"
   }
}