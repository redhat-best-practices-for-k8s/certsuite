# Standalone test executable

**Prerequisites**

The repo is cloned and all the commands should be run from the cloned repo.

```shell
mkdir ~/workspace
cd ~/workspace
git clone git@github.com:test-network-function/cnf-certification-test.git
cd cnf-certification-test
```


!!! note

    By default, `cnf-certification-test` emits results to `cnf-certification-test/cnf-certification-tests_junit.xml`.

### 1. Install dependencies

Run the following command to install the following dependencies.

```shell
make install-tools
```

Dependency|Minimum Version
---|---
[GoLang](https://golang.org/dl/)|1.19
[golangci-lint](https://golangci-lint.run/usage/install/)|1.51.1
[jq](https://stedolan.github.io/jq/)|1.6
[OpenShift Client](https://mirror.openshift.com/pub/openshift-v4/clients/ocp/)|4.11

Other binary dependencies required to run tests can be installed using the following command:

!!! note

    * You must also make sure that `$GOBIN` (default `$GOPATH/bin`) is on your `$PATH`.

    * Efforts to containerise this offering are considered a work in progress.


### 2. Build the Test Suite

In order to build the test executable, first make sure you have satisfied the [dependencies](#dependencies).
```shell
make build-cnf-tests
```

*Gotcha:* The `make build*` commands run unit tests where appropriate. They do NOT test the CNF.

### 3. Test a CNF

A CNF is tested by specifying which suites to run using the `run-cnf-suites.sh` helper
script.

Run any combination of the suites keywords listed at in the [General tests](test-spec.md#general-tests) section, e.g.

```shell
./run-cnf-suites.sh -l "lifecycle"
./run-cnf-suites.sh -l "networking,lifecycle"
./run-cnf-suites.sh -l "operator,networking"
./run-cnf-suites.sh -l "networking,platform-alteration"
./run-cnf-suites.sh -l "networking,lifecycle,affiliated-certification,operator"
```

!!! note

    As with "run-tnf-container.sh", if `-l` is not specified here, the tnf will run in 'diagnostic' mode.

    By default the claim file will be output into the same location as the test executable. The `-o` argument for
    `run-cnf-suites.sh` can be used to provide a new location that the output files will be saved to. For more detailed
    control over the outputs, see the output of `cnf-certification-test.test --help`.

    ```shell
    cd cnf-certification-test && ./cnf-certification-test.test --help
    ```

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
./run-cnf-suites.sh -l common,extended
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

### Build + Test a CNF

Refer [Developers' Guide](developers.md)
