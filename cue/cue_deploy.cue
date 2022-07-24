package cue

import (
   apps "kube/apps/v1"
)

parameter: {
    name:  string
}

output: apps.#Deployment
output: {
    metadata: name: parameter.name
}