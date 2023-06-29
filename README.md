# CNF Certification Test

![build](https://github.com/test-network-function/cnf-certification-test/actions/workflows/merge.yaml/badge.svg)
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

## Target Audience

* Partner
* Developer

## Technical Pre-requisites for Running the Test Suite

* OCP or Kubernetes Cluster
* Docker or Podman (if running the container-based version)
* `make install-tools` (if running the compiled binary version)

## Pre-requisites for Topics Covered

* Knowledge on Kubernetes
* OpenShift Container Platform
* Kubernetes Operator

## Language

Golang

## Linters for the Codebase

* [`checkmake`](https://github.com/mrtazz/checkmake)
* [`golangci-lint`](https://github.com/golangci/golangci-lint)
* [`hadolint`](https://github.com/hadolint/hadolint)
* [`markdownlint`](https://github.com/igorshubovych/markdownlint-cli)
* [`shellcheck`](https://github.com/koalaman/shellcheck)
* [`shfmt`](https://github.com/mvdan/sh)
* [`typos`](https://github.com/crate-ci/typos)
* [`yamllint`](https://github.com/adrienverge/yamllint)

## Show Results after finishing of running the test code

After the end of your run a claim.json file will be created for you with the
results of that specific test run. To see them in a good way that is clear for
you we created a parser that is local. You can open it by copying the path /
 #your-code-path/cnf-certification-test/script/results.html and open it in your
browser then upload the claim.json file of that run.
![overview](docs/assets/images/htmlpage.png)

## Upload a previous feedback to the html parser page

A feature on the html parser page that we did to allow users to upload a previous
feedback that they have to the parser page, they have an option to download the
feedback then to upload it for the next time from the html page another option users
can have that feedback file in the path #your-code-path/cnf-certification-test/script/
and it need to be named as feedback.js users need to convert the downloaded
feedback.json to feedback.js by running the command
./tnf generate catalog feedbackjs -f path-to-feedbackjson/feedback.json
the feedback.js will be copy to #your-code-path/cnf-certification-test/script/
users can see there feedback without upload it from the html page.

## License

CNF Certification Test is copyright [Red Hat, Inc.](https://www.redhat.com) and available
under an
[Apache 2 license](https://github.com/test-network-function/cnf-certification-test/blob/main/LICENSE).
