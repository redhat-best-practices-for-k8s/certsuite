# cnf-certification-test catalog

The catalog for cnf-certification-test contains a variety of `Test Cases`, as well as `Test Case Building Blocks`.
* Test Cases:  Traditional JUnit testcases, which are specified internally using `Ginkgo.It`.  Test cases often utilize several Test Case Building Blocks.
* Test Case Building Blocks:  Self-contained building blocks, which perform a small task usually making use of go clients to interact with the Openshift cluster.
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
Description|http://test-network-function.com/testcases/access-control/cluster-role-bindings tests that a Pod does not specify ClusterRoleBindings.
Result Type|normative
Suggested Remediation|In most cases, Pod's should not have ClusterRoleBindings.  The suggested remediation is to remove the need for ClusterRoleBindings, if possible.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2.10 and 6.3.6
#### container-host-port

Property|Description
---|---
Test Case Name|container-host-port
Test Case Label|access-control-container-host-port
Unique ID|http://test-network-function.com/testcases/access-control/container-host-port
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/container-host-port Verifies if containers define a hostPort.
Result Type|informative
Suggested Remediation|Remove hostPort configuration from the container
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.3.6
#### namespace

Property|Description
---|---
Test Case Name|namespace
Test Case Label|access-control-namespace
Unique ID|http://test-network-function.com/testcases/access-control/namespace
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/namespace tests that all CNF's resources (PUTs and CRs) belong to valid namespaces. A valid namespace meets the following conditions: (1) It was declared in the yaml config file under the targetNameSpaces tag. (2) It doesn't have any of the following prefixes: default, openshift-, istio- and aspenmesh-
Result Type|normative
Suggested Remediation|Ensure that your CNF utilizes namespaces declared in the yaml config file. Additionally, the namespaces should not start with "default, openshift-, istio- or aspenmesh-", except in rare cases.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2, 16.3.8 & 16.3.9
#### one-process-per-container

Property|Description
---|---
Test Case Name|one-process-per-container
Test Case Label|access-control-one-process-per-container
Unique ID|http://test-network-function.com/testcases/access-control/one-process-per-container
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/one-process-per-container check that all containers under test 		have only one process running
Result Type|informative
Suggested Remediation|launch only one process per container
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 11.8.3
#### pod-automount-service-account-token

Property|Description
---|---
Test Case Name|pod-automount-service-account-token
Test Case Label|access-control-pod-automount-service-account-token
Unique ID|http://test-network-function.com/testcases/access-control/pod-automount-service-account-token
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/pod-automount-service-account-token check that all pods under test have automountServiceAccountToken set to false
Result Type|normative
Suggested Remediation|check that pod has automountServiceAccountToken set to false or pod is attached to service account which has automountServiceAccountToken set to false
Best Practice Reference|
#### pod-host-ipc

Property|Description
---|---
Test Case Name|pod-host-ipc
Test Case Label|access-control-pod-host-ipc
Unique ID|http://test-network-function.com/testcases/access-control/pod-host-ipc
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/pod-host-ipc Verifies that the spec.HostIpc parameter is set to false
Result Type|informative
Suggested Remediation|Set the spec.HostIpc parameter to false in the pod configuration
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.3.6
#### pod-host-network

Property|Description
---|---
Test Case Name|pod-host-network
Test Case Label|access-control-pod-host-network
Unique ID|http://test-network-function.com/testcases/access-control/pod-host-network
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/pod-host-network Verifies that the spec.HostNetwork parameter is set to false
Result Type|informative
Suggested Remediation|Set the spec.HostNetwork parameter to false in the pod configuration
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.3.6
#### pod-host-path

Property|Description
---|---
Test Case Name|pod-host-path
Test Case Label|access-control-pod-host-path
Unique ID|http://test-network-function.com/testcases/access-control/pod-host-path
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/pod-host-path Verifies that the spec.HostPath parameter is not set (not present)
Result Type|informative
Suggested Remediation|Set the spec.HostPath parameter to false in the pod configuration
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.3.6
#### pod-host-pid

Property|Description
---|---
Test Case Name|pod-host-pid
Test Case Label|access-control-pod-host-pid
Unique ID|http://test-network-function.com/testcases/access-control/pod-host-pid
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/pod-host-pid Verifies that the spec.HostPid parameter is set to false
Result Type|informative
Suggested Remediation|Set the spec.HostPid parameter to false in the pod configuration
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.3.6
#### pod-role-bindings

