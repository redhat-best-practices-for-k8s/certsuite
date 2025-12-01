<!-- markdownlint-disable line-length no-bare-urls blanks-around-lists ul-indent blanks-around-headings no-trailing-spaces -->
# Red Hat Best Practices Test Suite for Kubernetes catalog

The catalog for the Red Hat Best Practices Test Suite for Kubernetes contains a list of test cases aiming at testing best practices in various areas. Test suites are defined in 10 areas : `platform-alteration`, `access-control`, `affiliated-certification`, `lifecycle`, `manageability`,`networking`, `observability`, `operator`, and `performance.`

Depending on the workload type, not all tests are required to pass to satisfy best practice requirements. The scenario section indicates which tests are mandatory or optional depending on the scenario. The following workload types / scenarios are defined: `Telco`, `Non-Telco`, `Far-Edge`, `Extended`.

## Test cases summary

### Total test cases: 121

### Total suites: 10

|Suite|Tests per suite|Link|
|---|---|---|
|access-control|28|[access-control](#access-control)|
|affiliated-certification|4|[affiliated-certification](#affiliated-certification)|
|lifecycle|19|[lifecycle](#lifecycle)|
|manageability|2|[manageability](#manageability)|
|networking|11|[networking](#networking)|
|observability|5|[observability](#observability)|
|operator|12|[operator](#operator)|
|performance|7|[performance](#performance)|
|platform-alteration|14|[platform-alteration](#platform-alteration)|
|preflight|19|[preflight](#preflight)|

### Extended specific tests only: 13

|Mandatory|Optional|
|---|---|---|
|10|3|

### Far-Edge specific tests only: 9

|Mandatory|Optional|
|---|---|---|
|8|1|

### Non-Telco specific tests only: 71

|Mandatory|Optional|
|---|---|---|
|43|28|

### Telco specific tests only: 28

|Mandatory|Optional|
|---|---|---|
|27|1|

## Test Case list

Test Cases are the specifications used to perform a meaningful test. Test cases may run once, or several times against several targets. The Red Hat Best Practices Test Suite for Kubernetes includes a number of normative and informative tests to ensure that workloads follow best practices. Here is the list of available Test Cases:

### access-control

#### access-control-bpf-capability-check

|Property|Description|
|---|---|
|Unique ID|access-control-bpf-capability-check|
|Description|Ensures that containers do not use BPF capability. Workloads should avoid loading eBPF filters|
|Suggested Remediation|Remove the following capability from the container/pod definitions: BPF|
|Best Practice Reference|No Doc Link - Telco|
|Exception Process|Exception can be considered. Must identify which container requires the capability and detail why.|
|Impact Statement|BPF capability allows kernel-level programming that can bypass security controls, monitor other processes, and potentially compromise the entire host system.|
|Tags|telco,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-cluster-role-bindings

|Property|Description|
|---|---|
|Unique ID|access-control-cluster-role-bindings|
|Description|Tests that a Pod does not specify ClusterRoleBindings.|
|Suggested Remediation|In most cases, Pod's should not have ClusterRoleBindings. The suggested remediation is to remove the need for ClusterRoleBindings, if possible. Cluster roles and cluster role bindings discouraged unless absolutely needed by the workload (often reserved for cluster admin only).|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-security-and-role-based-access-control|
|Exception Process|Exception possible only for workloads that's cluster wide in nature and absolutely needs cluster level roles & role bindings|
|Impact Statement|Cluster-wide role bindings grant excessive privileges that can be exploited for lateral movement and privilege escalation across the entire cluster.|
|Tags|telco,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-container-host-port

|Property|Description|
|---|---|
|Unique ID|access-control-container-host-port|
|Description|Verifies if containers define a hostPort.|
|Suggested Remediation|Remove hostPort configuration from the container. Workloads should avoid accessing host resources - containers should not configure HostPort.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-avoid-accessing-resource-on-host|
|Exception Process|Exception for host resource access tests will only be considered in rare cases where it is absolutely needed|
|Impact Statement|Host port usage can create port conflicts with host services and expose containers directly to the host network, bypassing network security controls.|
|Tags|common,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-crd-roles

|Property|Description|
|---|---|
|Unique ID|access-control-crd-roles|
|Description|If an application creates CRDs it must supply a role to access those CRDs and no other API resources/permission. This test checks that there is at least one role present in each namespaces under test that only refers to CRDs under test.|
|Suggested Remediation|Roles providing access to CRDs should not refer to any other api or resources. Change the generation of the CRD role accordingly|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-custom-role-to-access-application-crds|
|Exception Process|No exception needed for optional/extended tests.|
|Impact Statement|Improper CRD role configurations can grant excessive privileges, violate least-privilege principles, and create security vulnerabilities in custom resource access control.|
|Tags|extended,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### access-control-ipc-lock-capability-check

|Property|Description|
|---|---|
|Unique ID|access-control-ipc-lock-capability-check|
|Description|Ensures that containers do not use IPC_LOCK capability. Workloads should avoid accessing host resources - spec.HostIpc should be false.|
|Suggested Remediation|Exception possible if a workload uses mlock(), mlockall(), shmctl(), mmap(); exception will be considered for DPDK applications. Must identify which container requires the capability and detail why.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-ipc_lock|
|Exception Process|Exception possible if a workload uses mlock(), mlockall(), shmctl(), mmap(); exception will be considered for DPDK applications. Must identify which container requires the capability and detail why.|
|Impact Statement|IPC_LOCK capability can be exploited to lock system memory, potentially causing denial of service and affecting other workloads on the same node.|
|Tags|telco,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-namespace

|Property|Description|
|---|---|
|Unique ID|access-control-namespace|
|Description|Tests that all workload resources (PUTs and CRs) belong to valid namespaces. A valid namespace meets the following conditions: (1) It was declared in the yaml config file under the targetNameSpaces tag. (2) It does not have any of the following prefixes: default, openshift-, istio- and aspenmesh-|
|Suggested Remediation|Ensure that your workload utilizes namespaces declared in the yaml config file. Additionally, the namespaces should not start with "default, openshift-, istio- or aspenmesh-".|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-requirements-cnf-reqs|
|Exception Process|No exceptions|
|Impact Statement|Using inappropriate namespaces can lead to resource conflicts, security boundary violations, and administrative complexity in multi-tenant environments.|
|Tags|common,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### access-control-namespace-resource-quota

|Property|Description|
|---|---|
|Unique ID|access-control-namespace-resource-quota|
|Description|Checks to see if workload pods are running in namespaces that have resource quotas applied.|
|Suggested Remediation|Apply a ResourceQuota to the namespace your workload is running in. The workload's namespace should have resource quota defined.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-memory-allocation|
|Exception Process|No exception needed for optional/extended tests.|
|Impact Statement|Without resource quotas, workloads can consume excessive cluster resources, causing performance issues and potential denial of service for other applications.|
|Tags|extended,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### access-control-net-admin-capability-check

|Property|Description|
|---|---|
|Unique ID|access-control-net-admin-capability-check|
|Description|Ensures that containers do not use NET_ADMIN capability. Note: this test also ensures iptables and nftables are not configured by workload pods: - NET_ADMIN and NET_RAW are required to modify nftables (namespaced) which is not desired inside pods. nftables should be configured by an administrator outside the scope of the workload. nftables are usually configured by operators, for instance the Performance Addon Operator (PAO) or istio. - Privileged container are required to modify host iptables, which is not safe to perform inside pods. nftables should be configured by an administrator outside the scope of the workload. iptables are usually configured by operators, for instance the Performance Addon Operator (PAO) or istio.|
|Suggested Remediation|Exception possible if a workload uses mlock(), mlockall(), shmctl(), mmap(); exception will be considered for DPDK applications. Must identify which container requires the capability and detail why.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-net_admin|
|Exception Process|Exception will be considered for user plane or networking functions (e.g. SR-IOV, Multicast). Must identify which container requires the capability and detail why.|
|Impact Statement|NET_ADMIN capability allows network configuration changes that can compromise cluster networking, enable privilege escalation, and bypass network security controls.|
|Tags|telco,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-net-raw-capability-check

|Property|Description|
|---|---|
|Unique ID|access-control-net-raw-capability-check|
|Description|Ensures that containers do not use NET_RAW capability. Note: this test also ensures iptables and nftables are not configured by workload pods: - NET_ADMIN and NET_RAW are required to modify nftables (namespaced) which is not desired inside pods. nftables should be configured by an administrator outside the scope of the workload. nftables are usually configured by operators, for instance the Performance Addon Operator (PAO) or istio. - Privileged container are required to modify host iptables, which is not safe to perform inside pods. nftables should be configured by an administrator outside the scope of the workload. iptables are usually configured by operators, for instance the Performance Addon Operator (PAO) or istio.|
|Suggested Remediation|Exception possible if a workload uses mlock(), mlockall(), shmctl(), mmap(); exception will be considered for DPDK applications. Must identify which container requires the capability and detail why.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-user-plane-cnfs|
|Exception Process|Exception will be considered for user plane or networking functions. Must identify which container requires the capability and detail why.|
|Impact Statement|NET_RAW capability enables packet manipulation and network sniffing, which can be used for attacks against other workloads and compromise network security.|
|Tags|telco,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-no-1337-uid

|Property|Description|
|---|---|
|Unique ID|access-control-no-1337-uid|
|Description|Checks that all pods are not using the securityContext UID 1337|
|Suggested Remediation|Use another process UID that is not 1337.|
|Best Practice Reference|No Doc Link - Extended|
|Exception Process|No exception needed for optional/extended tests.|
|Impact Statement|UID 1337 is reserved for use by Istio service mesh components; using it for applications can cause conflicts with Istio sidecars and break service mesh functionality.|
|Tags|extended,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### access-control-one-process-per-container

|Property|Description|
|---|---|
|Unique ID|access-control-one-process-per-container|
|Description|Check that all containers under test have only one process running|
|Suggested Remediation|Launch only one process per container. Should adhere to 1 process per container best practice wherever possible.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-one-process-per-container|
|Exception Process|No exception needed for optional/extended tests. Not applicable to SNO applications.|
|Impact Statement|Multiple processes per container complicate monitoring, debugging, and security assessment, and can lead to zombie processes and resource leaks.|
|Tags|common,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### access-control-pod-automount-service-account-token

|Property|Description|
|---|---|
|Unique ID|access-control-pod-automount-service-account-token|
|Description|Check that all pods under test have automountServiceAccountToken set to false. Only pods that require access to the kubernetes API server should have automountServiceAccountToken set to true|
|Suggested Remediation|Check that pod has automountServiceAccountToken set to false or pod is attached to service account which has automountServiceAccountToken set to false, unless the pod needs access to the kubernetes API server. Pods which do not need API access should set automountServiceAccountToken to false in pod spec.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-automount-services-for-pods|
|Exception Process|Exception will be considered if container needs to access APIs which OCP does not offer natively. Must document which container requires which API(s) and detail why existing OCP APIs cannot be used.|
|Impact Statement|Auto-mounted service account tokens expose Kubernetes API credentials to application code, creating potential attack vectors if applications are compromised.|
|Tags|telco,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-pod-host-ipc

|Property|Description|
|---|---|
|Unique ID|access-control-pod-host-ipc|
|Description|Verifies that the spec.HostIpc parameter is set to false|
|Suggested Remediation|Set the spec.HostIpc parameter to false in the pod configuration. Workloads should avoid accessing host resources - spec.HostIpc should be false.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security|
|Exception Process|Exception for host resource access tests will only be considered in rare cases where it is absolutely needed|
|Impact Statement|Host IPC access allows containers to communicate with host processes, potentially exposing sensitive information and enabling privilege escalation.|
|Tags|common,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-pod-host-network

|Property|Description|
|---|---|
|Unique ID|access-control-pod-host-network|
|Description|Verifies that the spec.HostNetwork parameter is not set (not present)|
|Suggested Remediation|Set the spec.HostNetwork parameter to false in the pod configuration. Workloads should avoid accessing host resources - spec.HostNetwork should be false.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security|
|Exception Process|Exception for host resource access tests will only be considered in rare cases where it is absolutely needed|
|Impact Statement|Host network access removes network isolation, exposes containers to host network interfaces, and can compromise cluster networking security.|
|Tags|common,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-pod-host-path

|Property|Description|
|---|---|
|Unique ID|access-control-pod-host-path|
|Description|Verifies that the spec.HostPath parameter is not set (not present)|
|Suggested Remediation|Set the spec.HostPath parameter to false in the pod configuration. Workloads should avoid accessing host resources - spec.HostPath should be false.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security|
|Exception Process|Exception for host resource access tests will only be considered in rare cases where it is absolutely needed|
|Impact Statement|Host path mounts can expose sensitive host files to containers, enable container escape attacks, and compromise host system integrity.|
|Tags|common,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-pod-host-pid

|Property|Description|
|---|---|
|Unique ID|access-control-pod-host-pid|
|Description|Verifies that the spec.HostPid parameter is set to false|
|Suggested Remediation|Set the spec.HostPid parameter to false in the pod configuration. Workloads should avoid accessing host resources - spec.HostPid should be false.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security|
|Exception Process|Exception for host resource access tests will only be considered in rare cases where it is absolutely needed|
|Impact Statement|Host PID access allows containers to see and interact with all host processes, creating opportunities for privilege escalation and information disclosure.|
|Tags|common,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-pod-role-bindings

|Property|Description|
|---|---|
|Unique ID|access-control-pod-role-bindings|
|Description|Ensures that a workload does not utilize RoleBinding(s) in a non-workload Namespace.|
|Suggested Remediation|Ensure the workload is not configured to use RoleBinding(s) in a non-workload Namespace. Scope of role must <= scope of creator of role.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-security-and-role-based-access-control|
|Exception Process|No exceptions|
|Impact Statement|Cross-namespace role bindings can violate tenant isolation and create unintended privilege escalation paths.|
|Tags|common,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-pod-service-account

|Property|Description|
|---|---|
|Unique ID|access-control-pod-service-account|
|Description|Tests that each workload Pod utilizes a valid Service Account. Default or empty service account is not valid.|
|Suggested Remediation|Ensure that the each workload Pod is configured to use a valid Service Account|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-scc-permissions-for-an-application|
|Exception Process|No exceptions|
|Impact Statement|Default service accounts often have excessive privileges; improper usage can lead to unauthorized API access and security violations.|
|Tags|common,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-requests

|Property|Description|
|---|---|
|Unique ID|access-control-requests|
|Description|Check that containers have resource requests specified in their spec. Set proper resource requests based on container use case.|
|Suggested Remediation|Add requests to your container spec. See: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-requests-limits|
|Exception Process|Exceptions possible for platform and infrastructure containers. Must identify which container needs access and document why with details.|
|Impact Statement|Missing resource requests can lead to resource contention, node instability, and unpredictable application performance.|
|Tags|telco,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-security-context

|Property|Description|
|---|---|
|Unique ID|access-control-security-context|
|Description|Checks the security context matches one of the 4 categories|
|Suggested Remediation|Exception possible if a workload uses mlock(), mlockall(), shmctl(), mmap(); exception will be considered for DPDK applications. Must identify which container requires the capability and document why. If the container had the right configuration of the allowed category from the 4 approved list then the test will pass. The 4 categories are defined in Requirement ID 94118 [here](#security-context-categories)|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-linux-capabilities|
|Exception Process|no exception needed for optional/extended test|
|Impact Statement|Incorrect security context configurations can weaken container isolation, enable privilege escalation, and create exploitable attack vectors.|
|Tags|extended,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### access-control-security-context-non-root-user-id-check

|Property|Description|
|---|---|
|Unique ID|access-control-security-context-non-root-user-id-check|
|Description|Checks securityContext's runAsNonRoot and runAsUser fields at pod and container level to make sure containers are not run as root.|
|Suggested Remediation|Set the securityContext.runAsNonRoot field to true either at pod or container level. Alternatively, set a non-zero value to securityContext.runAsUser field either at pod or container level.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security|
|Exception Process|No exceptions - will only be considered under special circumstances. Must identify which container needs access and document why with details.|
|Impact Statement|Running containers as root increases the blast radius of security vulnerabilities and can lead to full host compromise if containers are breached.|
|Tags|common,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-security-context-privilege-escalation

|Property|Description|
|---|---|
|Unique ID|access-control-security-context-privilege-escalation|
|Description|Checks if privileged escalation is enabled (AllowPrivilegeEscalation=true).|
|Suggested Remediation|Configure privilege escalation to false. Privileged escalation should not be allowed (AllowPrivilegeEscalation=false).|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security|
|Exception Process|No exceptions|
|Impact Statement|Allowing privilege escalation can lead to containers gaining root access, compromising the security boundary between containers and hosts.|
|Tags|common,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-security-context-read-only-file-system

|Property|Description|
|---|---|
|Unique ID|access-control-security-context-read-only-file-system|
|Description|Checks the security context readOnlyFileSystem in containers is enabled. Containers should not try modify its own filesystem.|
|Suggested Remediation|No exceptions - will only be considered under special circumstances. Must identify which container needs access and document why with details.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-linux-capabilities|
|Exception Process|No exceptions|
|Impact Statement|Writable root filesystems increase the attack surface and can be exploited to modify container behavior or persist malware.|
|Tags|common,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### access-control-service-type

|Property|Description|
|---|---|
|Unique ID|access-control-service-type|
|Description|Tests that each workload Service does not utilize NodePort(s).|
|Suggested Remediation|Ensure Services are not configured to use NodePort(s). Workloads should avoid accessing host resources - tests that each workload Service does not utilize NodePort(s).|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security|
|Exception Process|Exception for host resource access tests will only be considered in rare cases where it is absolutely needed|
|Impact Statement|NodePort services expose applications directly on host ports, creating security risks and potential port conflicts with host services.|
|Tags|common,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-ssh-daemons

|Property|Description|
|---|---|
|Unique ID|access-control-ssh-daemons|
|Description|Check that pods do not run SSH daemons.|
|Suggested Remediation|Ensure that no SSH daemons are running inside a pod. Pods should not run as SSH Daemons (replicaset or statefulset only).|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-pod-interaction-and-configuration|
|Exception Process|No exceptions - special consideration can be given to certain containers which run as utility tool daemon|
|Impact Statement|SSH daemons in containers create additional attack surfaces, violate immutable infrastructure principles, and can be exploited for unauthorized access.|
|Tags|telco,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-sys-admin-capability-check

|Property|Description|
|---|---|
|Unique ID|access-control-sys-admin-capability-check|
|Description|Ensures that containers do not use SYS_ADMIN capability|
|Suggested Remediation|Exception possible if a workload uses mlock(), mlockall(), shmctl(), mmap(); exception will be considered for DPDK applications. Must identify which container requires the capability and detail why. Containers should not use the SYS_ADMIN Linux capability.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-avoid-sys_admin|
|Exception Process|No exceptions|
|Impact Statement|SYS_ADMIN capability provides extensive privileges that can compromise container isolation, enable host system access, and create serious security vulnerabilities.|
|Tags|common,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### access-control-sys-nice-realtime-capability

|Property|Description|
|---|---|
|Unique ID|access-control-sys-nice-realtime-capability|
|Description|Check that pods running on nodes with realtime kernel enabled have the SYS_NICE capability enabled in their spec. In the case that a workolad is running on a node using the real-time kernel, SYS_NICE will be used to allow DPDK application to switch to SCHED_FIFO.|
|Suggested Remediation|If pods are scheduled to realtime kernel nodes, they must add SYS_NICE capability to their spec.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-sys_nice|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Missing SYS_NICE capability on real-time nodes prevents applications from setting appropriate scheduling priorities, causing performance degradation.|
|Tags|telco,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### access-control-sys-ptrace-capability

|Property|Description|
|---|---|
|Unique ID|access-control-sys-ptrace-capability|
|Description|Check that if process namespace sharing is enabled for a Pod then the SYS_PTRACE capability is allowed. This capability is required when using Process Namespace Sharing. This is used when processes from one Container need to be exposed to another Container. For example, to send signals like SIGHUP from a process in a Container to another process in another Container. For more information on these capabilities refer to https://cloud.redhat.com/blog/linux-capabilities-in-openshift and https://kubernetes.io/docs/tasks/configure-pod-container/share-process-namespace/|
|Suggested Remediation|Allow the SYS_PTRACE capability when enabling process namespace sharing for a Pod|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-sys_ptrace|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Missing SYS_PTRACE capability when using shared process namespaces prevents inter-container process communication, breaking application functionality.|
|Tags|telco,access-control|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

### affiliated-certification

#### affiliated-certification-container-is-certified-digest

|Property|Description|
|---|---|
|Unique ID|affiliated-certification-container-is-certified-digest|
|Description|Tests whether container images that are autodiscovered have passed the Red Hat Container Certification Program by their digest(CCP).|
|Suggested Remediation|Ensure that your container has passed the Red Hat Container Certification Program (CCP).|
|Best Practice Reference|https://docs.redhat.com/en/documentation/red_hat_software_certification/2025/html/red_hat_software_certification_workflow_guide/index|
|Exception Process|There is no documented exception process for this. A partner can run the Red Hat Best Practices Test Suite before passing other certifications (Container/Operator/HelmChart) but the affiliated certification test cases in the Red Hat Best Practices Test Suite must be re-run once the other certifications have been granted.|
|Impact Statement|Uncertified containers may contain security vulnerabilities, lack enterprise support, and fail to meet compliance requirements.|
|Tags|common,affiliated-certification|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### affiliated-certification-helm-version

|Property|Description|
|---|---|
|Unique ID|affiliated-certification-helm-version|
|Description|Test to check if the helm chart is v3|
|Suggested Remediation|Check Helm Chart is v3 and not v2 which is not supported due to security risks associated with Tiller.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-helm|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Helm v2 has known security vulnerabilities and lacks proper RBAC controls, creating significant security risks in production environments.|
|Tags|common,affiliated-certification|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### affiliated-certification-helmchart-is-certified

|Property|Description|
|---|---|
|Unique ID|affiliated-certification-helmchart-is-certified|
|Description|Tests whether helm charts listed in the cluster passed the Red Hat Helm Certification Program.|
|Suggested Remediation|Ensure that the helm charts under test passed the Red Hat's helm Certification Program (e.g. listed in https://charts.openshift.io/index.yaml).|
|Best Practice Reference|https://docs.redhat.com/en/documentation/red_hat_software_certification/2025/html/red_hat_software_certification_workflow_guide/index|
|Exception Process|There is no documented exception process for this. A partner can run the Red Hat Best Practices Test Suite before passing other certifications (Container/Operator/HelmChart) but the affiliated certification test cases in the Red Hat Best Practices Test Suite must be re-run once the other certifications have been granted.|
|Impact Statement|Uncertified helm charts may contain security vulnerabilities, configuration errors, and lack proper testing, leading to deployment failures.|
|Tags|common,affiliated-certification|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### affiliated-certification-operator-is-certified

|Property|Description|
|---|---|
|Unique ID|affiliated-certification-operator-is-certified|
|Description|Tests whether the workload Operators listed in the configuration file have passed the Red Hat Operator Certification Program (OCP).|
|Suggested Remediation|Ensure that your Operator has passed Red Hat's Operator Certification Program (OCP).|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements|
|Exception Process|There is no documented exception process for this. A partner can run the Red Hat Best Practices Test Suite before passing other certifications (Container/Operator/HelmChart) but the affiliated certification test cases in the Red Hat Best Practices Test Suite must be re-run once the other certifications have been granted.|
|Impact Statement|Uncertified operators may have security flaws, compatibility issues, and lack enterprise support, creating operational risks.|
|Tags|common,affiliated-certification|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

### lifecycle

#### lifecycle-affinity-required-pods

|Property|Description|
|---|---|
|Unique ID|lifecycle-affinity-required-pods|
|Description|Checks that affinity rules are in place if AffinityRequired: 'true' labels are set on Pods.|
|Suggested Remediation|Pods which need to be co-located on the same node need Affinity rules. If a pod/statefulset/deployment is required to use affinity rules, please add AffinityRequired: 'true' as a label.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Missing affinity rules can cause incorrect pod placement, leading to performance issues and failure to meet co-location requirements.|
|Tags|telco,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-container-poststart

|Property|Description|
|---|---|
|Unique ID|lifecycle-container-poststart|
|Description|Ensure that the containers lifecycle postStart management feature is configured. A container must receive important events from the platform and conform/react to these events properly. For example, a container should catch SIGTERM or SIGKILL from the platform and shutdown as quickly as possible. Other typically important events from the platform are PostStart to initialize before servicing requests and PreStop to release resources cleanly before shutting down.|
|Suggested Remediation|PostStart is normally used to configure the container, set up dependencies, and record the new creation. You could use this event to check that a required API is available before the container’s main work begins. Kubernetes will not change the container’s state to Running until the PostStart script has executed successfully. For details, see https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks. PostStart is used to configure container, set up dependencies, record new creation. It can also be used to check that a required API is available before the container’s work begins.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cloud-native-design-best-practices|
|Exception Process|Identify which pod is not conforming to the process and submit information as to why it cannot use a postStart startup specification.|
|Impact Statement|Missing PostStart hooks can cause containers to start serving traffic before proper initialization, leading to application errors.|
|Tags|telco,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-container-prestop

|Property|Description|
|---|---|
|Unique ID|lifecycle-container-prestop|
|Description|Ensure that the containers lifecycle preStop management feature is configured. The most basic requirement for the lifecycle management of Pods in OpenShift are the ability to start and stop correctly. There are different ways a pod can stop on an OpenShift cluster. One way is that the pod can remain alive but non-functional. Another way is that the pod can crash and become non-functional. When pods are shut down by the platform they are sent a SIGTERM signal which means that the process in the container should start shutting down, closing connections and stopping all activity. If the pod doesn’t shut down within the default 30 seconds then the platform may send a SIGKILL signal which will stop the pod immediately. This method isn’t as clean and the default time between the SIGTERM and SIGKILL messages can be modified based on the requirements of the application. Containers should respond to SIGTERM/SIGKILL with graceful shutdown.|
|Suggested Remediation|The preStop can be used to gracefully stop the container and clean resources (e.g., DB connection). For details, see https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks. All pods must respond to SIGTERM signal and shutdown gracefully with a zero exit code.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cloud-native-design-best-practices|
|Exception Process|Identify which pod is not conforming to the process and submit information as to why it cannot use a preStop shutdown specification.|
|Impact Statement|Missing PreStop hooks can cause ungraceful shutdowns, data loss, and connection drops during container termination.|
|Tags|telco,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-cpu-isolation

|Property|Description|
|---|---|
|Unique ID|lifecycle-cpu-isolation|
|Description|CPU isolation requires: For each container within the pod, resource requests and limits must be identical. If cpu requests and limits are not identical and in whole units (Guaranteed pods with exclusive cpus), your pods will not be tested for compliance. The runTimeClassName must be specified. Annotations required disabling CPU and IRQ load-balancing.|
|Suggested Remediation|CPU isolation testing is enabled. Please ensure that all pods adhere to the CPU isolation requirements.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cpu-isolation|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Improper CPU isolation can cause performance interference between workloads and fail to provide guaranteed compute resources.|
|Tags|telco,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-crd-scaling

|Property|Description|
|---|---|
|Unique ID|lifecycle-crd-scaling|
|Description|Tests that a workload's CRD support scale in/out operations. First, the test starts getting the current replicaCount (N) of the crd/s with the Pod Under Test. Then, it executes the scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the crd/s. In case of crd that are managed by HPA the test is changing the min and max value to crd Replica - 1 during scale-in and the original replicaCount again for both min/max during the scale-out stage. Lastly its restoring the original min/max replica of the crd/s|
|Suggested Remediation|Ensure the workload's CRDs can scale in/out successfully.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations|
|Exception Process|There is no documented exception process for this. Not applicable to SNO applications.|
|Impact Statement|CRD scaling failures can prevent operator-managed applications from scaling properly, limiting application availability and performance.|
|Tags|common,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### lifecycle-deployment-scaling

|Property|Description|
|---|---|
|Unique ID|lifecycle-deployment-scaling|
|Description|Tests that workload deployments support scale in/out operations. First, the test starts getting the current replicaCount (N) of the deployment/s with the Pod Under Test. Then, it executes the scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the deployment/s. In case of deployments that are managed by HPA the test is changing the min and max value to deployment Replica - 1 during scale-in and the original replicaCount again for both min/max during the scale-out stage. Lastly its restoring the original min/max replica of the deployment/s|
|Suggested Remediation|Ensure the workload's deployments/replica sets can scale in/out successfully.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations|
|Exception Process|There is no documented exception process for this. Not applicable to SNO applications.|
|Impact Statement|Deployment scaling failures prevent horizontal scaling operations, limiting application elasticity and availability during high load.|
|Tags|common,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### lifecycle-image-pull-policy

|Property|Description|
|---|---|
|Unique ID|lifecycle-image-pull-policy|
|Description|Ensure that the containers under test are using IfNotPresent as Image Pull Policy. If there is a situation where the container dies and needs to be restarted, the image pull policy becomes important. PullIfNotPresent is recommended so that a loss of image registry access does not prevent the pod from restarting.|
|Suggested Remediation|Ensure that the containers under test are using IfNotPresent as Image Pull Policy.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-use-imagepullpolicy:-ifnotpresent|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Incorrect image pull policies can cause deployment failures when image registries are unavailable or during network issues.|
|Tags|telco,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-liveness-probe

|Property|Description|
|---|---|
|Unique ID|lifecycle-liveness-probe|
|Description|Check that all containers under test have liveness probe defined. The most basic requirement for the lifecycle management of Pods in OpenShift are the ability to start and stop correctly. When starting up, health probes like liveness and readiness checks can be put into place to ensure the application is functioning properly.|
|Suggested Remediation|Add a liveness probe to deployed containers. workloads shall self-recover from common failures like pod failure, host failure, and network failure. Kubernetes native mechanisms such as health-checks (Liveness, Readiness and Startup Probes) shall be employed at a minimum.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-liveness-readiness-and-startup-probes|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Missing liveness probes prevent Kubernetes from detecting and recovering from application deadlocks and hangs.|
|Tags|telco,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-persistent-volume-reclaim-policy

|Property|Description|
|---|---|
|Unique ID|lifecycle-persistent-volume-reclaim-policy|
|Description|Check that the persistent volumes the workloads pods are using have a reclaim policy of delete. Network Functions should clear persistent storage by deleting their PVs when removing their application from a cluster.|
|Suggested Remediation|Ensure that all persistent volumes are using the reclaim policy: delete|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-csi|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Incorrect reclaim policies can lead to data persistence after application removal, causing storage waste and potential data security issues.|
|Tags|telco,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-pod-high-availability

|Property|Description|
|---|---|
|Unique ID|lifecycle-pod-high-availability|
|Description|Ensures that workloads Pods specify podAntiAffinity rules and replica value is set to more than 1.|
|Suggested Remediation|In high availability cases, Pod podAntiAffinity rule should be specified for pod scheduling and pod replica value is set to more than 1 .|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations|
|Exception Process|There is no documented exception process for this. Not applicable to SNO applications.|
|Impact Statement|Missing anti-affinity rules can cause all pod replicas to be scheduled on the same node, creating single points of failure.|
|Tags|common,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### lifecycle-pod-owner-type

|Property|Description|
|---|---|
|Unique ID|lifecycle-pod-owner-type|
|Description|Tests that the workload Pods are deployed as part of a ReplicaSet(s)/StatefulSet(s).|
|Suggested Remediation|Deploy the workload using ReplicaSet/StatefulSet.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-no-naked-pods|
|Exception Process|There is no documented exception process for this. Pods should not be deployed as DaemonSet or naked pods.|
|Impact Statement|Naked pods and DaemonSets lack proper lifecycle management, making updates, scaling, and recovery operations difficult or impossible.|
|Tags|telco,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-pod-recreation

|Property|Description|
|---|---|
|Unique ID|lifecycle-pod-recreation|
|Description|Tests that a workload is configured to support High Availability. First, this test cordons and drains a Node that hosts the workload Pod. Next, the test ensures that OpenShift can re-instantiate the Pod on another Node, and that the actual replica count matches the desired replica count.|
|Suggested Remediation|Ensure that the workloads Pods utilize a configuration that supports High Availability. Additionally, ensure that there are available Nodes in the OpenShift cluster that can be utilized in the event that a host Node fails.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-upgrade-expectations|
|Exception Process|No exceptions - workloads should be able to be restarted/recreated.|
|Impact Statement|Failed pod recreation indicates poor high availability configuration, leading to potential service outages during node failures.|
|Tags|common,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### lifecycle-pod-scheduling

|Property|Description|
|---|---|
|Unique ID|lifecycle-pod-scheduling|
|Description|Ensures that workload Pods do not specify nodeSelector or nodeAffinity. In most cases, Pods should allow for instantiation on any underlying Node. Workloads shall not use node selectors nor taints/tolerations to assign pod location.|
|Suggested Remediation|In most cases, Pod's should not specify their host Nodes through nodeSelector or nodeAffinity. However, there are cases in which workloads require specialized hardware specific to a particular class of Node.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations|
|Exception Process|Exception will only be considered if application requires specialized hardware. Must specify which container requires special hardware and why.|
|Impact Statement|Node selectors can create scheduling constraints that reduce cluster flexibility and cause deployment failures when nodes are unavailable.|
|Tags|telco,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Mandatory|
|Telco|Optional|

#### lifecycle-pod-toleration-bypass

|Property|Description|
|---|---|
|Unique ID|lifecycle-pod-toleration-bypass|
|Description|Check that pods do not have NoExecute, PreferNoSchedule, or NoSchedule tolerations that have been modified from the default.|
|Suggested Remediation|Do not allow pods to bypass the NoExecute, PreferNoSchedule, or NoSchedule tolerations that are default applied by Kubernetes.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cpu-manager-pinning|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Modified tolerations can allow pods to be scheduled on inappropriate nodes, violating scheduling policies and causing performance issues.|
|Tags|telco,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-readiness-probe

|Property|Description|
|---|---|
|Unique ID|lifecycle-readiness-probe|
|Description|Check that all containers under test have readiness probe defined. There are different ways a pod can stop on on OpenShift cluster. One way is that the pod can remain alive but non-functional. Another way is that the pod can crash and become non-functional. In the first case, if the administrator has implemented liveness and readiness checks, OpenShift can stop the pod and either restart it on the same node or a different node in the cluster. For the second case, when the application in the pod stops, it should exit with a code and write suitable log entries to help the administrator diagnose what the issue was that caused the problem.|
|Suggested Remediation|Add a readiness probe to deployed containers|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-liveness-readiness-and-startup-probes|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Missing readiness probes can cause traffic to be routed to non-ready pods, resulting in failed requests and poor user experience.|
|Tags|telco,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-startup-probe

|Property|Description|
|---|---|
|Unique ID|lifecycle-startup-probe|
|Description|Check that all containers under test have startup probe defined. Workloads shall self-recover from common failures like pod failure, host failure, and network failure. Kubernetes native mechanisms such as health-checks (Liveness, Readiness and Startup Probes) shall be employed at a minimum.|
|Suggested Remediation|Add a startup probe to deployed containers|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-liveness-readiness-and-startup-probes|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Missing startup probes can cause slow-starting applications to be killed prematurely, preventing successful application startup.|
|Tags|telco,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### lifecycle-statefulset-scaling

|Property|Description|
|---|---|
|Unique ID|lifecycle-statefulset-scaling|
|Description|Tests that workload statefulsets support scale in/out operations. First, the test starts getting the current replicaCount (N) of the statefulset/s with the Pod Under Test. Then, it executes the scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the statefulset/s. In case of statefulsets that are managed by HPA the test is changing the min and max value to statefulset Replica - 1 during scale-in and the original replicaCount again for both min/max during the scale-out stage. Lastly its restoring the original min/max replica of the statefulset/s|
|Suggested Remediation|Ensure the workload's statefulsets/replica sets can scale in/out successfully.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations|
|Exception Process|There is no documented exception process for this. Not applicable to SNO applications.|
|Impact Statement|StatefulSet scaling issues can prevent proper data persistence and ordered deployment of stateful applications.|
|Tags|common,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### lifecycle-storage-provisioner

|Property|Description|
|---|---|
|Unique ID|lifecycle-storage-provisioner|
|Description|Checks that pods do not place persistent volumes on local storage in multinode clusters. Local storage is recommended for single node clusters, but only one type of local storage should be installed (lvms or noprovisioner).|
|Suggested Remediation|Use a non-local storage (e.g. no kubernetes.io/no-provisioner and no topolvm.io provisioners) in multinode clusters. Local storage are recommended for single node clusters only, but a single local provisioner should be installed.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-local-storage|
|Exception Process|No exceptions|
|Impact Statement|Inappropriate storage provisioners can cause data persistence issues, performance problems, and storage failures.|
|Tags|common,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### lifecycle-topology-spread-constraint

|Property|Description|
|---|---|
|Unique ID|lifecycle-topology-spread-constraint|
|Description|Ensures that Deployments using TopologySpreadConstraints include constraints for both hostname and zone topology keys. This helps telco workloads avoid needing to tweak PodDisruptionBudgets before platform upgrades. If TopologySpreadConstraints is not defined, the test passes as Kubernetes scheduler implicitly uses hostname and zone constraints. Not applicable to SNO applications.|
|Suggested Remediation|If using TopologySpreadConstraints in your Deployment, ensure you include constraints for both 'kubernetes.io/hostname' and 'topology.kubernetes.io/zone' topology keys. Alternatively, you can omit TopologySpreadConstraints entirely to let Kubernetes scheduler use implicit hostname and zone constraints. This helps maintain workload availability during platform upgrades without manually adjusting PodDisruptionBudgets.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Without proper topology spread constraints, pods may cluster on nodes causing PodDisruptionBudgets to block platform upgrades, requiring manual PDB adjustments and increasing operational complexity during maintenance windows.|
|Tags|telco,lifecycle|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Mandatory|

### manageability

#### manageability-container-port-name-format

|Property|Description|
|---|---|
|Unique ID|manageability-container-port-name-format|
|Description|Check that the container's ports name follow the naming conventions. Name field in ContainerPort section must be of form `<protocol>[-<suffix>]`. More naming convention requirements may be released in future|
|Suggested Remediation|Ensure that the container's ports name follow our partner naming conventions|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-requirements-cnf-reqs|
|Exception Process|No exception needed for optional/extended tests.|
|Impact Statement|Incorrect port naming conventions can cause service discovery issues and configuration management problems.|
|Tags|extended,manageability|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### manageability-containers-image-tag

|Property|Description|
|---|---|
|Unique ID|manageability-containers-image-tag|
|Description|Check that image tag exists on containers.|
|Suggested Remediation|Ensure that all the container images are tagged. Checks containers have image tags (e.g. latest, stable, dev).|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-image-tagging|
|Exception Process|No exception needed for optional/extended tests.|
|Impact Statement|Missing image tags make it difficult to track versions, perform rollbacks, and maintain deployment consistency.|
|Tags|extended,manageability|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

### networking

#### networking-dual-stack-service

|Property|Description|
|---|---|
|Unique ID|networking-dual-stack-service|
|Description|Checks that all services in namespaces under test are either ipv6 single stack or dual stack. This test case requires the deployment of the probe daemonset.|
|Suggested Remediation|Configure every workload service with either a single stack ipv6 or dual stack (ipv4/ipv6) load balancer.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-ipv4-&-ipv6|
|Exception Process|No exception needed for optional/extended tests.|
|Impact Statement|Single-stack IPv4 services limit network architecture flexibility and prevent migration to modern dual-stack infrastructures.|
|Tags|extended,networking|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### networking-icmpv4-connectivity

|Property|Description|
|---|---|
|Unique ID|networking-icmpv4-connectivity|
|Description|Checks that each workload Container is able to communicate via ICMPv4 on the Default OpenShift network. This test case requires the Deployment of the probe daemonset and at least 2 pods connected to each network under test(one source and one destination). If no network with more than 2 pods exists this test will be skipped.|
|Suggested Remediation|Ensure that the workload is able to communicate via the Default OpenShift network. In some rare cases, workloads may require routing table changes in order to communicate over the Default network. To exclude a particular pod from ICMPv4 connectivity tests, add the redhat-best-practices-for-k8s.com/skip_connectivity_tests label to it. The label value is trivial, only its presence.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-ipv4-&-ipv6|
|Exception Process|No exceptions - must be able to communicate on default network using IPv4|
|Impact Statement|Failure indicates potential network isolation issues that could prevent workload components from communicating, leading to service degradation or complete application failure.|
|Tags|common,networking|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### networking-icmpv4-connectivity-multus

|Property|Description|
|---|---|
|Unique ID|networking-icmpv4-connectivity-multus|
|Description|Checks that each workload Container is able to communicate via ICMPv4 on the Multus network(s). This test case requires the Deployment of the probe daemonset and at least 2 pods connected to each network under test(one source and one destination). If no network with more than 2 pods exists this test will be skipped.|
|Suggested Remediation|Ensure that the workload is able to communicate via the Multus network(s). In some rare cases, workloads may require routing table changes in order to communicate over the Multus network(s). To exclude a particular pod from ICMPv4 connectivity tests, add the redhat-best-practices-for-k8s.com/skip_connectivity_tests label to it. The label value is trivial, only its presence. Not applicable if MULTUS is not supported.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Multus network connectivity issues can isolate workloads from secondary networks, breaking multi-network applications and reducing network redundancy.|
|Tags|telco,networking|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### networking-icmpv6-connectivity

|Property|Description|
|---|---|
|Unique ID|networking-icmpv6-connectivity|
|Description|Checks that each workload Container is able to communicate via ICMPv6 on the Default OpenShift network. This test case requires the Deployment of the probe daemonset and at least 2 pods connected to each network under test(one source and one destination). If no network with more than 2 pods exists this test will be skipped.|
|Suggested Remediation|Ensure that the workload is able to communicate via the Default OpenShift network. In some rare cases, workloads may require routing table changes in order to communicate over the Default network. To exclude a particular pod from ICMPv6 connectivity tests, add the redhat-best-practices-for-k8s.com/skip_connectivity_tests label to it. The label value is trivial, only its presence. Not applicable if IPv6 is not supported.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-ipv4-&-ipv6|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|IPv6 connectivity failures can prevent dual-stack applications from functioning properly and limit future network architecture flexibility.|
|Tags|common,networking|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### networking-icmpv6-connectivity-multus

|Property|Description|
|---|---|
|Unique ID|networking-icmpv6-connectivity-multus|
|Description|Checks that each workload Container is able to communicate via ICMPv6 on the Multus network(s). This test case requires the Deployment of the probe daemonset and at least 2 pods connected to each network under test(one source and one destination). If no network with more than 2 pods exists this test will be skipped.|
|Suggested Remediation|Ensure that the workload is able to communicate via the Multus network(s). In some rare cases, workloads may require routing table changes in order to communicate over the Multus network(s). To exclude a particular pod from ICMPv6 connectivity tests, add the redhat-best-practices-for-k8s.com/skip_connectivity_tests label to it.The label value is trivial, only its presence. Not applicable if IPv6/MULTUS is not supported.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|IPv6 Multus connectivity problems can prevent dual-stack multi-network scenarios from working, limiting network scalability and future-proofing.|
|Tags|telco,networking|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### networking-network-attachment-definition-sriov-mtu

|Property|Description|
|---|---|
|Unique ID|networking-network-attachment-definition-sriov-mtu|
|Description|Ensures that MTU values are set correctly in NetworkAttachmentDefinitions for SRIOV network interfaces.|
|Suggested Remediation|Ensure that the MTU of the SR-IOV network attachment definition is set explicitly.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-multus-sr-iov---macvlan|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Incorrect MTU settings can cause packet fragmentation, network performance issues, and connectivity failures in high-performance networking scenarios.|
|Tags|faredge,networking|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### networking-network-policy-deny-all

|Property|Description|
|---|---|
|Unique ID|networking-network-policy-deny-all|
|Description|Check that network policies attached to namespaces running workload pods contain a default deny-all rule for both ingress and egress traffic|
|Suggested Remediation|Ensure that a NetworkPolicy with a default deny-all is applied. After the default is applied, apply a network policy to allow the traffic your application requires.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-vrfs-aka-routing-instances|
|Exception Process|No exception needed for optional/extended tests.|
|Impact Statement|Without default deny-all network policies, workloads are exposed to lateral movement attacks and unauthorized network access, compromising security posture and potentially enabling data breaches.|
|Tags|common,networking|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### networking-ocp-reserved-ports-usage

|Property|Description|
|---|---|
|Unique ID|networking-ocp-reserved-ports-usage|
|Description|Check that containers do not listen on ports that are reserved by OpenShift|
|Suggested Remediation|Ensure that workload's apps do not listen on ports that are reserved by OpenShift. The following ports are reserved by OpenShift and must NOT be used by any application: 22623, 22624.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-ports-reserved-by-openshift|
|Exception Process|No exceptions|
|Impact Statement|Using OpenShift-reserved ports can cause critical platform services to fail, potentially destabilizing the entire cluster.|
|Tags|common,networking|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### networking-reserved-partner-ports

|Property|Description|
|---|---|
|Unique ID|networking-reserved-partner-ports|
|Description|Checks that pods and containers are not consuming ports designated as reserved by partner|
|Suggested Remediation|Ensure ports are not being used that are reserved by our partner|
|Best Practice Reference|No Doc Link - Extended|
|Exception Process|No exception needed for optional/extended tests.|
|Impact Statement|Using reserved ports can cause port conflicts with essential platform services, leading to service startup failures and unpredictable application behavior.|
|Tags|extended,networking|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### networking-restart-on-reboot-sriov-pod

|Property|Description|
|---|---|
|Unique ID|networking-restart-on-reboot-sriov-pod|
|Description|Ensures that the label restart-on-reboot exists on pods that use SRIOV network interfaces.|
|Suggested Remediation|Ensure that the label restart-on-reboot exists on pods that use SRIOV network interfaces.|
|Best Practice Reference|No Doc Link - Far Edge|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Without restart-on-reboot labels, SRIOV-enabled pods may fail to recover from a race condition between kubernetes services startup and SR-IOV device plugin configuration on StarlingX AIO systems, causing SR-IOV devices to disappear from running pods when FPGA devices are reset.|
|Tags|faredge,networking|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Optional|

#### networking-undeclared-container-ports-usage

|Property|Description|
|---|---|
|Unique ID|networking-undeclared-container-ports-usage|
|Description|Check that containers do not listen on ports that weren't declared in their specification. Platforms may be configured to block undeclared ports.|
|Suggested Remediation|Ensure the workload's apps do not listen on undeclared containers' ports.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-requirements-cnf-reqs|
|Exception Process|No exception needed for optional/extended tests.|
|Impact Statement|Undeclared ports can be blocked by security policies, causing unexpected connectivity issues and making troubleshooting difficult.|
|Tags|extended,networking|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

### observability

#### observability-compatibility-with-next-ocp-release

|Property|Description|
|---|---|
|Unique ID|observability-compatibility-with-next-ocp-release|
|Description|Checks to ensure if the APIs the workload uses are compatible with the next OCP version|
|Suggested Remediation|Ensure the APIs the workload uses are compatible with the next OCP version|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-k8s-api-versions|
|Exception Process|No exceptions|
|Impact Statement|Deprecated API usage can cause applications to break during OpenShift upgrades, requiring emergency fixes.|
|Tags|common,observability|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### observability-container-logging

|Property|Description|
|---|---|
|Unique ID|observability-container-logging|
|Description|Check that all containers under test use standard input output and standard error when logging. A container must provide APIs for the platform to observe the container health and act accordingly. These APIs include health checks (liveness and readiness), logging to stderr and stdout for log aggregation (by tools such as Logstash or Filebeat), and integrate with tracing and metrics-gathering libraries (such as Prometheus or Metricbeat).|
|Suggested Remediation|Ensure containers are not redirecting stdout/stderr|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-logging|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Improper logging configuration prevents log aggregation and monitoring, making troubleshooting and debugging difficult.|
|Tags|telco,observability|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### observability-crd-status

|Property|Description|
|---|---|
|Unique ID|observability-crd-status|
|Description|Checks that all CRDs have a status sub-resource specification (Spec.versions[].Schema.OpenAPIV3Schema.Properties[“status”]).|
|Suggested Remediation|Ensure that all the CRDs have a meaningful status specification (Spec.versions[].Schema.OpenAPIV3Schema.Properties[“status”]).|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements|
|Exception Process|No exceptions|
|Impact Statement|Missing status subresources prevent proper monitoring and automation based on custom resource states.|
|Tags|common,observability|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### observability-pod-disruption-budget

|Property|Description|
|---|---|
|Unique ID|observability-pod-disruption-budget|
|Description|Checks to see if pod disruption budgets have allowed values for minAvailable and maxUnavailable|
|Suggested Remediation|Ensure minAvailable is not zero and maxUnavailable does not equal the number of pods in the replica|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-upgrade-expectations|
|Exception Process|No exceptions|
|Impact Statement|Improper disruption budgets can prevent necessary maintenance operations or allow too many pods to be disrupted simultaneously.|
|Tags|common,observability|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### observability-termination-policy

|Property|Description|
|---|---|
|Unique ID|observability-termination-policy|
|Description|Check that all containers are using terminationMessagePolicy: FallbackToLogsOnError. There are different ways a pod can stop on an OpenShift cluster. One way is that the pod can remain alive but non-functional. Another way is that the pod can crash and become non-functional. In the first case, if the administrator has implemented liveness and readiness checks, OpenShift can stop the pod and either restart it on the same node or a different node in the cluster. For the second case, when the application in the pod stops, it should exit with a code and write suitable log entries to help the administrator diagnose what the issue was that caused the problem.|
|Suggested Remediation|Ensure containers are all using FallbackToLogsOnError in terminationMessagePolicy|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-pod-exit-status|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Incorrect termination message policies can prevent proper error reporting and make failure diagnosis difficult.|
|Tags|telco,observability|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

### operator

#### operator-catalogsource-bundle-count

|Property|Description|
|---|---|
|Unique ID|operator-catalogsource-bundle-count|
|Description|Tests operator catalog source bundle count is less than 1000|
|Suggested Remediation|Ensure that the Operator's catalog source has a valid bundle count less than 1000.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements|
|Exception Process|No exceptions|
|Impact Statement|Large catalog sources can cause performance issues, slow operator resolution, and increase cluster resource usage.|
|Tags|common,operator|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### operator-crd-openapi-schema

|Property|Description|
|---|---|
|Unique ID|operator-crd-openapi-schema|
|Description|Tests whether an application Operator CRD is defined with OpenAPI spec.|
|Suggested Remediation|Ensure that the Operator CRD is defined with OpenAPI spec.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements|
|Exception Process|No exceptions|
|Impact Statement|Missing OpenAPI schemas prevent proper validation and can lead to configuration errors and runtime failures.|
|Tags|common,operator|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### operator-crd-versioning

|Property|Description|
|---|---|
|Unique ID|operator-crd-versioning|
|Description|Tests whether the Operator CRD has a valid versioning.|
|Suggested Remediation|Ensure that the Operator CRD has a valid version.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements|
|Exception Process|No exceptions|
|Impact Statement|Invalid CRD versioning can cause API compatibility issues and prevent proper schema evolution.|
|Tags|common,operator|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### operator-install-source

|Property|Description|
|---|---|
|Unique ID|operator-install-source|
|Description|Tests whether a workload Operator is installed via OLM.|
|Suggested Remediation|Ensure that your Operator is installed via OLM.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements|
|Exception Process|No exceptions|
|Impact Statement|Non-OLM operators bypass lifecycle management and dependency resolution, creating operational complexity and update issues.|
|Tags|common,operator|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### operator-install-status-no-privileges

|Property|Description|
|---|---|
|Unique ID|operator-install-status-no-privileges|
|Description|Checks whether the operator needs access to Security Context Constraints. Test passes if clusterPermissions is not present in the CSV manifest or is present with no RBAC rules related to SCCs.|
|Suggested Remediation|Ensure all the workload's operators have no privileges on cluster resources.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements|
|Exception Process|No exceptions|
|Impact Statement|Operators with SCC access have elevated privileges that can compromise cluster security and violate security policies.|
|Tags|common,operator|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### operator-install-status-succeeded

|Property|Description|
|---|---|
|Unique ID|operator-install-status-succeeded|
|Description|Ensures that the target workload operators report "Succeeded" as their installation status.|
|Suggested Remediation|Ensure all the workload's operators have been successfully installed by OLM.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements|
|Exception Process|No exceptions|
|Impact Statement|Failed operator installations can leave applications in incomplete states, causing functionality gaps and operational issues.|
|Tags|common,operator|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### operator-multiple-same-operators

|Property|Description|
|---|---|
|Unique ID|operator-multiple-same-operators|
|Description|Tests whether multiple instances of the same Operator CSV are installed.|
|Suggested Remediation|Ensure that only one Operator of the same type is installed in the cluster.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements|
|Exception Process|No exceptions|
|Impact Statement|Multiple operator instances can cause conflicts, resource contention, and unpredictable behavior.|
|Tags|common,operator|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### operator-olm-skip-range

|Property|Description|
|---|---|
|Unique ID|operator-olm-skip-range|
|Description|Test that checks the operator has a valid olm skip range.|
|Suggested Remediation|Ensure that the Operator has a valid OLM skip range. If the operator does not have another version to "skip", then ignore the result of this test.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements|
|Exception Process|If there is not a version of the operator that needs to be skipped, then an exception will be granted.|
|Impact Statement|Invalid skip ranges can prevent proper operator upgrades and cause version compatibility issues.|
|Tags|common,operator|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### operator-pods-no-hugepages

|Property|Description|
|---|---|
|Unique ID|operator-pods-no-hugepages|
|Description|Tests that the pods do not have hugepages enabled.|
|Suggested Remediation|Ensure that the pods are not using hugepages|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements|
|Exception Process|No exceptions|
|Impact Statement|Hugepage usage by operators can interfere with application hugepage allocation and cause resource contention.|
|Tags|common,operator|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### operator-semantic-versioning

|Property|Description|
|---|---|
|Unique ID|operator-semantic-versioning|
|Description|Tests whether an application Operator has a valid semantic versioning.|
|Suggested Remediation|Ensure that the Operator has a valid semantic versioning.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements|
|Exception Process|No exceptions|
|Impact Statement|Invalid semantic versioning prevents proper upgrade paths and dependency management, causing operational issues.|
|Tags|common,operator|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### operator-single-crd-owner

|Property|Description|
|---|---|
|Unique ID|operator-single-crd-owner|
|Description|Tests whether a CRD is owned by a single Operator.|
|Suggested Remediation|Ensure that a CRD is owned by only one Operator|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements|
|Exception Process|No exceptions|
|Impact Statement|Multiple CRD owners can cause conflicts, inconsistent behavior, and management complexity.|
|Tags|common,operator|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### operator-single-or-multi-namespaced-allowed-in-tenant-namespaces

|Property|Description|
|---|---|
|Unique ID|operator-single-or-multi-namespaced-allowed-in-tenant-namespaces|
|Description|Verifies that only single/multi namespaced operators are installed in a tenant-dedicated namespace. The test fails if this namespace contains any installed operator with Own/All-namespaced install mode, unlabeled operators, operands of any operator installed elsewhere, or pods unrelated to any operator.|
|Suggested Remediation|Ensure that operator with install mode SingleNamespaced or MultiNamespaced only is installed in the tenant namespace. Any installed operator with different install mode (AllNamespaced or OwnNamespaced) or pods not belonging to any operator must not be present in this namespace.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements|
|Exception Process|No exceptions|
|Impact Statement|Improperly scoped operators can violate tenant isolation and create unauthorized cross-namespace access.|
|Tags|extended,operator|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

### performance

#### performance-cpu-pinning-no-exec-probes

|Property|Description|
|---|---|
|Unique ID|performance-cpu-pinning-no-exec-probes|
|Description|Workloads utilizing CPU pinning (Guaranteed QoS with exclusive CPUs) should not use exec probes. Exec probes run a command within the container, which could interfere with latency-sensitive workloads and cause performance degradation.|
|Suggested Remediation|Workloads that use CPU pinning (Guaranteed QoS with exclusive CPUs) should not use exec probes. Use httpGet or tcpSocket probes instead, as exec probes can interfere with latency-sensitive workloads requiring non-interruptible task execution.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cpu-manager-pinning|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Exec probes on workloads with CPU pinning (exclusive CPUs) can cause performance degradation, interrupt latency-sensitive operations, and potentially crash applications due to resource contention. Any workload requiring exclusive CPUs inherently needs non-interruptible task execution.|
|Tags|telco,performance|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Mandatory|

#### performance-exclusive-cpu-pool

|Property|Description|
|---|---|
|Unique ID|performance-exclusive-cpu-pool|
|Description|Ensures that if one container in a Pod selects an exclusive CPU pool the rest select the same type of CPU pool|
|Suggested Remediation|Ensure that if one container in a Pod selects an exclusive CPU pool the rest also select this type of CPU pool|
|Best Practice Reference|No Doc Link - Far Edge|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Inconsistent CPU pool selection can cause performance interference and unpredictable latency in real-time applications.|
|Tags|faredge,performance|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Optional|

#### performance-exclusive-cpu-pool-rt-scheduling-policy

|Property|Description|
|---|---|
|Unique ID|performance-exclusive-cpu-pool-rt-scheduling-policy|
|Description|Ensures that if application workload runs in exclusive CPU pool, it chooses RT CPU schedule policy and set the priority less than 10.|
|Suggested Remediation|Ensure that the workload running in Application exclusive CPU pool can choose RT CPU scheduling policy, but should set priority less than 10|
|Best Practice Reference|No Doc Link - Far Edge|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Wrong scheduling policies in exclusive CPU pools can prevent real-time applications from meeting latency requirements.|
|Tags|faredge,performance|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Optional|

#### performance-isolated-cpu-pool-rt-scheduling-policy

|Property|Description|
|---|---|
|Unique ID|performance-isolated-cpu-pool-rt-scheduling-policy|
|Description|Ensures that a workload running in an application-isolated exclusive CPU pool selects a RT CPU scheduling policy|
|Suggested Remediation|Ensure that the workload running in an application-isolated exclusive CPU pool selects a RT CPU scheduling policy (such as SCHED_FIFO/SCHED_RR) with High priority.|
|Best Practice Reference|No Doc Link - Far Edge|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Incorrect scheduling policies in isolated CPU pools can cause performance degradation and violate real-time guarantees.|
|Tags|faredge,performance|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Optional|

#### performance-max-resources-exec-probes

|Property|Description|
|---|---|
|Unique ID|performance-max-resources-exec-probes|
|Description|Checks that less than 10 exec probes are configured in the cluster for this workload. Also checks that the periodSeconds parameter for each probe is superior or equal to 10.|
|Suggested Remediation|Reduce the number of exec probes in the cluster for this workload to less than 10. Increase the update period of the exec probe to be superior or equal to 10 seconds.|
|Best Practice Reference|No Doc Link - Far Edge|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Excessive exec probes can overwhelm system resources, degrade performance, and interfere with critical application operations in resource-constrained environments.|
|Tags|faredge,performance|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### performance-rt-apps-no-exec-probes

|Property|Description|
|---|---|
|Unique ID|performance-rt-apps-no-exec-probes|
|Description|Ensures that if one container runs a real time application exec probes are not used|
|Suggested Remediation|Ensure that if one container runs a real time application exec probes are not used|
|Best Practice Reference|No Doc Link - Far Edge|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Exec probes on real-time applications can cause latency spikes and interrupt time-critical operations.|
|Tags|faredge,performance|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Optional|

#### performance-shared-cpu-pool-non-rt-scheduling-policy

|Property|Description|
|---|---|
|Unique ID|performance-shared-cpu-pool-non-rt-scheduling-policy|
|Description|Ensures that if application workload runs in shared CPU pool, it chooses non-RT CPU schedule policy to always share the CPU with other applications and kernel threads.|
|Suggested Remediation|Ensure that the workload running in Application shared CPU pool should choose non-RT CPU schedule policy, like SCHED _OTHER to always share the CPU with other applications and kernel threads.|
|Best Practice Reference|No Doc Link - Far Edge|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Incorrect scheduling policies in shared CPU pools can cause performance interference and unfair resource distribution.|
|Tags|faredge,performance|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Optional|

### platform-alteration

#### platform-alteration-base-image

|Property|Description|
|---|---|
|Unique ID|platform-alteration-base-image|
|Description|Ensures that the Container Base Image is not altered post-startup. This test is a heuristic, and ensures that there are no changes to the following directories: 1) /var/lib/rpm 2) /var/lib/dpkg 3) /bin 4) /sbin 5) /lib 6) /lib64 7) /usr/bin 8) /usr/sbin 9) /usr/lib 10) /usr/lib64|
|Suggested Remediation|Ensure that Container applications do not modify the Container Base Image. In particular, ensure that the following directories are not modified: 1) /var/lib/rpm 2) /var/lib/dpkg 3) /bin 4) /sbin 5) /lib 6) /lib64 7) /usr/bin 8) /usr/sbin 9) /usr/lib 10) /usr/lib64 Ensure that all required binaries are built directly into the container image, and are not installed post startup.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-image-standards|
|Exception Process|No exceptions|
|Impact Statement|Modified base images can introduce security vulnerabilities, create inconsistent behavior, and violate immutable infrastructure principles.|
|Tags|common,platform-alteration|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### platform-alteration-boot-params

