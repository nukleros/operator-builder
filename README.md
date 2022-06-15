[![Go Reference](https://pkg.go.dev/badge/github.com/vmware-tanzu-labs/operator-builder.svg)](https://pkg.go.dev/github.com/vmware-tanzu-labs/operator-builder)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/vmware-tanzu-labs/operator-builder)](https://golang.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/vmware-tanzu-labs/operator-builder)](https://goreportcard.com/report/github.com/vmware-tanzu-labs/operator-builder)
[![GitHub](https://img.shields.io/github/license/vmware-tanzu-labs/operator-builder)](https://github.com/vmware-tanzu-labs/operator-builder/blob/main/LICENSE)[![GitHub release (latest by date)](https://img.shields.io/github/v/release/vmware-tanzu-labs/operator-builder)](https://github.com/vmware-tanzu-labs/operator-builder/releases)
[![Hombrew](https://img.shields.io/badge/dynamic/json.svg?url=https://raw.githubusercontent.com/vmware-tanzu-labs/homebrew-tap/master/Info/operator-builder.json&query=$.versions.stable&label=homebrew)](https://github.com/vmware-tanzu-labs/operator-builder/releases)
[![Get it from the Snap Store](https://badgen.net/snapcraft/v/operator-builder)](https://snapcraft.io/operator-builder)
![Github Downloads (by Release)](https://img.shields.io/github/downloads/vmware-tanzu-labs/operator-builder/total.svg)

# Operator Builder

**Accelerate the development of Kubernetes Operators.**

Operator Builder is a command line tool that ingests Kubernetes manifests and
generates the source code for a working Kubernetes operator based on the
resources defined in those manifests.

Operator Builder extends [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
to facilitate development and maintenance of Kubernetes operators.  It is especially
helpful if you need to take large numbers of resources defined with static or
templated yaml and migrate to managing those resources with a custom Kubernetes operator.

An operator built with Operator Builder has the following features:

- A defined API for a custom resource based on [markers](docs/markers.md) in
  static Kubernetes manifests.
- A functioning controller that will create, update and delete child resources
  to reconcile the state for the custom resource/s.
- A [companion CLI](docs/companion-cli.md) that helps end users with common
  operations.

Operator Builder uses a [workload configuration](docs/workloads.md) as the
primary configuration mechanism for providing attributes for the source code.

The custom resource defined in the source code can be cluster-scoped or
namespace-scoped based on the requirements of the project.  More info
[here](docs/resource-scope.md).

User Documentation:

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

Develpoer Documentation

* [Testing](docs/testing.md)

