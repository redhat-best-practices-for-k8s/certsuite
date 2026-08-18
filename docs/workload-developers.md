<!-- markdownlint-disable header-increment line-length no-bare-urls no-emphasis-as-heading -->
# Workload Guidelines for developers

Developers of Kubernetes workloads, particularly those targeting
[Red Hat certification on OpenShift](https://connect.redhat.com),
can use this suite to test the interaction of their workload with OpenShift.
If interested in certification, start with
[Red Hat Partner Connect](https://connect.redhat.com).

**Requirements**

- An [OpenShift cluster](https://docs.redhat.com/en/documentation/openshift_container_platform/4.21/html/welcome/index) in a currently supported version (QE covers 4.14 through 4.22)
- At least one extra machine to host the test suite, or run certsuite as a container or in-cluster pod

**Reference**

The [certsuite-sample-workload](https://github.com/redhat-best-practices-for-k8s/certsuite-sample-workload) repository provides a sample setup to model against.
