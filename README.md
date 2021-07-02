# Operator Builder

**Accelerate the development of Kubernetes Operators.**

Operator Builder extends [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
to facilitate development and maintenance of Kubernetes operators.  It is especially
helpful if you need to take large numbers of resources defined with static or
templated yaml and migrate to managing those resources with a custom Kubernetes operator.

An operator built with Operator Builder has the following features:

- A defined API for a custom resource based on [workload
  markers](docs/workload-markers.md) in static Kubernetes manifests.
- A functioning controller that will create, update and delete child resources
  to reconcile the state for the custom resource/s.
- A [companion CLI](docs/companion-cli.md) that helps end users with common
  operations.

Operator Builder uses a [workload configuration](docs/workloads.md) as the
primary configuration mechanism for providing attributes for the source code.

The custom resource defined in the source code can be cluster-scoped or
namespace-scoped based on the requirements of the project.  More info
[here](docs/resource-scope.md).

## Installation

Get the source code.

    git clone git@github.com:vmware-tanzu-labs/operator-builder.git
    cd operator-builder

> **NOTE:** Since the object-code-generator-for-k8s repository is still an internal Github repository you will have to do the following to build the project:  
`export  GOPRIVATE=github.com/vmware-tanzu-labs/object-code-generator-for-k8s`  
Add the following configuration to your ~/.gitconfig file:  
```
[url "git@github.com:"]
	insteadOf = https://github.com/
```

Build the binary.

    make build

Put the binary on your path.

    mv bin/operator-builder /usr/local/bin/

## Getting Started

This guide will walk you through the creation of a Kubernetes operator for a
single workload.  This workload can consist of any number of Kubernetes
resources and will be configured with a single custom resource.

### Prerequisites

- An available test cluster. A local kind or minikube cluster will work just
  fine in many cases.
- Operator Builder [installed](#installation).
- [kubectl installed](https://kubernetes.io/docs/tasks/tools/#kubectl).
- A set of static Kubernetes manifests that can be used to deploy
  your workload.  It is highly recommended that you apply these manifests to a
  test cluster and verify the resulting resources work as expected.
  If you don't have a workload of your own to use, you can use the examples
  provided in this guide.

This guide consists of the following steps:

1. [Create a repository](#step-1).
1. Determine what fields in your static manifests will need to be configurable for
   deployment into different environments. [Add commented markers to the
   manifests](#step-2). These will serve as instructions to Operator Builder.
1. [Create a workload configuration for your project](#step-3).
1. [Use the Operator Builder CLI to generate the source code for your operator](#step-4).
1. [Test the operator against your test cluster](#step-5).
1. [Build and install your operator's controller manager in your test cluster](#step-6).
1. [Build and test the operator's companion CLI](#step-7).

### Step 1

Create a new directory for your operator's source code.  We recommend you follow
the standard [code organization
guidelines](https://golang.org/doc/code#Organization).
In that directory initialize a new git repo.

    git init

And intialize a new go module.  The module should be the import path for your
project, usually something like `github.com/user-account/project-name`.  Use the
command `go help importpath` for more info.

    go mod init [module]

Lastly create a directory for your static manifests.  Operator Builder will use
these as a source for defining resources in your operator's codebase.  It must be a
hidden directory so as not to interfere with source code generation.

    mkdir .source-manifests

Put your static manifests in this `.source-manifests` directory.  In the next
step we will add commented markers to them.  Note that these static manifests
can be in one or more files.  And you can have one or more manifests (separated
by `---`) in each file.  Just organize them in a way that makes sense to you.

### Step 2

Look through your static manifests and determine which fields will need to be
configurable for deployment into different environments.  Let's look at a simple
example to illustrate.  Following is a Deployment, Ingress and Service that may
be used to deploy a workload.

    # app.yaml

    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: webstore-deploy
    spec:
      replicas: 2                       # <===== configurable
      selector:
        matchLabels:
          app: webstore
      template:
        metadata:
          labels:
            app: webstore
        spec:
          containers:
          - name: webstore-container
            image: nginx:1.17           # <===== configurable
            ports:
            - containerPort: 8080
    ---
    apiVersion: networking.k8s.io/v1beta1
    kind: Ingress
    metadata:
      name: webstore-ing
      annotations:
        nginx.ingress.kubernetes.io/rewrite-target: /
    spec:
      rules:
      - host: app.acme.com
        http:
          paths:
          - path: /
            backend:
              serviceName: webstorep-svc
              servicePort: 80
    ---
    kind: Service
    apiVersion: v1
    metadata:
      name: webstore-svc
    spec:
      selector:
        app: webstore
      ports:
      - protocol: TCP
        port: 80
        targetPort: 8080

There are two fields in the Deployment manifest that will need to be
configurable. They are noted with comments. The Deployment's replicas and the
Pod's container image will change between different environments.  For example,
in a dev environment the number of replicas will be low and a development
version of the app will be run.  In production, there will be more replicas and
a stable release of the app will be used. In this example we don't have any
configurable fields in the Ingress or Service.

Next we need to use `+workload` markers in comments to inform Operator Builder
that the operator will need to support configuration of these elements.
Following is the Deployment manifest with these markers in place.

    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: webstore-deploy
      labels:
        team: dev-team  # +workload:teamName:type=string
    spec:
      replicas: 2  # +workload:webStoreReplicas:default=2:type=int
      selector:
        matchLabels:
          app: webstore
      template:
        metadata:
          labels:
            app: webstore
            team: dev-team  # +workload:teamName:type=string
        spec:
          containers:
          - name: webstore-container
            image: nginx:1.17  # +workload:webStoreImage:type=string
            ports:
            - containerPort: 8080

These markers should always be provided as an in-line comment.  This retains the
functionality of the manifest and is what Operator Builder will look for.  The
marker always begins with `+workload`.

That is followed by fields separated by colons.  The first field in the marker
must be the name you want to use for the field in the custom resource that
Operator Builder will create.  If you're not sure what that means, it will
become clear shortly.

The other required field is the `type` field which specifies the data type for
the value.  The supported data types are:

- bool
- string
- int
- int32
- int64
- float32
- float64

You may also add a `default` field.  This will make configuration optional for
your operator's end user and, if not given, will use the supplied default value.
If a field has no default, it will be a required field in the custom resource.

Note that you can use a single custom resource field name to configure multiple
fields in the resource.  In the example above, the value for the `teamName`
field will be use for two different label values on the Deployment.

### Step 3

Operator Builder uses a workload configuration to provide important details for
your operator project.  This guide uses a [standalone
workload](docs/standalone-workloads.md). Save this file to your
`.source-manifests` directory.

    # .source-manifests/workload.yaml
    name: webstore
    kind: StandaloneWorkload
    spec:
      domain: acme.com
      apiGroup: apps
      apiVersion: v1alpha1
      apiKind: WebStore
      clusterScoped: false
      companionCliRootcmd:
        name: webstorectl
        description: Manage webstore application
      resources:
      - app.yaml

For a standalone workload the `kind` must be `StandaloneWorkload`.  The `name`
is arbitrary and can be whatever you like.

In the `spec`, the following fields are required:

- `domain`: This must be a globally unique name that will not be used by other
  organizations or groups.  It will contain groups of API types.
- `apiGroup`: This is a logical group of API types used as a namespacing
  mechanism for your APIs.
- `apiVersion`: Provide the intiial version for your API.
- `apiKind`: The name of the API type that will represent the workload you are
  managing with this operator.
- `resources`: An array of filenames where your static manifests live.  List the
  relative path from the workload manifest to all the files that contain the
  static manifests we talked about in step 2.

For more info about API groups, versions and kinds, checkout the [Kubebuilder
docs](https://kubebuilder.io/cronjob-tutorial/gvks.html).

The following fields in the `spec` are optional:

- `clusterScoped`: If your workload includes cluster-scoped resources like
  namespaces, this will need to be `true`.  The default is `false`.
- `companionCLIRootcmd`: If you wish to generate source code for a companion CLI
  for your operator, include this field.  We recommend you do.  Your end users
  will appreciate it.
  - `name`: The root command your end users will type when using the companion
    CLI.
  - `description`: The general information your end users will get if they use
    the `help` subcommand of your companion CLI.

At this point in our example, our `.source-manifests` directory looks as
follows:

    tree .source-manifests

    .source-manifests
    ├── app.yaml
    └── workload.yaml

Our StandaloneWorkload config is in `workload.yaml` and the Deployment, Ingress
and Service manifests are in `app.yaml` and referenced under `spec.resources` in
our StandaloneWorkload config.

We are now ready to generate our project's source code.

### Step 4

We first use the `init` command to create the general scaffolding.  We run this
command from the root of our repo and provide a single argument with the path to
our workload config.

    operator-builder init \
        --workload-config .source-manfiests/workload.yaml

With the basic project now set up, we can now run the `create api` command to
create a new custom API for our workload.

    operator-builder create api \
        --workload-config .source-manfiests/workload.yaml \
        --controller \
        --resource

We again provide the same workload config file.  Here we also added the
`--controller` and `--resource` arguments.  These indicate that we want both a
new controller and new custom resource created.

You now have a new working Kubernetes Operator!  Next, we will test it out.

### Step 5

Assuming you have a kubeconfig in place that allows you to interact with your
cluster with kubectl, you are ready to go.

First, install the new custom resource definition (CRD).

    make install

Now we can run the controller locally to test it out.

    make run

Operator Builder created a sample manifest in the `config/samples` directory.
For this example it looks like this:

    apiVersion: apps.acme.com/v1alpha1
    kind: WebStore
    metadata:
      name: webstore-sample
    spec:
      webStoreReplicas: 2
      webStoreImage: nginx:1.17
      teamName: dev-team

You will notice the fields and values in the `spec` were derived from the
markers you added to your static manifests.

Next, in another terminal, create a new instance of your workload with
the provided sample manifest.

    kubectl apply -f config/samples/

You should see your custom resource sample get created.  Now use `kubectl` to
inspect your cluster to confirm the workload's resources got created.  You should
find all the resources that were defined in your static manifests.

    kubectl get all

Clean up by stopping your controller with ctrl-c in that terminal and then
remove all the resources you just created.

    make uninstall

### Step 6

Now let's deploy your controller into the cluster.

First export an environment variable for your container image.

    export IMG=myrepo/acme-webstore-mgr:0.1.0

Run the rest of the commands in this step 6 in this same terminal as most of
them will need this `IMG` env var.

In order to run the controller in-cluster (as opposed to running locally with
`make run`) we will need to build a container image for it.

    make docker-build

Now we can push it to a registry that is accessible from the test cluster.

    make docker-push

Finally, we can deploy it to our test cluster.

    make deploy

Next, perform the same tests from step 5 to ensure proper operation of our
operator.

    kubectl apply -f config/sample/

Again, verify that all the resources you expect are created.

Once satisfied, remove the instance of your workload.

    kubectl delete -f config/sample/

For now, leave the controller running in your test cluster.  We'll use it in
Step 7.

### Step 7

Now let's build and test the companion CLI.

You will have a make target that includes the name of your CLI.  For this
example it is:

    make build-webstorectl

We can view the help info as follows.

    ./bin/webstorectl help

Your end users can use it to create a new custom resource manifest.

    ./bin/webstorectl init > /tmp/webstore.yaml

If you would like to change any of the default values, edit the file.

    vim /tmp/webstore.yaml

Then you can apply it to the cluster.

    kubectl apply -f /tmp/webstore.yaml

If your end users find they wish to make changes to the resources that aren't
supported by the operator, they can generate the resources from the custom
resource.

    ./bin/webstorectl generate --workload-manifest /tmp/webstore.yaml

This will print the resources to stdout.  These may be piped into an overlay
tool or written to disk and modified before applying to a cluster.

That's it!  You have a working operator without manually writing a single line
of code.  If you'd like to make any changes to your workload's API, you'll find
the code in the `apis` directory.  The controller's source code is in
`controllers` directory.  And the companion CLI code is in `cmd`.

Don't forget to clean up.  Remove the controller, CRD and the workload's
resources as follows.

    make undeploy

For more information, checkout the [Operator Builder docs](docs/) as
well as the [Kubebuilder docs](https://kubebuilder.io/).

## Collections

Operator Builder can generate source code for operators that manage multiple
workloads.  See [collections](docs/workload-collections.md) for more info.

## Licensing

Operator Builder can help manage licensing for the resulting project.  More
info [here](docs/license.md).

## Testing

Testing of Operator Builder is documented [here](docs/testing.md).