Property|Description
---|---
Test Case Name|pod-role-bindings
Test Case Label|access-control-pod-role-bindings
Unique ID|http://test-network-function.com/testcases/access-control/pod-role-bindings
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/pod-role-bindings ensures that a CNF does not utilize RoleBinding(s) in a non-CNF Namespace.
Result Type|normative
Suggested Remediation|Ensure the CNF is not configured to use RoleBinding(s) in a non-CNF Namespace.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.3.3 and 6.3.5
#### pod-service-account

Property|Description
---|---
Test Case Name|pod-service-account
Test Case Label|access-control-pod-service-account
Unique ID|http://test-network-function.com/testcases/access-control/pod-service-account
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/pod-service-account tests that each CNF Pod utilizes a valid Service Account.
Result Type|normative
Suggested Remediation|Ensure that the each CNF Pod is configured to use a valid Service Account
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2.3 and 6.2.7
#### security-context-capabilities-check

Property|Description
---|---
Test Case Name|security-context-capabilities-check
Test Case Label|access-control-security-context-capabilities-check
Unique ID|http://test-network-function.com/testcases/access-control/security-context-capabilities-check
Version|v1.0.0
Description|http://test-network-function.com/testcases/access-control/security-context-capabilities-check Tests that the following capabilities are not granted: 			- NET_ADMIN 			- SYS_ADMIN  			- NET_RAW 			- IPC_LOCK 
Result Type|normative
Suggested Remediation|Remove the following capabilities from the container/pod definitions: NET_ADMIN SCC, SYS_ADMIN SCC, NET_RAW SCC, IPC_LOCK SCC 
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2
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
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2
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
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2

### affiliated-certification

#### container-is-certified

Property|Description
---|---
Test Case Name|container-is-certified
Test Case Label|affiliated-certification-container-is-certified
Unique ID|http://test-network-function.com/testcases/affiliated-certification/container-is-certified
Version|v1.0.0
Description|http://test-network-function.com/testcases/affiliated-certification/container-is-certified tests whether container images listed in the configuration file have passed the Red Hat Container Certification Program (CCP).
Result Type|normative
Suggested Remediation|Ensure that your container has passed the Red Hat Container Certification Program (CCP).
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.3.7
#### helmchart-is-certified

Property|Description
---|---
Test Case Name|helmchart-is-certified
Test Case Label|affiliated-certification-helmchart-is-certified
Unique ID|http://test-network-function.com/testcases/affiliated-certification/helmchart-is-certified
Version|v1.0.0
Description|http://test-network-function.com/testcases/affiliated-certification/helmchart-is-certified tests whether helm charts listed in the cluster passed the Red Hat Helm Certification Program.
Result Type|normative
Suggested Remediation|Ensure that the helm charts under test passed the Red Hat's helm Certification Program (e.g. listed in https://charts.openshift.io/index.yaml).
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2.12 and Section 6.3.3
#### operator-is-certified

Property|Description
---|---
Test Case Name|operator-is-certified
Test Case Label|affiliated-certification-operator-is-certified
Unique ID|http://test-network-function.com/testcases/affiliated-certification/operator-is-certified
Version|v1.0.0
Description|http://test-network-function.com/testcases/affiliated-certification/operator-is-certified tests whether CNF Operators listed in the configuration file have passed the Red Hat Operator Certification Program (OCP).
Result Type|normative
Suggested Remediation|Ensure that your Operator has passed Red Hat's Operator Certification Program (OCP).
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2.12 and Section 6.3.3

### lifecycle

#### container-shutdown

Property|Description
---|---
Test Case Name|container-shutdown
Test Case Label|lifecycle-container-shutdown
Unique ID|http://test-network-function.com/testcases/lifecycle/container-shutdown
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/container-shutdown Ensure that the containers lifecycle pre-stop management feature is configured.
Result Type|normative
Suggested Remediation| 		It's considered best-practices to define prestop for proper management of container lifecycle. 		The prestop can be used to gracefully stop the container and clean resources (e.g., DB connection). 		 		The prestop can be configured using : 		 1) Exec : executes the supplied command inside the container 		 2) HTTP : executes HTTP request against the specified endpoint. 		 		When defined. K8s will handle shutdown of the container using the following: 		1) K8s first execute the preStop hook inside the container. 		2) K8s will wait for a grace period. 		3) K8s will clean the remaining processes using KILL signal.		 			
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2
#### deployment-scaling

Property|Description
---|---
Test Case Name|deployment-scaling
Test Case Label|lifecycle-deployment-scaling
Unique ID|http://test-network-function.com/testcases/lifecycle/deployment-scaling
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/deployment-scaling tests that CNF deployments support scale in/out operations.  			First, The test starts getting the current replicaCount (N) of the deployment/s with the Pod Under Test. Then, it executes the  			scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the deployment/s. 		    In case of deployments that are managed by HPA the test is changing the min and max value to deployment Replica - 1 during scale-in and the  			original replicaCount again for both min/max during the scale-out stage. lastly its restoring the original min/max replica of the deployment/s
Result Type|normative
Suggested Remediation|Make sure CNF deployments/replica sets can scale in/out successfully.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2
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
Best Practice Reference|https://docs.google.com/document/d/1wRHMk1ZYUSVmgp_4kxvqjVOKwolsZ5hDXjr5MLy-wbg/edit#  Section 15.6
#### liveness

