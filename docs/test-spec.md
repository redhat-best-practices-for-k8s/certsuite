<!-- markdownlint-disable line-length no-bare-urls -->
# Test Specifications

## Available Test Specs

These tests belong to multiple suites that can be run in any combination as is
appropriate for the workload under test. Use `--label-filter` with a suite
name, a test case name, or a category (`common`, `telco`, `faredge`,
`extended`).

!!! info

    Test suites group tests by topic. QE currently covers OpenShift 4.14 through 4.22.

Suite|Description
---|---
`access-control`|Service accounts, namespaces, role bindings, and pod/container security context.
`affiliated-certification`|Certification status of discovered containers, operators, and Helm charts.
`lifecycle`|Pod deployment, scaling, shutdown, and survivability.
`manageability`|Container port names and other manageability best practices.
`networking`|Connectivity and networking configuration best practices.
`observability`|Workload logging and CRD status fields.
`operator`|Operator lifecycle, installation, and related best practices.
`platform-alteration`|Key platform configuration is not modified by the workload under test.
`performance`|CPU pinning, exec probes, and related performance checks.
`preflight`|Red Hat preflight certification checks for containers and operators.

!!! info

    See [CATALOG.md](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md) for the full list of test cases and labels.
