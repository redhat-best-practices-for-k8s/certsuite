<!-- markdownlint-disable line-length no-bare-urls -->
# cnf-certification-test catalog

The catalog for cnf-certification-test contains a list of test cases aiming at testing CNF best practices in various areas. Test suites are defined in 10 areas : `platform-alteration`, `access-control`, `affiliated-certification`, `lifecycle`, `manageability`,`networking`, `observability`, `operator`, and `performance.`

Depending on the CNF type, not all tests are required to pass to satisfy best practice requirements. The scenario section indicates which tests are mandatory or optional depending on the scenario. The following CNF types / scenarios are defined: `Telco`, `Non-Telco`, `Far-Edge`, `Extended`.

## Test cases summary

### Total test cases: 89

### Total suites: 9

|Suite|Tests per suite|
|---|---|
|access-control|27|
|affiliated-certification|5|
|lifecycle|18|
|manageability|2|
|networking|11|
|observability|4|
|operator|3|
|performance|6|
|platform-alteration|13|

### Extended specific tests only: 12

|Mandatory|Optional|
|---|---|
|9|3|

### Far-Edge specific tests only: 8

|Mandatory|Optional|
|---|---|
|7|1|

### Non-Telco specific tests only: 42

|Mandatory|Optional|
|---|---|
|39|3|

### Telco specific tests only: 27

|Mandatory|Optional|
|---|---|
|27|0|

## Test Case list

Test Cases are the specifications used to perform a meaningful test. Test cases may run once, or several times against several targets. CNF Certification includes a number of normative and informative tests to ensure CNFs follow best practices. Here is the list of available Test Cases:

### access-control

#### access-control-bpf-capability-check

Property|Description
---|---
Unique ID|access-control-bpf-capability-check
Description|Ensures that containers do not use BFP capability. CNF should avoid loading eBPF filters
Suggested Remediation|Remove the following capability from the container/pod definitions: BPF
Best Practice Reference|No Doc Link - Telco
Exception Process|Exception can be considered. Must identify which container requires the capability and detail why.
Tags|telco,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-cluster-role-bindings

Property|Description
---|---
Unique ID|access-control-cluster-role-bindings
Description|Tests that a Pod does not specify ClusterRoleBindings.
Suggested Remediation|In most cases, Pod's should not have ClusterRoleBindings. The suggested remediation is to remove the need for ClusterRoleBindings, if possible. Cluster roles and cluster role bindings discouraged unless absolutely needed by CNF (often reserved for cluster admin only).
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-security-rbac
Exception Process|Exception possible only for workloads that's cluster wide in nature and absolutely needs cluster level roles & role bindings
Tags|telco,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-container-host-port

Property|Description
---|---
Unique ID|access-control-container-host-port
Description|Verifies if containers define a hostPort.
Suggested Remediation|Remove hostPort configuration from the container. CNF should avoid accessing host resources - containers should not configure HostPort.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-avoid-accessing-resource-on-host
Exception Process|Exception for host resource access tests will only be considered in rare cases where it is absolutely needed
Tags|common,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-crd-roles

Property|Description
---|---
Unique ID|access-control-crd-roles
Description|If an application creates CRDs it must supply a role to access those CRDs and no other API resources/permission. This test checks that there is at least one role present in each namespaces under test that only refers to CRDs under test.
Suggested Remediation|Roles providing access to CRDs should not refer to any other api or resources. Change the generation of the CRD role accordingly
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-custom-role-to-access-application-crds
Exception Process|No exception needed for optional/extended tests.
Tags|extended,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### access-control-ipc-lock-capability-check

Property|Description
---|---
Unique ID|access-control-ipc-lock-capability-check
Description|Ensures that containers do not use IPC_LOCK capability. CNF should avoid accessing host resources - spec.HostIpc should be false.
Suggested Remediation|Exception possible if CNF uses mlock(), mlockall(), shmctl(), mmap(); exception will be considered for DPDK applications. Must identify which container requires the capability and detail why.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-ipc_lock
Exception Process|Exception possible if CNF uses mlock(), mlockall(), shmctl(), mmap(); exception will be considered for DPDK applications. Must identify which container requires the capability and detail why.
Tags|telco,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-namespace

Property|Description
---|---
Unique ID|access-control-namespace
Description|Tests that all CNF's resources (PUTs and CRs) belong to valid namespaces. A valid namespace meets the following conditions: (1) It was declared in the yaml config file under the targetNameSpaces tag. (2) It does not have any of the following prefixes: default, openshift-, istio- and aspenmesh-
Suggested Remediation|Ensure that your CNF utilizes namespaces declared in the yaml config file. Additionally, the namespaces should not start with "default, openshift-, istio- or aspenmesh-".
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-requirements-cnf-reqs
Exception Process|No exceptions
Tags|common,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-namespace-resource-quota

Property|Description
---|---
Unique ID|access-control-namespace-resource-quota
Description|Checks to see if CNF workload pods are running in namespaces that have resource quotas applied.
Suggested Remediation|Apply a ResourceQuota to the namespace your CNF is running in. The CNF namespace should have resource quota defined.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-memory-allocation
Exception Process|No exception needed for optional/extended tests.
Tags|extended,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### access-control-net-admin-capability-check

Property|Description
---|---
Unique ID|access-control-net-admin-capability-check
Description|Ensures that containers do not use NET_ADMIN capability. Note: this test also ensures iptables and nftables are not configured by CNF pods: - NET_ADMIN and NET_RAW are required to modify nftables (namespaced) which is not desired inside pods. nftables should be configured by an administrator outside the scope of the CNF. nftables are usually configured by operators, for instance the Performance Addon Operator (PAO) or istio. - Privileged container are required to modify host iptables, which is not safe to perform inside pods. nftables should be configured by an administrator outside the scope of the CNF. iptables are usually configured by operators, for instance the Performance Addon Operator (PAO) or istio.
Suggested Remediation|Exception possible if CNF uses mlock(), mlockall(), shmctl(), mmap(); exception will be considered for DPDK applications. Must identify which container requires the capability and detail why.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-net_admin
Exception Process|Exception will be considered for user plane or networking functions (e.g. SR-IOV, Multicast). Must identify which container requires the capability and detail why.
Tags|telco,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-net-raw-capability-check

