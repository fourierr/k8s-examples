```yaml
apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  labels:
    project: lt31.04
  name: lt31-04-fourierapp01-dev-dev
  namespace: lt31-04-dev
spec:
  components:
    - name: fourierapp01-55259
      properties:
        componentEnv:
          clusterType: COMMON-APP
          netRegion: OA
        env:
          - name: CMB_CAAS_DEPLOYUNITID
            value: LT31.04_fourierapp01
          - name: CMB_CAAS_NAMESPACE
            value: lt31-04-dev-default
          - name: CMB_CAAS_APPLICATION
            value: fourierapp01-55259
          - name: CMB_CAAS_SERVICEUNITID
            value: LT31.04@fourierapp01_DEV_DEV
          - name: CMB_PAAS_ENV
            value: DEV
          - name: CMB_CAAS_PROJECT
            value: lt31.04
          - name: CMB_CLUSTER
            value: csdev-biz-06-qh
        image: csdev.registry.cmbchina.cn/lt31.04/fourierapp01:pl227952-1
        labels:
          app.alauda.io/name: fourierapp01-55259.lt31-04-dev-default
          cmb: lt31.04
          cmb-app-name: fourierapp01-55259
          cmb-log-tag: lt31-04-dev-default-Deployment-fourierapp01-55259
          cmb-org: lt31.04
          cmb-service-unit-id: LT31.04...fourierapp01_DEV_DEV
        matchLabels:
          app.alauda.io/name: fourierapp01-55259.lt31-04-dev-default
        namespace: lt31-04-dev-default-federation-csdev-biz-06
        ports:
          - port: 8080
      traits:
        # 资源规格和弹性配置
        - properties:
            limits:
              cpu: "1"
              memory: 1Gi
            replicas: 2
            requests:
              cpu: 100m
              memory: "357913941"
          type: resources
        # 外部路由和内部路由，以及域名
        - properties:
            domains:
              - domain: fourier01.paas.cmbchina.cn
                path: ""
                targetPort: "8080"
          type: route
        # java框架相关
        - properties:
            basePath: /actuator
            port: 10999
            types:
              - jvm
              - beeenv
          type: metrics
        - properties:
            url: /bee-scheduler-admin/index.html
          type: za21scheduler
        # labels 和 annos
        - properties:
            annotations:
              owners.alauda.io/info: '[{"name":"马祥博","phone":"17854227913","employee_id":"80310624@itc"}]'
            labels:
              app.alauda.io/name: fourierapp01-55259.lt31-04-dev-default
              cmb-app-name: fourierapp01-55259
              cmb-service-unit-id: LT31.04...fourierapp01_DEV_DEV
            matchLabels: null
          type: labels
        # 探针
        - properties:
            livenessProbe:
              failureThreshold: 3
              httpGet:
                path: /liveness
                port: 8080
                scheme: HTTP
              initialDelaySeconds: 30
              periodSeconds: 30
              successThreshold: 1
              timeoutSeconds: 5
            readinessProbe:
              failureThreshold: 3
              httpGet:
                path: /readiness
                port: 8080
                scheme: HTTP
              initialDelaySeconds: 30
              periodSeconds: 30
              successThreshold: 1
              timeoutSeconds: 5
          type: health-probe
        # 用户环境变量和配置
        - properties:
            envs:
              - name: JAVA-OPTS
                public: false
                value: "-Xss664K -Xms1296588K -XX:MetaspaceSize=199475K -Xmx1296588K -XX:MaxMetaspaceSize=199475K"
            refConfigs:
              - mountPath: /opt/test
                revisionHash: 077789d0b2319b8b45c7f3b2988b10dd
                uuid: 96293f752c4ebe0c
            restartTime: "2022-07-26T07:28:05.279075983Z"
          type: config
      type: webservice
  policies:
    - name: fourierapp01-55259
      properties:
        clusters:
          - csdev-biz-06-qh
          - csdev-biz-06-sk
        namespace: lt31-04-dev-default
      type: topology
  workflow:
    steps:
      - name: lt31-04-fourierapp01-dev-dev
        subSteps:
          - name: fourierapp01-55259
            properties:
              component: fourierapp01-55259
              topologyPolicies:
                - fourierapp01-55259
            type: deploy-cmb-upcase
        type: step-group

```