# Workload Collection Tutorial

This tutorial walks through all the steps to create an operator that manages
multiple distinct workloads using Operator Builder.

The operator that you'll build in this tutorial installs supporting services for
a cluster.  Specifically, it will install Cert Manager and the Nginx Ingress
Controller.

**Note**: all of the manifests and configurations in these instructions can be found
in the accompanying `.operator-builder` directory.

## Prerequisites

* [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
* [golang](https://go.dev/doc/install)
* [operator-builder](../installation.md)

## Setup

In this section we will create a new project and add the workload
configurations for operator-builder.

Make a new directory for the project and initialize git.

```bash
mkdir supporting-services-operator
cd supporting-services-operator
git init
```

Create a config directory and add the WorkloadCollection config.

```bash
mkdir .operator-builder
cd .operator-builder
operator-builder init-config collection > workload.yaml
```

Edit the sample config so that it looks as follows.

```yaml
kind: WorkloadCollection
name: supporting-services-collection
spec:
  api:
    clusterScoped: true
    domain: nukleros.io
    group: addons
    kind: SupportingServices
    version: v1alpha1
  companionCliRootcmd:
    description: "Manage a cluster's supporting service installations"
    name: ssctl
  companionCliSubcmd:
    description: Manage the collection of services
    name: collection
  componentFiles:
    - addons.nukleros.io_tls/workload.yaml
    - addons.nukleros.io_ingress/workload.yaml
  resources: []
```

Add the ComponentWorkload configs:

```bash
mkdir addons.nukleros.io_tls
operator-builder init-config component > addons.nukleros.io_tls/workload.yaml
mkdir addons.nukleros.io_ingress
operator-builder init-config component > addons.nukleros.io_ingress/workload.yaml
```

Edit the TLS component workload config so that it looks as follows.

```yaml
kind: ComponentWorkload
name: tls-component
spec:
  api:
    clusterScoped: true
    domain: nukleros.io
    group: addons
    kind: TLSComponent
    version: v1alpha1
  companionCliSubcmd:
    description: Manage the TLS management service component
    name: tls
  resources: []
```

Edit the ingress component workload config so that it looks as follows.  Note
that we have specified the name of the tls-component workload in the
`dependencies` section.  This indicates that the TLS component must be deployed
and ready before the Ingress component can be installed.  In this case, this is
because the Ingress component uses a Certificate resource which will not be
recognized by the cluster until the relevant CRD is created during installation
of the TLS component.

```yaml
kind: ComponentWorkload
name: ingress-component
spec:
  api:
    clusterScoped: true
    domain: nukleros.io
    group: addons
    kind: IngressComponent
    version: v1alpha1
  companionCliSubcmd:
    description: Manage the ingress service component
    name: ingress
  dependencies:
    - tls-component
  resources: []
```
## Resource Manifests

In this section we'll download the resource manifests.  These are Kubernetes
manifests for the resources that we want our operator to manage.  They will
contain operator-builder markers to indicate which fields need to be
configurable through custom resources.

Clone the operator-builder repo so as to access the manifests for this tutorial.

```bash
git clone git@github.com:nukleros/operator-builder.git /tmp/operator-builder
```

### Supporting Services Workload Collection

The only resource associated with the workload collection is a namespace.  No
other resources will be created in this namespace at this time.  The namespace
resource manifest will be used to create some fields for the SupportingServices
custom resource - these will be global configurations that can be used by
multiple component workloads.  And the values configured will be added to the
namespace as labels for easy reference to what those global configs are.

Copy the collection resource manifest for this tutorial.

```
cp /tmp/operator-builder/docs/workload-collection-tutorial/.operator-builder/namespace.yaml .
```

#### `.operator-builder/namespace.yaml`

The marker on line 4 makes the name of the namespace name configurable with the
name of the SupportServices custom resource.

The marker on line 6 creates a new field for the SupportingServices custom
resource called `cloudProvider`.  The Ingress Component will use this
field to determine the configuration for that component's Service resource.
Normally, we would use global configs for attributes that are shared by multiple
components.  However, in this case only the Ingress component uses it.  That is
because there is a reasonable expectation that other components will use it in
future.  This marker uses the description field to add a kubebuilder marker on
the following line that is used to inform the valid values for this field.  You
can learn more about kubebuilders validation markers in the [kubebuilder
docs](https://book.kubebuilder.io/reference/generating-crd.html?highlight=validation,marker#validation).

The marker on line 9 adds a `certProvider` field to the SupportingServices
custom resource.  It also includes a kubebuilder validation marker to inform the
valid values.  This field will configure the issuer name on the ingress default
servier Certificate resource as well as on the ClusterIssuer resources created
with the TLS component.

Update the WorkloadCollection config to include this file under `resources`.  It
should look as follows.

```yaml
kind: WorkloadCollection
name: supporting-services-collection
spec:
  api:
    clusterScoped: true
    domain: nukleros.io
    group: addons
    kind: SupportingServices
    version: v1alpha1
  companionCliRootcmd:
    description: "Manage a cluster's supporting service installations"
    name: ssctl
  companionCliSubcmd:
    description: Manage the collection of services
    name: collection
  componentFiles:
    - addons.nukleros.io_tls/workload.yaml
    - addons.nukleros.io_ingress/workload.yaml
  resources:
    - namespace.yaml
```

### TLS Component Workload

The resources for this component will deploy Cert Manager for TLS asset
management.

Copy the resource manifests for the TLS component for this tutorial.

```
cp -R /tmp/operator-builder/docs/workload-collection-tutorial/.operator-builder/addons.nukleros.io_tls .
```

#### `.operator-builder/addons.nukleros.io_tls/config.yaml`

This manifest has just a single marker on line 6.  This marker uses the namespace
field from the TLSComponent custom resource to set the namespace for this
ConfigMap.  The same will be done for all namespaced resources for this
component.  This field has a default value included which makes it an optional
field in the TLSComponent resource.

#### `.operator-builder/addons.nukleros.io_tls/crd.yaml`

This manifest has no operator-builder markers.  The CustomResourceDefinitions are
not configurable in any way.

#### `.operator-builder/addons.nukleros.io_tls/deployment.yaml`

This manifest contains 3 deployment manifests.  The namespace for each is
defined in the same way as the ConfigMap.

This manifest also includes markers to set versions and replica counts for each
Deployment.  For example, on line 12 a field is defined with
`field:name=caInjector.version` that uses dot notation to indicate a nested
field.  This field will be used to specify the version for the CA Injector
image.  Here it sets a label value, as it does on line 29.  On line 37 the same
field is used to add the version to the Deployment's image.  In this case
`replace="caInjectorVersion"` indicates the matching string in the image field's
value should be replaced with the value from the TLSComponents custom resource.

The number of replicas is similarly configured with a `caInjector.replicas`
field from the TLSComponent resource with the marker on line 15.

Separate corresponding markers are set on the other two Deployments in the same
file.

#### `.operator-builder/addons.nukleros.io_tls/issuers.yaml`

This file contains two ClusterIssuer resources.  The markers on lines 2 and 29
indicate which of the two resources will be used depending up on the value given
in the `certProvider` field of the collection SupportServices custom resource.
The `value="letsencrypt-staging",include` maker fields indicate that when the
value of `certProvider` matches this value, include this resource.  Otherwise it
is ommitted.  Therefore if the SupportingService resource includes
`certProvider: letsencrypt-staging` the first resource will be used (see the the
value of the `server` field on line 13.  If `certProvider:
letsencrypt-production` is in the SupportingServices resource, the second
ClusterIssuer resource will be used.  The value from this field is also used to
populate the value for the `cert-provider` as indicated by the marker on line 8.

Finally, the marker on line 15 allows the contact email for the cert provider
with a required `contactEmail` field.

#### `.operator-builder/addons.nukleros.io_tls/rbac.yaml`

The Namespace resource has its name field set by the name field in the spec of
the TLSComponent resource.  This field also informs the namespace field in all
other fields for the RBAC resources as required.

#### `.operator-builder/addons.nukleros.io_tls/service.yaml`

The namespace fields for the two Service resources are set as before.

#### `.operator-builder/addons.nukleros.io_tls/webhook.yaml`

Namespaces are again set using the same field.  However in this case we have to
use the `replace` marker field to set a part of a value, e.g. line 13.

Finally, update the tls-component ComponentWorkload config to include the
filenames of the manifests we added under `resources`.

```yaml
kind: ComponentWorkload
name: tls-component
spec:
  api:
    clusterScoped: true
    domain: nukleros.io
    group: addons
    kind: TLSComponent
    version: v1alpha1
  companionCliSubcmd:
    description: Manage the TLS management service component
    name: tls
  resources:
    - config.yaml
    - crd.yaml
    - deployment.yaml
    - issuers.yaml
    - rbac.yaml
    - service.yaml
    - webhook.yaml
```

### Ingress Component Workload

The resources for this component will install the Nginx Ingress Controller so
that tenant workloads can use Ingress resources to expose it to traffic from
outside the cluster.

Copy the resource manifests for the Ingress component for this tutorial.

```
cp -R /tmp/operator-builder/docs/workload-collection-tutorial/.operator-builder/addons.nukleros.io_ingress .
```

#### `.operator-builder/addons.nukleros.io_ingress/cert.yaml`

The namespace name on the Certificate resource is set using a namespace field as
before, this time using the IngressComponent custom resource.

The DNS name is set using a new field `domainName` that is created for the
IngressComponent resource.

The `certProvider` field from the SupportingServices resource is again used here
on the Certificate.

#### `.operator-builder/addons.nukleros.io_ingress/class.yaml`

The IngressClass resource has no markers and so is not configurable in any way.

#### `.operator-builder/addons.nukleros.io_ingress/config.yaml`

The ConfigMap's namespace is configured as the others are.

#### `.operator-builder/addons.nukleros.io_ingress/crd.yaml`

As with the TLSComponent, the CustomResourceDefinitions have no markers and are
not configurable.

#### `.operator-builder/addons.nukleros.io_ingress/deployment.yaml`

The Deployment's namespace is configured as the others are.

The replicas and container image are configured with markers on lines 7 and 23
in the same way the cert-manager Deployments were configured.  These markers
create new fields on the IngressComponent custom resource.

#### `.operator-builder/addons.nukleros.io_ingress/rbac.yaml`

The Namespace and RBAC resources for the Ingress component have their namespaces
configured the same way the TLS component did.

#### `.operator-builder/addons.nukleros.io_ingress/service.yaml`

There are four Services defined but only one is used in any given case.  The
marker at the beginning of each Service manifest determines which one is used
based on the SupportingService's `cloudProvider` field.

Again the namespace for the Service is defined as with the others.

Finally, update the ingress-component ComponentWorkload config to include the
filenames of the manifests we added under `resources`.

```yaml
kind: ComponentWorkload
name: ingress-component
spec:
  api:
    clusterScoped: true
    domain: nukleros.io
    group: addons
    kind: IngressComponent
    version: v1alpha1
  companionCliSubcmd:
    description: Manage the ingress service component
    name: ingress
  dependencies:
    - ../addons.nukleros.io_tls/workload.yaml
  resources:
    - cert.yaml
    - class.yaml
    - config.yaml
    - crd.yaml
    - deployment.yaml
    - rbac.yaml
    - service.yaml
```

## Code Generation

You'll often find you want re-generate the codebase when you find mistakes or
adjustments you want to make in the project.  For this reason it's helpful to
have a Makefile in your `.operator-builder` directory.

Copy the Makefile for this tutorial.

```
cp /tmp/operator-builder/docs/workload-collection-tutorial/.operator-builder/Makefile .
```

Now it's time to generate the code.

```
make operator-init
make operator-create
```

That's it.

## Operator Testing

Now it's time to test your operator.  For this you'll need a working Kubernetes
cluster.

Navigate up into the root of your operator project's codebase.

```
cd ..
```

Your new operator has a Makefile of it's own that has some handy make targets.

Install the CRDs into your cluster.

```
make install
```

Run the controller locally against your test cluster.

```
make run
```

In another terminal, install the custom resource samples.

```
kubectl apply -f config/samples
```

If everything has gone according to plan, your cluster will soon have Cert
Manager and the Nginx Ingress Controller installed.

Congratulations!  You just built a multi-workload operator.

**Note**: The Nginx Ingress Controller will require a valid DNS record that points
to your ingress public IP so as to validate your ownership of the domain and for
Let's Encrypt to issue a publicly trustable certificate.  Without this,
everything will spin up but the Nginx instance will fail to start due to the
inability to mount the secret with it's default server certificate.

