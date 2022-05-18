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

### Optionally configure POD affinity rules

Scan Jobs Manager can be configured to specify pod affinity for the jobs it creates.

Pod affinity can be specified by providing optional command line argument `-podconfig` poinining to `.yaml` or `.yml` file that describes affinity, similar to:

```yaml
affinity:
  nodeAffinity: ...
```

See the [detailed format for the nodes under `affinity` key here](https://kubernetes.io/docs/tasks/configure-pod-container/assign-pods-nodes-using-node-affinity/)

Typically the podconfig yaml file should be supplied via a config map and mounted insde Scan Jobs Manager container and path to it supplied throuh `args`:

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
volumes:
  - name: config
    configMap:
      name: podconfig
```

### Deployment

To deploy Scan Jobs Manager, run the following commands to create a separate namespace and apply the configuration you defined:

`kubectl create namespace scan`

`kubectl apply -n scan -f job-manager-config.yaml`
