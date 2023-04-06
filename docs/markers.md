# Markers

Operator Builder uses commented markers as the basis for defining a new API.
The fields for a custom resource kind are created when it finds a `+operator-builder`
marker in a source manifest.

A workload marker is commented out so the manifest is still valid and can be
used if needed.  The marker must begin with `+operator-builder` followed by some
colon-separated fields:

These markers should always be provided as an in-line comment or as a head
comment.  The marker always begins with `+operator-builder:field:` or
`+operator-builder:collection:field:` (more on this later).

That is followed by arguments separated by `,`.  Arguments can be given in any order.

## Arguments

Arguments come after the actual marker and are separated from the marker name 
with a `:`. They are given in the format of `argument=value` and separated by 
the `,`. Additionally, if the argument name is given by itself with no value, it 
is assumed to have an implict `=true` on the end and is treated as a flag.

Below you will find the supported markers and their supported arguments.

## Field Markers

Defined as `+operator-builder:field` this marker can be used to define a CRD
field for your workload.

| Field                                | Type                           | Required |
| ------------------------------------ | ------------------------------ | -------- |
| [name](#name-required)               | string                         | true     |
| [type](#type-required)               | string{string, int, bool}      | true     |
| [default](#default-optional)         | [type](#supported-field-types) | false    |
| [replace](#replace-optional)         | string                         | false    |
| [arbitrary](#arbitrary-optional)     | bool                           | false    |
| [description](#description-optional) | string                         | false    |

### Name (required if Parent is unspecified)

The name you want to use for the field in the custom resource that
Operator Builder will create.  If you're not sure what that means, it will
become clear shortly.

Example:

```
+operator-builder:field:name=myName,type=string
```

### Parent (required if Name is unspecified)

The parent field in which you wish to substitute.  Currently, only `metadata.name` is supported.
This will allow you to use the parent name as a value in the child resource.

Example:

```
+operator-builder:field:parent=metadata.name,type=string
```

The `metadata.name` field from the collection workload is also supported:

```
+operator-builder:collection:field:parent=metadata.name,type=string
```

### Type (required)

The other required field is the `type` field which specifies the data type for
the value.

[#supported-field-types]() The supported data types are:

- bool
- string
- int
- int32
- int64
- float32
- float64

ex. `+operator-builder:field:name=myName,type=string`

### Default (optional)

This will make configuration optional for your operator's end user. the supplied
value will be used for the default value. If a field has no default, it will be
a required field in the custom resource.  For example:

```
+operator-builder:field:name=myName,type=string,default=test
```

### Replace (optional)

There may be some instances where you only want a specific portion of a value
to be configurable (such as config maps). In these scenarios you can use the
replace argument to specify a search string (or regex) to target for configuration.

Consider the following example:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    # +operator-builder:field:name=environment,default=dev,type=string,replace="dev"
    app: myapp-dev
  name: contour-configmap
  namespace: ingress-system
data:
  # +operator-builder:field:name=configOption,default=myoption,type=string,replace="configuration2"
  # +operator-builder:field:name=yamlType,default=myoption,type=string,replace="multi.*yaml"
  config.yaml: |
    ---
    someoption: configuration2
    anotheroption: configuration1
    justtesting: multistringyaml
```

In this scenario three custom resource fields will be generated. The value from
the `environment` field will replace the `dev` portion of `myapp-dev`. For
example, if `prod` is provided as a value for the `environment` field, the
resulting config map will get the label `app: myapp-prod`. Values from the
`configOption` and `yamlType` fields will replace corresponding strings in the
content of `config.yaml`.  The resulting configmap will look as follows:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    app: myapp-prod
  name: contour-configmap
  namespace: ingress-system
data:
  config.yaml: |-
    ---
    someoption: myoption
    anotheroption: configuration1
    justtesting: myoption
```

### Arbitrary (optional)

If you wish to create a field for a custom resource that does not directly map
to a value in a child resource, mark a field as arbitrary.

Here is an example of how to mark a field as arbitrary:

```yaml
---
# +operator-builder:field:name=nginx.installType,arbitrary,default="deployment",type=string,description=`
# +kubebuilder:validation:Enum=deployment;daemonset
# Method of install nginx ingress controller.  One of: deployment | daemonset.`
apiVersion: v1
kind: Namespace
metadata:
  # +operator-builder:field:name=namespace,default="nukleros-ingress-system",type=string,description=`
  # Namespace to use for ingress support services.`
  name: nukleros-ingress-system
```

On the first line you can see the `nginx.installType` custom resource field is
marked as arbitrary with the `arbitrary` marker field.  Where you place this
marker is unimportant but it is recommended you put all arbitrary fields at the
beginning of one chosen manifest for ease of maintenance.

This will result in a custom resource sample that looks as follows:

```yaml
apiVersion: platform.addons.nukleros.io/v1alpha1
kind: IngressComponent
metadata:
  name: ingresscomponent-sample
spec:
  #collection:
    #name: "supportservices-sample"
    #namespace: ""
  nginx:
    installType: "deployment"            # <---- arbitary field
    image: "nginx/nginx-ingress"
    version: "2.3.0"
    replicas: 2
  namespace: "nukleros-ingress-system"
```

This arbitrary field will not map to any child resource value.  However it can
be leveraged by some custom mutation code or by a resource marker such as this:

```yaml
---
# +operator-builder:resource:field=nginx.installType,value="deployment",include
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-ingress
  namespace: nukleros-ingress-system # +operator-builder:field:name=namespace,default="nukleros-ingress-system",type=string
...
```

The marker on line one indicates the deployment resource only be created if
`nginx.installType` has a value of `deployment` (as shown in the custom resource
sample above).  In this example, we are providing an option to install the Nginx
Ingress Controller as a deployment _or_ a daemonset.

### Description (optional)

An optional description can be provided which will be used in the source code as
a Doc String, backticks `` ` `` may be used to capture multiline strings (head
comments only).

By injecting documentation to
the CRD, the consumer of the custom resource gets the added benefit by being
able to run `kubectl explain` against their resource and having documentation
right at their fingertips without having to navigate to API documentation in
order to see the usage of the API.  For example:

```
+operator-builder:field:name=myName,type=string,default=test,description="Hello World"
```

*Note: that you can use a single custom resource field name to configure multiple
fields in the resource.*

Consider the following Deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp-deploy
  labels:
    production: false  # +operator-builder:field:name=production,default=false,type=bool
spec:
  replicas: 2  # +operator-builder:field:name=webAppReplicas,default=2,type=int
  selector:
    matchLabels:
      app: webapp
  template:
    metadata:
      labels:
        app: webapp
    spec:
      containers:
      - name: webapp-container
        image: nginx:1.17  # +operator-builder:field:name=webAppImage,type=string
        ports:
        - containerPort: 8080
```

In this case, operator-builder will create and add three fields to the custom
resource:

- A `production` field that is a boolean.  It will have a default of `false` and
  will inform the value of the label when the deployment is configured.
- A `webAppReplicas` field that will default to `2` and allow the user to
  specify the number of replicas for the deployment in the custom resource
  manifest.
- A `webAppImage` field that will set the value for the images used in the pods.

Now the end-user of the operator will be able to define a custom resource
similar to the following to configure the deployment created:

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

## Collection Markers

A second marker type `+operator-builder:collection:field` can be used with the
same arguments as a Field Marker. These markers are used to define global fields
for your Collection and can be used in any of its associated components.

If you include any marker on a [collection
resource](workload-collections.md#collection-resources) it will be treated as a
collection marker and will configure a field in the collection's custom
resource.

## Resource Markers

Defined as `+operator-builder:resource` this marker can be used to control a specific
resource with arguments in the marker.

Note: a resource marker must reference a field defined by a field marker.  If
you include a resource marker with a unique field name that is not also defined
by a field marker you will get an error.  You may use an [arbitrary field](#arbitrary-optional)
on a field marker if you don't wish to associate the field with a value in a
child resource.

| Field                                               | Type                           | Required |
| --------------------------------------------------- | ------------------------------ | -------- |
| [field](#field--collectionfield-required)           | string                         | true     |
| [collectionField](#field--collectionfield-required) | string{string, int, bool}      | true     |
| [value](#value-required)                            | [type](#supported-field-types) | true     |
| [include](#include-required)                        | bool                           | true    |

### Field / CollectionField (required)

The conditional field to associate with an action (currently only [include](#include-required)).
One of `field` or `collectionField` must be provided depending upon if you are
checking a condition against a collection, or a component/standalone workload spec.
The field input relates directly to a given workload marker such as
`+operator-builder:field:name=provider` would produce a field of `provider` to be used
in a resource marker with argument `field=provider`.

ex. +operator-builder:resource:collectionField=provider,value="aws",include
ex. +operator-builder:resource:field=provider,value="aws",include=false

### Value (required)

The conditional value to associate with an action (currently only `include` - see
above).  The `value` input relates directly to the value of `field` as it exists
in the API spec requested by the user.

Examples:

```
+operator-builder:resource:collectionField=provider,value="aws",include
+operator-builder:resource:field=provider,value="aws",include=false
```

### Include (required)

The action to perform on the resource.  Include will include the resource for
deployment during a control loop given a `field` or `collectionField` and a `value`.
Using this means that the resource will **only be included** if this condition
is met.  If the condition is not met, the resource will not be deployed.

Here are some sample marker examples:

```
+operator-builder:resource:field=provider,value="aws",include
+operator-builder:resource:field=provider,value="aws",include=true
+operator-builder:resource:collectionField=provider,value="aws",include
+operator-builder:resource:collectionField=provider,value="aws",include=true
```

With include set to `false`, the opposite is true and the resource is
excluded from being deployed during a control loop if a condition is met:

Examples:

```
+operator-builder:resource:field=provider,value="aws",include=false
+operator-builder:resource:collectionField=provider,value="aws",include=false
```

At this time, the `include` argument with `field` and `value` can be simply thought of
as (pseudo-code):

```
if field == value {
  if include {
    includeResource()
  }
}
```

**IMPORTANT:** A resource marker is not required and should only be used when there is a desire
to act upon a resource.  If no resource marker is provided, a resource is always
deployed during a control loop.

#### Include Resource On Condition

Below is a sample of how to include a resource only if a condition is met.  If the
condition is not met, the resource is not deployed during the control loop:

```yaml
# +operator-builder:resource:field=provider,value="aws",include
---
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: aws-storage-class
  annotations:
    storageclass.kubernetes.io/is-default-class: true
  labels:
    provider: "aws" # +operator-builder:field:name=provider,type=string,default="aws"
allowVolumeExpansion: false
provisioner: "kubernetes.io/aws-ebs"
reclaimPolicy: "Delete"
volumeBindingMode: WaitForFirstConsumer
parameters:
  type: "gp2"
  iopsPerGB: 10
  fsType: "ext4"
  encrypted: false
```

Given the below CRD, the resource would be included:

```yaml
apiVersion: apps.acme.com/v1alpha1
kind: Sample
metadata:
  name: sample
  namespace: default
spec:
  provider: "aws"
```

Given the below CRD, the resource would **NOT** be included:

```yaml
apiVersion: apps.acme.com/v1alpha1
kind: Sample
metadata:
  name: sample
  namespace: default
spec:
  provider: "azure"
```

#### Exclude Resource On Condition

Below is a sample of how to exclude a resource only if a condition is met.  If the
condition is not met, the resource is not deployed during the control loop:

```yaml
# +operator-builder:resource:field=provider,value="azure",include=false
---
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: aws-storage-class
  annotations:
    storageclass.kubernetes.io/is-default-class: true
  labels:
    provider: "aws" # +operator-builder:field:name=provider,type=string,default="aws"
allowVolumeExpansion: false
provisioner: "kubernetes.io/aws-ebs"
reclaimPolicy: "Delete"
volumeBindingMode: WaitForFirstConsumer
parameters:
  type: "gp2"
  iopsPerGB: 10
  fsType: "ext4"
  encrypted: false
```

Given the below CRD, the resource would be included:

```yaml
apiVersion: apps.acme.com/v1alpha1
kind: Sample
metadata:
  name: sample
  namespace: default
spec:
  provider: "aws"
```

Given the below CRD, the resource would **NOT** be included:

```yaml
apiVersion: apps.acme.com/v1alpha1
kind: Sample
metadata:
  name: sample
  namespace: default
spec:
  provider: "azure"
```

### Stacking Resource Markers

You can include multiple resource markers on a particular resource.  For example:
```yaml
---
# +operator-builder:resource:field=nginx.include,value=true,include
# +operator-builder:resource:field=nginx.installType,value="deployment",include
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-ingress
...
```

The purpose of the first marker is to include *all* nginx ingress contoller
resources when `spec.nginx.include: true`.  The second gives users a choice
to install nginx ingress controller as a deployment or daemonset.  When
`spec.nginx.installType: deployment` the deployment resource is included.
Therefore the custom resource will need to look as follows for this deployment
resource to be created:

```yaml
apiVersion: platform.addons.nukleros.io/v1alpha1
kind: IngressComponent
metadata:
  name: ingresscomponent-sample
spec:
  nginx:
    installType: "deployment"  # if not "deployment" deployment resource excluded
    include: true  # if false, no nginx resources are created
    image: "nginx/nginx-ingress"
    version: "2.3.0"
    replicas: 2
```

The resulting source code looks as follows.  If either if-statement is evaluated
as true, the function will return without any object - hence the deployment will
not be included.

```go
// CreateDeploymentNamespaceNginxIngress creates the Deployment resource with name nginx-ingress.
func CreateDeploymentNamespaceNginxIngress(
	parent *platformv1alpha1.IngressComponent,
	collection *setupv1alpha1.SupportServices,
	reconciler workload.Reconciler,
	req *workload.Request,
) ([]client.Object, error) {

	if parent.Spec.Nginx.Include != true {
		return []client.Object{}, nil
	}

	if parent.Spec.Nginx.InstallType != "deployment" {
		return []client.Object{}, nil
	}

	var resourceObj = &unstructured.Unstructured{
		Object: map[string]interface{}{
			// +operator-builder:resource:field=nginx.include,value=true,include
			// +operator-builder:resource:field=nginx.installType,value="deployment",include
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
      ...
			}

	return mutate.MutateDeploymentNamespaceNginxIngress(resourceObj, parent, collection, reconciler, req)
}
```