Property|Description
---|---
Test Case Name|liveness
Test Case Label|lifecycle-liveness
Unique ID|http://test-network-function.com/testcases/lifecycle/liveness
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/liveness check that all containers under test 		have liveness probe defined
Result Type|normative
Suggested Remediation|add liveness probe to deployed containers
Best Practice Reference|
#### pod-high-availability

Property|Description
---|---
Test Case Name|pod-high-availability
Test Case Label|lifecycle-pod-high-availability
Unique ID|http://test-network-function.com/testcases/lifecycle/pod-high-availability
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/pod-high-availability ensures that CNF Pods specify podAntiAffinity rules and replica value is set to more than 1.
Result Type|informative
Suggested Remediation|In high availability cases, Pod podAntiAffinity rule should be specified for pod scheduling and pod replica value is set to more than 1 .
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2
#### pod-owner-type

Property|Description
---|---
Test Case Name|pod-owner-type
Test Case Label|lifecycle-pod-owner-type
Unique ID|http://test-network-function.com/testcases/lifecycle/pod-owner-type
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/pod-owner-type tests that CNF Pod(s) are deployed as part of a ReplicaSet(s)/StatefulSet(s).
Result Type|normative
Suggested Remediation|Deploy the CNF using ReplicaSet/StatefulSet.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.3.3 and 6.3.8
#### pod-recreation

Property|Description
---|---
Test Case Name|pod-recreation
Test Case Label|lifecycle-pod-recreation
Unique ID|http://test-network-function.com/testcases/lifecycle/pod-recreation
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/pod-recreation tests that a CNF is configured to support High Availability.   			First, this test cordons and drains a Node that hosts the CNF Pod.   			Next, the test ensures that OpenShift can re-instantiate the Pod on another Node,  			and that the actual replica count matches the desired replica count.
Result Type|normative
Suggested Remediation|Ensure that CNF Pod(s) utilize a configuration that supports High Availability.   			Additionally, ensure that there are available Nodes in the OpenShift cluster that can be utilized in the event that a host Node fails.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2
#### pod-scheduling

Property|Description
---|---
Test Case Name|pod-scheduling
Test Case Label|lifecycle-pod-scheduling
Unique ID|http://test-network-function.com/testcases/lifecycle/pod-scheduling
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/pod-scheduling ensures that CNF Pods do not specify nodeSelector or nodeAffinity.  In most cases, Pods should allow for instantiation on any underlying Node.
Result Type|informative
Suggested Remediation|In most cases, Pod's should not specify their host Nodes through nodeSelector or nodeAffinity.  However, there are cases in which CNFs require specialized hardware specific to a particular class of Node.  As such, this test is purely informative, and will not prevent a CNF from being certified. However, one should have an appropriate justification as to why nodeSelector and/or nodeAffinity is utilized by a CNF.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2
#### readiness

Property|Description
---|---
Test Case Name|readiness
Test Case Label|lifecycle-readiness
Unique ID|http://test-network-function.com/testcases/lifecycle/readiness
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/readiness check that all containers under test 		have readiness probe defined
Result Type|normative
Suggested Remediation|add readiness probe to deployed containers
Best Practice Reference|
#### scaling

Property|Description
---|---
Test Case Name|scaling
Test Case Label|lifecycle-scaling
Unique ID|http://test-network-function.com/testcases/lifecycle/scaling
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/scaling tests that CNF deployments support scale in/out operations.  			First, The test starts getting the current replicaCount (N) of the deployment/s with the Pod Under Test. Then, it executes the  			scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the deployment/s.
Result Type|normative
Suggested Remediation|Make sure CNF deployments/replica sets can scale in/out successfully.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2
#### statefulset-scaling

Property|Description
---|---
Test Case Name|statefulset-scaling
Test Case Label|lifecycle-statefulset-scaling
Unique ID|http://test-network-function.com/testcases/lifecycle/statefulset-scaling
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/statefulset-scaling tests that CNF statefulsets support scale in/out operations.  			First, The test starts getting the current replicaCount (N) of the statefulset/s with the Pod Under Test. Then, it executes the  			scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the statefulset/s. 			In case of statefulsets that are managed by HPA the test is changing the min and max value to statefulset Replica - 1 during scale-in and the  			original replicaCount again for both min/max during the scale-out stage. lastly its restoring the original min/max replica of the statefulset/s
Result Type|normative
Suggested Remediation|Make sure CNF statefulsets/replica sets can scale in/out successfully.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2

