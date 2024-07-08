# API Updates & Upgrades

This document deals with two things:
1. Updating an existing API.
2. Adding a new version of an API.

## Updating an Existing API

In this scenario, you are overwriting and changing an existing API specification.
You should only do this during development with an unreleased API that no end user
has started using.

For example, you have begun development on a brand new operator.  You have
generated the source code from a set of YAML manifests with markers.  While
testing, you discover that a field is misspelled, or that a default value should
be changed, or that a new field should be added.  The following instructions
describe how to overwrite an existing API to update the existing spec.  Please 
note that in the below example the `--resource=true` is not necessary and is 
only provided in the example for verbosity.  This option is set by default.

After making the necessary changes to your manifests run the following:

```bash
operator-builder create api \
    --workload-config [path/to/workload/config] \
    --controller=false \
    --resource=true \
    --force
```

You will pass the same workload config file.  The `--controller=false` flag will
skip generating controller code but `--resource` and `--force` will cause the
existing API to be overwritten.

Note: if you change any of the following fields in the workload config you will
get a new API rather than overwrite the existing one.
- `spec.api.domain`
- `spec.api.group`
- `spec.api.version`
- `spec.api.kind`

## Adding a New Version of an API

In this scenario, an existing version of an API is in use by end users.  You
need to add features to the API that require adding, removing or altering
fields.  Do not make breaking changes to an exising API that will render users'
existing custom resource manifests invalid.

Instead, you must add a *new* API version.  Now, you have two choices:
1. Maintain backward compatibility with previous API version/s.
2. Require an upgrade of API version with a new version of your controller and
   companion CLI.  In this case, ensure your users understand what to expect so
   they don't attempt to use an old API version with a new version of the
   controller.

### Kubernetes API Versions

Kubernetes has adopted detailed conventions for changing APIs.  To learn these
details visit [The Kubernetes
API](https://kubernetes.io/docs/concepts/overview/kubernetes-api/) and [API
Overview](https://kubernetes.io/docs/reference/using-api/) in the Kubernetes
docs.

While you are naturally free to use your own conventions, we encourage you to
model them on the Kubernetes system.  Your users will likely be familiar with
them and those upstream conventions have sound reasoning behind them.

For the purposes of this docuement, here are the important points:
- Do not maintain backward compatibility for alpha API verions.  The development
  cost of conversion between API versions is non-trivial.  Collaborate with the
  early adopters that use your alpha versions and clearly document which
  software versions support which API versions.
- Maintain backward compatibility for beta versions of your APIs for 9 months or
  3 releases (whichever is longer).  Do not release a beta version of your API
  until you are confident few changes will be required in the forseeable future.
- Maintain backward compatibility for stable versions for 12 months or 3
  releases (whichever is longer).  Do not release a stable version until the API
  is well tested and ready for production use.

Following is a table that provides an example of what this may look like.  The
first column shows the release version that applies to the controller and
companion CLI.  The versions for these two components will always be pinned
together - they share source code after all.  Notice that there is only one
major version change.  The 1.0 release coincides with the v1 release of the API
and signifies production readiness.  It does not signify any breaking change
as the API versions have their own compatibility lifecycle.  As such there is no
meaning to a 2.0 release of the controller and CLI.  This would only make sense
if entirely new API groups and types were implemented to break backward
compatibility.

| Controller, CLI Version | Supported API Versions   | Notes                                       |
|-------------------------|--------------------------|---------------------------------------------|
| 0.1                     | v1alpha1                 | Initial Release                             |
| 0.2                     | v1alpha2                 | v1alpha1 support removed                    |
| 0.3                     | v1alpha3                 | v1alpha2 support removed                    |
| 0.4                     | v1beta1                  | v1alpha3 support removed                    |
| 0.5                     | v1beta2, v1beta1         | v1beta1 deprecated                          |
| 0.6                     | v1beta2, v1beta1         | No change in API, software features added   |
| 1.0                     | v1, v1beta2              | v1beta1 support removed, v1beta2 deprecated |
| 1.1                     | v2alpha1, v1             |                                             |
| 1.2                     | v2alpha2, v1             | v1beta2 and v2alpha1 support removed        |
| 1.3                     | v2beta1, v1              | v2alpha2 support removed                    |
| 1.4                     | v2beta2, v2beta1, v1     | v2beta1 deprecated                          |
| 1.5                     | v2, v2beta2, v2beta1, v1 | v2beta2 and v1 deprecated                   |
| 1.6                     | v2, v2beta2, v1          | v2beta1 support removed                     |
| 1.7                     | v2, v1                   | v2beta2 support removed                     |
| 1.8                     | v2                       | v1 support removed                          |

Note there there is no change in supported API versions with version 0.6.  The
convention of increasing minor version when software features are added remains.
Any time an API version is added or removed, a new minor version should be
released, however an API change is not _needed_ to justify a new minor version of
the software when other features not related to an API are released.

To create a new version of an existing API, update the `spec.api.version` value
in your workload config, for example:

```yaml
name: webstore
kind: StandaloneWorkload
spec:
  api:
    domain: acme.com
    group: apps
    version: v1alpha2  # existing API version is v1alapha1
    kind: WebStore
    clusterScoped: false
  companionCliRootcmd:
    name: webstorectl
    description: Manage the webstore app
  resources:
  - app.yaml
```

Now reference the config in a new `create api` command:

```bash
operator-builder create api \
    --workload-config [path/to/workload/config] \
    --controller \
    --resource \
    --force
```

Note that we _do_ want to re-generate the controller in this case.
A new API definition will be create alongside the previous version.  If the
earlier version of the API is to be unsupported, you can delete the earlier
version.

For example if your APIs look as follows:

```bash
tree apis/apps
apis/apps
├── v1alpha1
│   ├── groupversion_info.go
│   ├── webstore
│   │   ├── app.go
│   │   └── resources.go
│   ├── webstore_types.go
│   └── zz_generated.deepcopy.go
└── v1alpha2
    ├── groupversion_info.go
    ├── webstore
    │   ├── app.go
    │   └── resources.go
    ├── webstore_types.go
    └── zz_generated.deepcopy.go

4 directories, 10 files
```

You will delete the earlier version with `rm -rf apis/apps/v1alpha1`.

If, however, you want to retain backward compatibility and support both versions
you will need to implement conversion between the APIs.  Operator Builder does
not yet support any scaffolding or code generation for this.  For details on how
to accomplish this, refer to the [Kubebuilder
docs](https://kubebuilder.io/multiversion-tutorial/conversion.html) on API
conversion.