Property|Description
---|---
Unique ID|access-control-net-raw-capability-check
Description|Ensures that containers do not use NET_RAW capability. Note: this test also ensures iptables and nftables are not configured by CNF pods: - NET_ADMIN and NET_RAW are required to modify nftables (namespaced) which is not desired inside pods. nftables should be configured by an administrator outside the scope of the CNF. nftables are usually configured by operators, for instance the Performance Addon Operator (PAO) or istio. - Privileged container are required to modify host iptables, which is not safe to perform inside pods. nftables should be configured by an administrator outside the scope of the CNF. iptables are usually configured by operators, for instance the Performance Addon Operator (PAO) or istio.
Suggested Remediation|Exception possible if CNF uses mlock(), mlockall(), shmctl(), mmap(); exception will be considered for DPDK applications. Must identify which container requires the capability and detail why.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-user-plane-cnfs
Exception Process|Exception will be considered for user plane or networking functions. Must identify which container requires the capability and detail why.
Tags|telco,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-no-1337-uid

Property|Description
---|---
Unique ID|access-control-no-1337-uid
Description|Checks that all pods are not using the securityContext UID 1337
Suggested Remediation|Use another process UID that is not 1337.
Best Practice Reference|No Doc Link - Extended
Exception Process|No exception needed for optional/extended tests.
Tags|extended,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### access-control-one-process-per-container

Property|Description
---|---
Unique ID|access-control-one-process-per-container
Description|Check that all containers under test have only one process running
Suggested Remediation|Launch only one process per container. Should adhere to 1 process per container best practice wherever possible.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-one-process-per-container
Exception Process|No exception needed for optional/extended tests. Not applicable to SNO applications.
Tags|common,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### access-control-pod-automount-service-account-token

Property|Description
---|---
Unique ID|access-control-pod-automount-service-account-token
Description|Check that all pods under test have automountServiceAccountToken set to false. Only pods that require access to the kubernetes API server should have automountServiceAccountToken set to true
Suggested Remediation|Check that pod has automountServiceAccountToken set to false or pod is attached to service account which has automountServiceAccountToken set to false, unless the pod needs access to the kubernetes API server. Pods which do not need API access should set automountServiceAccountToken to false in pod spec.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-automount-services-for-pods
Exception Process|Exception will be considered if container needs to access APIs which OCP does not offer natively. Must document which container requires which API(s) and detail why existing OCP APIs cannot be used.
Tags|telco,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-pod-host-ipc

Property|Description
---|---
Unique ID|access-control-pod-host-ipc
Description|Verifies that the spec.HostIpc parameter is set to false
Suggested Remediation|Set the spec.HostIpc parameter to false in the pod configuration. CNF should avoid accessing host resources - spec.HostIpc should be false.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-cnf-security
Exception Process|Exception for host resource access tests will only be considered in rare cases where it is absolutely needed
Tags|common,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-pod-host-network

Property|Description
---|---
Unique ID|access-control-pod-host-network
Description|Verifies that the spec.HostNetwork parameter is not set (not present)
Suggested Remediation|Set the spec.HostNetwork parameter to false in the pod configuration. CNF should avoid accessing host resources - spec.HostNetwork should be false.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-avoid-the-host-network-namespace
Exception Process|Exception for host resource access tests will only be considered in rare cases where it is absolutely needed
Tags|common,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-pod-host-path

Property|Description
---|---
Unique ID|access-control-pod-host-path
Description|Verifies that the spec.HostPath parameter is not set (not present)
Suggested Remediation|Set the spec.HostPath parameter to false in the pod configuration. CNF should avoid accessing host resources - spec.HostPath should be false.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-cnf-security
Exception Process|Exception for host resource access tests will only be considered in rare cases where it is absolutely needed
Tags|common,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-pod-host-pid

Property|Description
---|---
Unique ID|access-control-pod-host-pid
Description|Verifies that the spec.HostPid parameter is set to false
Suggested Remediation|Set the spec.HostPid parameter to false in the pod configuration. CNF should avoid accessing host resources - spec.HostPid should be false.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-cnf-security
Exception Process|Exception for host resource access tests will only be considered in rare cases where it is absolutely needed
Tags|common,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-pod-role-bindings

Property|Description
---|---
Unique ID|access-control-pod-role-bindings
Description|Ensures that a CNF does not utilize RoleBinding(s) in a non-CNF Namespace.
Suggested Remediation|Ensure the CNF is not configured to use RoleBinding(s) in a non-CNF Namespace. Scope of role must <= scope of creator of role.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-security-rbac
Exception Process|No exceptions
Tags|common,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-pod-service-account

Property|Description
---|---
Unique ID|access-control-pod-service-account
Description|Tests that each CNF Pod utilizes a valid Service Account. Default or empty service account is not valid.
Suggested Remediation|Ensure that the each CNF Pod is configured to use a valid Service Account
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-scc-permissions-for-an-application
Exception Process|No exceptions
Tags|common,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-requests-and-limits

Property|Description
---|---
Unique ID|access-control-requests-and-limits
Description|Check that containers have resource requests and limits specified in their spec.
Suggested Remediation|Add requests and limits to your container spec. See: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-requests/limits
Exception Process|There is no documented exception process for this.
Tags|telco,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-security-context

Property|Description
---|---
Unique ID|access-control-security-context
Description|Checks the security context matches one of the 4 categories
Suggested Remediation|Exception possible if CNF uses mlock(), mlockall(), shmctl(), mmap(); exception will be considered for DPDK applications. Must identify which container requires the capability and document why. If the container had the right configuration of the allowed category from the 4 approved list then the test will pass. The 4 categories are defined in Requirement ID 94118 of the Extended Best Practices guide (private repo)
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-cnf-security
Exception Process|no exception needed for optional/extended test
Tags|extended,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### access-control-security-context-non-root-user-check

Property|Description
---|---
Unique ID|access-control-security-context-non-root-user-check
Description|Checks the security context runAsUser parameter in pods and containers to make sure it is not set to uid root(0). Pods and containers should not run as root (runAsUser is not set to uid0).
Suggested Remediation|Change the pod and containers "runAsUser" uid to something other than root(0)
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-cnf-security
Exception Process|No exceptions - will only be considered under special circumstances. Must identify which container needs access and document why with details.
Tags|common,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-security-context-privilege-escalation