### networking

#### icmpv4-connectivity

Property|Description
---|---
Test Case Name|icmpv4-connectivity
Test Case Label|networking-icmpv4-connectivity
Unique ID|http://test-network-function.com/testcases/networking/icmpv4-connectivity
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/icmpv4-connectivity checks that each CNF Container is able to communicate via ICMPv4 on the Default OpenShift network.  This test case requires the Deployment of the debug daemonset.
Result Type|normative
Suggested Remediation|Ensure that the CNF is able to communicate via the Default OpenShift network. In some rare cases, CNFs may require routing table changes in order to communicate over the Default network. To exclude a particular pod from ICMPv4 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it. The label value is not important, only its presence.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2
#### icmpv4-connectivity-multus

Property|Description
---|---
Test Case Name|icmpv4-connectivity-multus
Test Case Label|networking-icmpv4-connectivity-multus
Unique ID|http://test-network-function.com/testcases/networking/icmpv4-connectivity-multus
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/icmpv4-connectivity-multus checks that each CNF Container is able to communicate via ICMPv4 on the Multus network(s).  This test case requires the Deployment of the debug daemonset.
Result Type|normative
Suggested Remediation|Ensure that the CNF is able to communicate via the Multus network(s). In some rare cases, CNFs may require routing table changes in order to communicate over the Multus network(s). To exclude a particular pod from ICMPv4 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it. The label value is not important, only its presence.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2
#### icmpv6-connectivity

Property|Description
---|---
Test Case Name|icmpv6-connectivity
Test Case Label|networking-icmpv6-connectivity
Unique ID|http://test-network-function.com/testcases/networking/icmpv6-connectivity
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/icmpv6-connectivity checks that each CNF Container is able to communicate via ICMPv6 on the Default OpenShift network.  This test case requires the Deployment of the debug daemonset.
Result Type|normative
Suggested Remediation|Ensure that the CNF is able to communicate via the Default OpenShift network. In some rare cases, CNFs may require routing table changes in order to communicate over the Default network. To exclude a particular pod from ICMPv6 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it. The label value is not important, only its presence.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2
#### icmpv6-connectivity-multus

Property|Description
---|---
Test Case Name|icmpv6-connectivity-multus
Test Case Label|networking-icmpv6-connectivity-multus
Unique ID|http://test-network-function.com/testcases/networking/icmpv6-connectivity-multus
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/icmpv6-connectivity-multus checks that each CNF Container is able to communicate via ICMPv6 on the Multus network(s).  This test case requires the Deployment of the debug daemonset.
Result Type|normative
Suggested Remediation|Ensure that the CNF is able to communicate via the Multus network(s). In some rare cases, CNFs may require routing table changes in order to communicate over the Multus network(s). To exclude a particular pod from ICMPv6 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it.The label value is not important, only its presence. 
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2
#### service-type

Property|Description
---|---
Test Case Name|service-type
Test Case Label|networking-service-type
Unique ID|http://test-network-function.com/testcases/networking/service-type
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/service-type tests that each CNF Service does not utilize NodePort(s).
Result Type|normative
Suggested Remediation|Ensure Services are not configured to use NodePort(s).
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.3.1
#### undeclared-container-ports-usage

Property|Description
---|---
Test Case Name|undeclared-container-ports-usage
Test Case Label|networking-undeclared-container-ports-usage
Unique ID|http://test-network-function.com/testcases/networking/undeclared-container-ports-usage
Version|v1.0.0
Description|http://test-network-function.com/testcases/networking/undeclared-container-ports-usage check that containers don't listen on ports that weren't declared in their specification
Result Type|normative
Suggested Remediation|ensure the CNF apps don't listen on undeclared containers' ports
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 16.3.1.1

### observability

#### container-logging

Property|Description
---|---
Test Case Name|container-logging
Test Case Label|observability-container-logging
Unique ID|http://test-network-function.com/testcases/observability/container-logging
Version|v1.0.0
Description|http://test-network-function.com/testcases/observability/container-logging check that all containers under test use standard input output and standard error when logging
Result Type|informative
Suggested Remediation|make sure containers are not redirecting stdout/stderr
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 11.1
#### crd-status

