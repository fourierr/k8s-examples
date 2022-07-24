package cue

parameter: {
    name:  string
    image: string
    env: [...{name:string,value:string}]
}
output: {
     spec: {
        containers: [{
            name:  parameter.name
            image: parameter.image
            env: [
                for _, v in parameter.env {
                    name:  v.name
                    value: v.value
                },
            ]
        }]
    }
}

parameter:{
   name: "mytest"
   image: "nginx:v1"
   env: [
   	{name:"a",value:"b"},
   	{name:"c",value:"d"},
   ]
}