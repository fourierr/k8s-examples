## 可基于k8s/sample-controller开发controller，在没有kubebuiler时:

（1）修改sample-controller为app-controller

（2）生成 clientset listers informers deepcopy

        go mod vendor

        ./hack/update-codegen.sh

（3）修改reconcile逻辑

（4）生成crd

        获取controller-tools的代码并执行

        go install ./cmd/{controller-gen,type-scaffold}

        controller-gen crd paths=./... output:crd:dir=artifacts/crds