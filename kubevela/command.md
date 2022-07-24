
### cd
```shell
cd /mnt/d/GoLand/workspace/src/k8s-examples/kubevela/example/
```

### apply
```shell
kubectl apply -f svc01.yaml
kubectl apply -f /mnt/d/GoLand/workspace/src/k8s-examples/kubevela/example/svc01.yaml
```

### get
```shell
kubectl get app -n fourier -o yaml fourierapp02
kubectl get deploy -n fourier fourierapp02-fouriercomponent-01 -o yaml
```

### edit
```shell
kubectl edit deploy -n vela-system kubevela-vela-core
```

### delete
```shell
kubectl delete app -n fourier fourierapp02
```

### scale
```shell
kubectl scale deploy -n vela-system kubevela-vela-core --replicas=0
```

### 备注
```shell
- --use-webhook=true
- --webhook-port=9443
- --webhook-cert-dir=/etc/k8s-webhook-certs
```