Property|Description
---|---
Test Case Name|crd-status
Test Case Label|observability-crd-status
Unique ID|http://test-network-function.com/testcases/observability/crd-status
Version|v1.0.0
Description|http://test-network-function.com/testcases/observability/crd-status checks that all CRDs have a status subresource specification (Spec.versions[].Schema.OpenAPIV3Schema.Properties["status"]).
Result Type|informative
Suggested Remediation|make sure that all the CRDs have a meaningful status specification (Spec.versions[].Schema.OpenAPIV3Schema.Properties["status"]).
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2
#### termination-policy

Property|Description
---|---
Test Case Name|termination-policy
Test Case Label|observability-termination-policy
Unique ID|http://test-network-function.com/testcases/observability/termination-policy
Version|v1.0.0
Description|http://test-network-function.com/testcases/observability/termination-policy check that all containers are using terminationMessagePolicy: FallbackToLogsOnError
Result Type|informative
Suggested Remediation|make sure containers are all using FallbackToLogsOnError in terminationMessagePolicy
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 15.1

### operator

#### install-source

Property|Description
---|---
Test Case Name|install-source
Test Case Label|operator-install-source
Unique ID|http://test-network-function.com/testcases/operator/install-source
Version|v1.0.0
Description|http://test-network-function.com/testcases/operator/install-source tests whether a CNF Operator is installed via OLM.
Result Type|normative
Suggested Remediation|Ensure that your Operator is installed via OLM.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2.12 and Section 6.3.3
#### install-status-no-privileges

Property|Description
---|---
Test Case Name|install-status-no-privileges
Test Case Label|operator-install-status-no-privileges
Unique ID|http://test-network-function.com/testcases/operator/install-status-no-privileges
Version|v1.0.0
Description|http://test-network-function.com/testcases/operator/install-status-no-privileges The operator is not installed with privileged rights. Test passes if clusterPermissions is not present in the CSV manifest or is present  with no resourceNames under its rules.
Result Type|normative
Suggested Remediation|Make sure all the CNF operators have no privileges on cluster resources.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2.12 and Section 6.3.3
#### install-status-succeeded

Property|Description
---|---
Test Case Name|install-status-succeeded
Test Case Label|operator-install-status-succeeded
Unique ID|http://test-network-function.com/testcases/operator/install-status-succeeded
Version|v1.0.0
Description|http://test-network-function.com/testcases/operator/install-status-succeeded Ensures that the target CNF operators report "Succeeded" as their installation status.
Result Type|normative
Suggested Remediation|Make sure all the CNF operators have been successfully installed by OLM.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2.12 and Section 6.3.3

### platform-alteration

#### base-image

Property|Description
---|---
Test Case Name|base-image
Test Case Label|platform-alteration-base-image
Unique ID|http://test-network-function.com/testcases/platform-alteration/base-image
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/base-image ensures that the Container Base Image is not altered post-startup.  This test is a heuristic, and ensures that there are no changes to the following directories: 1) /var/lib/rpm 2) /var/lib/dpkg 3) /bin 4) /sbin 5) /lib 6) /lib64 7) /usr/bin 8) /usr/sbin 9) /usr/lib 10) /usr/lib64
Result Type|normative
Suggested Remediation|Ensure that Container applications do not modify the Container Base Image.  In particular, ensure that the following directories are not modified: 1) /var/lib/rpm 2) /var/lib/dpkg 3) /bin 4) /sbin 5) /lib 6) /lib64 7) /usr/bin 8) /usr/sbin 9) /usr/lib 10) /usr/lib64 Ensure that all required binaries are built directly into the container image, and are not installed post startup.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2.2
#### boot-params

Property|Description
---|---
Test Case Name|boot-params
Test Case Label|platform-alteration-boot-params
Unique ID|http://test-network-function.com/testcases/platform-alteration/boot-params
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/boot-params tests that boot parameters are set through the MachineConfigOperator, and not set manually on the Node.
Result Type|normative
Suggested Remediation|Ensure that boot parameters are set directly through the MachineConfigOperator, or indirectly through the PerformanceAddonOperator.  Boot parameters should not be changed directly through the Node, as OpenShift should manage the changes for you.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2.13 and 6.2.14
#### hugepages-config

Property|Description
---|---
Test Case Name|hugepages-config
Test Case Label|platform-alteration-hugepages-config
Unique ID|http://test-network-function.com/testcases/platform-alteration/hugepages-config
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/hugepages-config checks to see that HugePage settings have been configured through MachineConfig, and not manually on the underlying Node.  This test case applies only to Nodes that are configured with the "worker" MachineConfigSet.  First, the "worker" MachineConfig is polled, and the Hugepage settings are extracted.  Next, the underlying Nodes are polled for configured HugePages through inspection of /proc/meminfo.  The results are compared, and the test passes only if they are the same.
Result Type|normative
Suggested Remediation|HugePage settings should be configured either directly through the MachineConfigOperator or indirectly using the PerformanceAddonOperator.  This ensures that OpenShift is aware of the special MachineConfig requirements, and can provision your CNF on a Node that is part of the corresponding MachineConfigSet.  Avoid making changes directly to an underlying Node, and let OpenShift handle the heavy lifting of configuring advanced settings.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2
#### is-selinux-enforcing