|Property|Description|
|---|---|
|Unique ID|platform-alteration-boot-params|
|Description|Tests that boot parameters are set through the MachineConfigOperator, and not set manually on the Node.|
|Suggested Remediation|Ensure that boot parameters are set directly through the MachineConfigOperator, or indirectly through the PerformanceAddonOperator. Boot parameters should not be changed directly through the Node, as OpenShift should manage the changes for you.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-host-os|
|Exception Process|No exceptions|
|Impact Statement|Manual boot parameter changes bypass cluster configuration management and can cause node instability and configuration drift.|
|Tags|common,platform-alteration|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### platform-alteration-cluster-operator-health

|Property|Description|
|---|---|
|Unique ID|platform-alteration-cluster-operator-health|
|Description|Tests that all cluster operators are healthy.|
|Suggested Remediation|Ensure each cluster operator is in an 'Available' state. If an operator is not in an 'Available' state, investigate the operator's logs and events to determine the cause of the failure.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements|
|Exception Process|No exceptions|
|Impact Statement|Unhealthy cluster operators can cause platform instability, feature failures, and degraded cluster functionality.|
|Tags|common,platform-alteration|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### platform-alteration-hugepages-1g-only

|Property|Description|
|---|---|
|Unique ID|platform-alteration-hugepages-1g-only|
|Description|Check that pods using hugepages only use 1Gi size|
|Suggested Remediation|Modify pod to consume 1Gi hugepages only|
|Best Practice Reference|No Doc Link - Far Edge|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Incorrect hugepage configuration can lead to memory fragmentation and application startup failures in memory-constrained environments.|
|Tags|faredge,platform-alteration|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Mandatory|
|Non-Telco|Optional|
|Telco|Optional|

