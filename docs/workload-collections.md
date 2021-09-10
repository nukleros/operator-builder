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
