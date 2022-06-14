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
| [description](#description-optional) | string                         | false    |

### Name (required if Parent is unspecified)

The name you want to use for the field in the custom resource that
Operator Builder will create.  If you're not sure what that means, it will
become clear shortly.

ex. +operator-builder:field:name=myName

### Parent (required if Name is unspecified)

The parent field in which you wish to substitute.  Currently, only `metadata.name` is supported.  This 
will allow you to use the parent name as a value in the child resource.

ex. +operator-builder:field:parent=metadata.name

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

    `operator-builder:field:name=myName,type=string,default=test`

### Replace (optional)

There may be some instances where you only want a specific portion of a value
to be configurable (such as config maps). In these scenarios you can use the
replace argument to specify a search string (or regex) to target for configuration.

Consider the following example:
```
---
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

```
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

### Description (optional)

An optional description can be provided which will be used in the source code as
a Doc String, backticks `` ` `` may be used to capture multiline strings (head
comments only).

By injecting documentation to
the CRD, the consumer of the custom resource gets the added benefit by being
able to run `kubectl explain` against their resource and having documentation
right at their fingertips without having to navigate to API documentation in
order to see the usage of the API.  For example:

    operator-builder:field:name=myName,type=string,default=test,description="Hello World"

*Note: that you can use a single custom resource field name to configure multiple
fields in the resource.*

Consider the following Deployment:

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

	apiVersion: product.apps.acme.com/v1alpha1
	kind: WebApp
	metadata:
	  name: dev-webapp
	spec:
      production: false
      webAppReplicas: 2
      webAppImage: acmerepo/webapp:3.5.3

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

ex. +operator-builder:resource:collectionField=provider,value="aws",include
ex. +operator-builder:resource:field=provider,value="aws",include=false

### Include (required)

The action to perform on the resource.  Include will include the resource for
deployment during a control loop given a `field` or `collectionField` and a `value`.  
Using this means that the resource will **only be included** if this condition 
is met.  If the condition is not met, the resource will not be deployed.

Here are some sample marker examples:

ex. +operator-builder:resource:field=provider,value="aws",include
ex. +operator-builder:resource:field=provider,value="aws",include=true
ex. +operator-builder:resource:collectionField=provider,value="aws",include
ex. +operator-builder:resource:collectionField=provider,value="aws",include=true

With include set to `false`, the opposite is true and the resource is 
excluded from being deployed during a control loop if a condition is met:

ex. +operator-builder:resource:field=provider,value="aws",include=false
ex. +operator-builder:resource:collectionField=provider,value="aws",include=false

At this time, the `include` argument with `field` and `value` can be simply thought of 
as (pseudo-code):

  if field == value {
    if include {
      includeResource()
    }
  }

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