Property|Description
---|---
Unique ID|access-control-security-context-privilege-escalation
Description|Checks if privileged escalation is enabled (AllowPrivilegeEscalation=true).
Suggested Remediation|Configure privilege escalation to false. Privileged escalation should not be allowed (AllowPrivilegeEscalation=false).
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-cnf-security
Exception Process|No exceptions
Tags|common,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-service-type

Property|Description
---|---
Unique ID|access-control-service-type
Description|Tests that each CNF Service does not utilize NodePort(s).
Suggested Remediation|Ensure Services are not configured to use NodePort(s).CNF should avoid accessing host resources - tests that each CNF Service does not utilize NodePort(s).
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-avoid-the-host-network-namespace
Exception Process|Exception for host resource access tests will only be considered in rare cases where it is absolutely needed
Tags|common,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-ssh-daemons

Property|Description
---|---
Unique ID|access-control-ssh-daemons
Description|Check that pods do not run SSH daemons.
Suggested Remediation|Ensure that no SSH daemons are running inside a pod. Pods should not run as SSH Daemons (replicaset or statefulset only).
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-pod-interaction/configuration
Exception Process|No exceptions - special consideration can be given to certain containers which run as utility tool daemon
Tags|telco,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-sys-admin-capability-check

Property|Description
---|---
Unique ID|access-control-sys-admin-capability-check
Description|Ensures that containers do not use SYS_ADMIN capability
Suggested Remediation|Exception possible if CNF uses mlock(), mlockall(), shmctl(), mmap(); exception will be considered for DPDK applications. Must identify which container requires the capability and detail why. Containers should not use the SYS_ADMIN Linux capability.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-avoid-sys_admin
Exception Process|No exceptions
Tags|common,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-sys-nice-realtime-capability

Property|Description
---|---
Unique ID|access-control-sys-nice-realtime-capability
Description|Check that pods running on nodes with realtime kernel enabled have the SYS_NICE capability enabled in their spec. In the case that a CNF is running on a node using the real-time kernel, SYS_NICE will be used to allow DPDK application to switch to SCHED_FIFO.
Suggested Remediation|If pods are scheduled to realtime kernel nodes, they must add SYS_NICE capability to their spec.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-sys_nice
Exception Process|There is no documented exception process for this.
Tags|telco,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-sys-ptrace-capability

Property|Description
---|---
Unique ID|access-control-sys-ptrace-capability
Description|Check that if process namespace sharing is enabled for a Pod then the SYS_PTRACE capability is allowed. This capability is required when using Process Namespace Sharing. This is used when processes from one Container need to be exposed to another Container. For example, to send signals like SIGHUP from a process in a Container to another process in another Container. For more information on these capabilities refer to https://cloud.redhat.com/blog/linux-capabilities-in-openshift and https://kubernetes.io/docs/tasks/configure-pod-container/share-process-namespace/
Suggested Remediation|Allow the SYS_PTRACE capability when enabling process namespace sharing for a Pod
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-sys_ptrace
Exception Process|There is no documented exception process for this.
Tags|telco,access-control
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

### affiliated-certification

#### affiliated-certification-container-is-certified

Property|Description
---|---
Unique ID|affiliated-certification-container-is-certified
Description|Tests whether container images listed in the configuration file have passed the Red Hat Container Certification Program (CCP).
Suggested Remediation|Ensure that your container has passed the Red Hat Container Certification Program (CCP).
Best Practice Reference|https://redhat-connect.gitbook.io/partner-guide-for-red-hat-openshift-and-container/certify-your-application/overview
Exception Process|There is no documented exception process for this.Partner can run CNF Certification test suite before passing other certifications (Container/Operator/HelmChart) but the affiliated certification test cases in CNF Certification test suite must be re-run once the other certifications have been granted.
Tags|common,affiliated-certification
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### affiliated-certification-container-is-certified-digest

Property|Description
---|---
Unique ID|affiliated-certification-container-is-certified-digest
Description|Tests whether container images that are autodiscovered have passed the Red Hat Container Certification Program by their digest(CCP).
Suggested Remediation|Ensure that your container has passed the Red Hat Container Certification Program (CCP).
Best Practice Reference|https://redhat-connect.gitbook.io/partner-guide-for-red-hat-openshift-and-container/certify-your-application/overview
Exception Process|There is no documented exception process for this.Partner can run CNF Certification test suite before passing other certifications (Container/Operator/HelmChart) but the affiliated certification test cases in CNF Certification test suite must be re-run once the other certifications have been granted.
Tags|common,affiliated-certification
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### affiliated-certification-helm-version

Property|Description
---|---
Unique ID|affiliated-certification-helm-version
Description|Test to check if the helm chart is v3
Suggested Remediation|Check Helm Chart is v3 and not v2 which is not supported due to security risks associated with Tiller.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-helm
Exception Process|There is no documented exception process for this.
Tags|common,affiliated-certification
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### affiliated-certification-helmchart-is-certified

Property|Description
---|---
Unique ID|affiliated-certification-helmchart-is-certified
Description|Tests whether helm charts listed in the cluster passed the Red Hat Helm Certification Program.
Suggested Remediation|Ensure that the helm charts under test passed the Red Hat's helm Certification Program (e.g. listed in https://charts.openshift.io/index.yaml).
Best Practice Reference|https://redhat-connect.gitbook.io/partner-guide-for-red-hat-openshift-and-container/certify-your-application/overview
Exception Process|There is no documented exception process for this.Partner can run CNF Certification test suite before passing other certifications (Container/Operator/HelmChart) but the affiliated certification test cases in CNF Certification test suite must be re-run once the other certifications have been granted.
Tags|common,affiliated-certification
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### affiliated-certification-operator-is-certified

Property|Description
---|---
Unique ID|affiliated-certification-operator-is-certified
Description|Tests whether CNF Operators listed in the configuration file have passed the Red Hat Operator Certification Program (OCP).
Suggested Remediation|Ensure that your Operator has passed Red Hat's Operator Certification Program (OCP).
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-cnf-operator-requirements
Exception Process|There is no documented exception process for this.Partner can run CNF Certification test suite before passing other certifications (Container/Operator/HelmChart) but the affiliated certification test cases in CNF Certification test suite must be re-run once the other certifications have been granted.
Tags|common,affiliated-certification
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

### lifecycle

#### lifecycle-affinity-required-pods

