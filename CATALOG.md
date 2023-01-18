# cnf-certification-test catalog

The catalog for cnf-certification-test contains a variety of `Test Cases`, which are traditional JUnit testcases specified internally using `Ginkgo.It`.
## Test Case Catalog

Test Cases are the specifications used to perform a meaningful test.  Test cases may run once, or several times against several targets.  CNF Certification includes a number of normative and informative tests to ensure CNFs follow best practices.  Here is the list of available Test Cases:

### access-control

#### cluster-role-bindings

Property|Description
---|---
Test Case Name|cluster-role-bindings
Test Case Label|access-control-cluster-role-bindings
Unique ID|http://test-network-function.com/testcases/access-control/cluster-role-bindings
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/cluster-role-bindings Tests that a Pod does not specify ClusterRoleBindings.
Result Type|normative
Suggested Remediation|In most cases, Pod's should not have ClusterRoleBindings.  The suggested remediation is to remove the need for 	ClusterRoleBindings, if possible.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2.10 and 5.3.6
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### container-host-port

Property|Description
---|---
Test Case Name|container-host-port
Test Case Label|access-control-container-host-port
Unique ID|http://test-network-function.com/testcases/access-control/container-host-port
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/container-host-port Verifies if containers define a hostPort.
Result Type|normative
Suggested Remediation|Remove hostPort configuration from the container
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.3.6
Exception Process|There is no documented exception process for this.
Tags|common
#### ipc-lock-capability-check

Property|Description
---|---
Test Case Name|ipc-lock-capability-check
Test Case Label|access-control-ipc-lock-capability-check
Unique ID|http://test-network-function.com/testcases/access-control/ipc-lock-capability-check
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/ipc-lock-capability-check Ensures that containers do not use IPC_LOCK capability
Result Type|normative
Suggested Remediation|Change the security context to be one of the 4 that are allowed on the documentation section 4.5
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|Identify the pod that is needing special capabilities and document why
Tags|common
#### namespace

Property|Description
---|---
Test Case Name|namespace
Test Case Label|access-control-namespace
Unique ID|http://test-network-function.com/testcases/access-control/namespace
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/namespace Tests that all CNF's resources (PUTs and CRs) belong to valid namespaces. A valid namespace meets the following conditions: (1) It was declared in the yaml config file under the targetNameSpaces tag. (2) It doesn't have any of the following prefixes: default, openshift-, istio- and aspenmesh-
Result Type|normative
Suggested Remediation|Ensure that your CNF utilizes namespaces declared in the yaml config file. Additionally, 	the namespaces should not start with "default, openshift-, istio- or aspenmesh-".
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2, 16.3.8 and 16.3.9
Exception Process|There is no documented exception process for this.
Tags|common
#### namespace-resource-quota

Property|Description
---|---
Test Case Name|namespace-resource-quota
Test Case Label|access-control-namespace-resource-quota
Unique ID|http://test-network-function.com/testcases/access-control/namespace-resource-quota
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/namespace-resource-quota Checks to see if CNF workload pods are running in namespaces that have resource quotas applied.
Result Type|normative
Suggested Remediation|Apply a ResourceQuota to the namespace your CNF is running in
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 4.6.8
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### net-admin-capability-check

Property|Description
---|---
Test Case Name|net-admin-capability-check
Test Case Label|access-control-net-admin-capability-check
Unique ID|http://test-network-function.com/testcases/access-control/net-admin-capability-check
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/net-admin-capability-check Ensures that containers do not use NET_ADMIN capability
Result Type|normative
Suggested Remediation|Change the security context to be one of the 4 that are allowed on the documentation section 4.5
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|Identify the pod that is needing special capabilities and document why
Tags|common
#### net-raw-capability-check

Property|Description
---|---
Test Case Name|net-raw-capability-check
Test Case Label|access-control-net-raw-capability-check
Unique ID|http://test-network-function.com/testcases/access-control/net-raw-capability-check
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/net-raw-capability-check Ensures that containers do not use NET_RAW capability
Result Type|normative
Suggested Remediation|Change the security context to be one of the 4 that are allowed on the documentation section 4.5
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|Identify the pod that is needing special capabilities and document why
Tags|common
#### no-1337-uid

Property|Description
---|---
Test Case Name|no-1337-uid
Test Case Label|access-control-no-1337-uid
Unique ID|http://test-network-function.com/testcases/access-control/no-1337-uid
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/no-1337-uid Checks that all pods are not using the securityContext UID 1337
Result Type|informative
Suggested Remediation|Use another process UID that is not 1337
Best Practice Reference|https://TODO Section 4.6.24
Exception Process|There is no documented exception process for this.
Tags|extended
#### one-process-per-container

Property|Description
---|---
Test Case Name|one-process-per-container
Test Case Label|access-control-one-process-per-container
Unique ID|http://test-network-function.com/testcases/access-control/one-process-per-container
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/one-process-per-container Check that all containers under test have only one process running
Result Type|informative
Suggested Remediation|Launch only one process per container
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 10.8.3
Exception Process|There is no documented exception process for this.
Tags|common
#### pod-automount-service-account-token

Property|Description
---|---
Test Case Name|pod-automount-service-account-token
Test Case Label|access-control-pod-automount-service-account-token
Unique ID|http://test-network-function.com/testcases/access-control/pod-automount-service-account-token
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/pod-automount-service-account-token Check that all pods under test have automountServiceAccountToken set to false. Only pods that require access to the kubernetes API server should have automountServiceAccountToken set to true
Result Type|informative
Suggested Remediation|Check that pod has automountServiceAccountToken set to false or pod is attached to service account which has automountServiceAccountToken set to false, unless the pod needs access to the kubernetes API server.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 12.7
Exception Process|There is no documented exception process for this.
Tags|common
#### pod-host-ipc

