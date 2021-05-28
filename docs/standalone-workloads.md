# Standalone Workloads

When you are building an operator for a single workload, you will just need a
single standalone WorkloadConfig along with the source manifests to define the
resources that will be created to fulfill an instance of your workload.

For example if your organization develops and maintains a web application as a
part of its core business, you may consider using an operator to deploy and
maintain that app in different environments.  We refer to all the various
resources that comprise that web app collectively as a "workload."

This is the simplest and most common implemetation of operator-builder.

If you have multiple workloads that have dependencies upon one another, and it
makes sense to orchestrate them with an operator, a standalone workload will not
suffice.  For that you will need to leverage a [workload
collection](workload-collections.md).

