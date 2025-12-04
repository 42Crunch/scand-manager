# API Conformance Scan Jobs Manager for Kubernetes

API Conformance Scan Jobs Manager provides a convenient way to run 42Crunch API Conformance Scan on-premises as a [Kubernetes Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/) in your Kubernetes cluster.

## API Conformance Scan on premises

API Conformance Scan lets you perform dynamic application security testing (DAST) on your OpenAPI definitions. The scan tests that your API implementation matches the contract your API sets out in its OpenAPI definition.

The Docker image `scand-agent` lets you deploy and run Conformance Scan on premises rather than in 42Crunch API Security Platform. This way you can, for example, integrate Conformance Scan as a task that your CI/CD pipeline runs on every push to your repository to automate the testing.

## Kubernetes Jobs manager for Conformance Scan

Scan Jobs Manager is a service that exposes a REST API for starting and deleting Kubernetes Jobs and retrieving the job logs. You can use it to initiate Conformance Scans from your CI/CD pipeline, and combine it with with the 42Crunch CI/CD plugins.

## Installation

To add Scan Jobs Manager to your Kubernetes environment, you must first configure some details for the service and then deploy it.

### Configure the service

1. Create a file called `job-manager-config.yaml` with the following contents:

```yaml
# service account
apiVersion: v1
kind: ServiceAccount
metadata:
  name: api-sa
---
# role
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: api-role
rules:
  - apiGroups: ["batch", "extensions"]
    resources: ["jobs"]
    verbs: ["get", "create", "delete", "list"]
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["pods/log"]
    verbs: ["get"]
---
# role binding
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: api-sa-rolebinding
  labels:
    app: tools-rbac
subjects:
  - kind: ServiceAccount
    name: api-sa
roleRef:
  kind: Role
  name: api-role
  apiGroup: rbac.authorization.k8s.io
---
# deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
  name: scand-job-manager
spec:
  replicas: 1
  selector:
    matchLabels:
      app: scand-job-manager
  template:
    metadata:
      labels:
        app: scand-job-manager
    spec:
      serviceAccountName: api-sa
      containers:
        - image: 42crunch/scand-manager:v1
          name: scand-job-manager
          ports:
            - containerPort: 8090
          livenessProbe:
            httpGet:
              path: /health
              port: 8090
            initialDelaySeconds: 5
            periodSeconds: 10
            timeoutSeconds: 2
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /health
              port: 8090
            initialDelaySeconds: 3
            periodSeconds: 10
            timeoutSeconds: 2
            failureThreshold: 3
          env:
            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: PLATFORM_SERVICE
              value: services.us.42crunch.cloud:8001
            - name: SCAND_IMAGE
              value: 42crunch/scand-agent:latest
            - name: EXPIRATION_TIME
              value: "86400"
            # Example proxy settings that can be overridden as needed during job creation. 
            # These will be passed to scand-agent jobs by default if specified
            - name: HTTP_PROXY
              value: http://corp-proxy.example:8080
            - name: HTTPS_PROXY
              value: http://corp-proxy.example:8443
            - name: HTTP_PROXY_API
              value: http://api-proxy.example:8080
            - name: HTTPS_PROXY_API
              value: http://api-proxy.example:8443
      imagePullSecrets:
          # Pull secret for scand-manager container, if required
          # NOT for scand-agent jobs, that should be in podconfig.yaml
  		  - name: privatepullsecret 
---
# service
apiVersion: v1
kind: Service
metadata:
  name: scand-job-manager
spec:
  type: NodePort
  ports:
    - port: 8090
      targetPort: 8090
  selector:
    app: scand-job-manager
```

2. Under `env`, configure the following environment variables for Scan Jobs Manager to suit your environment:

| Variable           | Description                                                                                                                                                                                                                                                                                                    |
| ------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `NAMESPACE`        | The namespace where to create the Conformance Scan Jobs.                                                                                                                                                                                                                                                       |
| `PLATFORM_SERVICE` | The hostname and port for that Conformance Scan uses to connect to 42Crunch Platform. The default hostname for most users is `services.us.42crunch.cloud:8001`.                                                                                                                                                |
| `SCAND_IMAGE`      | The version of the Docker image `scand-agent` that the service pulls and runs for the on-premises scan. The default is `42crunch/scand-agent:latest`. For more details on the available images, see the [release notes of 42Crunch Platform](https://docs.42crunch.com/latest/content/whatsnew/whats_new.htm). |
| `EXPIRATION_TIME`  | The expiration time for the jobs (in seconds). Completed jobs are deleted after the specified time. The default value is `86400` (24 hours). Requires Kubernetes v1.21 or newer, for older Kubernetes versions jobs must be cleaned up manually using provided API or `kubectl`.                               |
| `HTTP(S)_PROXY(_API)`| Here you can specify `HTTP_PROXY`, `HTTPS_PROXY`, `HTTP_PROXY_API`, and/or `HTTPS_PROXY_API` as the default values for these environmental runtime variables.  These can be changed at runtime in the job submission env as well. |

### Healthcheck

If you would like to configure a healthcheck, there is a /health endpoint that will return 200 `{"status":"OK"}` if the pod is healthy.  In the example above, we added the following to  `spec: containers`

``` yaml
          livenessProbe:
            httpGet:
              path: /health
              port: 8090
            initialDelaySeconds: 5
            periodSeconds: 10
            timeoutSeconds: 2
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /health
              port: 8090
            initialDelaySeconds: 3
            periodSeconds: 10
            timeoutSeconds: 2
            failureThreshold: 3
```


### Optionally configure POD rules

Scand Manager can be configured to specify pod affinity, securityContext, resources, or imagePullSecrets for the jobs it creates.

These can be specified by providing optional command line argument `-podconfig` poinining to `.yaml` or `.yml` file that describes affinity, similar to:

```yaml
affinity:
  nodeAffinity: ...
```

See the [detailed format for the nodes under `affinity` key here](https://kubernetes.io/docs/tasks/configure-pod-container/assign-pods-nodes-using-node-affinity/)

We can also specify a pull secret array for the pod

```yaml
imagePullSecrets: 
  name: ...
```
See the docs for [`imagePullSecrets` key here](https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod)

Launched scand-agent container restrictions can also be supplied

```yaml
containers:
  - securityContext:
      runAsUser: 1000
      runAsGroup: 1000
      capabilities:
        drop:
          - ALL
    resources:
      limits:
        memory: "512Mi"
        cpu: "500m"
      requests:
        memory: "256Mi"
        cpu: "200m"
```
See the docs for [`securityContext` key here](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) and [`resources` key here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)

An example podconfig.yaml would be:

```yaml
apiVersion: v1
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
        - matchExpressions:
            - key: scandallowed
              operator: In
              values:
                - "true"
imagePullSecrets:
  - name: secret1
  - name: privatepullsecret
containers:
  - securityContext:
      runAsUser: 1000
    resources:
      limits:
        cpu: "1"
        memory: "512Mi"

```
This would have an affinity for any cluster node tagged with `scandallowed: true` and would attempt to use the k8s secrets `secret1` or `privatepullsecret` to pull `SCAND_IMAGE` from your private container registry.  This would also specify resource limits and associated securityContexts.

Typically the podconfig yaml file should be supplied via a config map and mounted inside Scan Jobs Manager container and path to it supplied through `args`: 

```yaml
containers:
  - image: 42crunch/scand-manager:v1
    ...
    args:
      - -podconfig
      - /config/podconfig.yaml
    ...
    volumeMounts:
      - name: config
        mountPath: /config
        readOnly: true
volumes:
  - name: config
    configMap:
      name: scandpodconfig

```

In this example, we would create the configmap (assuming we deployed in the scand-manager namespace):



### Deployment

To deploy Scan Jobs Manager, run the following commands to create a separate namespace and apply the configuration you defined:

Create the namespace:

`kubectl create namespace scand-manager`

Create the configmap:

`kubectl create configmap scandpodconfig --from-file=podconfig.yaml -n scand-manager`

Create a secret, if required:

* For example, this would be for a private Docker Hub repo

	`kubectl create secret docker-registry privatepullsecret --docker-username={Your Username} --docker-password={Access Token} --docker-email={Your Email} -n scand-manager`

Deploy:

`kubectl apply -n scand-manager -f job-manager-config.yaml`
