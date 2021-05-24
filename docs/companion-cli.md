# Companion CLI

Generate source code for a companion CLI to a Kubernetes operator.

The companion CLI does three things:
1. Generate Sample Manifests: the `init` command will save a sample manifest to
   disk for a custom resource.  This gives the end user a convenient way to get
   started with defining configuration variables.
2. Generate Child Resource Manifests: the `generate` command prints the
   manifests for all of the custom resources children - the Kubernetes resources
   that are created and managed when a custom resource is created.  This offers
   the end user a workaround when they need to configure changes that are not
   exposed by the operator.
3. Install the Operator: the `install` command installs the operator, CRDs and
   necessary resources in a Kubernetes cluster.

These are the CLI configurations:
1. No CLI: Don't define any companion CLI data and no CLI source code will be
   scaffolded.
2. A single root command: define the `spec.companionCliRootcmd` fields in a the
   `Workload` manifest.
3. A root command with subcommands: define the `spec.companionCliRootcmd` in a
   `WorkloadCollection` manifest.  Then define `spec.companionCliSubcmd` in one
   or more `Workload` manifests.

## Root Command

The root command for the CLI can be defined in a standalone workload or in a
[collection workload](collections.md).

## Subcommands

If a workload belongs to a collection you may define a subcommand for that
workload.

