<!-- markdownlint-disable code-block-style line-length no-bare-urls no-emphasis-as-heading -->
# Run the Test Suite

The Test Suite can be run with the Certsuite tool either using the binary executable directly or through a container.

## Common context

### Test labels

### Selected flags description


## Using the Certsuite tool executable

### Fetch the Certsuite executable

To be explained when the Certsuite binary release method is in place.

### Launch the Test Suite



## Using the container image

The only pre-requisite for running the Test Suite in container mode is having Docker or Podman installed.

### Pull the test image

The test image is available at this repository [repository](https://quay.io/repository/testnetworkfunction/cnf-certification-test) and can be pulled using:

```shell
docker pull quay.io/testnetworkfunction/cnf-certification-test
```

### Launch the Test Suite

The Test Suite requires 3 files that must be provided to the test container:
* The _Kubeconfig_ for the target cluster.
* The _Dockerconfig_ of the local Docker installation (only for the Preflight test suite).
* The `tnf_config.yml`.

To reduce the number of shared volumes with the test container in the example below those files are copied into a folder called "config". Also, another folder to contain the output files called "results" has been created. The files saved in the output directory after the test run are:

* A `claim.json` file with the test results.
* A `certsuite.log` file with the execution logs.
* A `.tar.gz` file with the above two files and an addional `results.html` file to visualize the results in a website.

```shell
docker run --rm --network host 
  -v <path-to-local-dir>/config:/usr/tnf/config:Z 
  -v <path-to-local-dir>/results:/usr/tnf/results:Z 
  
  quay.io/testnetworkfunction/cnf-certification-test:latest
  
  ./cnf-certification-test/certsuite run 
  --kubeconfig=/usr/tnf/config/kubeconfig
  --preflight-dockerconfig=/usr/tnf/config/dockerconfig
  --config-file=/usr/tnf/config/tnf_config.yml
  --output-dir=/usr/tnf/results
  --label-filter=all
```

The CLI output will show the following information:

* Details of the Certsuite and claim file versions, the test case filter used and the location of the output files.
* The results for each test case grouped into test suites (the most recent log line is shown live as each test executes).
* Table with the number of test cases that have passed/failed or been skipped per test suite.
* The log lines produced by each test case that has failed.

Once the test run has completed, the test results can be visualized by opening the `results.html` website in a web browser and loading the `claim.json` file.

**Selected flags description**

The following is a non-exhaustive list of the most common flags that the `certsuite run` command accepts. To see the complete list use the `-h, --help` flag.

* `-l, --label-filter`: Label expression to filter test cases. Can be a test suite or list or test suites, such as `"observability,access-control"` or a more complex expression with logical operators such as `"access-control && !access-control-sys-admin-capability"`.

!!! note

    If `-l` is not specified, the Test Suite will run in 'diagnostic' mode. In this mode, no test case will run: it will only get information from the cluster (PUTs, CRDs, nodes info, etc…) to save it in the claim file. This can be used to make sure the configuration was properly set and the autodiscovery found the right pods/crds…

* `-o, --output-dir`: Path of the local directory where test results (claim.json), the execution logs (certsuite.log), and the results artifacts file (results.tar.gz) will be available from after the container exits.

* `-k, --kubeconfig`: Path to the Kubeconfig file of the target cluster.

* `-c, --config-file`: Path to the `tnf_config.yml` file.

* `--preflight-dockerconfig`: Path to the Dockerconfig file to be used by the Preflight test suite

* `--offline-db`: Path to an offline DB to check the certification status of container images, operators and helm charts. Defaults to the DB included in the test container image.

!!! note

    See the [OCT tool](https://github.com/test-network-function/oct) for more information on how to create this DB.