Property|Description
---|---
Test Case Name|is-selinux-enforcing
Test Case Label|platform-alteration-is-selinux-enforcing
Unique ID|http://test-network-function.com/testcases/platform-alteration/is-selinux-enforcing
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/is-selinux-enforcing verifies that all openshift platform/cluster nodes have selinux in "Enforcing" mode.
Result Type|normative
Suggested Remediation|configure selinux and enable enforcing mode.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 11.3 Pod Security
#### isredhat-release

Property|Description
---|---
Test Case Name|isredhat-release
Test Case Label|platform-alteration-isredhat-release
Unique ID|http://test-network-function.com/testcases/platform-alteration/isredhat-release
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/isredhat-release verifies if the container base image is redhat.
Result Type|normative
Suggested Remediation|build a new docker image that's based on UBI (redhat universal base image).
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2
#### sysctl-config

Property|Description
---|---
Test Case Name|sysctl-config
Test Case Label|platform-alteration-sysctl-config
Unique ID|http://test-network-function.com/testcases/platform-alteration/sysctl-config
Version|v1.0.0
Description|http://test-network-function.com/testcases/lifecycle/pod-recreation tests that no one has changed the node's sysctl configs after the node 			was created, the tests works by checking if the sysctl configs are consistent with the 			MachineConfig CR which defines how the node should be configured
Result Type|normative
Suggested Remediation|You should recreate the node or change the sysctls, recreating is recommended because there might be other unknown changes
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2
#### tainted-node-kernel

Property|Description
---|---
Test Case Name|tainted-node-kernel
Test Case Label|platform-alteration-tainted-node-kernel
Unique ID|http://test-network-function.com/testcases/platform-alteration/tainted-node-kernel
Version|v1.0.0
Description|http://test-network-function.com/testcases/platform-alteration/tainted-node-kernel ensures that the Node(s) hosting CNFs do not utilize tainted kernels. This test case is especially important to support Highly Available CNFs, since when a CNF is re-instantiated on a backup Node, that Node's kernel may not have the same hacks.'
Result Type|normative
Suggested Remediation|Test failure indicates that the underlying Node's' kernel is tainted.  Ensure that you have not altered underlying Node(s) kernels in order to run the CNF.
Best Practice Reference|[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf) Section 6.2.14

## Test Case Building Blocks Catalog

A number of Test Case Building Blocks, or `tnf.Test`s, are included out of the box.  This is a summary of the available implementations:### automountservice
Property|Description
---|---
Test Name|automountservice
Unique ID|http://test-network-function.com/tests/automountservice
Version|v1.0.0
Description|check if automount service account token is set to false
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### clusterVersion
Property|Description
---|---
Test Name|clusterVersion
Unique ID|http://test-network-function.com/tests/clusterVersion
Version|v1.0.0
Description|Extracts OCP versions from the cluster
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### clusterrolebinding
Property|Description
---|---
Test Name|clusterrolebinding
Unique ID|http://test-network-function.com/tests/clusterrolebinding
Version|v1.0.0
Description|A generic test used to test ClusterRoleBindings of CNF pod's ServiceAccount.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### command
Property|Description
---|---
Test Name|command
Unique ID|http://test-network-function.com/tests/command
Version|v1.0.0
Description|A generic test used with any command and would match any output. The caller is responsible for interpreting the output and extracting data from it.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|

### container-pod
Property|Description
---|---
Test Name|container-pod
Unique ID|http://test-network-function.com/tests/container/pod
Version|v1.0.0
Description|A container-specific test suite used to verify various aspects of the underlying container.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`jq`, `oc`

### crdStatusExistence
Property|Description
---|---
Test Name|crdStatusExistence
Unique ID|http://test-network-function.com/tests/crdStatusExistence
Version|v1.0.0
Description|Checks whether a give CRD has status subresource specification.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `jq`

### csiDriver
Property|Description
---|---
Test Name|csiDriver
Unique ID|http://test-network-function.com/tests/csiDriver
Version|v1.0.0
Description|extracts the csi driver info in the cluster
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### currentKernelCmdlineArgs
Property|Description
---|---
Test Name|currentKernelCmdlineArgs
Unique ID|http://test-network-function.com/tests/currentKernelCmdlineArgs
Version|v1.0.0
Description|A generic test used to get node's /proc/cmdline
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`cat`