Property|Description
---|---
Test Case Name|pod-host-ipc
Test Case Label|access-control-pod-host-ipc
Unique ID|http://test-network-function.com/testcases/access-control/pod-host-ipc
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/pod-host-ipc Verifies that the spec.HostIpc parameter is set to false
Result Type|normative
Suggested Remediation|Set the spec.HostIpc parameter to false in the pod configuration
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.3.6
Exception Process|There is no documented exception process for this.
Tags|common
#### pod-host-network

Property|Description
---|---
Test Case Name|pod-host-network
Test Case Label|access-control-pod-host-network
Unique ID|http://test-network-function.com/testcases/access-control/pod-host-network
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/pod-host-network Verifies that the spec.HostNetwork parameter is set to false
Result Type|normative
Suggested Remediation|Set the spec.HostNetwork parameter to false in the pod configuration
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.3.6
Exception Process|There is no documented exception process for this.
Tags|common
#### pod-host-path

Property|Description
---|---
Test Case Name|pod-host-path
Test Case Label|access-control-pod-host-path
Unique ID|http://test-network-function.com/testcases/access-control/pod-host-path
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/pod-host-path Verifies that the spec.HostPath parameter is not set (not present)
Result Type|normative
Suggested Remediation|Set the spec.HostNetwork parameter to false in the pod configuration
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.3.6
Exception Process|There is no documented exception process for this.
Tags|common
#### pod-host-pid

Property|Description
---|---
Test Case Name|pod-host-pid
Test Case Label|access-control-pod-host-pid
Unique ID|http://test-network-function.com/testcases/access-control/pod-host-pid
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/pod-host-pid Verifies that the spec.HostPid parameter is set to false
Result Type|normative
Suggested Remediation|Set the spec.HostPid parameter to false in the pod configuration
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.3.6
Exception Process|There is no documented exception process for this.
Tags|common
#### pod-role-bindings

Property|Description
---|---
Test Case Name|pod-role-bindings
Test Case Label|access-control-pod-role-bindings
Unique ID|http://test-network-function.com/testcases/access-control/pod-role-bindings
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/pod-role-bindings Ensures that a CNF does not utilize RoleBinding(s) in a non-CNF Namespace.
Result Type|normative
Suggested Remediation|Ensure the CNF is not configured to use RoleBinding(s) in a non-CNF Namespace.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.3.3 and 5.3.5
Exception Process|There is no documented exception process for this.
Tags|common
#### pod-service-account

Property|Description
---|---
Test Case Name|pod-service-account
Test Case Label|access-control-pod-service-account
Unique ID|http://test-network-function.com/testcases/access-control/pod-service-account
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/pod-service-account Tests that each CNF Pod utilizes a valid Service Account.
Result Type|normative
Suggested Remediation|Ensure that the each CNF Pod is configured to use a valid Service Account
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2.3 and 5.2.7
Exception Process|There is no documented exception process for this.
Tags|common
#### requests-and-limits

Property|Description
---|---
Test Case Name|requests-and-limits
Test Case Label|access-control-requests-and-limits
Unique ID|http://test-network-function.com/testcases/access-control/requests-and-limits
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/requests-and-limits Check that containers have resource requests and limits specified in their spec.
Result Type|informative
Suggested Remediation|Add requests and limits to your container spec.  See: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits
Best Practice Reference|https://TODO Section 4.6.11
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### security-context

Property|Description
---|---
Test Case Name|security-context
Test Case Label|access-control-security-context
Unique ID|http://test-network-function.com/testcases/access-control/security-context
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/security-context Checks the security context matches one of the 4 categories
Result Type|normative
Suggested Remediation|Change the security context to be one of the 4 that are allowed on the documentation section 4.5
Best Practice Reference|https://TODO Section 4.5
Exception Process|If the container had the right configuration of the allowed category from the 4 list so the test will pass the  	list is on page 51 on the CNF Security Context Constraints (SCC) section 4.5(Allowed categories are category 1 and category 0),  	Applications MUST use one of the approved Security Context Constraints.
Tags|extended
#### security-context-non-root-user-check

Property|Description
---|---
Test Case Name|security-context-non-root-user-check
Test Case Label|access-control-security-context-non-root-user-check
Unique ID|http://test-network-function.com/testcases/access-control/security-context-non-root-user-check
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/security-context-non-root-user-check Checks the security context runAsUser parameter in pods and containers to make sure it is not set to uid root(0)
Result Type|normative
Suggested Remediation|Change the pod and containers "runAsUser" uid to something other than root(0)
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|If your application needs root user access, please document why your application cannot be ran as 											non-root and supply the reasoning for exception.
Tags|common
#### security-context-privilege-escalation

Property|Description
---|---
Test Case Name|security-context-privilege-escalation
Test Case Label|access-control-security-context-privilege-escalation
Unique ID|http://test-network-function.com/testcases/access-control/security-context-privilege-escalation
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/security-context-privilege-escalation Checks if privileged escalation is enabled (AllowPrivilegeEscalation=true)
Result Type|normative
Suggested Remediation|Configure privilege escalation to false
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|There is no documented exception process for this.
Tags|common
#### ssh-daemons

Property|Description
---|---
Test Case Name|ssh-daemons
Test Case Label|access-control-ssh-daemons
Unique ID|http://test-network-function.com/testcases/access-control/ssh-daemons
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/ssh-daemons Check that pods do not run SSH daemons.
Result Type|normative
Suggested Remediation|Ensure that no SSH daemons are running inside a pod
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 4.6.12
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### sys-admin-capability-check

Property|Description
---|---
Test Case Name|sys-admin-capability-check
Test Case Label|access-control-sys-admin-capability-check
Unique ID|http://test-network-function.com/testcases/access-control/sys-admin-capability-check
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/sys-admin-capability-check Ensures that containers do not use SYS_ADMIN capability
Result Type|normative
Suggested Remediation|Change the security context to be one of the 4 that are allowed on the documentation section 4.5
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|Identify the pod that is needing special capabilities and document why
Tags|common
#### sys-nice-realtime-capability