Property|Description
---|---
Unique ID|lifecycle-affinity-required-pods
Description|Checks that affinity rules are in place if AffinityRequired: 'true' labels are set on Pods.
Suggested Remediation|Pods which need to be co-located on the same node need Affinity rules. If a pod/statefulset/deployment is required to use affinity rules, please add AffinityRequired: 'true' as a label.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-high-level-cnf-expectations
Exception Process|There is no documented exception process for this.
Tags|telco,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-container-shutdown

Property|Description
---|---
Unique ID|lifecycle-container-shutdown
Description|Ensure that the containers lifecycle preStop management feature is configured. The most basic requirement for the lifecycle management of Pods in OpenShift are the ability to start and stop correctly. There are different ways a pod can stop on an OpenShift cluster. One way is that the pod can remain alive but non-functional. Another way is that the pod can crash and become non-functional. When pods are shut down by the platform they are sent a SIGTERM signal which means that the process in the container should start shutting down, closing connections and stopping all activity. If the pod doesn’t shut down within the default 30 seconds then the platform may send a SIGKILL signal which will stop the pod immediately. This method isn’t as clean and the default time between the SIGTERM and SIGKILL messages can be modified based on the requirements of the application. Containers should respond to SIGTERM/SIGKILL with graceful shutdown.
Suggested Remediation|The preStop can be used to gracefully stop the container and clean resources (e.g., DB connection). For details, see https://www.containiq.com/post/kubernetes-container-lifecycle-events-and-hooks and https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks. All pods must respond to SIGTERM signal and shutdown gracefully with a zero exit code.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-cloud-native-design-best-practices
Exception Process|Identify which pod is not conforming to the process and submit information as to why it cannot use a preStop shutdown specification.
Tags|telco,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-container-startup

Property|Description
---|---
Unique ID|lifecycle-container-startup
Description|Ensure that the containers lifecycle postStart management feature is configured. A container must receive important events from the platform and conform/react to these events properly. For example, a container should catch SIGTERM or SIGKILL from the platform and shutdown as quickly as possible. Other typically important events from the platform are PostStart to initialize before servicing requests and PreStop to release resources cleanly before shutting down.
Suggested Remediation|PostStart is normally used to configure the container, set up dependencies, and record the new creation. You could use this event to check that a required API is available before the container’s main work begins. Kubernetes will not change the container’s state to Running until the PostStart script has executed successfully. For details, see https://www.containiq.com/post/kubernetes-container-lifecycle-events-and-hooks and https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks. PostStart is used to configure container, set up dependencies, record new creation. It can also be used to check that a required API is available before the container’s work begins.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-cloud-native-design-best-practices
Exception Process|Identify which pod is not conforming to the process and submit information as to why it cannot use a postStart startup specification.
Tags|telco,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-cpu-isolation

Property|Description
---|---
Unique ID|lifecycle-cpu-isolation
Description|CPU isolation requires: For each container within the pod, resource requests and limits must be identical. If cpu requests and limits are not identical and in whole units (Guaranteed pods with exclusive cpus), your pods will not be tested for compliance. The runTimeClassName must be specified. Annotations required disabling CPU and IRQ load-balancing.
Suggested Remediation|CPU isolation testing is enabled. Please ensure that all pods adhere to the CPU isolation requirements.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-cpu-isolation
Exception Process|There is no documented exception process for this.
Tags|telco,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-crd-scaling

Property|Description
---|---
Unique ID|lifecycle-crd-scaling
Description|Tests that CNF crd support scale in/out operations. First, the test starts getting the current replicaCount (N) of the crd/s with the Pod Under Test. Then, it executes the scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the crd/s. In case of crd that are managed by HPA the test is changing the min and max value to crd Replica - 1 during scale-in and the original replicaCount again for both min/max during the scale-out stage. Lastly its restoring the original min/max replica of the crd/s
Suggested Remediation|Ensure CNF crd/replica sets can scale in/out successfully.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-high-level-cnf-expectations
Exception Process|There is no documented exception process for this. Not applicable to SNO applications.
Tags|common,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### lifecycle-deployment-scaling

Property|Description
---|---
Unique ID|lifecycle-deployment-scaling
Description|Tests that CNF deployments support scale in/out operations. First, the test starts getting the current replicaCount (N) of the deployment/s with the Pod Under Test. Then, it executes the scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the deployment/s. In case of deployments that are managed by HPA the test is changing the min and max value to deployment Replica - 1 during scale-in and the original replicaCount again for both min/max during the scale-out stage. Lastly its restoring the original min/max replica of the deployment/s
Suggested Remediation|Ensure CNF deployments/replica sets can scale in/out successfully.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-high-level-cnf-expectations
Exception Process|There is no documented exception process for this. Not applicable to SNO applications.
Tags|common,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### lifecycle-image-pull-policy

Property|Description
---|---
Unique ID|lifecycle-image-pull-policy
Description|Ensure that the containers under test are using IfNotPresent as Image Pull Policy. If there is a situation where the container dies and needs to be restarted, the image pull policy becomes important. PullIfNotPresent is recommended so that a loss of image registry access does not prevent the pod from restarting.
Suggested Remediation|Ensure that the containers under test are using IfNotPresent as Image Pull Policy.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-use-imagepullpolicy-if-not-present
Exception Process|There is no documented exception process for this.
Tags|telco,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-liveness-probe

Property|Description
---|---
Unique ID|lifecycle-liveness-probe
Description|Check that all containers under test have liveness probe defined. The most basic requirement for the lifecycle management of Pods in OpenShift are the ability to start and stop correctly. When starting up, health probes like liveness and readiness checks can be put into place to ensure the application is functioning properly.
Suggested Remediation|Add a liveness probe to deployed containers. CNFs shall self-recover from common failures like pod failure, host failure, and network failure. Kubernetes native mechanisms such as health-checks (Liveness, Readiness and Startup Probes) shall be employed at a minimum.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-high-level-cnf-expectations
Exception Process|There is no documented exception process for this.
Tags|telco,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-persistent-volume-reclaim-policy

Property|Description
---|---
Unique ID|lifecycle-persistent-volume-reclaim-policy
Description|Check that the persistent volumes the CNF pods are using have a reclaim policy of delete. Network Functions should clear persistent storage by deleting their PVs when removing their application from a cluster.
Suggested Remediation|Ensure that all persistent volumes are using the reclaim policy: delete
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-csi
Exception Process|There is no documented exception process for this.
Tags|telco,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-pod-high-availability

