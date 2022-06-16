# Companion CLI

When you generate the source code for a Kubernetes operator with Operator
Builder, it can include the code for a companion CLI.  The source code for the
companion CLI will be found in the `cmd` directory of the generated codebase.

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
2. A single root command: define the `spec.companionCliRootcmd` fields in a
   standalone `WorkloadConfig` manifest.
3. A root command with subcommands: define the `spec.companionCliRootcmd` in a
   collection `WorkloadConfig` manifest.  Then define `spec.companionCliSubcmd`
   in one or more component `WorkloadConfig` manifests.

## Root Command

The root command for the CLI can be defined in a standalone workload or in a
[workload collection](workload-collections.md).

## Subcommands

If a workload belongs to a collection you may define a subcommand for that
workload.