Property|Description
---|---
Test Case Name|sys-nice-realtime-capability
Test Case Label|access-control-sys-nice-realtime-capability
Unique ID|http://test-network-function.com/testcases/access-control/sys-nice-realtime-capability
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/sys-nice-realtime-capability Check that pods running on nodes with realtime kernel enabled have the SYS_NICE capability enabled in their spec.
Result Type|informative
Suggested Remediation|If pods are scheduled to realtime kernel nodes, they must add SYS_NICE capability to their spec.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 2.7.4
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### sys-ptrace-capability

Property|Description
---|---
Test Case Name|sys-ptrace-capability
Test Case Label|access-control-sys-ptrace-capability
Unique ID|http://test-network-function.com/testcases/access-control/sys-ptrace-capability
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/sys-ptrace-capability Check that if process namespace sharing is enabled for a Pod then the SYS_PTRACE capability is allowed
Result Type|informative
Suggested Remediation|Allow the SYS_PTRACE capability when enabling process namespace sharing for a Pod
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 2.7.5
Exception Process|There is no documented exception process for this.
Tags|common,telco

### affiliated-certification

#### container-is-certified

Property|Description
---|---
Test Case Name|container-is-certified
Test Case Label|affiliated-certification-container-is-certified
Unique ID|http://test-network-function.com/testcases/affiliated-certification/container-is-certified
Version|v1.0.0
Description|http://test-network-function.com/testcases/affiliated-certification/container-is-certified Tests whether container images listed in the configuration file have passed the Red Hat Container Certification Program (CCP).
Result Type|normative
Suggested Remediation|Ensure that your container has passed the Red Hat Container Certification Program (CCP).
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.3.7
Exception Process|There is no documented exception process for this.
Tags|common
#### container-is-certified-digest

Property|Description
---|---
Test Case Name|container-is-certified-digest
Test Case Label|affiliated-certification-container-is-certified-digest
Unique ID|http://test-network-function.com/testcases/affiliated-certification/container-is-certified-digest
Version|v1.0.0
Description|http://test-network-function.com/testcases/affiliated-certification/container-is-certified-digest Tests whether container images that are autodiscovered have passed the Red Hat Container Certification Program by their digest(CCP).
Result Type|normative
Suggested Remediation|Ensure that your container has passed the Red Hat Container Certification Program (CCP).
Best Practice Reference|https://TODO Section 5.3.7
Exception Process|There is no documented exception process for this.
Tags|extended,telco
#### helmchart-is-certified

Property|Description
---|---
Test Case Name|helmchart-is-certified
Test Case Label|affiliated-certification-helmchart-is-certified
Unique ID|http://test-network-function.com/testcases/affiliated-certification/helmchart-is-certified
Version|v1.0.0
Description|http://test-network-function.com/testcases/affiliated-certification/helmchart-is-certified Tests whether helm charts listed in the cluster passed the Red Hat Helm Certification Program.
Result Type|normative
Suggested Remediation|Ensure that the helm charts under test passed the Red Hat's helm Certification Program (e.g. listed in https://charts.openshift.io/index.yaml).
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2.12 and 5.3.3
Exception Process|There is no documented exception process for this.
Tags|common
#### operator-is-certified

Property|Description
---|---
Test Case Name|operator-is-certified
Test Case Label|affiliated-certification-operator-is-certified
Unique ID|http://test-network-function.com/testcases/affiliated-certification/operator-is-certified
Version|v1.0.0
Description|http://test-network-function.com/testcases/affiliated-certification/operator-is-certified Tests whether CNF Operators listed in the configuration file have passed the Red Hat Operator Certification Program (OCP).
Result Type|normative
Suggested Remediation|Ensure that your Operator has passed Red Hat's Operator Certification Program (OCP).
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2.12 and 5.3.3
Exception Process|There is no documented exception process for this.
Tags|common

### lifecycle

#### affinity-required-pods

Property|Description
---|---
Test Case Name|affinity-required-pods
Test Case Label|lifecycle-affinity-required-pods
Unique ID|http://test-network-function.com/testcases/lifecycle/affinity-required-pods
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/affinity-required-pods Checks that affinity rules are in place if AffinityRequired: 'true' labels are set on Pods.
Result Type|informative
Suggested Remediation|If a pod/statefulset/deployment is required to use affinity rules, please add AffinityRequired: 'true' as a label
Best Practice Reference|https://TODO Section 4.6.24
Exception Process|There is no documented exception process for this.
Tags|extended,telco
#### container-shutdown

Property|Description
---|---
Test Case Name|container-shutdown
Test Case Label|lifecycle-container-shutdown
Unique ID|http://test-network-function.com/testcases/lifecycle/container-shutdown
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/container-shutdown Ensure that the containers lifecycle preStop management feature is configured.
Result Type|normative
Suggested Remediation|The preStop can be used to gracefully stop the container and clean resources (e.g., DB connection). For details, see https://www.containiq.com/post/kubernetes-container-lifecycle-events-and-hooks and https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.1.3, 12.2 and 12.5
Exception Process|Identify which pod is not conforming to the process and submit information as to why it cannot use a preStop shutdown specification.
Tags|common,telco
#### container-startup

Property|Description
---|---
Test Case Name|container-startup
Test Case Label|lifecycle-container-startup
Unique ID|http://test-network-function.com/testcases/lifecycle/container-startup
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/container-startup Ensure that the containers lifecycle postStart management feature is configured.
Result Type|normative
Suggested Remediation|PostStart is normally used to configure the container, set up dependencies, and record the new creation. You could use this event to check that a required API is available before the container’s main work begins. Kubernetes will not change the container’s state to Running until the PostStart script has executed successfully. For details, see https://www.containiq.com/post/kubernetes-container-lifecycle-events-and-hooks and https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.1.3, 12.2 and 12.5
Exception Process|Identify which pod is not conforming to the process and submit information as to why it cannot use a postStart startup specification.
Tags|common,telco
#### cpu-isolation