Property|Description
---|---
Unique ID|lifecycle-pod-high-availability
Description|Ensures that CNF Pods specify podAntiAffinity rules and replica value is set to more than 1.
Suggested Remediation|In high availability cases, Pod podAntiAffinity rule should be specified for pod scheduling and pod replica value is set to more than 1 .
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-high-level-cnf-expectations
Exception Process|There is no documented exception process for this. Not applicable to SNO applications.
Tags|common,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### lifecycle-pod-owner-type

Property|Description
---|---
Unique ID|lifecycle-pod-owner-type
Description|Tests that CNF Pod(s) are deployed as part of a ReplicaSet(s)/StatefulSet(s).
Suggested Remediation|Deploy the CNF using ReplicaSet/StatefulSet.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-no-naked-pods
Exception Process|There is no documented exception process for this. Pods should not be deployed as DaemonSet or naked pods.
Tags|telco,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-pod-recreation

Property|Description
---|---
Unique ID|lifecycle-pod-recreation
Description|Tests that a CNF is configured to support High Availability. First, this test cordons and drains a Node that hosts the CNF Pod. Next, the test ensures that OpenShift can re-instantiate the Pod on another Node, and that the actual replica count matches the desired replica count.
Suggested Remediation|Ensure that CNF Pod(s) utilize a configuration that supports High Availability. Additionally, ensure that there are available Nodes in the OpenShift cluster that can be utilized in the event that a host Node fails.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-upgrade-expectations
Exception Process|No exceptions - workloads should be able to be restarted/recreated.
Tags|common,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### lifecycle-pod-scheduling

Property|Description
---|---
Unique ID|lifecycle-pod-scheduling
Description|Ensures that CNF Pods do not specify nodeSelector or nodeAffinity. In most cases, Pods should allow for instantiation on any underlying Node. CNFs shall not use node selectors nor taints/tolerations to assign pod location.
Suggested Remediation|In most cases, Pod's should not specify their host Nodes through nodeSelector or nodeAffinity. However, there are cases in which CNFs require specialized hardware specific to a particular class of Node.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-high-level-cnf-expectations
Exception Process|Exception will only be considered if application requires specialized hardware. Must specify which container requires special hardware and why.
Tags|telco,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-pod-toleration-bypass

Property|Description
---|---
Unique ID|lifecycle-pod-toleration-bypass
Description|Check that pods do not have NoExecute, PreferNoSchedule, or NoSchedule tolerations that have been modified from the default.
Suggested Remediation|Do not allow pods to bypass the NoExecute, PreferNoSchedule, or NoSchedule tolerations that are default applied by Kubernetes.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-taints-and-tolerations
Exception Process|There is no documented exception process for this.
Tags|telco,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-readiness-probe

Property|Description
---|---
Unique ID|lifecycle-readiness-probe
Description|Check that all containers under test have readiness probe defined. There are different ways a pod can stop on on OpenShift cluster. One way is that the pod can remain alive but non-functional. Another way is that the pod can crash and become non-functional. In the first case, if the administrator has implemented liveness and readiness checks, OpenShift can stop the pod and either restart it on the same node or a different node in the cluster. For the second case, when the application in the pod stops, it should exit with a code and write suitable log entries to help the administrator diagnose what the issue was that caused the problem.
Suggested Remediation|Add a readiness probe to deployed containers
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-high-level-cnf-expectations
Exception Process|There is no documented exception process for this.
Tags|telco,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-startup-probe

Property|Description
---|---
Unique ID|lifecycle-startup-probe
Description|Check that all containers under test have startup probe defined. CNFs shall self-recover from common failures like pod failure, host failure, and network failure. Kubernetes native mechanisms such as health-checks (Liveness, Readiness and Startup Probes) shall be employed at a minimum.
Suggested Remediation|Add a startup probe to deployed containers
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-pod-exit-status
Exception Process|There is no documented exception process for this.
Tags|telco,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-statefulset-scaling

Property|Description
---|---
Unique ID|lifecycle-statefulset-scaling
Description|Tests that CNF statefulsets support scale in/out operations. First, the test starts getting the current replicaCount (N) of the statefulset/s with the Pod Under Test. Then, it executes the scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the statefulset/s. In case of statefulsets that are managed by HPA the test is changing the min and max value to statefulset Replica - 1 during scale-in and the original replicaCount again for both min/max during the scale-out stage. Lastly its restoring the original min/max replica of the statefulset/s
Suggested Remediation|Ensure CNF statefulsets/replica sets can scale in/out successfully.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-high-level-cnf-expectations
Exception Process|There is no documented exception process for this. Not applicable to SNO applications.
Tags|common,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### lifecycle-storage-required-pods

Property|Description
---|---
Unique ID|lifecycle-storage-required-pods
Description|Checks that pods do not place persistent volumes on local storage.
Suggested Remediation|If pod is StatefulSet, make sure servicename is not local-storage (persistent volumes should not be on local storage).
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-local-storage
Exception Process|No exceptions
Tags|common,lifecycle
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

### manageability

#### manageability-container-port-name-format

Property|Description
---|---
Unique ID|manageability-container-port-name-format
Description|Check that the container's ports name follow the naming conventions. Name field in ContainerPort section must be of form `<protocol>[-<suffix>]`. More naming convention requirements may be released in future
Suggested Remediation|Ensure that the container's ports name follow our partner naming conventions
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-requirements-cnf-reqs
Exception Process|No exception needed for optional/extended tests.
Tags|extended,manageability
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### manageability-containers-image-tag

Property|Description
---|---
Unique ID|manageability-containers-image-tag
Description|Check that image tag exists on containers.
Suggested Remediation|Ensure that all the container images are tagged. Checks containers have image tags (e.g. latest, stable, dev).
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-image-tagging
Exception Process|No exception needed for optional/extended tests.
Tags|extended,manageability
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

### networking

#### networking-dpdk-cpu-pinning-exec-probe

Property|Description
---|---
Unique ID|networking-dpdk-cpu-pinning-exec-probe
Description|If a CNF is doing CPU pinning, exec probes may not be used.
Suggested Remediation|If the CNF is doing CPU pinning and running a DPDK process do not use exec probes (executing a command within the container) as it may pile up and block the node eventually.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-cpu-manager-pinning
Exception Process|There is no documented exception process for this.
Tags|telco,networking
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### networking-dual-stack-service

