# Workloads

Operator Builder uses configurations called "workloads."  A workload represents
any group of Kubernetes resources.  For example, a workload could be an
application with Deployment, Service and Ingress resources.  A workload may also
be a group of resources that provide a platform service.  They can be
any group of resources that make sense to deploy together.  Here's an example of
a simple workload definition for a hypothetical web application called "webapp":

    name: webapp
    spec:
      group: product
      version: v1alpha1
      kind: WebApp
      clusterScoped: false
      companionCliRootcmd:
        name: webappctl
        description: Manage webapp stuff like a boss
      resources:
      - deploy.yaml
      - service.yaml
      - ingress.yaml

This tells Operator Builder to create the source code for a new Kubernetes API
called "WebApp."  This is the API kind.  The resources that comprise this webapp
are a deployment, service and ingress.  Those resources are defined in the files
referenced under `resources`.  Operator Builder uses those manifests to generate
source code for those resources.  Those resource manifests can contain [workload
markers](workload-markers.md) to help define the API.

With the source code generated, the [companion CLI](companion-cli.md) can be
built for end users.  The controller container image can also be built and made
available to end users.

End users then use the CLI `install` command to deploy the operator and then
use the `init` command to generate a sample custom resource that might look
something like this:

	apiVersion: product.apps.acme.com/v1alpha1
	kind: WebApp
	metadata:
	  name: dev-webapp
	spec:
      production: false
      webAppReplicas: 2
      WebAppImage: acmerepo/webapp:3.5.3

## Required Fields

The following are required fields:
- spec.group
- spec.version
- spec.kind

All other fields are optional.  The default value for `clusterScoped` if not
defined is `false`.

## Collections

The `spec.children` field can only be defined if `collection: true`.  See
[collections](collections.md) for more information.