### daemonset
Property|Description
---|---
Test Name|daemonset
Unique ID|http://test-network-function.com/tests/daemonset
Version|v1.0.0
Description|check whether a given daemonset was deployed successfully
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### deploymentsnodes
Property|Description
---|---
Test Name|deploymentsnodes
Unique ID|http://test-network-function.com/tests/deploymentsnodes
Version|v1.0.0
Description|A generic test used to drain node from its deployment pods
Result Type|normative
Intrusive|true
Modifications Persist After Test|true
Runtime Binaries Required|`jq`, `echo`

### deploymentsnodes
Property|Description
---|---
Test Name|deploymentsnodes
Unique ID|http://test-network-function.com/tests/deploymentsnodes
Version|v1.0.0
Description|A generic test used to read node names of pods owned by deployments in namespace
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `grep`

### generic-cnf_fs_diff
Property|Description
---|---
Test Name|generic-cnf_fs_diff
Unique ID|http://test-network-function.com/tests/generic/cnf_fs_diff
Version|v1.0.0
Description|A test used to check if there were no installation during container runtime
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`grep`, `cut`

### generic-containerId
Property|Description
---|---
Test Name|generic-containerId
Unique ID|http://test-network-function.com/tests/generic/containerId
Version|v1.0.0
Description|A test used to check what is the id of the crio generated container this command is run from
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`cat`

### generic-version
Property|Description
---|---
Test Name|generic-version
Unique ID|http://test-network-function.com/tests/generic/version
Version|v1.0.0
Description|A generic test used to determine if a target container/machine is based on RHEL.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`cat`

### gracePeriod
Property|Description
---|---
Test Name|gracePeriod
Unique ID|http://test-network-function.com/tests/gracePeriod
Version|v1.0.0
Description|A generic test used to extract the CNF pod's terminationGracePeriod.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`grep`, `cut`

### grubKernelCmdlineArgs
Property|Description
---|---
Test Name|grubKernelCmdlineArgs
Unique ID|http://test-network-function.com/tests/grubKernelCmdlineArgs
Version|v1.0.0
Description|A generic test used to get node's next boot kernel args
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`ls`, `sort`, `head`, `cut`, `oc`

### hostname
Property|Description
---|---
Test Name|hostname
Unique ID|http://test-network-function.com/tests/hostname
Version|v1.0.0
Description|A generic test used to check the hostname of a target machine/container.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`hostname`

### hugepages
Property|Description
---|---
Test Name|hugepages
Unique ID|http://test-network-function.com/tests/hugepages
Version|v1.0.0
Description|A generic test used to read cluster's hugepages configuration
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`grep`, `cut`, `oc`, `grep`

### imagepullpolicy
Property|Description
---|---
Test Name|imagepullpolicy
Unique ID|http://test-network-function.com/tests/imagepullpolicy
Version|v1.0.0
Description|A generic test used to get Image Pull Policy type.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### ipaddr
Property|Description
---|---
Test Name|ipaddr
Unique ID|http://test-network-function.com/tests/ipaddr
Version|v1.0.0
Description|A generic test used to derive the default network interface IP address of a target container.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`ip`

### logging
Property|Description
---|---
Test Name|logging
Unique ID|http://test-network-function.com/tests/logging
Version|v1.0.0
Description|A test used to check logs are redirected to stderr/stdout
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `wc`

### mckernelarguments
Property|Description
---|---
Test Name|mckernelarguments
Unique ID|http://test-network-function.com/tests/mckernelarguments
Version|v1.0.0
Description|A generic test used to get an mc's kernel arguments
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `jq`, `echo`

### node-uncordon
Property|Description
---|---
Test Name|node-uncordon
Unique ID|http://test-network-function.com/tests/node/uncordon
Version|v1.0.0
Description|A generic test used to uncordon a node
Result Type|normative
Intrusive|true
Modifications Persist After Test|true
Runtime Binaries Required|`oc`

### nodedebug
Property|Description
---|---
Test Name|nodedebug
Unique ID|http://test-network-function.com/tests/nodedebug
Version|v1.0.0
Description|A generic test used to execute a command in a node
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `echo`

### nodehugepages
Property|Description
---|---
Test Name|nodehugepages
Unique ID|http://test-network-function.com/tests/nodehugepages
Version|v1.0.0
Description|A generic test used to verify a node's hugepages configuration
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `grep`

### nodemcname
Property|Description
---|---
Test Name|nodemcname
Unique ID|http://test-network-function.com/tests/nodemcname
Version|v1.0.0
Description|A generic test used to get a node's current mc
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `grep`