Property|Description
---|---
Test Case Name|cpu-isolation
Test Case Label|lifecycle-cpu-isolation
Unique ID|http://test-network-function.com/testcases/lifecycle/cpu-isolation
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/cpu-isolation CPU isolation requires: For each container within the pod, resource requests and limits must be identical. 		Request and Limits are in the form of whole CPUs. The runTimeClassName must be specified. Annotations required disabling CPU and IRQ load-balancing.
Result Type|informative
Suggested Remediation|CPU isolation testing is enabled.  Please ensure that all pods adhere to the CPU isolation requirements
Best Practice Reference|https://TODO Section 3.5.5
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### deployment-scaling

Property|Description
---|---
Test Case Name|deployment-scaling
Test Case Label|lifecycle-deployment-scaling
Unique ID|http://test-network-function.com/testcases/lifecycle/deployment-scaling
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/deployment-scaling Tests that CNF deployments support scale in/out operations. 			First, The test starts getting the current replicaCount (N) of the deployment/s with the Pod Under Test. Then, it executes the 			scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the deployment/s. 		    In case of deployments that are managed by HPA the test is changing the min and max value to deployment Replica - 1 during scale-in and the 			original replicaCount again for both min/max during the scale-out stage. lastly its restoring the original min/max replica of the deployment/s
Result Type|normative
Suggested Remediation|Ensure CNF deployments/replica sets can scale in/out successfully.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|There is no documented exception process for this.
Tags|common
#### image-pull-policy

Property|Description
---|---
Test Case Name|image-pull-policy
Test Case Label|lifecycle-image-pull-policy
Unique ID|http://test-network-function.com/testcases/lifecycle/image-pull-policy
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/image-pull-policy Ensure that the containers under test are using IfNotPresent as Image Pull Policy..
Result Type|normative
Suggested Remediation|Ensure that the containers under test are using IfNotPresent as Image Pull Policy.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf  Section 12.6
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### liveness-probe

Property|Description
---|---
Test Case Name|liveness-probe
Test Case Label|lifecycle-liveness-probe
Unique ID|http://test-network-function.com/testcases/lifecycle/liveness-probe
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/liveness-probe Check that all containers under test have liveness probe defined
Result Type|normative
Suggested Remediation|Add a liveness probe to deployed containers
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2.16, 12.1 and 12.5
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### persistent-volume-reclaim-policy

Property|Description
---|---
Test Case Name|persistent-volume-reclaim-policy
Test Case Label|lifecycle-persistent-volume-reclaim-policy
Unique ID|http://test-network-function.com/testcases/lifecycle/persistent-volume-reclaim-policy
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/persistent-volume-reclaim-policy Check that the persistent volumes the CNF pods are using have a reclaim policy of delete.
Result Type|informative
Suggested Remediation|Ensure that all persistent volumes are using the reclaim policy: delete
Best Practice Reference|https://TODO Section 3.3.4
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### pod-high-availability

Property|Description
---|---
Test Case Name|pod-high-availability
Test Case Label|lifecycle-pod-high-availability
Unique ID|http://test-network-function.com/testcases/lifecycle/pod-high-availability
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/pod-high-availability Ensures that CNF Pods specify podAntiAffinity rules and replica value is set to more than 1.
Result Type|informative
Suggested Remediation|In high availability cases, Pod podAntiAffinity rule should be specified for pod scheduling and pod replica value is set to more than 1 .
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|There is no documented exception process for this.
Tags|common
#### pod-owner-type

Property|Description
---|---
Test Case Name|pod-owner-type
Test Case Label|lifecycle-pod-owner-type
Unique ID|http://test-network-function.com/testcases/lifecycle/pod-owner-type
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/pod-owner-type Tests that CNF Pod(s) are deployed as part of a ReplicaSet(s)/StatefulSet(s).
Result Type|normative
Suggested Remediation|Deploy the CNF using ReplicaSet/StatefulSet.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.3.3 and 5.3.8
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### pod-recreation

Property|Description
---|---
Test Case Name|pod-recreation
Test Case Label|lifecycle-pod-recreation
Unique ID|http://test-network-function.com/testcases/lifecycle/pod-recreation
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/pod-recreation Tests that a CNF is configured to support High Availability. 			First, this test cordons and drains a Node that hosts the CNF Pod. 			Next, the test ensures that OpenShift can re-instantiate the Pod on another Node, 			and that the actual replica count matches the desired replica count.
Result Type|normative
Suggested Remediation|Ensure that CNF Pod(s) utilize a configuration that supports High Availability.   	Additionally, ensure that there are available Nodes in the OpenShift cluster that can be utilized in the event that a host Node fails.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|There is no documented exception process for this.
Tags|common
#### pod-scheduling

Property|Description
---|---
Test Case Name|pod-scheduling
Test Case Label|lifecycle-pod-scheduling
Unique ID|http://test-network-function.com/testcases/lifecycle/pod-scheduling
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/pod-scheduling Ensures that CNF Pods do not specify nodeSelector or nodeAffinity.  In most cases, Pods should allow for instantiation on any underlying Node.
Result Type|normative
Suggested Remediation|In most cases, Pod's should not specify their host Nodes through nodeSelector or nodeAffinity.  However, there are 	cases in which CNFs require specialized hardware specific to a particular class of Node.  As such, this test is purely 	informative, and will not prevent a CNF from being certified. However, one should have an appropriate justification as 	to why nodeSelector and/or nodeAffinity is utilized by a CNF.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### pod-toleration-bypass