#### platform-alteration-hugepages-2m-only

|Property|Description|
|---|---|
|Unique ID|platform-alteration-hugepages-2m-only|
|Description|Check that pods using hugepages only use 2Mi size|
|Suggested Remediation|Modify pod to consume 2Mi hugepages only|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-huge-pages|
|Exception Process|No exception needed for optional/extended tests.|
|Impact Statement|Using inappropriate hugepage sizes can cause memory allocation failures and reduce overall system performance and stability.|
|Tags|extended,platform-alteration|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### platform-alteration-hugepages-config

|Property|Description|
|---|---|
|Unique ID|platform-alteration-hugepages-config|
|Description|Checks to see that HugePage settings have been configured through MachineConfig, and not manually on the underlying Node. This test case applies only to Nodes that are labeled as workers with the standard label "node-role.kubernetes.io/worker". First, the MachineConfig is inspected for hugepage settings in systemd units. If not, the MC's .spec.kernelArguments are inspected for hugepage settings. The sizes and page numbers are compared, and the test passes only if they are the same than then ones in node's /sys/kernel/mm/hugepages/hugepages-X folders.|
|Suggested Remediation|HugePage settings for worker nodes must be configured either directly through the MachineConfigOperator or indirectly using the PerformanceAddonOperator. Avoid making changes directly to an underlying Node, and let OpenShift handle the heavy lifting of configuring advanced settings.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-huge-pages|
|Exception Process|No exceptions|
|Impact Statement|Manual hugepage configuration bypasses cluster management, can cause node instability, and creates configuration drift issues.|
|Tags|common,platform-alteration|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### platform-alteration-hyperthread-enable

