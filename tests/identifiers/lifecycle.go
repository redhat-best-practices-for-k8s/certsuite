// Copyright (C) 2021-2026 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package identifiers

import (
	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
)

var (
	TestAffinityRequiredPods                    claim.Identifier
	TestCPUIsolationIdentifier                  claim.Identifier
	TestContainerPostStartIdentifier            claim.Identifier
	TestContainerPrestopIdentifier              claim.Identifier
	TestCrdScalingIdentifier                    claim.Identifier
	TestDeploymentScalingIdentifier             claim.Identifier
	TestImagePullPolicyIdentifier               claim.Identifier
	TestLivenessProbeIdentifier                 claim.Identifier
	TestPersistentVolumeReclaimPolicyIdentifier claim.Identifier
	TestPodDeploymentBestPracticesIdentifier    claim.Identifier
	TestPodHighAvailabilityBestPractices        claim.Identifier
	TestPodNodeSelectorAndAffinityBestPractices claim.Identifier
	TestPodRecreationIdentifier                 claim.Identifier
	TestPodTolerationBypassIdentifier           claim.Identifier
	TestReadinessProbeIdentifier                claim.Identifier
	TestStartupProbeIdentifier                  claim.Identifier
	TestStatefulSetScalingIdentifier            claim.Identifier
	TestStorageProvisioner                      claim.Identifier
	TestTopologySpreadConstraint                claim.Identifier
)