Property|Description
---|---
Unique ID|networking-dual-stack-service
Description|Checks that all services in namespaces under test are either ipv6 single stack or dual stack. This test case requires the deployment of the debug daemonset.
Suggested Remediation|Configure every CNF services with either a single stack ipv6 or dual stack (ipv4/ipv6) load balancer.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-ipv4-&-ipv6
Exception Process|No exception needed for optional/extended tests.
Tags|extended,networking
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### networking-icmpv4-connectivity

Property|Description
---|---
Unique ID|networking-icmpv4-connectivity
Description|Checks that each CNF Container is able to communicate via ICMPv4 on the Default OpenShift network. This test case requires the Deployment of the debug daemonset and at least 2 pods connected to each network under test(one source and one destination). If no network with more than 2 pods exists this test will be skipped.
Suggested Remediation|Ensure that the CNF is able to communicate via the Default OpenShift network. In some rare cases, CNFs may require routing table changes in order to communicate over the Default network. To exclude a particular pod from ICMPv4 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it. The label value is trivial, only its presence.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-ipv4-&-ipv6
Exception Process|No exceptions - must be able to communicate on default network using IPv4
Tags|common,networking
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### networking-icmpv4-connectivity-multus

Property|Description
---|---
Unique ID|networking-icmpv4-connectivity-multus
Description|Checks that each CNF Container is able to communicate via ICMPv4 on the Multus network(s). This test case requires the Deployment of the debug daemonset and at least 2 pods connected to each network under test(one source and one destination). If no network with more than 2 pods exists this test will be skipped.
Suggested Remediation|Ensure that the CNF is able to communicate via the Multus network(s). In some rare cases, CNFs may require routing table changes in order to communicate over the Multus network(s). To exclude a particular pod from ICMPv4 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it. The label value is trivial, only its presence. Not applicable if MULTUS is not supported.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-high-level-cnf-expectations
Exception Process|There is no documented exception process for this.
Tags|telco,networking
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### networking-icmpv6-connectivity

Property|Description
---|---
Unique ID|networking-icmpv6-connectivity
Description|Checks that each CNF Container is able to communicate via ICMPv6 on the Default OpenShift network. This test case requires the Deployment of the debug daemonset and at least 2 pods connected to each network under test(one source and one destination). If no network with more than 2 pods exists this test will be skipped.
Suggested Remediation|Ensure that the CNF is able to communicate via the Default OpenShift network. In some rare cases, CNFs may require routing table changes in order to communicate over the Default network. To exclude a particular pod from ICMPv6 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it. The label value is trivial, only its presence. Not applicable if IPv6 is not supported.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-ipv4-&-ipv6
Exception Process|There is no documented exception process for this.
Tags|common,networking
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### networking-icmpv6-connectivity-multus

Property|Description
---|---
Unique ID|networking-icmpv6-connectivity-multus
Description|Checks that each CNF Container is able to communicate via ICMPv6 on the Multus network(s). This test case requires the Deployment of the debug daemonset and at least 2 pods connected to each network under test(one source and one destination). If no network with more than 2 pods exists this test will be skipped.
Suggested Remediation|Ensure that the CNF is able to communicate via the Multus network(s). In some rare cases, CNFs may require routing table changes in order to communicate over the Multus network(s). To exclude a particular pod from ICMPv6 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it.The label value is trivial, only its presence. Not applicable if IPv6/MULTUS is not supported.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-high-level-cnf-expectations
Exception Process|There is no documented exception process for this.
Tags|telco,networking
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### networking-network-policy-deny-all

Property|Description
---|---
Unique ID|networking-network-policy-deny-all
Description|Check that network policies attached to namespaces running CNF pods contain a default deny-all rule for both ingress and egress traffic
Suggested Remediation|Ensure that a NetworkPolicy with a default deny-all is applied. After the default is applied, apply a network policy to allow the traffic your application requires.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-vrfs-aka-routing-instances
Exception Process|No exception needed for optional/extended tests.
Tags|common,networking
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### networking-ocp-reserved-ports-usage

Property|Description
---|---
Unique ID|networking-ocp-reserved-ports-usage
Description|Check that containers do not listen on ports that are reserved by OpenShift
Suggested Remediation|Ensure that CNF apps do not listen on ports that are reserved by OpenShift. The following ports are reserved by OpenShift and must NOT be used by any application: 22623, 22624.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-ports-reserved-by-openshift
Exception Process|No exceptions
Tags|common,networking
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### networking-reserved-partner-ports

Property|Description
---|---
Unique ID|networking-reserved-partner-ports
Description|Checks that pods and containers are not consuming ports designated as reserved by partner
Suggested Remediation|Ensure ports are not being used that are reserved by our partner
Best Practice Reference|No Doc Link - Extended
Exception Process|No exception needed for optional/extended tests.
Tags|extended,networking
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### networking-restart-on-reboot-sriov-pod

Property|Description
---|---
Unique ID|networking-restart-on-reboot-sriov-pod
Description|Ensures that the label restart-on-reboot exists on pods that use SRIOV network interfaces.
Suggested Remediation|Ensure that the label restart-on-reboot exists on pods that use SRIOV network interfaces.
Best Practice Reference|No Doc Link - Far Edge
Exception Process|There is no documented exception process for this.
Tags|faredge,networking
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Optional|

#### networking-undeclared-container-ports-usage

Property|Description
---|---
Unique ID|networking-undeclared-container-ports-usage
Description|Check that containers do not listen on ports that weren't declared in their specification. Platforms may be configured to block undeclared ports.
Suggested Remediation|Ensure the CNF apps do not listen on undeclared containers' ports.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-requirements-cnf-reqs
Exception Process|No exception needed for optional/extended tests.
Tags|extended,networking
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

### observability

#### observability-container-logging

Property|Description
---|---
Unique ID|observability-container-logging
Description|Check that all containers under test use standard input output and standard error when logging. A container must provide APIs for the platform to observe the container health and act accordingly. These APIs include health checks (liveness and readiness), logging to stderr and stdout for log aggregation (by tools such as Logstash or Filebeat), and integrate with tracing and metrics-gathering libraries (such as Prometheus or Metricbeat).
Suggested Remediation|Ensure containers are not redirecting stdout/stderr
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-logging
Exception Process|There is no documented exception process for this.
Tags|telco,observability
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### observability-crd-status