### nodenames
Property|Description
---|---
Test Name|nodenames
Unique ID|http://test-network-function.com/tests/nodenames
Version|v1.0.0
Description|A generic test used to get node names
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### nodeport
Property|Description
---|---
Test Name|nodeport
Unique ID|http://test-network-function.com/tests/nodeport
Version|v1.0.0
Description|A generic test used to test services of CNF pod's namespace.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `grep`

### nodes
Property|Description
---|---
Test Name|nodes
Unique ID|http://test-network-function.com/tests/nodes
Version|v1.0.0
Description|Polls the state of the OpenShift cluster nodes using "oc get nodes -o json".
Result Type|
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### nodeselector
Property|Description
---|---
Test Name|nodeselector
Unique ID|http://test-network-function.com/tests/nodeselector
Version|v1.0.0
Description|A generic test used to verify a pod's nodeSelector and nodeAffinity configuration
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `grep`

### nodetainted
Property|Description
---|---
Test Name|nodetainted
Unique ID|http://test-network-function.com/tests/nodetainted
Version|v1.0.0
Description|A generic test used to test whether node is tainted
Result Type|informative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `cat`, `echo`

### operator
Property|Description
---|---
Test Name|operator
Unique ID|http://test-network-function.com/tests/operator
Version|v1.0.0
Description|An operator-specific test used to exercise the behavior of a given operator.  In the current offering, we check if the operator ClusterServiceVersion (CSV) is installed properly.  A CSV is a YAML manifest created from Operator metadata that assists the Operator Lifecycle Manager (OLM) in running the Operator.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`jq`, `oc`

### operator-check-subscription
Property|Description
---|---
Test Name|operator-check-subscription
Unique ID|http://test-network-function.com/tests/operator/check-subscription
Version|v1.0.0
Description|A test used to check the subscription of a given operator
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### owners
Property|Description
---|---
Test Name|owners
Unique ID|http://test-network-function.com/tests/owners
Version|v1.0.0
Description|A generic test used to verify pod is managed by a ReplicaSet/StatefulSet
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`cat`

### ping
Property|Description
---|---
Test Name|ping
Unique ID|http://test-network-function.com/tests/ping
Version|v1.0.0
Description|A generic test used to test ICMP connectivity from a source machine/container to a target destination.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`ping`

### podnodename
Property|Description
---|---
Test Name|podnodename
Unique ID|http://test-network-function.com/tests/podnodename
Version|v1.0.0
Description|A generic test used to get a pod's node
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### podsets
Property|Description
---|---
Test Name|podsets
Unique ID|http://test-network-function.com/tests/podsets
Version|v1.0.0
Description|A generic test used to read namespace's deployments/statefulsets
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### readRemoteFile
Property|Description
---|---
Test Name|readRemoteFile
Unique ID|http://test-network-function.com/tests/readRemoteFile
Version|v1.0.0
Description|A generic test used to read a specified file at a specified node
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`echo`

### rolebinding
Property|Description
---|---
Test Name|rolebinding
Unique ID|http://test-network-function.com/tests/rolebinding
Version|v1.0.0
Description|A generic test used to test RoleBindings of CNF pod's ServiceAccount.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`cat`, `oc`

### scaling
Property|Description
---|---
Test Name|scaling
Unique ID|http://test-network-function.com/tests/scaling
Version|v1.0.0
Description|A test to check the deployments scale in/out. The tests issues the oc scale command on a deployment for a given number of replicas and checks whether the command output is valid.
Result Type|normative
Intrusive|true
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### serviceaccount
Property|Description
---|---
Test Name|serviceaccount
Unique ID|http://test-network-function.com/tests/serviceaccount
Version|v1.0.0
Description|A generic test used to extract the CNF pod's ServiceAccount name.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`grep`, `cut`

### shutdown
Property|Description
---|---
Test Name|shutdown
Unique ID|http://test-network-function.com/tests/shutdown
Version|v1.0.0
Description|A test used to check pre-stop lifecycle is defined
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### sysctlAllConfigsArgs
Property|Description
---|---
Test Name|sysctlAllConfigsArgs
Unique ID|http://test-network-function.com/tests/sysctlAllConfigsArgs
Version|v1.0.0
Description|A test used to find all sysctl configuration args
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`sysctl`

### sysctlConfigFilesList
Property|Description
---|---
Test Name|sysctlConfigFilesList
Unique ID|http://test-network-function.com/tests/sysctlConfigFilesList
Version|v1.0.0
Description|A generic test used to get node's list of sysctl config files
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`cat`

### testPodHighAvailability
Property|Description
---|---
Test Name|testPodHighAvailability
Unique ID|http://test-network-function.com/tests/testPodHighAvailability
Version|v1.0.0
Description|A generic test used to check pod's replica and podAntiAffinity configuration in high availability mode
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

