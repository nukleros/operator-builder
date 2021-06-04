# Operator Builder

Accelerate the development of Kubernetes Operators.

Operator Builder extends [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
to facilitate development and maintenance of Kubernetes operators.  It is especially
helpful if you need to take large numbers of resources defined with static or
templated yaml and migrate to managing those resources with a custom Kubernetes operator.

An operator built with Operator Builder has the following features:
- A defined API for a custom resource based on [workload
  markers](docs/workload-markers.md).
- A functioning controller that will create, update and delete child resources
  to reconcile the state for the custom resource/s.
- A [companion CLI](docs/companion-cli.md) that helps end users with common
  operations.

Operator Builder uses a [workload configuration](docs/workloads.md) as the
primary configuration mechanism for providing attributes for the source code.

The custom resource defined in the source code can be cluster-scoped or
namespace-scoped based on the requirements of the project.  More info
[here](docs/resource-scope.md).

## Collections

Operator Builder can generate source code for operators that manage multiple
workloads.  See [collections](docs/collections.md) for more info.

## Licensing

Operator Builder can help manage licensing for the resulting project.  More
info [here](docs/licensing.md).

## Testing

Testing of Operator Builder is documented [here](docs/testing.md).

