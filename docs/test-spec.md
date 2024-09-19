<!-- markdownlint-disable line-length no-bare-urls -->
# Test Specifications

## Available Test Specs

There are two categories for workload tests.

- **General**

These tests are designed to test any commodity workload running on OpenShift, and include specifications such as
`Default` network connectivity.

- **Workload-Specific**

These tests are designed to test some unique aspects of the workload under test are behaving correctly. This could
include specifications such as issuing a `GET` request to a web server, or passing traffic through an IPSEC tunnel.

### General tests

These tests belong to multiple suites that can be run in any combination as is
appropriate for the workload under test.

!!! info

    Test suites group tests by the topic areas.

Suite|Test Spec Description|Minimum OpenShift Version
---|---|---
`access-control`|The access-control test suite is used to test  service account, namespace and cluster/pod role binding for the pods under test. It also tests the pods/containers configuration.|4.6.0
`affiliated-certification`|The affiliated-certification test suite verifies that the containers and operators discovered or listed in the configuration file are certified by Redhat|4.6.0
`lifecycle`| The lifecycle test suite verifies the pods deployment, creation, shutdown and  survivability. |4.6.0
`networking`|The networking test suite contains tests that check connectivity and networking config related best practices.|4.6.0
`operator`|The operator test suite is designed to test basic Kubernetes Operator functionality.|4.6.0
`platform-alteration`| verifies that key platform configuration is not modified by the workload under test|4.6.0
`observability`|  the observability test suite contains tests that check workload logging is following best practices and that CRDs have status fields|4.6.0

!!! info

    Please refer [CATALOG.md](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md) for more details.

### Workload-specific tests

TODO
