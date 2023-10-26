<!-- markdownlint-disable code-block-style line-length no-bare-urls -->
# CNF Certification configuration

The CNF Certification Test uses a YAML configuration file to certify a specific CNF workload. This file specifies the CNF resources to be certified, as well as any exceptions or other general configuration options.

By default a file named _tnf_config.yml_ will be used. Here's an [example](https://github.com/test-network-function/cnf-certification-test/blob/main/cnf-certification-test/tnf_config.yml) of the CNF Config File. For a description of each config option see the section [CNF Config File options](#cnf-config-file-options).

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
<img src="../assets/images/demo-config.svg">
</object>
<!-- markdownlint-enable MD033 -->

## CNF Config File options

### CNF resources

These options allow configuring the workload resources of the CNF to be verified. Only the resources that the CNF uses are required to be configured. The rest can be left empty. Usually a basic configuration includes _Namespaces_ and _Pods_ at least.

#### targetNameSpaces

The namespaces in which the CNF under test will be deployed.

``` { .yaml .annotate }
targetNameSpaces:
  - name: tnf
```

#### podsUnderTestLabels

The labels that each Pod of the CNF under test must have to be verified by the CNF Certification Suite.

!!! note "Highly recommended"

    The labels should be defined in Pod definition rather than added after the Pod is created, as labels added later on will be lost in case the Pod gets rescheduled. In the case of Pods defined as part of a Deployment, it's best to use the same label as the one defined in the _spec.selector.matchLabels_ section of the Deployment YAML. The prefix field can be used to avoid naming collision with other labels.

``` { .yaml .annotate }
podsUnderTestLabels:
  - "test-network-function.com/generic: target"
```

#### operatorsUnderTestLabels

The labels that each operator's CSV of the CNF under test must have to be verified by the CNF Certification Suite.

If a new label is used for this purpose make sure it is added to the CNF operator's CSVs.

``` { .yaml .annotate }
operatorsUnderTestLabels:
  - "test-network-function.com/operator: target" 
```

#### targetCrdFilters

The CRD name suffix used to filter the CNF's CRDs among all the CRDs present in the cluster. For each CRD it can also be specified if it's scalable or not in order to avoid some lifecycle test cases.

``` { .yaml .annotate }
targetCrdFilters:
 - nameSuffix: "group1.tnf.com"
   scalable: false
 - nameSuffix: "anydomain.com"
   scalable: true
```

With the config show above, all CRD names in the cluster whose names have the suffix _group1.tnf.com_ or _anydomain.com_ ( e.g. _crd1.group1.tnf.com_ or _mycrd.mygroup.anydomain.com_) will be tested.

#### managedDeployments / managedStatefulSets

The Deployments/StatefulSets managed by a Custom Resource whose scaling is controlled using the "scale" subresource of the CR.

The CRD defining that CR should be included in the CRD filters with the scalable property set to true. If so, the test case _lifecycle-{deployment/statefulset}-scaling_ will be skipped, otherwise it will fail.

``` { .yaml .annotate }
managedDeployments:
  - name: jack
managedStatefulsets:
  - name: jack
```

### Exceptions

These options allow adding exceptions to skip several checks for different resources. The exceptions must be justified in order to pass the CNF Certification.

#### acceptedKernelTaints

The list of kernel modules loaded by the CNF that make the Linux kernel mark itself as _tainted_ but that should skip verification.

Test cases affected: _platform-alteration-tainted-node-kernel_.

``` { .yaml .annotate }
acceptedKernelTaints:
  - module: vboxsf
  - module: vboxguest
```

#### skipHelmChartList

The list of Helm charts that the CNF uses whose certification status will not be verified.

If no exception is configured, the certification status for all Helm charts will be checked in the [OpenShift Helms Charts repository](https://charts.openshift.io/).

Test cases affected: _affiliated-certification-helmchart-is-certified_.

``` { .yaml .annotate }
skipHelmChartList:
  - name: coredns
```

#### validProtocolNames

The list of allowed protocol names to be used for container port names.

The name field of a container port must be of the form _protocol[-suffix]_ where _protocol_ must be allowed by default or added to this list. The optional _suffix_ can be chosen by the application. Protocol names allowed by default: _grpc_, _grpc-web_, _http_, _http2_, _tcp_, _udp_.

Test cases affected: _manageability-container-port-name-format_.

``` { .yaml .annotate }
validProtocolNames:
  - "http3"
  - "sctp"
```

#### servicesIgnoreList

The list of Services that will skip verification.

Services included in this list will be filtered out at the autodiscovery stage and will not be subject to checks in any test case.

Tests cases affected: _networking-dual-stack-service_, _access-control-service-type_.

``` { .yaml .annotate }
servicesignorelist:
  - "hazelcast-platform-controller-manager-service"
  - "hazelcast-platform-webhook-service"
  - "new-pro-controller-manager-metrics-service"
```

#### skipScalingTestDeployments / skipScalingTestStatefulSets

The list of Deployments/StatefulSets that do not support scale in/out operations.

Deployments/StatefulSets included in this list will skip any scaling operation check.

Test cases affected: _lifecycle-deployment-scaling_, _lifecycle-statefulset-scaling_.

``` { .yaml .annotate }
skipScalingTestDeployments:
  - name: deployment1
    namespace: tnf
skipScalingTestStatefulSetNames:
  - name: statefulset1
    namespace: tnf
```

### CNF Certification settings

#### debugDaemonSetNamespace

This is an optional field with the name of the namespace where a privileged DaemonSet will be deployed. The namespace will be created in case it does not exist. In case this field is not set, the default namespace for this DaemonSet is _cnf-suite_.

``` { .yaml .annotate }
debugDaemonSetNamespace: cnf-cert
```

This DaemonSet, called _tnf-debug_ is deployed and used internally by the CNF Certification tool to issue some shell commands that are needed in certain test cases. Some of these test cases might fail or be skipped in case it wasn't deployed correctly.

### Other settings

The autodiscovery mechanism will attempt to identify the default network device and all the IP addresses of the Pods it needs for network connectivity tests, though that information can be explicitly set using annotations if needed.

#### Pod IPs

- The _k8s.v1.cni.cncf.io/networks-status_ annotation is checked and all IPs from it are used. This annotation is automatically managed in OpenShift but may not be present in K8s.
- If it is not present, then only known IPs associated with the Pod are used (the Pod _.status.ips_ field).

#### Network Interfaces

- The _k8s.v1.cni.cncf.io/networks-status_ annotation is checked and the _interface_ from the first entry found with _“default”=true_ is used. This annotation is automatically managed in OpenShift but may not be present in K8s.

The label _test-network-function.com/skip_connectivity_tests_ excludes Pods from all connectivity tests.

The label _test-network-function.com/skip_multus_connectivity_tests_ excludes Pods from [Multus](https://github.com/k8snetworkplumbingwg/multus-cni) connectivity tests. Tests on the default interface are still run.

#### Affinity requirements

For CNF workloads that require Pods to use Pod or Node Affinity rules, the label _AffinityRequired: true_ must be included on the Pod YAML. This will ensure that the affinity best practices are tested and prevent any test cases for anti-affinity to fail.