Property|Description
---|---
Unique ID|observability-crd-status
Description|Checks that all CRDs have a status sub-resource specification (Spec.versions[].Schema.OpenAPIV3Schema.Properties[“status”]).
Suggested Remediation|Ensure that all the CRDs have a meaningful status specification (Spec.versions[].Schema.OpenAPIV3Schema.Properties[“status”]).
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-cnf-operator-requirements
Exception Process|No exceptions
Tags|common,observability
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### observability-pod-disruption-budget

Property|Description
---|---
Unique ID|observability-pod-disruption-budget
Description|Checks to see if pod disruption budgets have allowed values for minAvailable and maxUnavailable
Suggested Remediation|Ensure minAvailable is not zero and maxUnavailable does not equal the number of pods in the replica
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-upgrade-expectations
Exception Process|No exceptions
Tags|common,observability
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### observability-termination-policy

Property|Description
---|---
Unique ID|observability-termination-policy
Description|Check that all containers are using terminationMessagePolicy: FallbackToLogsOnError. There are different ways a pod can stop on an OpenShift cluster. One way is that the pod can remain alive but non-functional. Another way is that the pod can crash and become non-functional. In the first case, if the administrator has implemented liveness and readiness checks, OpenShift can stop the pod and either restart it on the same node or a different node in the cluster. For the second case, when the application in the pod stops, it should exit with a code and write suitable log entries to help the administrator diagnose what the issue was that caused the problem.
Suggested Remediation|Ensure containers are all using FallbackToLogsOnError in terminationMessagePolicy
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-pod-exit-status
Exception Process|There is no documented exception process for this.
Tags|telco,observability
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

### operator

#### operator-install-source

Property|Description
---|---
Unique ID|operator-install-source
Description|Tests whether a CNF Operator is installed via OLM.
Suggested Remediation|Ensure that your Operator is installed via OLM.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-cnf-operator-requirements
Exception Process|No exceptions
Tags|common,operator
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### operator-install-status-no-privileges

Property|Description
---|---
Unique ID|operator-install-status-no-privileges
Description|The operator is not installed with privileged rights. Test passes if clusterPermissions is not present in the CSV manifest or is present with no resourceNames under its rules.
Suggested Remediation|Ensure all the CNF operators have no privileges on cluster resources.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-cnf-operator-requirements
Exception Process|No exceptions
Tags|common,operator
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### operator-install-status-succeeded

Property|Description
---|---
Unique ID|operator-install-status-succeeded
Description|Ensures that the target CNF operators report "Succeeded" as their installation status.
Suggested Remediation|Ensure all the CNF operators have been successfully installed by OLM.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-cnf-operator-requirements
Exception Process|No exceptions
Tags|common,operator
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

### performance

#### performance-exclusive-cpu-pool

Property|Description
---|---
Unique ID|performance-exclusive-cpu-pool
Description|Ensures that if one container in a Pod selects an exclusive CPU pool the rest select the same type of CPU pool
Suggested Remediation|Ensure that if one container in a Pod selects an exclusive CPU pool the rest also select this type of CPU pool
Best Practice Reference|No Doc Link - Far Edge
Exception Process|There is no documented exception process for this.
Tags|faredge,performance
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Optional|

#### performance-exclusive-cpu-pool-rt-scheduling-policy

Property|Description
---|---
Unique ID|performance-exclusive-cpu-pool-rt-scheduling-policy
Description|Ensures that if application workload runs in exclusive CPU pool, it chooses RT CPU schedule policy and set the priority less than 10.
Suggested Remediation|Ensure that the workload running in Application exclusive CPU pool can choose RT CPU scheduling policy, but should set priority less than 10
Best Practice Reference|No Doc Link - Far Edge
Exception Process|There is no documented exception process for this.
Tags|faredge,performance
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Optional|

#### performance-isolated-cpu-pool-rt-scheduling-policy

Property|Description
---|---
Unique ID|performance-isolated-cpu-pool-rt-scheduling-policy
Description|Ensures that a workload running in an application-isolated exclusive CPU pool selects a RT CPU scheduling policy
Suggested Remediation|Ensure that the workload running in an application-isolated exclusive CPU pool selects a RT CPU scheduling policy (such as SCHED_FIFO/SCHED_RR) with High priority.
Best Practice Reference|No Doc Link - Far Edge
Exception Process|There is no documented exception process for this.
Tags|faredge,performance
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Optional|

#### performance-max-resources-exec-probes

Property|Description
---|---
Unique ID|performance-max-resources-exec-probes
Description|Checks that less than 10 exec probes are configured in the cluster for this CNF. Also checks that the periodSeconds parameter for each probe is superior or equal to 10.
Suggested Remediation|Reduce the number of exec probes in the cluster for this CNF to less than 10. Increase the update period of the exec probe to be superior or equal to 10 seconds.
Best Practice Reference|No Doc Link - Far Edge
Exception Process|There is no documented exception process for this.
Tags|faredge,performance
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### performance-rt-apps-no-exec-probes

Property|Description
---|---
Unique ID|performance-rt-apps-no-exec-probes
Description|Ensures that if one container runs a real time application exec probes are not used
Suggested Remediation|Ensure that if one container runs a real time application exec probes are not used
Best Practice Reference|No Doc Link - Far Edge
Exception Process|There is no documented exception process for this.
Tags|faredge,performance
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Optional|

#### performance-shared-cpu-pool-non-rt-scheduling-policy

Property|Description
---|---
Unique ID|performance-shared-cpu-pool-non-rt-scheduling-policy
Description|Ensures that if application workload runs in shared CPU pool, it chooses non-RT CPU schedule policy to always share the CPU with other applications and kernel threads.
Suggested Remediation|Ensure that the workload running in Application shared CPU pool should choose non-RT CPU schedule policy, like SCHED _OTHER to always share the CPU with other applications and kernel threads.
Best Practice Reference|No Doc Link - Far Edge
Exception Process|There is no documented exception process for this.
Tags|faredge,performance
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Optional|

### platform-alteration

#### platform-alteration-base-image