Property|Description
---|---
Test Case Name|pod-toleration-bypass
Test Case Label|lifecycle-pod-toleration-bypass
Unique ID|http://test-network-function.com/testcases/lifecycle/pod-toleration-bypass
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/pod-toleration-bypass Check that pods do not have NoExecute, PreferNoSchedule, or NoSchedule tolerations that have been modified from the default.
Result Type|normative
Suggested Remediation|Do not allow pods to bypass the NoExecute, PreferNoSchedule, or NoSchedule tolerations that are default applied by Kubernetes.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 10.6
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### readiness-probe

Property|Description
---|---
Test Case Name|readiness-probe
Test Case Label|lifecycle-readiness-probe
Unique ID|http://test-network-function.com/testcases/lifecycle/readiness-probe
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/readiness-probe Check that all containers under test have readiness probe defined
Result Type|normative
Suggested Remediation|Add a readiness probe to deployed containers
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2.16, 12.1 and 12.5
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### startup-probe

Property|Description
---|---
Test Case Name|startup-probe
Test Case Label|lifecycle-startup-probe
Unique ID|http://test-network-function.com/testcases/lifecycle/startup-probe
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/startup-probe Check that all containers under test have startup probe defined
Result Type|normative
Suggested Remediation|Add a startup probe to deployed containers
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 4.6.12
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### statefulset-scaling

Property|Description
---|---
Test Case Name|statefulset-scaling
Test Case Label|lifecycle-statefulset-scaling
Unique ID|http://test-network-function.com/testcases/lifecycle/statefulset-scaling
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/statefulset-scaling Tests that CNF statefulsets support scale in/out operations. 			First, The test starts getting the current replicaCount (N) of the statefulset/s with the Pod Under Test. Then, it executes the 			scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the statefulset/s. 			In case of statefulsets that are managed by HPA the test is changing the min and max value to statefulset Replica - 1 during scale-in and the 			original replicaCount again for both min/max during the scale-out stage. lastly its restoring the original min/max replica of the statefulset/s
Result Type|normative
Suggested Remediation|Ensure CNF statefulsets/replica sets can scale in/out successfully.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|There is no documented exception process for this.
Tags|common
#### storage-required-pods

Property|Description
---|---
Test Case Name|storage-required-pods
Test Case Label|lifecycle-storage-required-pods
Unique ID|http://test-network-function.com/testcases/lifecycle/storage-required-pods
Version|https://TODO Section 4.6.24
Description|http://test-network-function.com/testcases/lifecycle/storage-required-pods Checks that pods do not place persistent volumes on local storage.
Result Type|informative
Suggested Remediation|If the kind of pods is StatefulSet, so we need to make sure that servicename is not local-storage.
Best Practice Reference|extended
Exception Process|v1.0.0
Tags|common

### manageability

#### container-port-name-format

Property|Description
---|---
Test Case Name|container-port-name-format
Test Case Label|manageability-container-port-name-format
Unique ID|http://test-network-function.com/testcases/manageability/container-port-name-format
Version|v1.0.0
Description|http://test-network-function.com/testcases/manageability/container-port-name-format Check that the container's ports name follow the naming conventions.
Result Type|normative
Suggested Remediation|Ensure that the container's ports name follow our partner naming conventions
Best Practice Reference|https://TODO Section 4.6.20
Exception Process|There is no documented exception process for this.
Tags|extended
#### containers-image-tag

Property|Description
---|---
Test Case Name|containers-image-tag
Test Case Label|manageability-containers-image-tag
Unique ID|http://test-network-function.com/testcases/manageability/containers-image-tag
Version|v1.0.0
Description|http://test-network-function.com/testcases/manageability/containers-image-tag Check that image tag exists on containers.
Result Type|informative
Suggested Remediation|Ensure that all the container images are tagged
Best Practice Reference|https://TODO Section 4.6.12
Exception Process|There is no documented exception process for this.
Tags|extended

### networking

#### dpdk-cpu-pinning-exec-probe

Property|Description
---|---
Test Case Name|dpdk-cpu-pinning-exec-probe
Test Case Label|networking-dpdk-cpu-pinning-exec-probe
Unique ID|http://test-network-function.com/testcases/networking/dpdk-cpu-pinning-exec-probe
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/dpdk-cpu-pinning-exec-probe If a CNF is doing CPI pinning, exec probes may not be used.
Result Type|informative
Suggested Remediation|If the CNF is doing CPU pinning and running a DPDK process do not use exec probes (executing a command within the container) as it may pile up and block the node eventually.
Best Practice Reference|https://TODO Section 4.6.24
Exception Process|There is no documented exception process for this.
Tags|extended,telco
#### dual-stack-service

Property|Description
---|---
Test Case Name|dual-stack-service
Test Case Label|networking-dual-stack-service
Unique ID|http://test-network-function.com/testcases/networking/dual-stack-service
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/dual-stack-service Checks that all services in namespaces under test are either ipv6 single stack or dual stack. This test case requires the deployment of the debug daemonset.
Result Type|normative
Suggested Remediation|Configure every CNF services with either a single stack ipv6 or dual stack (ipv4/ipv6) load balancer
Best Practice Reference|https://TODO Section 3.5.7
Exception Process|There is no documented exception process for this.
Tags|common,extended
#### icmpv4-connectivity

Property|Description
---|---
Test Case Name|icmpv4-connectivity
Test Case Label|networking-icmpv4-connectivity
Unique ID|http://test-network-function.com/testcases/networking/icmpv4-connectivity
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/icmpv4-connectivity Checks that each CNF Container is able to communicate via ICMPv4 on the Default OpenShift network. This test case requires the Deployment of the debug daemonset.
Result Type|normative
Suggested Remediation|Ensure that the CNF is able to communicate via the Default OpenShift network. In some rare cases, CNFs may require routing table changes in order to communicate over the Default network. To exclude a particular pod from ICMPv4 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it. The label value is not important, only its presence.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|There is no documented exception process for this.
Tags|common
#### icmpv4-connectivity-multus