|Property|Description|
|---|---|
|Unique ID|platform-alteration-hyperthread-enable|
|Description|Check that baremetal workers have hyperthreading enabled|
|Suggested Remediation|Check that baremetal workers have hyperthreading enabled|
|Best Practice Reference|No Doc Link - Extended|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Disabled hyperthreading reduces CPU performance and can affect workload scheduling and resource utilization efficiency.|
|Tags|extended,platform-alteration|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### platform-alteration-is-selinux-enforcing

|Property|Description|
|---|---|
|Unique ID|platform-alteration-is-selinux-enforcing|
|Description|verifies that all openshift platform/cluster nodes have selinux in "Enforcing" mode.|
|Suggested Remediation|Configure selinux and enable enforcing mode.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-pod-security|
|Exception Process|No exceptions|
|Impact Statement|Non-enforcing SELinux reduces security isolation and can allow privilege escalation attacks and unauthorized resource access.|
|Tags|common,platform-alteration|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### platform-alteration-isredhat-release

|Property|Description|
|---|---|
|Unique ID|platform-alteration-isredhat-release|
|Description|verifies if the container base image is redhat.|
|Suggested Remediation|Build a new container image that is based on UBI (Red Hat Universal Base Image).|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security|
|Exception Process|No exceptions|
|Impact Statement|Non-Red Hat base images may lack security updates, enterprise support, and compliance certifications required for production use.|
|Tags|common,platform-alteration|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### platform-alteration-ocp-lifecycle

