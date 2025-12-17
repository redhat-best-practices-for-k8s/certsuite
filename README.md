# Red Hat Best Practices Test Suite for Kubernetes

![build](https://github.com/redhat-best-practices-for-k8s/certsuite/actions/workflows/merge.yaml/badge.svg)
[![QE OCP 4.14 Testing](https://github.com/redhat-best-practices-for-k8s/certsuite/actions/workflows/qe-ocp-414.yaml/badge.svg)](https://github.com/redhat-best-practices-for-k8s/certsuite/actions/workflows/qe-ocp-414.yaml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/redhat-best-practices-for-k8s/certsuite/badge)](https://scorecard.dev/viewer/?uri=github.com/)
[![go-doc](https://godoc.org/github.com/?status.svg)](https://godoc.org/github.com/)
[![release)](https://img.shields.io/github/v/release/redhat-best-practices-for-k8s/certsuite?color=blue&label=%20&logo=semver&logoColor=white&style=flat)](https://github.com/redhat-best-practices-for-k8s/certsuite/releases)
[![red hat](https://img.shields.io/badge/red%20hat---?color=gray&logo=redhat&logoColor=red&style=flat)](https://www.redhat.com)
[![openshift](https://img.shields.io/badge/openshift---?color=gray&logo=redhatopenshift&logoColor=red&style=flat)](https://www.redhat.com/en/technologies/cloud-computing/openshift)
[![license](https://img.shields.io/github/license/redhat-best-practices-for-k8s/certsuite?color=blue&labelColor=gray&logo=apache&logoColor=lightgray&style=flat)](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/LICENSE)

## Objective

To provide a set of test cases for the Containerized Network Functions/Cloud
Native Functions (CNFs) to verify if best practices for deployment on Red Hat
OpenShift clusters are followed.

* The test suite can be run as a standalone (after compiling the Golang code)
or as a container.
* The **full documentation** is published in the
[official documentation site](https://redhat-best-practices-for-k8s.github.io/certsuite/).
Please contact us in case the documentation is broken.

* The catalog of all the available test cases can be found in the [test catalog](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md).

## Demo

<!-- markdownlint-disable MD033 MD045 -->
<object type="image/svg+xml" data="docs/assets/images/demo-certsuite.svg">
<img src="docs/assets/images/demo-certsuite.svg">
</object>
<!-- markdownlint-enable MD033 MD045 -->

## Target Audience

* Partner
* Developer

## Pre-requisites for Running the Test Suite

* OCP or Kubernetes Cluster
* Docker or Podman (if running the container-based version)

## Pre-requisites Knowledge

* Basics of Kubernetes
* OpenShift Container Platform (OCP)
* Kubernetes Operator

## License

Red Hat Best Practices Test Suite for Kubernetes is copyright
[Red Hat, Inc.](https://www.redhat.com) and available under an [Apache 2 license](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/LICENSE).