Property|Description
---|---
Unique ID|platform-alteration-base-image
Description|Ensures that the Container Base Image is not altered post-startup. This test is a heuristic, and ensures that there are no changes to the following directories: 1) /var/lib/rpm 2) /var/lib/dpkg 3) /bin 4) /sbin 5) /lib 6) /lib64 7) /usr/bin 8) /usr/sbin 9) /usr/lib 10) /usr/lib64
Suggested Remediation|Ensure that Container applications do not modify the Container Base Image. In particular, ensure that the following directories are not modified: 1) /var/lib/rpm 2) /var/lib/dpkg 3) /bin 4) /sbin 5) /lib 6) /lib64 7) /usr/bin 8) /usr/sbin 9) /usr/lib 10) /usr/lib64 Ensure that all required binaries are built directly into the container image, and are not installed post startup.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-image-standards
Exception Process|No exceptions
Tags|common,platform-alteration
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### platform-alteration-boot-params

Property|Description
---|---
Unique ID|platform-alteration-boot-params
Description|Tests that boot parameters are set through the MachineConfigOperator, and not set manually on the Node.
Suggested Remediation|Ensure that boot parameters are set directly through the MachineConfigOperator, or indirectly through the PerformanceAddonOperator. Boot parameters should not be changed directly through the Node, as OpenShift should manage the changes for you.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-host-os
Exception Process|No exceptions
Tags|common,platform-alteration
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### platform-alteration-hugepages-1g-only

Property|Description
---|---
Unique ID|platform-alteration-hugepages-1g-only
Description|Check that pods using hugepages only use 1Gi size
Suggested Remediation|Modify pod to consume 1Gi hugepages only
Best Practice Reference|No Doc Link - Far Edge
Exception Process|There is no documented exception process for this.
Tags|faredge,platform-alteration
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Optional|

#### platform-alteration-hugepages-2m-only

Property|Description
---|---
Unique ID|platform-alteration-hugepages-2m-only
Description|Check that pods using hugepages only use 2Mi size
Suggested Remediation|Modify pod to consume 2Mi hugepages only
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-huge-pages
Exception Process|No exception needed for optional/extended tests.
Tags|extended,platform-alteration
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### platform-alteration-hugepages-config

Property|Description
---|---
Unique ID|platform-alteration-hugepages-config
Description|Checks to see that HugePage settings have been configured through MachineConfig, and not manually on the underlying Node. This test case applies only to Nodes that are configured with the "worker" MachineConfigSet. First, the "worker" MachineConfig is polled, and the Hugepage settings are extracted. Next, the underlying Nodes are polled for configured HugePages through inspection of /proc/meminfo. The results are compared, and the test passes only if they are the same.
Suggested Remediation|HugePage settings should be configured either directly through the MachineConfigOperator or indirectly using the PerformanceAddonOperator. This ensures that OpenShift is aware of the special MachineConfig requirements, and can provision your CNF on a Node that is part of the corresponding MachineConfigSet. Avoid making changes directly to an underlying Node, and let OpenShift handle the heavy lifting of configuring advanced settings. This test case applies only to Nodes that are configured with the "worker" MachineConfigSet.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-huge-pages
Exception Process|No exceptions
Tags|common,platform-alteration
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### platform-alteration-hyperthread-enable

Property|Description
---|---
Unique ID|platform-alteration-hyperthread-enable
Description|Check that baremetal workers have hyperthreading enabled
Suggested Remediation|Check that baremetal workers have hyperthreading enabled
Best Practice Reference|No Doc Link - Extended
Exception Process|There is no documented exception process for this.
Tags|extended,platform-alteration
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### platform-alteration-is-selinux-enforcing

Property|Description
---|---
Unique ID|platform-alteration-is-selinux-enforcing
Description|verifies that all openshift platform/cluster nodes have selinux in "Enforcing" mode.
Suggested Remediation|Configure selinux and enable enforcing mode.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-pod-security
Exception Process|No exceptions
Tags|common,platform-alteration
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### platform-alteration-isredhat-release

Property|Description
---|---
Unique ID|platform-alteration-isredhat-release
Description|verifies if the container base image is redhat.
Suggested Remediation|Build a new container image that is based on UBI (Red Hat Universal Base Image).
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-base-images
Exception Process|No exceptions
Tags|common,platform-alteration
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### platform-alteration-ocp-lifecycle

Property|Description
---|---
Unique ID|platform-alteration-ocp-lifecycle
Description|Tests that the running OCP version is not end of life.
Suggested Remediation|Please update your cluster to a version that is generally available.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-k8s
Exception Process|No exceptions
Tags|common,platform-alteration
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### platform-alteration-ocp-node-os-lifecycle

Property|Description
---|---
Unique ID|platform-alteration-ocp-node-os-lifecycle
Description|Tests that the nodes running in the cluster have operating systems that are compatible with the deployed version of OpenShift.
Suggested Remediation|Please update your workers to a version that is supported by your version of OpenShift
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-host-os
Exception Process|No exceptions
Tags|common,platform-alteration
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### platform-alteration-service-mesh-usage

Property|Description
---|---
Unique ID|platform-alteration-service-mesh-usage
Description|Checks if the istio namespace ("istio-system") is present. If it is present, checks that the istio sidecar is present in all pods under test.
Suggested Remediation|Ensure all the CNF pods are using service mesh if the cluster provides it.
Best Practice Reference|No Doc Link - Extended
Exception Process|No exception needed for optional/extended tests.
Tags|extended,platform-alteration
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### platform-alteration-sysctl-config

Property|Description
---|---
Unique ID|platform-alteration-sysctl-config
Description|Tests that no one has changed the node's sysctl configs after the node was created, the tests works by checking if the sysctl configs are consistent with the MachineConfig CR which defines how the node should be configured
Suggested Remediation|You should recreate the node or change the sysctls, recreating is recommended because there might be other unknown changes
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-cnf-security
Exception Process|No exceptions
Tags|common,platform-alteration
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### platform-alteration-tainted-node-kernel

Property|Description
---|---
Unique ID|platform-alteration-tainted-node-kernel
Description|Ensures that the Node(s) hosting CNFs do not utilize tainted kernels. This test case is especially important to support Highly Available CNFs, since when a CNF is re-instantiated on a backup Node, that Node's kernel may not have the same hacks.'
Suggested Remediation|Test failure indicates that the underlying Node's kernel is tainted. Ensure that you have not altered underlying Node(s) kernels in order to run the CNF.
Best Practice Reference|https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-high-level-cnf-expectations
Exception Process|If taint is necessary, document details of the taint and why it's needed by workload or environment.
Tags|common,platform-alteration
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|