|Property|Description|
|---|---|
|Unique ID|platform-alteration-ocp-lifecycle|
|Description|Tests that the running OCP version is not end of life.|
|Suggested Remediation|Please update your cluster to a version that is generally available.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-k8s|
|Exception Process|No exceptions|
|Impact Statement|End-of-life OpenShift versions lack security updates and support, creating significant security and operational risks.|
|Tags|common,platform-alteration|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### platform-alteration-ocp-node-os-lifecycle

|Property|Description|
|---|---|
|Unique ID|platform-alteration-ocp-node-os-lifecycle|
|Description|Tests that the nodes running in the cluster have operating systems that are compatible with the deployed version of OpenShift.|
|Suggested Remediation|Please update your workers to a version that is supported by your version of OpenShift|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-host-os|
|Exception Process|No exceptions|
|Impact Statement|Incompatible node operating systems can cause stability issues, security vulnerabilities, and lack of vendor support.|
|Tags|common,platform-alteration|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### platform-alteration-service-mesh-usage

|Property|Description|
|---|---|
|Unique ID|platform-alteration-service-mesh-usage|
|Description|Checks if the istio namespace ("istio-system") is present. If it is present, checks that the istio sidecar is present in all pods under test.|
|Suggested Remediation|Ensure all the workload pods are using service mesh if the cluster provides it.|
|Best Practice Reference|No Doc Link - Extended|
|Exception Process|No exception needed for optional/extended tests.|
|Impact Statement|Inconsistent service mesh configuration can create security gaps, monitoring blind spots, and traffic management issues.|
|Tags|extended,platform-alteration|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### platform-alteration-sysctl-config

