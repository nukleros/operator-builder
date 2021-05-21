# Operator Builder

Accelerate the development of Kubernetes Operators.

Operator Builder extends [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
to facilitate development and maintenance of Kubernetes operators.  It is especially
helpful if you need to take large numbers of resources defined with static or
templated yaml and migrate to managing those resources with a custom Kubernetes operator.

## TODO

* add companion cli build make targets (see https://gitlab.eng.vmware.com/landerr/rpk-operator)
* add validation for Workload vs WorkloadCollection (see pkg/plugins/workload/v1/init.go)

