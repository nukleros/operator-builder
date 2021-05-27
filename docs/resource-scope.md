# Resource Scope

Kubernetes resources are either cluster-scoped or namespace scoped.  The custom
resources (CRs) created with workload-api can be either.  By default, they are
namespace-scoped.  If you wish to create a cluster-scoped CR, you
will need to specify in the WorkloadConfig manifest as shown in this example:

    name: webapp
    spec:
      group: apps
      version: v1alpha1
      kind: WebApp
      clusterScoped: false  # <-- indicates custom resource should be cluster-scoped
      companionCliRootcmd:
        name: webappctl
        description: Manage webapp stuff like a boss
      resources:
      - app.yaml


In general, you will want to use the default namespace-scoped CR
unless your CR will be a parent for cluster-scoped resources, e.g. namespaces or
CRDs.

NOTE: The scope of your CR will have a bearing on the the
namespace for your CR's child resources, i.e. the resources that are owned and
configured by your CR.
1. Namespace-scoped: If your CR is namespace-scoped, all child resources will be
   created in the namespace your CR is created in.  In this case the lifecycle
   of the namespace is managed outside of your operator.
2. Cluster-scoped: If your CR is cluster-scoped, you will need to include a
   `metadata.namespace` field in all applicable source manifests and include a
   field in your CR to set that value.  This is because your CR will not have a namespace
   and so the namespace must be assigned by the operator.  In this case the
   lifecycle of the namespace may be managed by your operator.