|Property|Description|
|---|---|
|Unique ID|platform-alteration-sysctl-config|
|Description|Tests that no one has changed the node's sysctl configs after the node was created, the tests works by checking if the sysctl configs are consistent with the MachineConfig CR which defines how the node should be configured|
|Suggested Remediation|You should recreate the node or change the sysctls, recreating is recommended because there might be other unknown changes|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security|
|Exception Process|No exceptions|
|Impact Statement|Manual sysctl modifications can cause system instability, security vulnerabilities, and unpredictable kernel behavior.|
|Tags|common,platform-alteration|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

#### platform-alteration-tainted-node-kernel

|Property|Description|
|---|---|
|Unique ID|platform-alteration-tainted-node-kernel|
|Description|Ensures that the Node(s) hosting workloads do not utilize tainted kernels. This test case is especially important to support Highly Available workloads, since when a workload is re-instantiated on a backup Node, that Node's kernel may not have the same hacks.'|
|Suggested Remediation|Test failure indicates that the underlying Node's kernel is tainted. Ensure that you have not altered underlying Node(s) kernels in order to run the workload.|
|Best Practice Reference|https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations|
|Exception Process|If taint is necessary, document details of the taint and why it's needed by workload or environment.|
|Impact Statement|Tainted kernels indicate unauthorized modifications that can introduce instability, security vulnerabilities, and support issues.|
|Tags|common,platform-alteration|
|**Scenario**|**Optional/Mandatory**|
|Extended|Mandatory|
|Far-Edge|Mandatory|
|Non-Telco|Mandatory|
|Telco|Mandatory|

