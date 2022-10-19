The tests can be run within a prebuilt container in the OCP cluster.

**Prerequisites for the OCP cluster**

* The cluster should allow interactive shell sessions to pods/containers to stay alive when being idle for more than a few minutes. To enable it, consult the maintainer of the cluster infrastructure. Also, make sure the firewalls/load balancers on the path do not timeout idle connections too quickly.
* The cluster should have enough resources to drain nodes and reschedule pods. If that is not the case, then ``lifecycle-pod-recreation`` test should be skipped.

## With quay test container image
### Pull test image
The test image is available at this repository in [quay.io](https://quay.io/repository/testnetworkfunction/cnf-certification-test) and can be pulled using
The image can be pulled using :

```shell
podman pull quay.io/testnetworkfunction/cnf-certification-test
```

### Check cluster resources
Some tests suites such as `platform-alteration` require node access to get node configuration like `hugepage`.
In order to get the required information, the test suite does not `ssh` into nodes, but instead rely on [oc debug tools ](https://docs.openshift.com/container-platform/3.7/cli_reference/basic_cli_operations.html#debug). This tool makes it easier to fetch information from nodes and also to debug running pods.

`oc debug tool` will launch a new container ending with **-debug** suffix, and the container will be destroyed once the debug session is done. Ensure that the cluster should have enough resources to create debug pod, otherwise those tests would fail.

!!! note

    It's **recommended** to clean up disk space and make sure there's enough resources to deploy another container image in every node before starting the tests.


### Run the tests

```shell
./run-tnf-container.sh
```

**Required arguments**

* `-t` to provide the path of the local directory that contains tnf config files
* `-o` to provide the path of the local directory that the test results will be available after the container exits.

!!! warning

    This directory must exist in order for the claim file to be written.

**Optional arguments**

* `-l` to list the labels to be run. See [Ginkgo Spec Labels](https://onsi.github.io/ginkgo/#spec-labels) for more information on how to filter tests with labels.

!!! note

    If `-l` is not specified, the tnf will run in 'diagnostic' mode. In this mode, no test case will run: it will only get information from the cluster (PUTs, CRDs, nodes info, etc...) to save it in the claim file. This can be used to make sure the configuration was properly set and the autodiscovery found the right pods/crds...

* `-i` to provide a name to a custom TNF container image. Supports local images, as well as images from external registries.


* `-k` to set a path to one or more kubeconfig files to be used by the container to authenticate with the cluster. Paths must be separated by a colon.

!!! note

    If `-k` is not specified, autodiscovery is performed.

    The autodiscovery first looks for paths in the `$KUBECONFIG` environment variable on the host system, and if the variable is not set or is empty, the default configuration stored in `$HOME/.kube/config` is checked.

* `-n` to give the network mode of the container. Defaults set to `host`, which requires selinux to be disabled. Alternatively, `bridge` mode can be used with selinux if TNF_CONTAINER_CLIENT is set to `docker` or running the test as root.

!!! note

    See the [docker run --network parameter reference](https://docs.docker.com/engine/reference/run/#network-settings) for more information on how to configure network settings.


**Command to run**

```shell
./run-tnf-container.sh -k ~/.kube/config -t ~/tnf/config
-o ~/tnf/output -l "networking,access-control"
```

See [General tests](#general-tests) for a list of available keywords.

### Run with `docker`

By default, `run-container.sh` utilizes `podman`. However, an alternate container virtualization
client using `TNF_CONTAINER_CLIENT` can be configured. This is particularly useful for operating systems that do not readily support
`podman`.

In order to configure the test harness to use `docker`, issue the following prior to
`run-tnf-container.sh`:

```shell
export TNF_CONTAINER_CLIENT=docker
```

## With local test container image


### Build locally

```shell
podman build -t cnf-certification-test:v1.0.5 \
  --build-arg TNF_VERSION=v1.0.5 \
  --build-arg OPENSHIFT_VERSION=4.7.55 .
```

  - `TNF_VERSION` value is set to a branch, a tag, or a hash of a commit that will be installed into the image
  -  `OPENSHIFT_VERSION` value points to the OCP version of the cluster in which the workloads to be tested are deployed.


### Build from an unofficial source

The unofficial source could be a fork of the TNF repository.

Use the `TNF_SRC_URL` build argument to override the URL to a source repository.

```shell
podman build -t cnf-certification-test:v1.0.5 \
  --build-arg TNF_VERSION=v1.0.5 \
  --build-arg TNF_SRC_URL=https://github.com/test-network-function/cnf-certification-test \
  --build-arg OPENSHIFT_VERSION=4.7.55 .
```

### Run the tests

Specify the custom TNF image using the `-i` parameter.

```shell
./run-tnf-container.sh -i test-network-function:v1.0.5
-t ~/tnf/config -o ~/tnf/output -l "networking,access-control"
```
 Note: see [General tests](#general-tests) for a list of available keywords.