Property|Description
---|---
Test Case Name|icmpv4-connectivity-multus
Test Case Label|networking-icmpv4-connectivity-multus
Unique ID|http://test-network-function.com/testcases/networking/icmpv4-connectivity-multus
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/icmpv4-connectivity-multus Checks that each CNF Container is able to communicate via ICMPv4 on the Multus network(s).  This test case requires the Deployment of the debug daemonset.
Result Type|normative
Suggested Remediation|Ensure that the CNF is able to communicate via the Multus network(s). In some rare cases, 	CNFs may require routing table changes in order to communicate over the Multus network(s). To exclude a particular pod 	from ICMPv4 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it. The label value is not important, only its presence.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|There is no documented exception process for this.
Tags|
#### icmpv6-connectivity

Property|Description
---|---
Test Case Name|icmpv6-connectivity
Test Case Label|networking-icmpv6-connectivity
Unique ID|http://test-network-function.com/testcases/networking/icmpv6-connectivity
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/icmpv6-connectivity Checks that each CNF Container is able to communicate via ICMPv6 on the Default OpenShift network.  This test case requires the Deployment of the debug daemonset.
Result Type|normative
Suggested Remediation|Ensure that the CNF is able to communicate via the Default OpenShift network. In some rare cases, 	CNFs may require routing table changes in order to communicate over the Default network. To exclude a particular pod 	from ICMPv6 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it. The label value is not important, only its presence.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|There is no documented exception process for this.
Tags|common
#### icmpv6-connectivity-multus

Property|Description
---|---
Test Case Name|icmpv6-connectivity-multus
Test Case Label|networking-icmpv6-connectivity-multus
Unique ID|http://test-network-function.com/testcases/networking/icmpv6-connectivity-multus
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/icmpv6-connectivity-multus Checks that each CNF Container is able to communicate via ICMPv6 on the Multus network(s).  This test case requires the Deployment of the debug daemonset.
Result Type|normative
Suggested Remediation|Ensure that the CNF is able to communicate via the Multus network(s). In some rare cases, 	CNFs may require routing table changes in order to communicate over the Multus network(s). To exclude a particular pod 	from ICMPv6 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it.The label value is not important, only its presence.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|There is no documented exception process for this.
Tags|common
#### iptables

Property|Description
---|---
Test Case Name|iptables
Test Case Label|networking-iptables
Unique ID|http://test-network-function.com/testcases/networking/iptables
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/iptables Checks that the output of "iptables-save" is empty, e.g. there is no iptables configuration on any CNF containers.
Result Type|normative
Suggested Remediation|Do not configure iptables on any CNF container.
Best Practice Reference|https://TODO Section 4.6.23
Exception Process|There is no documented exception process for this.
Tags|common,extended
#### network-policy-deny-all

Property|Description
---|---
Test Case Name|network-policy-deny-all
Test Case Label|networking-network-policy-deny-all
Unique ID|http://test-network-function.com/testcases/networking/network-policy-deny-all
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/network-policy-deny-all Check that network policies attached to namespaces running CNF pods contain a default deny-all rule for both ingress and egress traffic
Result Type|informative
Suggested Remediation|Ensure that a NetworkPolicy with a default deny-all is applied. After the default is applied, apply a network policy to allow the traffic your application requires.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 10.6
Exception Process|There is no documented exception process for this.
Tags|common
#### nftables

Property|Description
---|---
Test Case Name|nftables
Test Case Label|networking-nftables
Unique ID|http://test-network-function.com/testcases/networking/nftables
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/nftables Checks that the output of "nft list ruleset" is empty, e.g. there is no nftables configuration on any CNF containers.
Result Type|normative
Suggested Remediation|Do not configure nftables on any CNF container.
Best Practice Reference|https://TODO Section 4.6.23
Exception Process|There is no documented exception process for this.
Tags|common,extended
#### ocp-reserved-ports-usage

Property|Description
---|---
Test Case Name|ocp-reserved-ports-usage
Test Case Label|networking-ocp-reserved-ports-usage
Unique ID|http://test-network-function.com/testcases/networking/ocp-reserved-ports-usage
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/ocp-reserved-ports-usage Check that containers do not listen on ports that are reserved by OpenShift
Result Type|normative
Suggested Remediation|Ensure that CNF apps do not listen on ports that are reserved by OpenShift
Best Practice Reference|https://TODO Section 3.5.9
Exception Process|There is no documented exception process for this.
Tags|common
#### reserved-partner-ports

Property|Description
---|---
Test Case Name|reserved-partner-ports
Test Case Label|networking-reserved-partner-ports
Unique ID|http://test-network-function.com/testcases/networking/reserved-partner-ports
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/reserved-partner-ports Checks that pods and containers are not consuming ports designated as reserved by partner
Result Type|informative
Suggested Remediation|Ensure ports are not being used that are reserved by our partner
Best Practice Reference|https://TODO Section 4.6.24
Exception Process|There is no documented exception process for this.
Tags|extended
#### service-type

Property|Description
---|---
Test Case Name|service-type
Test Case Label|networking-service-type
Unique ID|http://test-network-function.com/testcases/networking/service-type
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/service-type Tests that each CNF Service does not utilize NodePort(s).
Result Type|normative
Suggested Remediation|Ensure Services are not configured to use NodePort(s).
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.3.1
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### undeclared-container-ports-usage

Property|Description
---|---
Test Case Name|undeclared-container-ports-usage
Test Case Label|networking-undeclared-container-ports-usage
Unique ID|http://test-network-function.com/testcases/networking/undeclared-container-ports-usage
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/undeclared-container-ports-usage Check that containers do not listen on ports that weren't declared in their specification
Result Type|normative
Suggested Remediation|Ensure the CNF apps do not listen on undeclared containers' ports
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 16.3.1.1
Exception Process|There is no documented exception process for this.
Tags|common,telco

### observability

#### container-logging