### preflight

#### preflight-AllImageRefsInRelatedImages

|Property|Description|
|---|---|
|Unique ID|preflight-AllImageRefsInRelatedImages|
|Description|Check that all images in the CSV are listed in RelatedImages section. Currently, this check is not enforced.|
|Suggested Remediation|Either manually or with a tool, populate the RelatedImages section of the CSV|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Missing or incorrect image references in related images can cause deployment failures and broken operator functionality.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-BasedOnUbi

|Property|Description|
|---|---|
|Unique ID|preflight-BasedOnUbi|
|Description|Checking if the container's base image is based upon the Red Hat Universal Base Image (UBI)|
|Suggested Remediation|Change the FROM directive in your Dockerfile or Containerfile, for the latest list of images and details refer to: https://catalog.redhat.com/software/base-images|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Non-UBI base images may lack security updates, enterprise support, and compliance certifications required for production use.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-BundleImageRefsAreCertified

|Property|Description|
|---|---|
|Unique ID|preflight-BundleImageRefsAreCertified|
|Description|Checking that all images referenced in the CSV are certified. Currently, this check is not enforced.|
|Suggested Remediation|Ensure that any images referenced in the CSV, including the relatedImages section, have been certified.|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Uncertified bundle image references can introduce security vulnerabilities and compatibility issues in production deployments.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-DeployableByOLM

