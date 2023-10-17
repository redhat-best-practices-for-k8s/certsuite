<!-- markdownlint-disable code-block-style line-length no-bare-urls -->
# Test configuration

The certification test suite supports autodiscovery using labels and annotations.

These can be configured through the following config file.

- `tnf_config.yml`

[Sample](https://github.com/test-network-function/cnf-certification-test/blob/main/cnf-certification-test/tnf_config.yml)

As per the requirement the following fields can be changed.

## CNF Config Generator

The CNF config file can be created using the CNF Config Generator, which is part of the TNF tool shipped with the CNF Certification. The purpose of this particular tool is to help users configuring the CNF Certification providing a logical structure of the available options as well as the information required to make use of them. The result is a CNF config file in YAML format that the CNF Certification will parse to adapt the certification process to a specific CNF workload.

To compile the TNF tool:

```shell
make build-tnf-tool
```

To launch the CNF Config Generator:

```shell
./tnf generate config
```

Here's an example of how to use the tool:

<!-- markdownlint-disable MD033 -->
<object type="image/svg+xml" data="./assets/images/demo-config.svg">
<img width="600" src="../assets/images/demo-config.svg">
</object>
<!-- markdownlint-enable MD033 -->

## targetNameSpaces

Multiple namespaces can be specified to deploy partner pods for testing through `targetNameSpaces` in the config file.

``` { .yaml .annotate }
targetNameSpaces:
  - name: firstnamespace
  - name: secondnamespace
```

## targetPodLabels

The goal of this section is to specify the labels to be used to identify the CNF resources under test.

!!! note "Highly recommended"

    The labels should be defined in pod definition rather than added after pod is created, as labels added later on will be lost in case the pod gets rescheduled. In case of pods defined as part of a deployment, it's best to use the same label as the one defined in the `spec.selector.matchLabels` section of the deployment yaml. The prefix field can be used to avoid naming collision with other labels.

``` { .yaml .annotate }
targetPodLabels:
  - prefix: test-network-function.com
    name: generic
    value: target
```

The corresponding pod label used to match pods is:

``` { .yaml .annotate }
test-network-function.com/generic: target
```

Once the pods are found, all of their containers are also added to the target container list. A target deployment list will also be created with all the deployments which the test pods belong to.

## targetCrds

In order to autodiscover the CRDs to be tested, an array of search filters can be set under the "targetCrdFilters" label. The autodiscovery mechanism will iterate through all the filters to look for all the CRDs that match it. Currently, filters only work by name suffix.

``` { .yaml .annotate }
targetCrdFilters:
 - nameSuffix: "group1.tnf.com"
 - nameSuffix: "anydomain.com"
```

The autodiscovery mechanism will create a list of all CRD names in the cluster whose names have the suffix `group1.tnf.com` or `anydomain.com`, e.g. `crd1.group1.tnf.com` or `mycrd.mygroup.anydomain.com`.

## testTarget

### podsUnderTest / containersUnderTest

The autodiscovery mechanism will attempt to identify the default network device and all the IP addresses of the pods it needs for network connectivity tests, though that information can be explicitly set using annotations if needed.

#### Pod IPs

- The `k8s.v1.cni.cncf.io/networks-status` annotation is checked and all IPs from it are used. This annotation is automatically managed in OpenShift but may not be present in K8s.
- If it is not present, then only known IPs associated with the pod are used (the pod `.status.ips` field).

#### Network Interfaces

- The `k8s.v1.cni.cncf.io/networks-status` annotation is checked and the `interface` from the first entry found with `“default”=true` is used. This annotation is automatically managed in OpenShift but may not be present in K8s.

The label `test-network-function.com/skip_connectivity_tests` excludes pods from all connectivity tests. The label value is trivial, only its presence.
The label `test-network-function.com/skip_multus_connectivity_tests` excludes pods from [Multus](https://github.com/k8snetworkplumbingwg/multus-cni) connectivity tests. Tests on default interface are still done. The label value is trivial, but its presence.

## AffinityRequired

For CNF workloads that require pods to use Pod or Node Affinity rules, the label `AffinityRequired: true` must be included on the Pod YAML. This will prevent any tests for anti-affinity to fail as well as test your workloads for affinity rules that support your CNF's use-case.

## certifiedcontainerinfo

The `certifiedcontainerinfo` section contains information about CNFs containers that are
to be checked for certification status on Red Hat catalogs.

## Operators

The CSV of the installed Operators can be tested by the `operator` and `affiliated-certification` specs are identified with the `test-network-function.com/operator=target`
label. Any value is permitted here but `target` is used here for consistency with the other specs.

## AllowedProtocolNames

This name of protocols that allowed.
If we want to add another name, we just need to write the name in the yaml file.

for example: if we want to add new protocol - "http4", we add in "tnf_config.yml"  below "validProtocolNames" and then this protocol ("http4") add to map allowedProtocolNames and finally "http4"  will be allow protocol.

## ServicesIgnoreList

This is a list of service names present in the namespace under test and that should not be tested.

## skipScalingTestDeployments and skipScalingTestStatefulSetNames

This section of the TNF config allows the user to skip the scaling tests that potentially cause known problems with workloads that do not like being scaled up and scaled down.

Example:

``` { .yaml .annotate }
skipScalingTestDeployments:
  - name: "deployment1"
    namespace: "tnf"
skipScalingTestStatefulSetNames:
  - name: "statefulset1"
    namespace: "tnf"
```

## debugDaemonSetNamespace

This is an optional field with the name of the namespace where a privileged DaemonSet will be deployed. The namespace will be created in case it does not exist. In case this field is not set, the default namespace for this DaemonSet is "cnf-suite".

```sh
debugDaemonSetNamespace: cnf-cert
```

This DaemonSet, called "tnf-debug" is deployed and used internally by the CNF Certification tool to issue some shell commands that are needed in certain test cases. Some of these test cases might fail or be skipped in case it wasn't deployed correctly.
