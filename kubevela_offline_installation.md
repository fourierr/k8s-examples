# KubeVela Offline Installation
## Vela CLi
- Download the vela cli binary file via [release log](https://github.com/oam-dev/kubevela/releases).
- Unzip the binary file, and configure the environment variables in $PATH.
  - unzip the binary file via 
    - `tar -zxvf vela-v1.2.5-linux-amd64.tar.gz`
    - `mv ./linux-amd64/vela /usr/local/bin/vela`
  - set env variables via 
    - `vi /etc/profile`
    - `export PATH="$PATH:/usr/local/bin"`
    - `source /etc/profile`
  - validate cli installation via `vela version`, and check out the outcome
  ```shell script
    CLI Version: v1.2.5
    Core Version:
    GitRevision: git-ef80b66
    GolangVersion: go1.17.7
  ```
## Vela Core
- install helm in private environment and helm v3.2.0+ required
  - install helm via [installing helm](https://helm.sh/docs/intro/install/)
  - check helm version via `helm version`
- prepare docker image, vela core used five image
  - pull image from dockerhub via
    - `docker pull oamdev/vela-core:v1.2.5`
    - `docker pull oamdev/cluster-gateway:v1.1.7`
    - `docker pull oamdev/kube-webhook-certgen:v2.3`
    - `docker pull oamdev/alpine-k8s:1.18.2`
    - `docker pull oamdev/hello-world:v1`
  - save image to local via
    - `docker save -o vela-core.tar oamdev/vela-core:v1.2.5`
    - `docker save -o cluster-gateway.tar oamdev/cluster-gateway:v1.1.7`
    - `docker save -o kube-webhook-certgen.tar oamdev/kube-webhook-certgen:v2.3`
    - `docker save -o alpine-k8s.tar oamdev/alpine-k8s:1.18.2`
    - `docker save -o hello-world.tar oamdev/hello-world:v1`
  - load image to private environment
    - `docker load vela-core.tar`
    - `docker load cluster-gateway.tar`
    - `docker load kube-webhook-certgen.tar`
    - `docker load alpine-k8s.tar`
    - `docker load hello-world.tar`
- download the source code via [KubeVela Source Code](https://github.com/oam-dev/kubevela/releases) and repackage helm chart
  - repackage the source via `helm package kubevela/charts/vela-core --destination kubevela/charts`
  - install the package offline via `helm install --create-namespace -n vela-system kubevela kubevela/charts/vela-core-0.1.0.tgz --wait`
  - check out the outcome
    ```shell script
      KubeVela control plane has been successfully set up on your cluster.
    ```
## VelaUX
- download the source code via [Catalog Source Code](https://github.com/oam-dev/catalog) and copy to private environment
- prepare docker image, vela ux used two image
  - pull image via
    - `docker pull oamdev/vela-apiserver:v1.2.5`
    - `docker pull oamdev/velaux:v1.2.5`
  - save image to local via
    - `docker save -o vela-apiserver.tar oamdev/vela-apiserver:v1.2.5`
    - `docker save -o velaux.tar oamdev/velaux:v1.2.5`
  - load image to private env
      - `docker load vela-apiserver.tar`
      - `docker load velaux.tar`
- enable velaUX in local
  - enable velaUX locally via `vela addon enable catalog-master/addons/velaux`
  - checkout the outcome
    ```shell script
      Addon: velaux enabled Successfully.
    ```
  - setup with ingress or openshift route and domain, e.g.:
    ```yaml
      apiVersion: route.openshift.io/v1
      kind: Route
      metadata:
      labels:
        cmb-route-default: 'no'
        cmb-route-sharding: sharding-1
      name: velaux-route
      namespace: vela-system
      spec:
       host: velaux.xxx.xxx.cn
       port:
         targetPort: 80
       to:
         kind: Service
         name: velaux
         weight: 100
       wildcardPolicy: None
    ```
  - validate velaUX installation
    ```shell script
      curl -I -m 10 -o /dev/null -s -w %{http_code} http://velaux.xxx.xxx.cn/applications
    ```
    expected output
    ```shell script
      200
    ```