|Property|Description|
|---|---|
|Unique ID|preflight-DeployableByOLM|
|Description|Checking if the operator could be deployed by OLM|
|Suggested Remediation|Follow the guidelines on the operator-sdk website to learn how to package your operator https://sdk.operatorframework.io/docs/olm-integration/cli-overview/|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Operators not deployable by OLM cannot be properly managed, updated, or integrated into OpenShift lifecycle management.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-FollowsRestrictedNetworkEnablementGuidelines

|Property|Description|
|---|---|
|Unique ID|preflight-FollowsRestrictedNetworkEnablementGuidelines|
|Description|Checks for indicators that this bundle has implemented guidelines to indicate readiness for running in a disconnected cluster, or a cluster with a restricted network.|
|Suggested Remediation|If consumers of your operator may need to do so on a restricted network, implement the guidelines outlined in OCP documentation: https://docs.redhat.com/en/documentation/openshift_container_platform/latest/html/disconnected_environments/olm-restricted-networks|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Non-compliance with restricted network guidelines can prevent deployment in air-gapped environments and violate security policies.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-HasLicense

|Property|Description|
|---|---|
|Unique ID|preflight-HasLicense|
|Description|Checking if terms and conditions applicable to the software including open source licensing information are present. The license must be at /licenses|
|Suggested Remediation|Create a directory named /licenses and include all relevant licensing and/or terms and conditions as text file(s) in that directory.|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Missing license information can create legal compliance issues and prevent proper software asset management.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-HasModifiedFiles

|Property|Description|
|---|---|
|Unique ID|preflight-HasModifiedFiles|
|Description|Checks that no files installed via RPM in the base Red Hat layer have been modified|
|Suggested Remediation|Do not modify any files installed by RPM in the base Red Hat layer|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Modified files in containers can introduce security vulnerabilities, create inconsistent behavior, and violate immutable infrastructure principles.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-HasNoProhibitedLabels

|Property|Description|
|---|---|
|Unique ID|preflight-HasNoProhibitedLabels|
|Description|Checking if the labels (name, vendor, maintainer) violate Red Hat trademark.|
|Suggested Remediation|Ensure the name, vendor, and maintainer label on your image do not violate the Red Hat trademark.|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Misuse of Red Hat trademarks in name, vendor, or maintainer labels creates legal and compliance risks that can block certification and publication.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-HasNoProhibitedPackages

|Property|Description|
|---|---|
|Unique ID|preflight-HasNoProhibitedPackages|
|Description|Checks to ensure that the image in use does not include prohibited packages, such as Red Hat Enterprise Linux (RHEL) kernel packages.|
|Suggested Remediation|Remove any RHEL packages that are not distributable outside of UBI|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Prohibited packages can introduce security vulnerabilities, licensing issues, and compliance violations.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-HasProhibitedContainerName

|Property|Description|
|---|---|
|Unique ID|preflight-HasProhibitedContainerName|
|Description|Checking if the container-name violates Red Hat trademark.|
|Suggested Remediation|Update container-name ie (quay.io/repo-name/container-name) to not violate Red Hat trademark.|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Prohibited container names can cause conflicts with system components and violate naming conventions.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-HasRequiredLabel

|Property|Description|
|---|---|
|Unique ID|preflight-HasRequiredLabel|
|Description|Checking if the required labels (name, vendor, version, release, summary, description, maintainer) are present in the container metadata|
|Suggested Remediation|Add the following labels to your Dockerfile or Containerfile: name, vendor, version, release, summary, description, maintainer.|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Missing required labels prevent proper metadata management and can cause deployment and management issues.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-HasUniqueTag

|Property|Description|
|---|---|
|Unique ID|preflight-HasUniqueTag|
|Description|Checking if container has a tag other than 'latest', so that the image can be uniquely identified.|
|Suggested Remediation|Add a tag to your image. Consider using Semantic Versioning. https://semver.org/|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Non-unique tags can cause version conflicts and deployment inconsistencies, making rollbacks and troubleshooting difficult.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-LayerCountAcceptable

|Property|Description|
|---|---|
|Unique ID|preflight-LayerCountAcceptable|
|Description|Checking if container has less than 40 layers.  Too many layers within the container images can degrade container performance.|
|Suggested Remediation|Optimize your Dockerfile to consolidate and minimize the number of layers. Each RUN command will produce a new layer. Try combining RUN commands using && where possible.|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Excessive image layers can cause poor performance, increased storage usage, and longer deployment times.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-RequiredAnnotations

|Property|Description|
|---|---|
|Unique ID|preflight-RequiredAnnotations|
|Description|Checks that the CSV has all of the required feature annotations.|
|Suggested Remediation|Add all of the required annotations, and make sure the value is set to either 'true' or 'false'|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Missing required annotations can prevent proper operator lifecycle management and cause deployment failures.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-RunAsNonRoot

|Property|Description|
|---|---|
|Unique ID|preflight-RunAsNonRoot|
|Description|Checking if container runs as the root user because a container that does not specify a non-root user will fail the automatic certification, and will be subject to a manual review before the container can be approved for publication|
|Suggested Remediation|Indicate a specific USER in the dockerfile or containerfile|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Running containers as root increases the blast radius of security vulnerabilities and can lead to full host compromise if containers are breached.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-ScorecardBasicSpecCheck

|Property|Description|
|---|---|
|Unique ID|preflight-ScorecardBasicSpecCheck|
|Description|Check to make sure that all CRs have a spec block.|
|Suggested Remediation|Make sure that all CRs have a spec block|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Failing basic scorecard checks indicates fundamental operator implementation issues that can cause runtime failures.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-ScorecardOlmSuiteCheck

|Property|Description|
|---|---|
|Unique ID|preflight-ScorecardOlmSuiteCheck|
|Description|Operator-sdk scorecard OLM Test Suite Check|
|Suggested Remediation|See scorecard output for details, artifacts/operator_bundle_scorecard_OlmSuiteCheck.json|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Failing OLM suite checks indicates operator lifecycle management issues that can prevent proper installation and updates.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-SecurityContextConstraintsInCSV

|Property|Description|
|---|---|
|Unique ID|preflight-SecurityContextConstraintsInCSV|
|Description|Evaluates the csv and logs a message if a non default security context constraint is needed by the operator|
|Suggested Remediation|If no scc is detected the default restricted scc will be used.|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Incorrect SCC definitions in CSV can cause security policy violations and deployment failures.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

#### preflight-ValidateOperatorBundle

|Property|Description|
|---|---|
|Unique ID|preflight-ValidateOperatorBundle|
|Description|Validating Bundle image that checks if it can validate the content and format of the operator bundle|
|Suggested Remediation|Valid bundles are defined by bundle spec, so make sure that this bundle conforms to that spec. More Information: https://github.com/operator-framework/operator-registry/blob/master/docs/design/operator-bundle.md|
|Best Practice Reference|No Doc Link|
|Exception Process|There is no documented exception process for this.|
|Impact Statement|Invalid operator bundles can cause deployment failures, update issues, and operational instability.|
|Tags|common,preflight|
|**Scenario**|**Optional/Mandatory**|
|Extended|Optional|
|Far-Edge|Optional|
|Non-Telco|Optional|
|Telco|Optional|

## Security Context Categories

Security context categories referred here are applicable to the [access control test case](#access-control-security-context).

### 1st Category
Default SCC for all users if namespace does not use service mesh.

Workloads under this category should: 
 - Use default CNI (OVN) network interface
 - Not request NET_ADMIN or NET_RAW for advanced networking functions

### 2nd Category
For workloads which utilize Service Mesh sidecars for mTLS or load balancing. These workloads must utilize an alternative SCC “restricted-no-uid0” to workaround a service mesh UID limitation. Workloads under this category should not run as root (UID0).

### 3rd Category
For workloads with advanced networking functions/requirements (e.g. CAP_NET_RAW, CAP_NET_ADMIN, may run as root).

For example:
  - Manipulate the low-level protocol flags, such as the 802.1p priority, VLAN tag, DSCP value, etc.
  - Manipulate the interface IP addresses or the routing table or the firewall rules on-the-fly.
  - Process Ethernet packets
Workloads under this category may
  - Use Macvlan interface to sending and receiving Ethernet packets
  - Request CAP_NET_RAW for creating raw sockets
  - Request CAP_NET_ADMIN for
    - Modify the interface IP address on-the-fly
    - Manipulating the routing table on-the-fly
    - Manipulating firewall rules on-the-fly
    - Setting packet DSCP value

### 4th Category
For workloads handling user plane traffic or latency-sensitive payloads at line rate, such as load balancing, routing, deep packet inspection etc. Workloads under this category may also need to process the packets at a lower level.

These workloads shall 
  - Use SR-IOV interfaces 
  - Fully or partially bypassing kernel networking stack with userspace networking technologies,such as DPDK, F-stack, VPP, OpenFastPath, etc. A userspace networking stack not only improvesthe performance but also reduces the need for CAP_NET_ADMIN and CAP_NET_RAW.
CAP_IPC_LOCK is mandatory for allocating hugepage memory, hence shall be granted to DPDK applications. If the workload is latency-sensitive and needs a real-time kernel, CAP_SYS_NICE would be required.
