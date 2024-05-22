# Getting Started

## Prerequisites

- Make
- Go version 1.22 or later
- Docker (for building/pushing controller images)
- An available test cluster. A local kind or minikube cluster will work just
  fine in many cases.
- Operator Builder [installed](#installation).
- [kubectl installed](https://kubernetes.io/docs/tasks/tools/#kubectl).
- A set of static Kubernetes manifests that can be used to deploy
  your workload.  It is highly recommended that you apply these manifests to a
  test cluster and verify the resulting resources work as expected.
  If you don't have a workload of your own to use, you can use the examples
  provided in this guide.

## Guide

This guide will walk you through the creation of a Kubernetes operator for a
single workload.  This workload can consist of any number of Kubernetes
resources and will be configured with a single custom resource.  Please review
the [prerequisites](#prerequisites) prior to attempting to follow this guide.

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

### Step 1: Create a Repo

Create a new directory for your operator's source code.  We recommend you follow
the standard [code organization
guidelines](https://golang.org/doc/code#Organization).
In that directory initialize a new git repo.

```bash
git init
```

And intialize a new go module.  The module should be the import path for your
project, usually something like `github.com/user-account/project-name`.  Use the
command `go help importpath` for more info.

```bash
go mod init [module]
```

Lastly create a directory for your static manifests.  Operator Builder will use
these as a source for defining resources in your operator's codebase.  It must be a
hidden directory so as not to interfere with source code generation.

```bash
mkdir .source-manifests
```

Put your static manifests in this `.source-manifests` directory.  In the next
step we will add commented markers to them.  Note that these static manifests
can be in one or more files.  And you can have one or more manifests (separated
by `---`) in each file.  Just organize them in a way that makes sense to you.

### Step 2: Add Manifest Markers

Look through your static manifests and determine which fields will need to be
configurable for deployment into different environments.  Let's look at a simple
example to illustrate.  Following is a Deployment, Ingress and Service that may
be used to deploy a workload.

```yaml
# .source-manifests/app.yaml

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
```

There are two fields in the Deployment manifest that will need to be
configurable. They are noted with comments. The Deployment's replicas and the
Pod's container image will change between different environments.  For example,
in a dev environment the number of replicas will be low and a development
version of the app will be run.  In production, there will be more replicas and
a stable release of the app will be used. In this example we don't have any
configurable fields in the Ingress or Service.

Next we need to use `+operator-builder:field` markers in comments to inform Operator Builder
that the operator will need to support configuration of these elements.
Following is the Deployment manifest with these markers in place.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webstore-deploy
  labels:
    team: dev-team  # +operator-builder:field:name=teamName,type=string
spec:
  replicas: 2  # +operator-builder:field:name=webStoreReplicas,default=2,type=int
  selector:
    matchLabels:
      app: webstore
  template:
    metadata:
      labels:
        app: webstore
        team: dev-team  # +operator-builder:field:name=teamName,type=string
    spec:
      containers:
      - name: webstore-container
        image: nginx:1.17  # +operator-builder:field:name=webStoreImage,type=string
        ports:
        - containerPort: 8080
```

These markers should always be provided as an in-line comment or as a head
comment.  The marker always begins with `+operator-builder:field:` or
`+operator-builder:collection:field:` See [Markers](docs/markers.md) to learn
more.

### Step 3: Create a Workload Config

Operator Builder uses a workload configuration to provide important details for
your operator project.  This guide uses a [standalone
workload](docs/standalone-workloads.md). Save a workload config to your
`.source-manifests` directory by using one of the following commands (or 
simply copy/pasting the YAML below the commands):

```bash
# generate a workload config with the path (-p) flag
operator-builder init-config standalone -p .source-manifests/workload.yaml

# generate a workload config from stdout
operator-builder init-config standalone > .source-manifests/workload.yaml
```

This will generate the following YAML:

```yaml
# .source-manifests/workload.yaml
name: webstore
kind: StandaloneWorkload
spec:
  api:
    domain: acme.com
    group: apps
    version: v1alpha1
    kind: WebStore
    clusterScoped: false
  companionCliRootcmd:
    name: webstorectl
    description: Manage webstore application
  resources:
  - app.yaml
```

The `name` is arbitrary and can be whatever you like.

In the `spec`, the following fields are required:

- `api.domain`: This must be a globally unique name that will not be used by other
  organizations or groups.  It will contain groups of API types.
- `api.group`: This is a logical group of API types used as a namespacing
  mechanism for your APIs.
- `api.version`: Provide the intiial version for your API.
- `api.kind`: The name of the API type that will represent the workload you are
  managing with this operator.
- `resources`: An array of filenames where your static manifests live.  List the
  relative path from the workload manifest to all the files that contain the
  static manifests we talked about in step 2.

For more info about API groups, versions and kinds, check out the [Kubebuilder
docs](https://kubebuilder.io/cronjob-tutorial/gvks.html).

The following fields in the `spec` are optional:

- `api.clusterScoped`: If your workload includes cluster-scoped resources like
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

```bash
tree .source-manifests

.source-manifests
├── app.yaml
└── workload.yaml
```

Our StandaloneWorkload config is in `workload.yaml` and the Deployment, Ingress
and Service manifests are in `app.yaml` and referenced under `spec.resources` in
our StandaloneWorkload config.

We are now ready to generate our project's source code.

### Step 4: Generate Operator Source Code

We first use the `init` command to create the general scaffolding.  We run this
command from the root of our repo and provide a single argument with the path to
our workload config.

```bash
operator-builder init \
    --workload-config .source-manfiests/workload.yaml
```

With the basic project now set up, we can now run the `create api` command to
create a new custom API for our workload.

```bash
operator-builder create api \
    --workload-config .source-manfiests/workload.yaml \
    --controller \
    --resource
```

We again provide the same workload config file.  Here we also added the
`--controller` and `--resource` arguments.  These indicate that we want both a
new controller and new custom resource created.

You now have a new working Kubernetes Operator!  Next, we will test it out.

### Step 5: Run & Test the Operator

Assuming you have a kubeconfig in place that allows you to interact with your
cluster with kubectl, you are ready to go.

First, install the new custom resource definition (CRD).

```bash
make install
```

Now we can run the controller locally to test it out.

```bash
make run
```

Operator Builder created a sample manifest in the `config/samples` directory.
For this example it looks like this:

```yaml
apiVersion: apps.acme.com/v1alpha1
kind: WebStore
metadata:
  name: webstore-sample
spec:
  webStoreReplicas: 2
  webStoreImage: nginx:1.17
  teamName: dev-team
```

You will notice the fields and values in the `spec` were derived from the
markers you added to your static manifests.

Next, in another terminal, create a new instance of your workload with
the provided sample manifest.

```bash
kubectl apply -f config/samples/
```

You should see your custom resource sample get created.  Now use `kubectl` to
inspect your cluster to confirm the workload's resources got created.  You should
find all the resources that were defined in your static manifests.

```bash
kubectl get all
```

Clean up by stopping your controller with ctrl-c in that terminal and then
remove all the resources you just created.

```bash
make uninstall
```

### Step 6: Build & Deploy the Controller Manager

Now let's deploy your controller into the cluster.

First export an environment variable for your container image.

```bash
export IMG=myrepo/acme-webstore-mgr:0.1.0
```

Run the rest of the commands in this step 6 in this same terminal as most of
them will need this `IMG` env var.

In order to run the controller in-cluster (as opposed to running locally with
`make run`) we will need to build a container image for it.

```bash
make docker-build
```

Now we can push it to a registry that is accessible from the test cluster.

```bash
make docker-push
```

Finally, we can deploy it to our test cluster.

```bash
make deploy
```

Next, perform the same tests from step 5 to ensure proper operation of our
operator.

```bash
kubectl apply -f config/sample/
```

Again, verify that all the resources you expect are created.

Once satisfied, remove the instance of your workload.

```bash
kubectl delete -f config/sample/
```

For now, leave the controller running in your test cluster.  We'll use it in
Step 7.

### Step 7: Build & Test Companion CLI

Now let's build and test the companion CLI.

You will have a make target that includes the name of your CLI.  For this
example it is:

```bash
make build-webstorectl
```

We can view the help info as follows.

```bash
./bin/webstorectl help
```

Your end users can use it to create a new custom resource manifest.

```bash
./bin/webstorectl init > /tmp/webstore.yaml
```

If you would like to change any of the default values, edit the file.

```bash
vim /tmp/webstore.yaml
```

Then you can apply it to the cluster.

```bash
kubectl apply -f /tmp/webstore.yaml
```

If your end users find they wish to make changes to the resources that aren't
supported by the operator, they can generate the resources from the custom
resource.

```bash
./bin/webstorectl generate --workload-manifest /tmp/webstore.yaml
```

This will print the resources to stdout.  These may be piped into an overlay
tool or written to disk and modified before applying to a cluster.

That's it!  You have a working operator without manually writing a single line
of code.  If you'd like to make any changes to your workload's API, you'll find
the code in the `apis` directory.  The controller's source code is in
`controllers` directory.  And the companion CLI code is in `cmd`.

Don't forget to clean up.  Remove the controller, CRD and the workload's
resources as follows.

```bash
make undeploy
```

For more information, checkout the [Operator Builder docs](docs/) as
well as the [Kubebuilder docs](https://kubebuilder.io/).

