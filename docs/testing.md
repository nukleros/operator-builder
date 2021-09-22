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
* `debug`: Runs operator-builder to generate an operator codebase using the
  [delve debugger](https://github.com/go-delve/delve).
* `debug-clean`: Use with caution. Deletes the contents of `TEST_WORKLOAD_PATH`
  where the test operator codebase is generated.
* `func-test`: Runs a functional test of operator-builder.  It builds and uses
  the built binary to run `operator-builder init` and `operator-builder create api`.
  It generates a new operator codebase in `FUNC_TEST_PATH`.
* `func-test-clean`: Use with caution. Deletes the contents of `FUNC_TEST_PATH`.

## Run Functional Testing

To run the default `application` test in the default `FUNC_TEST_PATH`:

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