//nolint:funlen
func init() {
	TestAffinityRequiredPods = AddCatalogEntry(
		"affinity-required-pods",
		common.LifecycleTestKey,
		`Checks that affinity rules are in place if AffinityRequired: 'true' labels are set on Pods.`,
		AffinityRequiredRemediation,
		NoDocumentedProcess,
		TestAffinityRequiredPodsDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestCPUIsolationIdentifier = AddCatalogEntry(
		"cpu-isolation",
		common.LifecycleTestKey,
		`CPU isolation requires: For each container within the pod, resource requests and limits must be identical. If cpu requests and limits are not identical and in whole units (Guaranteed pods with exclusive cpus), your pods will not be tested for compliance. The runTimeClassName must be specified. Annotations required disabling CPU and IRQ load-balancing.`, //nolint:lll
		CPUIsolationRemediation,
		NoDocumentedProcess,
		TestCPUIsolationIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestContainerPostStartIdentifier = AddCatalogEntry(
		"container-poststart",
		common.LifecycleTestKey,
		`Ensure that the containers lifecycle postStart management feature is configured. A container must receive important events from the platform and conform/react to these events properly. For example, a container should catch SIGTERM from the platform and shutdown as quickly as possible. Other typically important events from the platform are PostStart to initialize before servicing requests and PreStop to release resources cleanly before shutting down.`,                                                                                                                                                   //nolint:lll
		`PostStart is normally used to configure the container, set up dependencies, and record the new creation. You could use this event to check that a required API is available before the container's main work begins. Kubernetes will not change the container's state to Running until the PostStart script has executed successfully. For details, see https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks. PostStart is used to configure container, set up dependencies, record new creation. It can also be used to check that a required API is available before the container's work begins.`, //nolint:lll
		ContainerPostStartIdentifierRemediation,
		TestContainerPostStartIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestContainerPrestopIdentifier = AddCatalogEntry(
		"container-prestop",
		common.LifecycleTestKey,
		`Ensure that the containers lifecycle preStop management feature is configured. The most basic requirement for the lifecycle management of Pods in OpenShift are the ability to start and stop correctly. There are different ways a pod can stop on an OpenShift cluster. One way is that the pod can remain alive but non-functional. Another way is that the pod can crash and become non-functional. When pods are shut down by the platform they are sent a SIGTERM signal which means that the process in the container should start shutting down, closing connections and stopping all activity. If the pod doesn't shut down within the default 30 seconds then the platform will send a SIGKILL signal which will stop the pod immediately. This method isn't as clean and the default time between the SIGTERM and SIGKILL can be modified based on the requirements of the application. Containers must handle SIGTERM and shut down gracefully.`, //nolint:lll
		`The preStop can be used to gracefully stop the container and clean resources (e.g., DB connection). For details, see https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks. All pods must respond to SIGTERM signal and shutdown gracefully with a zero exit code.`, //nolint:lll
		ContainerPrestopIdentifierRemediation,
		TestContainerPrestopIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestCrdScalingIdentifier = AddCatalogEntry(
		"crd-scaling",
		common.LifecycleTestKey,
		`Tests that a workload's CRD support scale in/out operations. First, the test starts getting the current replicaCount (N) of the crd/s with the Pod Under Test. Then, it executes the scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the crd/s. In case of crd that are managed by HPA the test is changing the min and max value to crd Replica - 1 during scale-in and the original replicaCount again for both min/max during the scale-out stage. Lastly its restoring the original min/max replica of the crd/s`, //nolint:lll
		CrdScalingRemediation,
		NoDocumentedProcess+NotApplicableSNO,
		TestCrdScalingIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon,
	)
	TestDeploymentScalingIdentifier = AddCatalogEntry(
		"deployment-scaling",
		common.LifecycleTestKey,
		`Tests that workload deployments support scale in/out operations. First, the test starts getting the current replicaCount (N) of the deployment/s with the Pod Under Test. Then, it executes the scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the deployment/s. In case of deployments that are managed by HPA the test is changing the min and max value to deployment Replica - 1 during scale-in and the original replicaCount again for both min/max during the scale-out stage. Lastly its restoring the original min/max replica of the deployment/s`, //nolint:lll
		DeploymentScalingRemediation,
		NoDocumentedProcess+NotApplicableSNO,
		TestDeploymentScalingIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestImagePullPolicyIdentifier = AddCatalogEntry(
		"image-pull-policy",
		common.LifecycleTestKey,
		`Ensure that the containers under test are using IfNotPresent as Image Pull Policy. If there is a situation where the container dies and needs to be restarted, the image pull policy becomes important. PullIfNotPresent is recommended so that a loss of image registry access does not prevent the pod from restarting.`, //nolint:lll
		ImagePullPolicyRemediation,
		NoDocumentedProcess,
		TestImagePullPolicyIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestLivenessProbeIdentifier = AddCatalogEntry(
		"liveness-probe",
		common.LifecycleTestKey,
		`Check that all containers under test have liveness probe defined. The most basic requirement for the lifecycle management of Pods in OpenShift are the ability to start and stop correctly. When starting up, health probes like liveness and readiness checks can be put into place to ensure the application is functioning properly.`, //nolint:lll
		LivenessProbeRemediation+` workloads shall self-recover from common failures like pod failure, host failure, and network failure. Kubernetes native mechanisms such as health-checks (Liveness, Readiness and Startup Probes) shall be employed at a minimum.`,                                                                            //nolint:lll
		NoDocumentedProcess,
		TestLivenessProbeIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestPersistentVolumeReclaimPolicyIdentifier = AddCatalogEntry(
		"persistent-volume-reclaim-policy",
		common.LifecycleTestKey,
		`Check that the persistent volumes the workloads pods are using have a reclaim policy of delete. Network Functions should clear persistent storage by deleting their PVs when removing their application from a cluster.`,
		PersistentVolumeReclaimPolicyRemediation,
		NoDocumentedProcess,
		TestPersistentVolumeReclaimPolicyIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestPodDeploymentBestPracticesIdentifier = AddCatalogEntry(
		"pod-owner-type",
		common.LifecycleTestKey,
		`Tests that the workload Pods are deployed as part of a ReplicaSet(s)/StatefulSet(s).`,
		PodDeploymentBestPracticesRemediation,
		NoDocumentedProcess+` Pods should not be deployed as DaemonSet or naked pods.`,
		TestPodDeploymentBestPracticesIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestPodHighAvailabilityBestPractices = AddCatalogEntry(
		"pod-high-availability",
		common.LifecycleTestKey,
		`Ensures that workloads Pods specify podAntiAffinity rules and replica value is set to more than 1.`,
		PodHighAvailabilityBestPracticesRemediation,
		NoDocumentedProcess+NotApplicableSNO,
		TestPodHighAvailabilityBestPracticesDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestPodNodeSelectorAndAffinityBestPractices = AddCatalogEntry(
		"pod-scheduling",
		common.LifecycleTestKey,
		`Ensures that workload Pods do not specify nodeSelector or nodeAffinity. In most cases, Pods should allow for instantiation on any underlying Node. Workloads shall not use node selectors nor taints/tolerations to assign pod location.`,
		PodNodeSelectorAndAffinityBestPracticesRemediation,
		`Exception will only be considered if application requires specialized hardware. Must specify which container requires special hardware and why.`,
		TestPodNodeSelectorAndAffinityBestPracticesDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Mandatory,
			Extended: Optional,
		},
		TagTelco)

	TestPodRecreationIdentifier = AddCatalogEntry(
		"pod-recreation",
		common.LifecycleTestKey,
		`Tests that a workload is configured to support High Availability. First, this test cordons and drains a Node that hosts the workload Pod. Next, the test ensures that OpenShift can re-instantiate the Pod on another Node, and that the actual replica count matches the desired replica count.`, //nolint:lll
		PodRecreationRemediation,
		`No exceptions - workloads should be able to be restarted/recreated.`,
		TestPodRecreationIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestPodTolerationBypassIdentifier = AddCatalogEntry(
		"pod-toleration-bypass",
		common.LifecycleTestKey,
		`Check that pods do not have NoExecute, PreferNoSchedule, or NoSchedule tolerations that have been modified from the default.`,
		PodTolerationBypassRemediation,
		NoDocumentedProcess,
		TestPodTolerationBypassIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestReadinessProbeIdentifier = AddCatalogEntry(
		"readiness-probe",
		common.LifecycleTestKey,
		`Check that all containers under test have readiness probe defined. There are different ways a pod can stop on on OpenShift cluster. One way is that the pod can remain alive but non-functional. Another way is that the pod can crash and become non-functional. In the first case, if the administrator has implemented liveness and readiness checks, OpenShift can stop the pod and either restart it on the same node or a different node in the cluster. For the second case, when the application in the pod stops, it should exit with a code and write suitable log entries to help the administrator diagnose what the issue was that caused the problem.`, //nolint:lll
		ReadinessProbeRemediation,
		NoDocumentedProcess,
		TestReadinessProbeIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestStartupProbeIdentifier = AddCatalogEntry(
		"startup-probe",
		common.LifecycleTestKey,
		`Check that all containers under test have startup probe defined. Workloads shall self-recover from common failures like pod failure, host failure, and network failure. Kubernetes native mechanisms such as health-checks (Liveness, Readiness and Startup Probes) shall be employed at a minimum.`, //nolint:lll
		StartupProbeRemediation,
		NoDocumentedProcess,
		TestStartupProbeIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestStatefulSetScalingIdentifier = AddCatalogEntry(
		"statefulset-scaling",
		common.LifecycleTestKey,
		`Tests that workload statefulsets support scale in/out operations. First, the test starts getting the current replicaCount (N) of the statefulset/s with the Pod Under Test. Then, it executes the scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the statefulset/s. In case of statefulsets that are managed by HPA the test is changing the min and max value to statefulset Replica - 1 during scale-in and the original replicaCount again for both min/max during the scale-out stage. Lastly its restoring the original min/max replica of the statefulset/s`, //nolint:lll
		StatefulSetScalingRemediation,
		NoDocumentedProcess+NotApplicableSNO,
		TestStatefulSetScalingIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestStorageProvisioner = AddCatalogEntry(
		"storage-provisioner",
		common.LifecycleTestKey,
		`Checks that pods do not place persistent volumes on local storage in multinode clusters. Local storage is recommended for single node clusters, but only one type of local storage should be installed (lvms or noprovisioner).`,
		CheckStorageProvisionerRemediation,
		NoExceptions,
		TestStorageProvisionerDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestTopologySpreadConstraint = AddCatalogEntry(
		"topology-spread-constraint",
		common.LifecycleTestKey,
		`Ensures that Deployments using TopologySpreadConstraints include constraints for both hostname and zone topology keys. This helps telco workloads avoid needing to tweak PodDisruptionBudgets before platform upgrades. If TopologySpreadConstraints is not defined, the test passes as Kubernetes scheduler implicitly uses hostname and zone constraints.`+NotApplicableSNO, //nolint:lll
		TopologySpreadConstraintRemediation,
		NoDocumentedProcess,
		TestTopologySpreadConstraintDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagTelco)
}
