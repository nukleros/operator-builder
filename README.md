[![Go Reference](https://pkg.go.dev/badge/github.com/nukleros/operator-builder.svg)](https://pkg.go.dev/github.com/nukleros/operator-builder)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/nukleros/operator-builder)](https://golang.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/nukleros/operator-builder)](https://goreportcard.com/report/github.com/nukleros/operator-builder)
[![GitHub](https://img.shields.io/github/license/nukleros/operator-builder)](https://github.com/nukleros/operator-builder/blob/main/LICENSE)[![GitHub release (latest by date)](https://img.shields.io/github/v/release/nukleros/operator-builder)](https://github.com/nukleros/operator-builder/releases)
[![Hombrew](https://img.shields.io/badge/dynamic/json.svg?url=https://raw.githubusercontent.com/nukleros/homebrew-tap/master/Info/operator-builder.json&query=$.versions.stable&label=homebrew)](https://github.com/nukleros/operator-builder/releases)
<!---[![Get it from the Snap Store](https://badgen.net/snapcraft/v/operator-builder)](https://snapcraft.io/operator-builder)-->
![Github Downloads (by Release)](https://img.shields.io/github/downloads/nukleros/operator-builder/total.svg)

# Operator Builder

**Accelerate the development of Kubernetes Operators.**

## Types of Operators

There is a vast amount of functionality that can be implemented in a Kubernetes
Operator.  You are limited by only by the Kuberetnes API and the things you can
do with Go.  That said, there are general categories we can break Kubernetes
Operators into.

* Resource Managers:  Use a custom resource to trigger the creation of a
  collection of other Kubernetes resources.  This kind of operator is an
  abstraction mechanism that implements a custom resource to represent an entire
  application that, when created, triggers the operator to create all the
  Kubernetes resources that constitue that application.  Some applications
  consist of dozens of distinct resources, so this kind of abstraction can be
  very helpful.  Popular examples include the [Prometheus
  Operator](https://github.com/prometheus-operator/prometheus-operator) and
  various database operators.  These are a very common type of Kubernetes
  Operator.
* External Integrators:  This kind of operator uses custom resources to define
  resources external to Kubernetes such as cloud provider resources.  The [AWS
  Controllers for Kubernetes](https://github.com/aws-controllers-k8s/community)
  is a good example of this.
* Configuration Controllers:  Some operators don't directly manage Kubernetes or
  external resources, but instead provide configuration support services for
  applications.  They often watch other resource kinds and take config actions
  to support different workloads.  [cert-manager](https://cert-manager.io/) is a
  good example of this when used to manage TLS assets based on other resources
  such as Ingresses.

## When to Use Operator Builder

Operator Builder speeds up the development of the first kind of operator:
Resource Managers.  It is a command line tool that ingests Kubernetes manifests
and generates the source code for a working Kubernetes Operator based on the
resources defined in those manifests.  These are the general steps to
building a Resource Manager Operator with Operator Builder:

* Construct the Kubernetes manifests for the application you want to manage and
  test them in a Kubernetes cluster.  You can also use Helm and the `helm template`
  command to create these resources if a helm chart exists.
* Determine which fields in the manifests need to be mutable and managed by the
  operator, then add [markers](docs/markers.md) to the manifests.
* Create an [workload configuration](docs/workloads.md) to give it some details,
  such as what you would like to call your custom resource.
* Run the Operator Builder CLI in a new repository and provide it the marked up
  manifests and config.

That's it!  You will now have a Kubernetes Operator that will create, update and
delete the resources that constitute your application in response to creating,
updating or deleting a custom resource instance.

An operator built with Operator Builder has the following features:

* A defined API for a custom resource based on [markers](docs/markers.md) in
  static Kubernetes manifests.
* A functioning controller that will create, update and delete child resources
  to reconcile the state for the custom resource/s.
* A [companion CLI](docs/companion-cli.md) that helps end users with common
  operations.

The custom resource defined in the source code can be cluster-scoped or
namespace-scoped based on the requirements of the project.  More info
[here](docs/resource-scope.md).

## Built Atop Kubebuilder

Operator Builder is a [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
plugin.  Kubebuilder provides excellent scaffolding for Kubernetes Operators but
anyone who has built a Resource Manager operator using Kubebuilder can attest to
the amount of time and effort required to define the managed resources in Go,
not to mention the logic for creating, updating and deleting those resources.
Operator Builder adds those resource definitions and other code to get you up
and running in short order.

## Documentation

### User Docs

* [Installation](docs/installation.md)
* [Getting Started](docs/getting-started.md)
* [Workloads](docs/workloads.md)
    * [Standalone Workloads](docs/standalone-workloads.md)
    * [Workload Collections](docs/workload-collections.md)
* [Markers](docs/markers.md)
* [Resource Scope](docs/resource-scope.md)
* [Companion CLI](docs/companion-cli.md)
* [API Updates & Upgrades](docs/api-updates-upgrades.md)
* [License Manaagement](docs/license.md)

### Developer Docs

* [Testing](docs/testing.md)

