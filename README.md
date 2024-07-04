# Red Hat Best Practices Test Suite for Kubernetes

![build](https://github.com/test-network-function/cnf-certification-test/actions/workflows/merge.yaml/badge.svg)
[![QE OCP 4.14 Testing](https://github.com/test-network-function/cnf-certification-test/actions/workflows/qe-ocp-414.yaml/badge.svg)](https://github.com/test-network-function/cnf-certification-test/actions/workflows/qe-ocp-414.yaml)
[![QE OCP 4.15 Testing](https://github.com/test-network-function/cnf-certification-test/actions/workflows/qe-ocp-415.yaml/badge.svg)](https://github.com/test-network-function/cnf-certification-test/actions/workflows/qe-ocp-415.yaml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/test-network-function/cnf-certification-test/badge)](https://scorecard.dev/viewer/?uri=github.com/test-network-function/cnf-certification-test)
[![go report](https://goreportcard.com/badge/github.com/test-network-function/test-network-function)](https://goreportcard.com/report/github.com/test-network-function/cnf-certification-test)
[![go-doc](https://godoc.org/github.com/test-network-function/cnf-certification-test?status.svg)](https://godoc.org/github.com/test-network-function/cnf-certification-test)
[![release)](https://img.shields.io/github/v/release/test-network-function/cnf-certification-test?color=blue&label=%20&logo=semver&logoColor=white&style=flat)](https://github.com/test-network-function/cnf-certification-test/releases)
[![red hat](https://img.shields.io/badge/red%20hat---?color=gray&logo=redhat&logoColor=red&style=flat)](https://www.redhat.com)
[![openshift](https://img.shields.io/badge/openshift---?color=gray&logo=redhatopenshift&logoColor=red&style=flat)](https://www.redhat.com/en/technologies/cloud-computing/openshift)
[![license](https://img.shields.io/github/license/test-network-function/cnf-certification-test?color=blue&labelColor=gray&logo=apache&logoColor=lightgray&style=flat)](https://github.com/test-network-function/cnf-certification-test/blob/main/LICENSE)

## Objective

To provide a set of test cases for the Containerized Network Functions/Cloud
Native Functions (CNFs) to verify if best practices for deployment on Red Hat
OpenShift clusters are followed.

* The test suite can be run as a standalone (after compiling the Golang code)
or as a container.
* The **full documentation** is published
[here](https://test-network-function.github.io/cnf-certification-test/).
Please contact us in case the documentation is broken.

* The catalog of all the available test cases can be found [here](https://github.com/test-network-function/cnf-certification-test/blob/main/CATALOG.md).

## Demo

<!-- markdownlint-disable MD033 -->
<object type="image/svg+xml" data="docs/assets/images/demo-certsuite.svg">
<img src="docs/assets/images/demo-certsuite.svg">
</object>
<!-- markdownlint-enable MD033 -->

## Target Audience

* Partner
* Developer

## Technical Pre-requisites for Running the Test Suite

* OCP or Kubernetes Cluster
* Docker or Podman (if running the container-based version)

## Pre-requisites for Topics Covered

* Knowledge on Kubernetes
* OpenShift Container Platform
* Kubernetes Operator

## Linters for the Codebase

* [`checkmake`](https://github.com/mrtazz/checkmake)
* [`golangci-lint`](https://github.com/golangci/golangci-lint)
* [`hadolint`](https://github.com/hadolint/hadolint)
* [`markdownlint`](https://github.com/igorshubovych/markdownlint-cli)
* [`shellcheck`](https://github.com/koalaman/shellcheck)
* [`shfmt`](https://github.com/mvdan/sh)
* [`typos`](https://github.com/crate-ci/typos)
* [`yamllint`](https://github.com/adrienverge/yamllint)

## License

Red Hat Best Practices Test Suite for Kubernetes is copyright
[Red Hat, Inc.](https://www.redhat.com) and available under an [Apache 2 license](https://github.com/test-network-function/cnf-certification-test/blob/main/LICENSE).