Property|Description
---|---
Test Case Name|container-logging
Test Case Label|observability-container-logging
Unique ID|http://test-network-function.com/testcases/observability/container-logging
Version|v1.0.0
Description|http://test-network-function.com/testcases/observability/container-logging Check that all containers under test use standard input output and standard error when logging
Result Type|informative
Suggested Remediation|Ensure containers are not redirecting stdout/stderr
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 10.1
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### crd-status

Property|Description
---|---
Test Case Name|crd-status
Test Case Label|observability-crd-status
Unique ID|http://test-network-function.com/testcases/observability/crd-status
Version|v1.0.0
Description|http://test-network-function.com/testcases/observability/crd-status Checks that all CRDs have a status subresource specification (Spec.versions[].Schema.OpenAPIV3Schema.Properties["status"]).
Result Type|informative
Suggested Remediation|Ensure that all the CRDs have a meaningful status specification (Spec.versions[].Schema.OpenAPIV3Schema.Properties["status"]).
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### pod-disruption-budget

Property|Description
---|---
Test Case Name|pod-disruption-budget
Test Case Label|observability-pod-disruption-budget
Unique ID|http://test-network-function.com/testcases/observability/pod-disruption-budget
Version|v1.0.0
Description|http://test-network-function.com/testcases/observability/pod-disruption-budget Checks to see if pod disruption budgets have allowed values for minAvailable and maxUnavailable
Result Type|normative
Suggested Remediation|Ensure minAvailable is not zero and maxUnavailable does not equal the number of pods in the replica
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 4.6.20
Exception Process|There is no documented exception process for this.
Tags|common,telco
#### termination-policy

Property|Description
---|---
Test Case Name|termination-policy
Test Case Label|observability-termination-policy
Unique ID|http://test-network-function.com/testcases/observability/termination-policy
Version|v1.0.0
Description|http://test-network-function.com/testcases/observability/termination-policy Check that all containers are using terminationMessagePolicy: FallbackToLogsOnError
Result Type|informative
Suggested Remediation|Ensure containers are all using FallbackToLogsOnError in terminationMessagePolicy
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 12.1
Exception Process|There is no documented exception process for this.
Tags|common

### operator

#### install-source

Property|Description
---|---
Test Case Name|install-source
Test Case Label|operator-install-source
Unique ID|http://test-network-function.com/testcases/operator/install-source
Version|v1.0.0
Description|http://test-network-function.com/testcases/operator/install-source Tests whether a CNF Operator is installed via OLM.
Result Type|normative
Suggested Remediation|Ensure that your Operator is installed via OLM.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2.12 and 5.3.3
Exception Process|There is no documented exception process for this.
Tags|common
#### install-status-no-privileges

Property|Description
---|---
Test Case Name|install-status-no-privileges
Test Case Label|operator-install-status-no-privileges
Unique ID|http://test-network-function.com/testcases/operator/install-status-no-privileges
Version|v1.0.0
Description|http://test-network-function.com/testcases/operator/install-status-no-privileges The operator is not installed with privileged rights. Test passes if clusterPermissions is not present in the CSV manifest or is present with no resourceNames under its rules.
Result Type|normative
Suggested Remediation|Ensure all the CNF operators have no privileges on cluster resources.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2.12 and 5.3.3
Exception Process|There is no documented exception process for this.
Tags|common
#### install-status-succeeded

Property|Description
---|---
Test Case Name|install-status-succeeded
Test Case Label|operator-install-status-succeeded
Unique ID|http://test-network-function.com/testcases/operator/install-status-succeeded
Version|v1.0.0
Description|http://test-network-function.com/testcases/operator/install-status-succeeded Ensures that the target CNF operators report "Succeeded" as their installation status.
Result Type|normative
Suggested Remediation|Ensure all the CNF operators have been successfully installed by OLM.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2.12 and 5.3.3
Exception Process|There is no documented exception process for this.
Tags|common

### platform-alteration

#### base-image

Property|Description
---|---
Test Case Name|base-image
Test Case Label|platform-alteration-base-image
Unique ID|http://test-network-function.com/testcases/platform-alteration/base-image
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/base-image Ensures that the Container Base Image is not altered post-startup.  This test is a heuristic, and ensures that there are no changes to the following directories: 1) /var/lib/rpm 2) /var/lib/dpkg 3) /bin 4) /sbin 5) /lib 6) /lib64 7) /usr/bin 8) /usr/sbin 9) /usr/lib 10) /usr/lib64
Result Type|normative
Suggested Remediation|Ensure that Container applications do not modify the Container Base Image.  In particular, ensure that the following 	directories are not modified: 	1) /var/lib/rpm 	2) /var/lib/dpkg 	3) /bin 	4) /sbin 	5) /lib 	6) /lib64 	7) /usr/bin 	8) /usr/sbin 	9) /usr/lib 	10) /usr/lib64 	Ensure that all required binaries are built directly into the container image, and are not installed post startup.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.1.4
Exception Process|Images should not be changed during runtime.  There is no exception process for this.
Tags|common
#### boot-params

Property|Description
---|---
Test Case Name|boot-params
Test Case Label|platform-alteration-boot-params
Unique ID|http://test-network-function.com/testcases/platform-alteration/boot-params
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/boot-params Tests that boot parameters are set through the MachineConfigOperator, and not set manually on the Node.
Result Type|normative
Suggested Remediation|Ensure that boot parameters are set directly through the MachineConfigOperator, or indirectly through the PerformanceAddonOperator.   	Boot parameters should not be changed directly through the Node, as OpenShift should manage the changes for you.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2.13 and 5.2.14
Exception Process|There is no documented exception process for this.
Tags|common
#### hugepages-1g-only

Property|Description
---|---
Test Case Name|hugepages-1g-only
Test Case Label|platform-alteration-hugepages-1g-only
Unique ID|http://test-network-function.com/testcases/platform-alteration/hugepages-1g-only
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/hugepages-1g-only Check that pods using hugepages only use 1Gi size
Result Type|informative
Suggested Remediation|Modify pod to consume 1Gi hugepages only
Best Practice Reference|https://TODO
Exception Process|There is no documented exception process for this.
Tags|faredge
#### hugepages-2m-only

