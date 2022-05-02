# Testing

When making changes to Operator Builder, you can test your changes with
functional tests that generate the codebase for a new operator.  You can also
use [delve](https://github.com/go-delve/delve) to test changes.

Keep in mind you are testing a source code generator
that end-user engineers will use to manage Kubernetes operator projects.
Operator Builder is used to generate code for a distinct code repository - so the
testing is conducted as such.  It stamps out and/or modifies source code for an
operator project when a functional or debug test is run.

At this time, manual verification of results is needed.  In future, automated
integration tests will be added to test the generated operator.

## Make Targets

* `test`: Reserved for unit tests.
* `test-e2e`: Run the E2E tests as indicated by the `e2e_test` tag and stored in the 
  `test/e2e` path from the generated repository.
* `debug`: Runs operator-builder to generate an operator codebase using the
  [delve debugger](https://github.com/go-delve/delve).
* `debug-clean`: Use with caution. Deletes the contents of `TEST_WORKLOAD_PATH`
  where the test operator codebase is generated.
* `func-test`: Runs a functional test of operator-builder.  It builds and uses
  the built binary to run `operator-builder init` and `operator-builder create api`.
  It generates a new operator codebase in `FUNC_TEST_PATH`.
* `func-test-clean`: Use with caution. Deletes the contents of `FUNC_TEST_PATH`.

## Run Functional Testing

To run the default `application` (based on a standalone use case) test in the default `FUNC_TEST_PATH`:

    make func-test

This will generate the codebase for an operator that uses a [standalone
workload](standalone-workloads.md) in your local `/tmp/operator-builder-func-test`
directory.

To run the `platform` test in a non-default directory:

    FUNC_TEST_PATH=/path/to/platform-operator \
      TEST_WORKLOAD_PATH=test/platform \
      make func-test

This will generate the cdoebase for an operator that uses a [workload
collection](workload-collections.md) in the `/path/to/platform-operator`
directory.

## New Functional Tests

Follow these steps to create a new test case:

1. Create a descriptive name for the test by creating a new directory under
   the `test/` directory.
2. Create a `.workloadConfig` directory within your newly created directory.
3. Add the YAML files for your workload under the newly created `.workloadConfig`
   directory.
4. Create a [workload configuration](workloads.md) in the `.workloadConfig` directory
   with the name `workload.yaml`.


## Run E2E Testing

As part of the generated code repo, operator-builder will lay down a set of E2E tests 
that are meant to be run either during normal operating conditions, to check status 
of the operator, or while make changes to the operator, perhaps as part of a CI/CD or 
GitOps workflow.

There are a few different scenarios to consider when running E2E tests:

### Scenario 1 (Default): CRDs Installed in Cluster, Controller Running

This is the default scenario as it is generally the quickest and easiest to spin up.  It is 
also the least invasive and will cause the least amount of headache in the instance that 
the `make test-e2e` target is run, accidentally, against an operational cluster.  In this 
scenario, the E2E test assumes that the custom CRDs are installed into the cluster 
and the controller is either deployed in the cluster, or running with the `make run` target.

To test this scenario, simply run:

    make test-e2e

### Scenario 2: CRDs Not Installed in Cluster, Controller Running

This scenario allows you to deploy the CRDs into the cluster as part of E2E testing, but 
also makes the assumption that the the controller is either deployed in the cluster, or 
running with the `make run` target.

To test this scenario, simply run:

    DEPLOY="true" make test-e2e

### Scenario 3: CRDs Not Installed in Cluster, Controller Not Running

This scenario allows you to deploy the CRDs into the cluster as part of E2E testing and 
additionally run through the workflow to deploy the controller into the cluster.  This 
is the most robust of all of the tests and should be considered for CI/CD or GitOps 
workflows for testing.

To test this scenario, simply run:

    DEPLOY="true" DEPLOY_IN_CLUSTER="true" IMG="my-repo/my-image:my-tag" make test-e2e

### Additional Options

Additionally, the following environment variables are available as options for the 
E2E Testing:

* `TEARDOWN`: when set to "true", the teardown procedures are run and tested.  You 
  may not want to set this if you need to inspect the cluster in the case of failed 
  deployments.
