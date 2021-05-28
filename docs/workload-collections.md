# Workload Collections

If you are building an operator to manage a collection of workloads that have
dependencies upon one another, you will need to use a workload collection.  If,
instead, you have just a single workload to manage, you will want to use a
[standalone workload](standalone-workloads.md).

A workload collection is defined by a WorkloadConfig - just like any other
workload used with Operator Builder.  The only differences are that it will
include `collection: true` and an array of workload definitions under the
`children` field.

    name: acme-app-platform
    spec:
      group: platform
      version: v1alpha1
      kind: AcmeAppPlatform
      clusterScoped: true
      companionCliRootcmd:
        name: platformctl
        description: Manage app platform services like a boss
      children:
      - ingress-workload.yaml
      - metrics-worklaod.yaml
      - logging-workload.yaml
      - admission-control-workload.yaml

Each of the `children` are component WorkloadConfigs that may have dependencies
upon one another which the operator will manage.  This project will include a
custom resource for each of the component workloads as well as a distinct
controller for each component workload.  Each of these controllers will run in a
single containerized controller manager in the cluster when it is deployed.

This collection WorkloadConfig includes a root command that will be used as a
companion CLI for the Acme App Platform operator.  Each of the component
workloads may also include a subcommand for use with that component.  It will
look something like this:

    name: metrics-component
    spec:
      group: platform
      version: v1alpha1
      kind: MetricsComponent
      clusterScoped: false
      companionCliSubcmd:
        name: metrics
        description: Manage metrics in for the app platform like a boss
      resources:
      - prom-operator.yaml
      - prometheus.yaml
      - alertmanager.yaml
      - grafana.yaml
      - prom-adapter.yaml
      - kube-state-metrics.yaml
      - node-exporter.yaml

Now the end uers of this operator - the platform operators - will be able to use
the companion CLI and issue commands like `platformctl meteics init` to
initialize a new MetricsComponent custom resource to use to deploy the
metrics-component workload.  Or they can use `platformctl meteics generate` to
ouptut a set of Kubernetes manifests configured by a supplied MetricsComponent
custom resource.

