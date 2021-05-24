# Workload Markers

Operator Builder uses commented markers as the basis for defining a new API.
The fields for a custom resource kind are created when it finds a `+workload`
marker in a manifest.

A workload marker is commented out so the manifest is still valid and can be
used if needed.  The marker must begin with `+worload` followed by some
colon-separated fields:
- API Field: The first field is requried.  It should be provided as you want it
  to be in the spec of the resulting custom resource.
- Type Field: This field is provided as `type=[value]`.  It is also a required
  field.  The supported data types:
  - bool
  - string
  - int
- Default Field; This field is provided `default=[value]`.  It is an optional
  field.  If provided it will make the field optional in the custom resource and
  when not included will get the default value provided.

Consider the following Deployment:

    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: webapp-deploy
      labels:
        production: false  # +workload:production:default=false:type=bool
    spec:
      replicas: 2  # +workload:webAppReplicas:default=2:type=int
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
            image: nginx:1.17  # +workload:webAppImage:type=string
            ports:
            - containerPort: 8080

In this case, operator-builder will create add three fields to the custom
resource:
- A `production` field that is a boolean.  It will have a default of `false` and
  will inform the value of the label when the Deployment is configured.
- A `webAppReplicas` field that will default to `2` and allow the user to
  specify the number of replicas for the deployment in the custom resource
  manifest.
- A `webAppImage` field that will set the value for the images used in the pods.

Now the end-user of the operator will be able to define a custom resource
similar to the following to configure the deployment created:

	apiVersion: apps.acme.com/v1alpha1
	kind: WebApp
	metadata:
	  name: dev-webapp
	spec:
      production: false
      webAppReplicas: 2
      WebAppImage: acmerepo/webapp:3.5.3

