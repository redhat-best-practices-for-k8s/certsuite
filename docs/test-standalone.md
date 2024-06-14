<!-- markdownlint-disable code-block-style link-fragments line-length no-bare-urls no-emphasis-as-heading -->
# Run the Test Suite in standalone mode

The Test Suite can be run with the Certsuite tool either directly using its binary or through a container. Here we describe how to run it using the binary.

#### Run a single test

All tests have unique labels, which can be used to filter which tests are to be run. This is useful when debugging
a single test.

To select the test to be executed when running `run-cnf-suites.sh` with the following command-line:

```shell
./run-cnf-suites.sh -l operator-install-source
```

!!! note

    The test labels work the same as the suite labels, so you can select more than one test with the filtering mechanism shown before.

### Run all of the tests

You can run all of the tests (including the intrusive tests and the extended suite) with the following commands:

```shell
./run-cnf-suites.sh -l all
```

#### Run a subset

You can find all the labels attached to the tests by running the following command:

```shell
./run-cnf-suites.sh --list
```

You can also check the [CATALOG.md](CATALOG.md) to find all test labels.

#### Labels for offline environments

Some tests do require connectivity to Red Hat servers to validate certification status.
To run the tests in an offline environment, skip the tests using the `l` option.

```shell
./run-cnf-suites.sh -l '!online'
```

Alternatively, if an offline DB for containers, helm charts and operators is available, there is no need to skip those tests if the environment variable `TNF_OFFLINE_DB` is set to the DB location. This DB can be generated using the [OCT tool](https://github.com/test-network-function/oct).

Note: Only partner certified images are stored in the offline database. If Red Hat images are checked against the offline database, they will show up as not certified. The online database includes both Partner and Redhat images.

### Build + Test a workload

Refer [Developers' Guide](developers.md)
