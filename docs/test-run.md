<!-- markdownlint-disable code-block-style line-length no-bare-urls no-emphasis-as-heading -->
# Run the Test Suite

The Test Suite can be run using the Certsuite tool directly or through a container.

To run the Test Suite direct use:

```shell
certsuite run -l <label-filter> -c <certsuite-config> -k <kubeconfig> -o <output-dir> [<flags>]
```

If the _kubeconfig_ is not provided the value of the `KUBECONFIG` environment variable will be taken by default.

The CLI output will show the following information:

* Details of the Certsuite and claim file versions, the test case filter used and the location of the output files.
* The results for each test case grouped into test suites (the most recent log line is shown live as each test executes).
* Table with the number of test cases that have passed/failed or been skipped per test suite.
* The log lines produced by each test case that has failed.

Once the test run has completed, the test results can be visualized by opening the `results.html` website in a web browser and loading the `claim.json` file.

For more information on how to analyze the results see [Test Output](test-output.md).

## Building the Certsuite tool executable

The Certsuite binary can be built as follows:

```shell
make build-certsuite-tool
```

## Test labels

The test cases cases have several labels to allow for different types of groupings when selecting which to run. These are the following:

* The name of the test case
* The name of the test suite
* The category of the test case (common, telco, faredge, extended)

These labels can be combined with some operators to create label filters that match any condition. For example:

* The label filter "observability,access-control" will match the test suites _observability_ and _access-control_.
* The label filter "operator && !operator-crd-versioning" will match the _operator_ test suite without the _operator_crd_versioning_ test case.
* To select all the test cases the _all_ label filter can be used.

To view which test cases will run for a specific label or label filter use the flag `--list`.

See the [CATALOG.md](CATALOG.md) to find all test labels.

## Disable intrusive tests

To skip intrusive tests which may disrupt cluster operations, issue the
following:

```shell
certsuite run --intrusive=false
```

The intrusive test cases are:

* [lifecycle-deployment-scaling](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-deployment-scaling)
* [lifecycle-statefulset-scaling](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-statefulset-scaling)
* [lifecycle-crd-scaling](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-crd-scaling)
* [lifecycle-pod-recreation](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-pod-recreation)

Likewise, to enable intrusive tests, set the following:

```shell
certsuite run --intrusive=true
```

Intrusive tests are enabled by default.

## Selected flags description

The following is a non-exhaustive list of the most common flags that the `certsuite run` command accepts. To see the complete list use the `-h, --help` flag.

* `-l, --label-filter`: Label expression to filter test cases. Can be a test suite or list or test suites, such as `"observability,access-control"` or a more complex expression with logical operators such as `"access-control && !access-control-sys-admin-capability"`.

!!! note

    If `-l` is not specified, the Test Suite will run in 'diagnostic' mode. In this mode, no test case will run: it will only get information from the cluster (PUTs, CRDs, nodes info, etc…) to save it in the claim file. This can be used to make sure the configuration was properly set and the autodiscovery found the right pods/crds…

* `-o, --output-dir`: Path of the local directory where test results (claim.json), the execution logs (certsuite.log), and the results artifacts file (results.tar.gz) will be available from after the container exits.

* `-k, --kubeconfig`: Path to the Kubeconfig file of the target cluster.

* `-c, --config-file`: Path to the `certsuite_config.yml` file.

* `--preflight-dockerconfig`: Path to the Dockerconfig file to be used by the Preflight test suite

* `--offline-db`: Path to an offline DB to check the certification status of container images, operators and helm charts. Defaults to the DB included in the test container image.

!!! note

    See the [OCT tool](https://github.com/redhat-best-practices-for-k8s/oct) for more information on how to create this DB.

* `--cleanup-probe`: Controls whether the probe daemonset and its namespace are deleted at the end of the test run. By default (true), the probe daemonset is cleaned up after tests complete. Set to `--cleanup-probe=false` to keep the probe daemonset running on the cluster for debugging or repeated test runs.

```shell
certsuite run --cleanup-probe=false
```

When running in container mode, add the flag to the certsuite command:

```shell
docker run --rm --network host \
  -v <path-to-local-dir>/config:/usr/certsuite/config:Z \
  -v <path-to-local-dir>/results:/usr/certsuite/results:Z \
  quay.io/redhat-best-practices-for-k8s/certsuite:latest \
  certsuite run \
  --kubeconfig=/usr/certsuite/config/kubeconfig \
  --config-file=/usr/certsuite/config/certsuite_config.yml \
  --output-dir=/usr/certsuite/results \
  --label-filter=all \
  --cleanup-probe=false
```

## Using the container image

The only prerequisite for running the Test Suite in container mode is having Docker or Podman installed.

### Pull the test image

The test image is available at this [repository](https://quay.io/repository/redhat-best-practices-for-k8s/certsuite) and can be pulled using:

```shell
docker pull quay.io/redhat-best-practices-for-k8s/certsuite:<image-tag>
```

The image tag can be `latest` to select the latest release, `unstable` to fetch the image built with the latest commit in the repository or any existing version number such as `v5.2.1`.

### Launch the Test Suite

The Test Suite requires 3 files that must be provided to the test container:

* The _Kubeconfig_ for the target cluster.
* The _Dockerconfig_ of the local Docker installation (only for the Preflight test suite).
* The `certsuite_config.yml`.

To reduce the number of shared volumes with the test container in the example below those files are copied into a folder called "config". Also, another folder to contain the output files called "results" has been created. The files saved in the output directory after the test run are:

* A `claim.json` file with the test results.
* A `certsuite.log` file with the execution logs.
* A `.tar.gz` file with the above two files and an additional `results.html` file to visualize the results in a website.

```shell
docker run --rm --network host \
  -v <path-to-local-dir>/config:/usr/certsuite/config:Z \
  -v <path-to-local-dir>/results:/usr/certsuite/results:Z \
  quay.io/redhat-best-practices-for-k8s/certsuite:latest \
  certsuite run \
  --kubeconfig=/usr/certsuite/config/kubeconfig \
  --preflight-dockerconfig=/usr/certsuite/config/dockerconfig \
  --config-file=/usr/certsuite/config/certsuite_config.yml \
  --output-dir=/usr/certsuite/results \
  --label-filter=all
```
