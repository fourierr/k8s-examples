
Dispatch resource within a component sequentially.

**Is your feature request related to a problem? Please describe.**
- In our scenario, we define several traits to generate `Custom Resource` for completing specific tasks. 
- In some cases, trait-A need to dispatch before trait-B and workload.
- In k8s's case, when we modify cpu parameter and configmap parameter, the configmap should be dispatched before deployment.
So, the ability that dispatch resource within a component sequentially is necessary.


**Describe alternatives you've considered**

We want to add a field `inOrder` align with `dependsOn`,
and serially dispatch resource between steps and dispatch resource within steps in parallel.
There is an example for `inOrder` as follows
```yaml
apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: fourierapp03
  namespace: fourier
spec:
  components:
    - name: fourierapp03-comp-01
      type: webservice
      inOrder:
        - step:
            - config
        - step:
            - workload
            - cpuscaler
      properties:
        image: nginx:latest
        ports:
          - expose: true
            port: 80
            protocol: TCP
        imagePullPolicy: IfNotPresent
      traits:
        - type: cpuscaler
          properties:
            min: 1
            max: 10
        - type: config
          properties:
            refConfigs:
              - name: fourierapp03-comp-01
                encrypt: false
                mountPath: /mount/test/cm/tomcat.yaml
 
  policies:
    - name: fourierapp03-comp-01
      type: topology
      properties:
        clusters: [ "local" ]
        namespace: fourier01
 
  workflow:
    steps:
      - name: step-group
        type: step-group
        subSteps:
          - name: fourierapp03-deploy-01
            type: deploy
            properties:
              policies: [ "fourierapp03-comp-01" ]
```
