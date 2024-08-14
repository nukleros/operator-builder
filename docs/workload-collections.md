# Workload Collections

If you are building an operator to manage a collection of workloads that have
dependencies upon one another, you will need to use a workload collection.  If,
instead, you have just a single workload to manage, you will want to use a
[standalone workload](standalone-workloads.md).

A workload collection is defined by a configuration of kind `WorkloadCollection`.
The only differences are that it will
include an array of children `ComponentWorkloads` which are references
to the individual, managed workloads via the `componentFiles` field.

```yaml
name: acme-app-platform
kind: WorkloadCollection
spec:
  api:
    domain: apps.acme.com
    group: platform
    version: v1alpha1
    kind: AcmeAppPlatform
    clusterScoped: true
  companionCliRootcmd:
    name: platformctl
    description: Manage app platform services like a boss
  componentFiles:
    - ingress-workload.yaml
    - metrics-worklaod.yaml
    - logging-workload.yaml
    - admission-control-workload.yaml
```

Each of the `componentFiles` are `ComponentWorkload` configs that may have dependencies
upon one another which the operator will manage as identified by the
`dependencies` field (see below).  This project will include a
custom resource for each of the component workloads as well as a distinct
controller for each component workload.  Each of these controllers will run in a
single containerized controller manager in the cluster when it is deployed.

This collection WorkloadConfig includes a root command that will be used as a
companion CLI for the Acme App Platform operator.  Each of the component
workloads may also include a subcommand for use with that component.  It will
look something like this:

```yaml
name: metrics-component
kind: ComponentWorkload
spec:
  api:
    group: platform
    version: v1alpha1
    kind: MetricsComponent
    clusterScoped: false
  companionCliSubcmd:
    name: metrics
    description: Manage metrics in for the app platform like a boss
  dependencies:
    - ingress-component
  resources:
    - prom-operator.yaml
    - prometheus.yaml
    - alertmanager.yaml
    - grafana.yaml
    - prom-adapter.yaml
    - kube-state-metrics.yaml
    - node-exporter.yaml
```

Now the end uers of this operator - the platform operators - will be able to use
the companion CLI and issue commands like `platformctl metrics init` to
initialize a new MetricsComponent custom resource to use to deploy the
metrics-component workload.  Or they can use `platformctl metrics generate` to
ouptut a set of Kubernetes manifests configured by a supplied MetricsComponent
custom resource.

## Collection Resources

Collections can also have resources associated directly with them.  This is
useful for resources such as namespaces that are shared by the components in the
collection.  Consider the following workload collection with a resource
included.

```yaml
name: acme-complex-app
kind: WorkloadCollection
spec:
  api:
    domain: apps.acme.com
    group: tenant
    version: v1alpha1
    kind: AcmeComplexApp
    clusterScoped: true
  companionCliRootcmd:
    name: appctl
    description: Manage a really complex app
  resources:
    - namespace.yaml  # collection resource identified here
  componentFiles:
    - frontend-component.yaml
    - backend-component.yaml
    - service-x-component.yaml
    - service-y-component.yaml
    - service-z-component.yaml
```

Any markers included in these collection resources will configure fields for the
collection's custom resource.  For example the following marker will result in a
`namespace` field being included in the spec for a `AcmeComplexApp` resource:

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: complex-app # +operator-builder:field:name=namespace,type=string
```

If you want to add leverage that same `namespace` field from the
`AcmeComplexApp` field in a component resources ensure that you use a collection
marker.  In the following example this component service resource will derive it's
namespace from the `AcmeComplexApp` `spec.namespace` field.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: frontend-svc
  namespace: complex-app # +operator-builder:collection:field:name=namespace,type=string
spec:
  ports:
    - name: https
      port: 443
      protocol: TCP
  selector:
    app: frontend
```

## Next Step

Follow the [workload collection
tutorial](workload-collection-tutorial/workload-collection-tutorial.md).

