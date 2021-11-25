# Workloads

Operator Builder uses WorkloadConfig manifests to define "workloads."  A workload
represents any group of Kubernetes resources.  For example, a workload could be an
application with Deployment, Service and Ingress resources.  A workload may also
be a group of resources that provide a platform service.  They can be
any group of resources that logically group together.  Here's an example of
a simple WorkloadConfig for a hypothetical web application called "webapp":

```yaml
name: webapp
kind: StandaloneWorkload
spec:
  api:
    domain: apps.acme.com
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
```

This tells Operator Builder to create the source code for a new Kubernetes API
called "WebApp."  This is the API kind.  The resources that comprise this webapp
are a deployment, service and ingress.  Those resources are defined in the
source manifest files referenced under `resources`.  Operator Builder uses those
source manifests to generate source code for those resources.  Those source
manifests can contain [markers](markers.md) to help define the
API.

With the source code generated, the [companion CLI](companion-cli.md) can be
built for end users.  The controller container image can also be built and made
available to end users.

End users then use the CLI `install` command to deploy the operator and then
use the `init` command to generate a sample custom resource that might look
something like this:

```yaml
apiVersion: product.apps.acme.com/v1alpha1
kind: WebApp
metadata:
  name: dev-webapp
spec:
  production: false
  webAppReplicas: 2
  webAppImage: acmerepo/webapp:3.5.3
```

## Required Fields

The following are required fields:
- spec.api.domain   # required for 'operator-builder init'
- spec.api.group    # required for 'operator-builder create api'
- spec.api.version  # required for 'operator-builder create api'
- spec.api.kind     # required for 'operator-builder create api'

All other fields are optional.  The default value for `clusterScoped` if not
defined is `false`.  Alternatively, the above fields can be defined
imperatively via the `domain`, `group`, `version`, and `kind` flags
when running either `operator-builder init` or `operater-builder create api` (see above for correct context).

## Resources

When specifying resource manifest files under `spec.resources`, in addition to
providing specific filenames, you may use glob pattern matching to collect files.
For example:

```yaml
name: webapp
kind: StandaloneWorkload
spec:
  api:
    domain: apps.acme.com
    group: product
    version: v1alpha1
    kind: WebApp
    clusterScoped: false
  companionCliRootcmd:
    name: webappctl
    description: Manage webapp stuff like a boss
  resources:
    - rbac/*.yaml       # get all .yaml files in the rbac directory
    - workload/**.yaml  # get all .yaml files recursively in the workload
                        # directory and subdirectories therein
```

## Collections

The `spec.componentFiles` field can only be defined in a `WorkloadCollection`.
See [workload collections](workload-collections.md) for more information.