Property|Description
---|---
Test Case Name|hugepages-2m-only
Test Case Label|platform-alteration-hugepages-2m-only
Unique ID|http://test-network-function.com/testcases/platform-alteration/hugepages-2m-only
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/hugepages-2m-only Check that pods using hugepages only use 2Mi size
Result Type|normative
Suggested Remediation|Modify pod to consume 2Mi hugepages only
Best Practice Reference|https://TODO Section 3.5.4
Exception Process|There is no documented exception process for this.
Tags|extended
#### hugepages-config

Property|Description
---|---
Test Case Name|hugepages-config
Test Case Label|platform-alteration-hugepages-config
Unique ID|http://test-network-function.com/testcases/platform-alteration/hugepages-config
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/hugepages-config Checks to see that HugePage settings have been configured through MachineConfig, and not manually on the underlying Node.  This test case applies only to Nodes that are configured with the "worker" MachineConfigSet.  First, the "worker" MachineConfig is polled, and the Hugepage settings are extracted.  Next, the underlying Nodes are polled for configured HugePages through inspection of /proc/meminfo.  The results are compared, and the test passes only if they are the same.
Result Type|normative
Suggested Remediation|HugePage settings should be configured either directly through the MachineConfigOperator or indirectly using the 	PerformanceAddonOperator.  This ensures that OpenShift is aware of the special MachineConfig requirements, and can 	provision your CNF on a Node that is part of the corresponding MachineConfigSet.  Avoid making changes directly to an 	underlying Node, and let OpenShift handle the heavy lifting of configuring advanced settings.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|There is no documented exception process for this.
Tags|common
#### is-selinux-enforcing

Property|Description
---|---
Test Case Name|is-selinux-enforcing
Test Case Label|platform-alteration-is-selinux-enforcing
Unique ID|http://test-network-function.com/testcases/platform-alteration/is-selinux-enforcing
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/is-selinux-enforcing verifies that all openshift platform/cluster nodes have selinux in "Enforcing" mode.
Result Type|normative
Suggested Remediation|Configure selinux and enable enforcing mode.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 10.3 Pod Security
Exception Process|There is no documented exception process for this.
Tags|common
#### isredhat-release

Property|Description
---|---
Test Case Name|isredhat-release
Test Case Label|platform-alteration-isredhat-release
Unique ID|http://test-network-function.com/testcases/platform-alteration/isredhat-release
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/isredhat-release verifies if the container base image is redhat.
Result Type|normative
Suggested Remediation|Build a new container image that is based on UBI (Red Hat Universal Base Image).
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|Document which containers are not able to meet the RHEL-based container  											requirement and if/when the base image can be updated.
Tags|common
#### ocp-lifecycle

Property|Description
---|---
Test Case Name|ocp-lifecycle
Test Case Label|platform-alteration-ocp-lifecycle
Unique ID|http://test-network-function.com/testcases/platform-alteration/ocp-lifecycle
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/ocp-lifecycle Tests that the running OCP version is not end of life.
Result Type|normative
Suggested Remediation|Please update your cluster to a version that is generally available.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 7.9
Exception Process|There is no documented exception process for this.
Tags|common
#### ocp-node-os-lifecycle

Property|Description
---|---
Test Case Name|ocp-node-os-lifecycle
Test Case Label|platform-alteration-ocp-node-os-lifecycle
Unique ID|http://test-network-function.com/testcases/platform-alteration/ocp-node-os-lifecycle
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/ocp-node-os-lifecycle Tests that the nodes running in the cluster have operating systems 			that are compatible with the deployed version of OpenShift.
Result Type|normative
Suggested Remediation|Please update your workers to a version that is supported by your version of OpenShift
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 7.9
Exception Process|There is no documented exception process for this.
Tags|common
#### service-mesh-usage

Property|Description
---|---
Test Case Name|service-mesh-usage
Test Case Label|platform-alteration-service-mesh-usage
Unique ID|http://test-network-function.com/testcases/platform-alteration/service-mesh-usage
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/service-mesh-usage Checks if the istio namespace ("istio-system") is present. If it is present, checks that the istio sidecar is present in all pods under test.
Result Type|normative
Suggested Remediation|Ensure all the CNF pods are using service mesh if the cluster provides it.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|There is no documented exception process for this.
Tags|common
#### sysctl-config

Property|Description
---|---
Test Case Name|sysctl-config
Test Case Label|platform-alteration-sysctl-config
Unique ID|http://test-network-function.com/testcases/platform-alteration/sysctl-config
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/pod-recreation Tests that no one has changed the node's sysctl configs after the node 			was created, the tests works by checking if the sysctl configs are consistent with the 			MachineConfig CR which defines how the node should be configured
Result Type|normative
Suggested Remediation|You should recreate the node or change the sysctls, recreating is recommended because there might be other unknown changes
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2
Exception Process|There is no documented exception process for this.
Tags|common
#### tainted-node-kernel

Property|Description
---|---
Test Case Name|tainted-node-kernel
Test Case Label|platform-alteration-tainted-node-kernel
Unique ID|http://test-network-function.com/testcases/platform-alteration/tainted-node-kernel
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/tainted-node-kernel Ensures that the Node(s) hosting CNFs do not utilize tainted kernels. This test case is especially important to support Highly Available CNFs, since when a CNF is re-instantiated on a backup Node, that Node's kernel may not have the same hacks.'
Result Type|normative
Suggested Remediation|Test failure indicates that the underlying Node's kernel is tainted.  Ensure that you have not altered underlying 	Node(s) kernels in order to run the CNF.
Best Practice Reference|https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf Section 5.2.14
Exception Process|There is no documented exception process for this.
Tags|common

