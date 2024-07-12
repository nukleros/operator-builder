# How to Contribute
1. The operator-builder team welcomes contributions from the community.  Before you start working with operator-builder, please read our [Developer Certificate of Origin](https://cla.vmware.com/dco).  All contributions to this repository must be signed as described on that page.  Your signature certifies that you wrote the patch or have the right to pass it on as an open-source patch.
1. Open an issue with what you intend to fix or would like to add into the project
1. A project maintainer will triage and assess the impact of this feature or issue, and whether it should be brought into the project
1. Once approval of the issue occurs, work on the issue should start by first forking the project
1. A feature branch should be created in the forked repository
1. A Pull Request (PR) should be created to indicate (WIP) work in progress
1. Once the work has been completed the (WIP) on the PR should be removed
1. The work will then be reviewed either by an automated CI process or manual testing
1. Once all tests have passed, the code will be reviewed
1. Once code review has been completed, the PR will either be approved, or further changes will be made on the feature branch until it is determined to function as expected, and either fixes or adds the feature that the issue initially raised.
1. Occasionally, exceptions will be allowed to this process, but only in rare circumstances when the maintainer deems it necessary.

## Commit Messages

We use the [Conventional Commits 1.0.0
spec](https://www.conventionalcommits.org/en/v1.0.0/).  This helps keep things
standardized and allows us to automate generating CHANGELOGs.

## Testing

In order to test the effect of changes made to Operator Builder, use `make
func-test` or `make debug`.

At a minimum, ensure your changes work for:
- standalone: This tests a basic standalone workload use case.
- collection: This tests a basic workload collection use case.
- edge-standalone: This tests standalone workloads which contain edge cases.
- edge-collection: This tests a collection workload which contains edge cases.

See the [testing docs](docs/testing.md) for instructions on how to run these
tests.

