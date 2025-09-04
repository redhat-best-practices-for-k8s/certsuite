// Copyright (C) 2020-2024 Red Hat, Inc.
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

package provider

import (
	"context"
	"regexp"
	"time"

	"fmt"
	"strings"

	"encoding/json"

	nadClient "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	configv1 "github.com/openshift/api/config/v1"
	mcv1 "github.com/openshift/api/machineconfiguration/v1"
	olmv1 "github.com/operator-framework/api/pkg/operators/v1"
	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmpkgv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/autodiscover"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	k8sPrivilegedDs "github.com/redhat-best-practices-for-k8s/privileged-daemonset"
	plibRuntime "github.com/redhat-openshift-ecosystem/openshift-preflight/certification"
	"helm.sh/helm/v3/pkg/release"
	scalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CentOS Stream CoreOS starts being used instead of rhcos from OCP 4.13 latest.
const (
	AffinityRequiredKey              = "AffinityRequired"
	containerName                    = "container-00"
	DaemonSetName                    = "certsuite-probe"
	probePodsTimeout                 = 5 * time.Minute
	CniNetworksStatusKey             = "k8s.v1.cni.cncf.io/network-status"
	skipConnectivityTestsLabel       = "redhat-best-practices-for-k8s.com/skip_connectivity_tests"
	skipMultusConnectivityTestsLabel = "redhat-best-practices-for-k8s.com/skip_multus_connectivity_tests"
	rhcosName                        = "Red Hat Enterprise Linux CoreOS"
	cscosName                        = "CentOS Stream CoreOS"
	rhelName                         = "Red Hat Enterprise Linux"
)

// Node's roles labels. Node is role R if it has **any** of the labels of each list.
// Master's role label "master" is deprecated since k8s 1.20.
var (
	WorkerLabels = []string{"node-role.kubernetes.io/worker"}
	MasterLabels = []string{"node-role.kubernetes.io/master", "node-role.kubernetes.io/control-plane"}
)

// TestEnvironment Provides runtime information for test execution
//
// This struct holds configuration, cluster state, and collected resources
// needed during tests. It tracks pods, nodes, operators, catalogs, and various
// Kubernetes objects while exposing helper methods to filter them by
// characteristics such as CPU isolation or affinity requirements. The data is
// populated from the test harness and can be refreshed when the underlying
// environment changes.
type TestEnvironment struct { // rename this with testTarget
	Namespaces     []string `json:"testNamespaces"`
	AbnormalEvents []*Event

	// Pod Groupings
	Pods            []*Pod                 `json:"testPods"`
	ProbePods       map[string]*corev1.Pod // map from nodename to probePod
	AllPods         []*Pod                 `json:"AllPods"`
	CSVToPodListMap map[string][]*Pod      `json:"CSVToPodListMap"`
	PodStates       autodiscover.PodStates `json:"podStates"`

	// Deployment Groupings
	Deployments []*Deployment `json:"testDeployments"`
	// StatefulSet Groupings
	StatefulSets []*StatefulSet `json:"testStatefulSets"`

	// Note: Containers is a filtered list of objects based on a block list of disallowed container names.
	Containers             []*Container `json:"testContainers"`
	Operators              []*Operator  `json:"testOperators"`
	AllOperators           []*Operator  `json:"AllOperators"`
	AllOperatorsSummary    []string     `json:"AllOperatorsSummary"`
	AllCsvs                []*olmv1Alpha.ClusterServiceVersion
	PersistentVolumes      []corev1.PersistentVolume
	PersistentVolumeClaims []corev1.PersistentVolumeClaim
	ClusterRoleBindings    []rbacv1.ClusterRoleBinding
	RoleBindings           []rbacv1.RoleBinding
	Roles                  []rbacv1.Role

	Config  configuration.TestConfiguration
	params  configuration.TestParameters
	Crds    []*apiextv1.CustomResourceDefinition `json:"testCrds"`
	AllCrds []*apiextv1.CustomResourceDefinition

	HorizontalScaler             []*scalingv1.HorizontalPodAutoscaler `json:"testHorizontalScaler"`
	Services                     []*corev1.Service                    `json:"testServices"`
	AllServices                  []*corev1.Service                    `json:"testAllServices"`
	ServiceAccounts              []*corev1.ServiceAccount             `json:"testServiceAccounts"`
	AllServiceAccounts           []*corev1.ServiceAccount             `json:"AllServiceAccounts"`
	AllServiceAccountsMap        map[string]*corev1.ServiceAccount
	Nodes                        map[string]Node    `json:"-"`
	K8sVersion                   string             `json:"-"`
	OpenshiftVersion             string             `json:"-"`
	OCPStatus                    string             `json:"-"`
	HelmChartReleases            []*release.Release `json:"testHelmChartReleases"`
	ResourceQuotas               []corev1.ResourceQuota
	PodDisruptionBudgets         []policyv1.PodDisruptionBudget
	NetworkPolicies              []networkingv1.NetworkPolicy
	AllInstallPlans              []*olmv1Alpha.InstallPlan   `json:"AllInstallPlans"`
	AllSubscriptions             []olmv1Alpha.Subscription   `json:"AllSubscriptions"`
	AllCatalogSources            []*olmv1Alpha.CatalogSource `json:"AllCatalogSources"`
	AllPackageManifests          []*olmpkgv1.PackageManifest `json:"AllPackageManifests"`
	OperatorGroups               []*olmv1.OperatorGroup      `json:"OperatorGroups"`
	SriovNetworks                []unstructured.Unstructured
	AllSriovNetworks             []unstructured.Unstructured
	SriovNetworkNodePolicies     []unstructured.Unstructured
	AllSriovNetworkNodePolicies  []unstructured.Unstructured
	NetworkAttachmentDefinitions []nadClient.NetworkAttachmentDefinition
	ClusterOperators             []configv1.ClusterOperator
	IstioServiceMeshFound        bool
	ValidProtocolNames           []string
	DaemonsetFailedToSpawn       bool
	ScaleCrUnderTest             []ScaleObject
	StorageClassList             []storagev1.StorageClass
	ExecutedBy                   string
	PartnerName                  string
	CollectorAppPassword         string
	CollectorAppEndpoint         string
	ConnectAPIKey                string
	ConnectProjectID             string
	ConnectAPIBaseURL            string
	ConnectAPIProxyURL           string
	ConnectAPIProxyPort          string
	SkipPreflight                bool
}

// MachineConfig Encapsulates a machine configuration including systemd unit definitions
//
// The structure embeds the core machine configuration type from the Kubernetes
// API, adding a Config field that contains systemd unit information. It holds
// an array of unit descriptors, each specifying a name and contents for a
// systemd service file. This representation is used to unmarshal the raw JSON
// of a MachineConfig resource into usable Go objects.
type MachineConfig struct {
	*mcv1.MachineConfig
	Config struct {
		Systemd struct {
			Units []struct {
				Contents string `json:"contents"`
				Name     string `json:"name"`
			} `json:"units"`
		} `json:"systemd"`
	} `json:"config"`
}

// CniNetworkInterface Represents a network interface configured by CNI
//
// This struct holds details about a pod’s network attachment, including the
// interface name, assigned IP addresses, whether it is the default route, DNS
// settings, and additional device metadata. The fields are populated from the
// Kubernetes annotation that lists all attached networks for a pod.
type CniNetworkInterface struct {
	Name       string                 `json:"name"`
	Interface  string                 `json:"interface"`
	IPs        []string               `json:"ips"`
	Default    bool                   `json:"default"`
	DNS        map[string]interface{} `json:"dns"`
	DeviceInfo deviceInfo             `json:"device-info"`
}

// ScaleObject Represents a Kubernetes custom resource scaling configuration
//
// This struct holds the desired scale for a custom resource along with its
// group and resource identifiers. The Scale field contains the target number of
// replicas, while GroupResourceSchema specifies which API group and kind it
// applies to. It is used by provider functions to adjust or query resource
// scaling settings.
type ScaleObject struct {
	Scale               CrScale
	GroupResourceSchema schema.GroupResource
}

// deviceInfo Holds low-level device details
//
// This struct stores information about a device, including its type and version
// strings as well as a PCI configuration structure. The PCI field contains the
// specific bus, device, and function identifiers that enable precise hardware
// identification. Together, these fields provide a compact representation of
// the device’s identity for use in diagnostics or policy enforcement.
type deviceInfo struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	PCI     pci    `json:"pci"`
}

// pci Represents a PCI device address
//
// This type holds the string representation of a PCI bus, device, and function
// identifier used by the provider to locate hardware resources. The single
// field contains the address formatted as "domain:bus:device.function" or a
// simplified form compatible with the system's PCI enumeration. It is utilized
// internally when mapping certificates or configurations to specific hardware
// components.
type pci struct {
	PciAddress string `json:"pci-address"`
}

// PreflightTest Represents the outcome of a pre‑flight check
//
// This structure holds information about a single test performed before
// deployment, including its name, a description of what it verifies, an
// optional error if the test failed, and suggested remediation steps. When the
// Error field is nil, the test succeeded; otherwise the value explains why it
// did not pass. The struct can be used to report results in logs or user
// interfaces.
type PreflightTest struct {
	Name        string
	Description string
	Remediation string
	Error       error
}

// PreflightResultsDB Stores the outcomes of preflight checks for a container image
//
// This structure holds lists of tests that passed, failed, or encountered
// errors during a preflight run. Each entry contains the test name,
// description, remediation guidance, and any error message if applicable. The
// data is used to report results back to callers and can be cached for reuse.
type PreflightResultsDB struct {
	Passed []PreflightTest
	Failed []PreflightTest
	Errors []PreflightTest
}

var (
	env    = TestEnvironment{}
	loaded = false
)

// deployDaemonSet Deploys the privileged probe daemonset
//
// This function first configures a Kubernetes client for privileged daemonset
// operations and checks whether the target daemonset is already running with
// the correct image. If it is not ready, it creates the daemonset using the
// specified image and resource limits from configuration parameters. After
// creation, it waits until all pods of the daemonset are ready or times out,
// returning an error if any step fails.
func deployDaemonSet(namespace string) error {
	k8sPrivilegedDs.SetDaemonSetClient(clientsholder.GetClientsHolder().K8sClient)

	dsImage := env.params.CertSuiteProbeImage
	if k8sPrivilegedDs.IsDaemonSetReady(DaemonSetName, namespace, dsImage) {
		return nil
	}

	matchLabels := make(map[string]string)
	matchLabels["name"] = DaemonSetName
	matchLabels["redhat-best-practices-for-k8s.com/app"] = DaemonSetName
	_, err := k8sPrivilegedDs.CreateDaemonSet(DaemonSetName, namespace, containerName, dsImage, matchLabels, probePodsTimeout,
		configuration.GetTestParameters().DaemonsetCPUReq,
		configuration.GetTestParameters().DaemonsetCPULim,
		configuration.GetTestParameters().DaemonsetMemReq,
		configuration.GetTestParameters().DaemonsetMemLim,
		corev1.PullIfNotPresent,
	)
	if err != nil {
		return fmt.Errorf("could not deploy certsuite daemonset, err=%v", err)
	}
	err = k8sPrivilegedDs.WaitDaemonsetReady(namespace, DaemonSetName, probePodsTimeout)
	if err != nil {
		return fmt.Errorf("timed out waiting for certsuite daemonset, err=%v", err)
	}

	return nil
}

// buildTestEnvironment initializes the test environment state
//
// The function starts by resetting the global environment structure and loading
// configuration parameters from a file. It then attempts to deploy a probe
// daemonset; if that fails it records the failure but continues with limited
// tests. Next, it performs autodiscovery of cluster resources such as
// operators, pods, services, CRDs, and more, populating many fields in the
// environment struct. Throughout the process, it logs progress, handles errors
// by terminating on critical failures, and measures the total time taken.
func buildTestEnvironment() { //nolint:funlen,gocyclo
	start := time.Now()
	env = TestEnvironment{}

	env.params = *configuration.GetTestParameters()
	config, err := configuration.LoadConfiguration(env.params.ConfigFile)
	if err != nil {
		log.Fatal("Cannot load configuration file: %v", err)
	}
	log.Debug("CERTSUITE configuration: %+v", config)

	// Wait for the probe pods to be ready before the autodiscovery starts.
	if err := deployDaemonSet(config.ProbeDaemonSetNamespace); err != nil {
		log.Error("The TNF daemonset could not be deployed, err: %v", err)
		// Because of this failure, we are only able to run a certain amount of tests that do not rely
		// on the existence of the daemonset probe pods.
		env.DaemonsetFailedToSpawn = true
	}

	data := autodiscover.DoAutoDiscover(&config)
	// OpenshiftVersion needs to be set asap, as other helper functions will use it here.
	env.OpenshiftVersion = data.OpenshiftVersion
	env.Config = config
	env.Crds = data.Crds
	env.AllInstallPlans = data.AllInstallPlans
	env.OperatorGroups, err = GetAllOperatorGroups()
	if err != nil {
		log.Fatal("Cannot get OperatorGroups: %v", err)
	}
	env.AllSubscriptions = data.AllSubscriptions
	env.AllCatalogSources = data.AllCatalogSources
	env.AllPackageManifests = data.AllPackageManifests
	env.AllOperators = createOperators(data.AllCsvs, data.AllSubscriptions, data.AllPackageManifests, data.AllInstallPlans, data.AllCatalogSources, false, true)
	env.ClusterOperators = data.ClusterOperators
	env.AllCsvs = data.AllCsvs
	env.AllOperatorsSummary = getSummaryAllOperators(env.AllOperators)
	env.AllCrds = data.AllCrds
	env.Namespaces = data.Namespaces
	env.Nodes = createNodes(data.Nodes.Items)
	env.IstioServiceMeshFound = data.IstioServiceMeshFound
	env.ValidProtocolNames = append(env.ValidProtocolNames, data.ValidProtocolNames...)
	for i := range data.AbnormalEvents {
		aEvent := NewEvent(&data.AbnormalEvents[i])
		env.AbnormalEvents = append(env.AbnormalEvents, &aEvent)
	}

	// Service accounts
	env.ServiceAccounts = data.ServiceAccounts
	env.AllServiceAccounts = data.AllServiceAccounts
	env.AllServiceAccountsMap = make(map[string]*corev1.ServiceAccount)
	for i := 0; i < len(data.AllServiceAccounts); i++ {
		mapIndex := data.AllServiceAccounts[i].Namespace + data.AllServiceAccounts[i].Name
		env.AllServiceAccountsMap[mapIndex] = data.AllServiceAccounts[i]
	}
	// Pods
	pods := data.Pods
	for i := 0; i < len(pods); i++ {
		aNewPod := NewPod(&pods[i])
		aNewPod.AllServiceAccountsMap = &env.AllServiceAccountsMap
		env.Pods = append(env.Pods, &aNewPod)
	}
	pods = data.AllPods
	for i := 0; i < len(pods); i++ {
		aNewPod := NewPod(&pods[i])
		aNewPod.AllServiceAccountsMap = &env.AllServiceAccountsMap
		env.AllPods = append(env.AllPods, &aNewPod)
	}
	env.ProbePods = make(map[string]*corev1.Pod)
	for i := 0; i < len(data.ProbePods); i++ {
		nodeName := data.ProbePods[i].Spec.NodeName
		env.ProbePods[nodeName] = &data.ProbePods[i]
	}

	env.PodStates = data.PodStates

	csvPods := []*Pod{}
	env.CSVToPodListMap = make(map[string][]*Pod)
	for csv, podList := range data.CSVToPodListMap {
		var pods []*Pod
		for i := 0; i < len(podList); i++ {
			aNewPod := NewPod(podList[i])
			aNewPod.AllServiceAccountsMap = &env.AllServiceAccountsMap
			aNewPod.IsOperator = true
			pods = append(pods, &aNewPod)
			log.Info("CSV: %v, Operator Pod: %v/%v", csv, podList[i].Namespace, podList[i].Name)
		}
		env.CSVToPodListMap[csv.String()] = pods
		csvPods = append(csvPods, pods...)
	}

	// Add operator pods to list of normal pods to test.
	addOperatorPodsToTestPods(csvPods, &env)

	// Best effort mode autodiscovery for operand pods.
	operandPods := []*Pod{}
	for _, pod := range data.OperandPods {
		aNewPod := NewPod(pod)
		aNewPod.AllServiceAccountsMap = &env.AllServiceAccountsMap
		aNewPod.IsOperand = true
		operandPods = append(operandPods, &aNewPod)
	}

	addOperandPodsToTestPods(operandPods, &env)
	// Add operator pods' containers to the list.
	for _, pod := range env.Pods {
		// Note: 'getPodContainers' is returning a filtered list of Container objects.
		env.Containers = append(env.Containers, getPodContainers(pod.Pod, true)...)
	}

	log.Info("Found pods in %d csvs", len(env.CSVToPodListMap))

	env.OCPStatus = data.OCPStatus
	env.K8sVersion = data.K8sVersion
	env.ResourceQuotas = data.ResourceQuotaItems
	env.PodDisruptionBudgets = data.PodDisruptionBudgets
	env.PersistentVolumes = data.PersistentVolumes
	env.PersistentVolumeClaims = data.PersistentVolumeClaims
	env.ClusterRoleBindings = data.ClusterRoleBindings
	env.RoleBindings = data.RoleBindings
	env.Roles = data.Roles
	env.Services = data.Services
	env.AllServices = data.AllServices
	env.NetworkPolicies = data.NetworkPolicies
	for _, nsHelmChartReleases := range data.HelmChartReleases {
		for _, helmChartRelease := range nsHelmChartReleases {
			if !isSkipHelmChart(helmChartRelease.Name, config.SkipHelmChartList) {
				env.HelmChartReleases = append(env.HelmChartReleases, helmChartRelease)
			}
		}
	}
	for i := range data.Deployments {
		aNewDeployment := &Deployment{
			&data.Deployments[i],
		}
		env.Deployments = append(env.Deployments, aNewDeployment)
	}
	for i := range data.StatefulSet {
		aNewStatefulSet := &StatefulSet{
			&data.StatefulSet[i],
		}
		env.StatefulSets = append(env.StatefulSets, aNewStatefulSet)
	}

	env.ScaleCrUnderTest = updateCrUnderTest(data.ScaleCrUnderTest)
	env.HorizontalScaler = data.Hpas
	env.StorageClassList = data.StorageClasses

	env.ExecutedBy = data.ExecutedBy
	env.PartnerName = data.PartnerName
	env.CollectorAppPassword = data.CollectorAppPassword
	env.CollectorAppEndpoint = data.CollectorAppEndpoint
	env.ConnectAPIKey = data.ConnectAPIKey
	env.ConnectProjectID = data.ConnectProjectID
	env.ConnectAPIProxyURL = data.ConnectAPIProxyURL
	env.ConnectAPIProxyPort = data.ConnectAPIProxyPort
	env.ConnectAPIBaseURL = data.ConnectAPIBaseURL

	operators := createOperators(data.Csvs, data.AllSubscriptions, data.AllPackageManifests,
		data.AllInstallPlans, data.AllCatalogSources, false, true)
	env.Operators = operators
	log.Info("Operators found: %d", len(env.Operators))
	// SR-IOV
	env.SriovNetworks = data.SriovNetworks
	env.SriovNetworkNodePolicies = data.SriovNetworkNodePolicies
	env.AllSriovNetworks = data.AllSriovNetworks
	env.AllSriovNetworkNodePolicies = data.AllSriovNetworkNodePolicies
	env.NetworkAttachmentDefinitions = data.NetworkAttachmentDefinitions
	for _, pod := range env.Pods {
		isCreatedByDeploymentConfig, err := pod.CreatedByDeploymentConfig()
		if err != nil {
			log.Warn("Pod %q failed to get parent resource: %v", pod, err)
			continue
		}

		if isCreatedByDeploymentConfig {
			log.Warn("Pod %q has been deployed using a DeploymentConfig, please use Deployment or StatefulSet instead.", pod.String())
		}
	}

	log.Info("Completed the test environment build process in %.2f seconds", time.Since(start).Seconds())
}

// updateCrUnderTest Transforms raw scale objects into internal representation
//
// The function receives a slice of autodiscover.ScaleObject items, converts
// each entry into the provider's ScaleObject type by copying its scaling
// information and resource schema, and accumulates them in a new slice. It
// returns this populated slice for use elsewhere in the test environment
// construction.
func updateCrUnderTest(scaleCrUnderTest []autodiscover.ScaleObject) []ScaleObject {
	var scaleCrUndeTestTemp []ScaleObject
	for i := range scaleCrUnderTest {
		aNewScaleCrUnderTest := ScaleObject{Scale: CrScale{scaleCrUnderTest[i].Scale},
			GroupResourceSchema: scaleCrUnderTest[i].GroupResourceSchema}
		scaleCrUndeTestTemp = append(scaleCrUndeTestTemp, aNewScaleCrUnderTest)
	}
	return scaleCrUndeTestTemp
}

// getPodContainers Collects relevant container information from a pod while optionally filtering ignored containers
//
// The function iterates over the pod’s declared containers, matching each
// with its status to extract runtime details and image identifiers. It logs
// warnings for containers that are not ready or not running, providing reasons
// and restart counts. If the caller enables ignore mode, containers whose names
// match predefined patterns are skipped; otherwise they are added to the
// returned slice.
func getPodContainers(aPod *corev1.Pod, useIgnoreList bool) (containerList []*Container) {
	for j := 0; j < len(aPod.Spec.Containers); j++ {
		cut := &(aPod.Spec.Containers[j])

		var cutStatus corev1.ContainerStatus
		// get Status for current container
		for index := range aPod.Status.ContainerStatuses {
			if aPod.Status.ContainerStatuses[index].Name == cut.Name {
				cutStatus = aPod.Status.ContainerStatuses[index]
				break
			}
		}
		aRuntime, uid := GetRuntimeUID(&cutStatus)
		container := Container{Podname: aPod.Name, Namespace: aPod.Namespace,
			NodeName: aPod.Spec.NodeName, Container: cut, Status: cutStatus, Runtime: aRuntime, UID: uid,
			ContainerImageIdentifier: buildContainerImageSource(aPod.Spec.Containers[j].Image, cutStatus.ImageID)}

		// Warn if readiness probe did not succeeded yet.
		if !cutStatus.Ready {
			log.Warn("Container %q is not ready yet.", &container)
		}

		// Warn if container state is not running.
		if state := &cutStatus.State; state.Running == nil {
			reason := ""
			switch {
			case state.Waiting != nil:
				reason = "waiting - " + state.Waiting.Reason
			case state.Terminated != nil:
				reason = "terminated - " + state.Terminated.Reason
			default:
				// When no state was explicitly set, it's assumed to be in "waiting state".
				reason = "waiting state reason unknown"
			}

			log.Warn("Container %q is not running (reason: %s, restarts %d): some test cases might fail.",
				&container, reason, cutStatus.RestartCount)
		}

		// Build slices of containers based on whether or not we are "ignoring" them or not.
		if useIgnoreList && container.HasIgnoredContainerName() {
			continue
		}
		containerList = append(containerList, &container)
	}
	return containerList
}

// isSkipHelmChart determines whether a Helm chart should be excluded from processing
//
// The function receives the name of a Helm release and a list of names to skip.
// It checks if the list is empty, returning false immediately. Otherwise it
// iterates through each entry; if a match is found it logs that the chart was
// skipped and returns true. If no match is found after the loop, it returns
// false.
func isSkipHelmChart(helmName string, skipHelmChartList []configuration.SkipHelmChartList) bool {
	if len(skipHelmChartList) == 0 {
		return false
	}
	for _, helm := range skipHelmChartList {
		if helmName == helm.Name {
			log.Info("Helm chart with name %s was skipped", helmName)
			return true
		}
	}
	return false
}

// GetTestEnvironment Retrieves the test environment configuration
//
// This function returns a TestEnvironment instance used throughout the suite.
// It lazily builds the environment on first call by invoking
// buildTestEnvironment and caches it for future invocations. Subsequent calls
// simply return the cached environment without re‑initialising resources.
func GetTestEnvironment() TestEnvironment {
	if !loaded {
		buildTestEnvironment()
		loaded = true
	}
	return env
}

// IsOCPCluster Determines if the current cluster is an OpenShift installation
//
// The function checks whether the test environment’s OpenshiftVersion field
// differs from a predefined constant that represents non‑OpenShift clusters.
// It returns true when the cluster is recognized as OpenShift, and false
// otherwise.
func IsOCPCluster() bool {
	return env.OpenshiftVersion != autodiscover.NonOpenshiftClusterVersion
}

// buildContainerImageSource Extracts registry, repository, tag, and digest information from image strings
//
// The function parses a container image URL to obtain the registry, repository,
// and optional tag using a regular expression. It then extracts the image
// digest from an image ID string with another regex. The parsed values are
// assembled into a ContainerImageIdentifier structure and returned for use
// elsewhere in the program.
func buildContainerImageSource(urlImage, urlImageID string) (source ContainerImageIdentifier) {
	const regexImageWithTag = `^([^/]*)/*([^@]*):(.*)`
	const regexImageDigest = `^([^/]*)/(.*)@(.*:.*)`

	// get image repository, Name and tag if present
	re := regexp.MustCompile(regexImageWithTag)
	match := re.FindStringSubmatch(urlImage)

	if match != nil {
		if match[2] != "" {
			source.Registry = match[1]
			source.Repository = match[2]
			source.Tag = match[3]
		} else {
			source.Repository = match[1]
			source.Tag = match[3]
		}
	}

	// get image Digest based on imageID only
	re = regexp.MustCompile(regexImageDigest)
	match = re.FindStringSubmatch(urlImageID)

	if match != nil {
		source.Digest = match[3]
	}

	log.Debug("Parsed image, repo: %s, name:%s, tag: %s, digest: %s",
		source.Registry,
		source.Repository,
		source.Tag,
		source.Digest)

	return source
}

// GetRuntimeUID Extracts runtime type and unique identifier from a container status
//
// The function splits the ContainerID string at "://" to separate the runtime
// prefix from the unique ID. If a split occurs, it assigns the first part as
// the runtime name and the last part as the UID. It returns these two values
// for use in higher‑level logic.
func GetRuntimeUID(cs *corev1.ContainerStatus) (runtime, uid string) {
	split := strings.Split(cs.ContainerID, "://")
	if len(split) > 0 {
		uid = split[len(split)-1]
		runtime = split[0]
	}
	return runtime, uid
}

// GetPodIPsPerNet Retrieves pod IP addresses from a CNI annotation
//
// This function takes the JSON string stored in the
// "k8s.v1.cni.cncf.io/networks-status" annotation and parses it into a slice of
// network interface structures. It then builds a map keyed by each
// non‑default network name, associating each key with its corresponding
// interface information that includes IP addresses. If the annotation is empty
// or missing, an empty map is returned without error; if parsing fails, an
// error is reported.
func GetPodIPsPerNet(annotation string) (ips map[string]CniNetworkInterface, err error) {
	// This is a map indexed with the network name (network attachment) and
	// listing all the IPs created in this subnet and belonging to the pod namespace
	// The list of ips pr net is parsed from the content of the "k8s.v1.cni.cncf.io/networks-status" annotation.
	ips = make(map[string]CniNetworkInterface)

	// Sanity check: if the annotation is missing or empty, return empty result without error
	if strings.TrimSpace(annotation) == "" {
		return ips, nil
	}

	var cniInfo []CniNetworkInterface
	err = json.Unmarshal([]byte(annotation), &cniInfo)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal network-status annotation, err: %v", err)
	}
	// If this is the default interface, skip it as it is tested separately
	// Otherwise add all non default interfaces
	for _, cniInterface := range cniInfo {
		if !cniInterface.Default {
			ips[cniInterface.Name] = cniInterface
		}
	}
	return ips, nil
}

// GetPciPerPod Retrieves PCI addresses associated with a pod's network interfaces
//
// The function accepts the CNI networks status annotation string, checks for
// emptiness, and parses it as JSON into a slice of network interface objects.
// It iterates over each interface, extracting any non-empty PCI address from
// the device information and appends it to the result slice. If parsing fails,
// an error is returned; otherwise the collected PCI addresses are returned.
func GetPciPerPod(annotation string) (pciAddr []string, err error) {
	// Sanity check: if the annotation is missing or empty, return empty result without error
	if strings.TrimSpace(annotation) == "" {
		return []string{}, nil
	}

	var cniInfo []CniNetworkInterface
	err = json.Unmarshal([]byte(annotation), &cniInfo)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal network-status annotation, err: %v", err)
	}
	for _, cniInterface := range cniInfo {
		if cniInterface.DeviceInfo.PCI.PciAddress != "" {
			pciAddr = append(pciAddr, cniInterface.DeviceInfo.PCI.PciAddress)
		}
	}
	return pciAddr, nil
}

// TestEnvironment.SetNeedsRefresh Marks the test environment as needing a reload
//
// When invoked, this method clears the internal flag that tracks whether the
// environment has been initialized or loaded. It ensures subsequent operations
// will reinitialize necessary resources before use. The function does not
// return any value and performs no additional side effects.
func (env *TestEnvironment) SetNeedsRefresh() {
	loaded = false
}

// TestEnvironment.IsIntrusive Indicates if the test environment is running in intrusive mode
//
// The method checks a configuration flag stored in the environment's parameters
// and returns true when intrusive testing is enabled, otherwise false. It
// performs no other side effects or computations.
func (env *TestEnvironment) IsIntrusive() bool {
	return env.params.Intrusive
}

// TestEnvironment.IsPreflightInsecureAllowed Indicates whether insecure Preflight connections are permitted
//
// This method returns the value of the AllowPreflightInsecure flag stored in
// the TestEnvironment parameters. It is used to decide if insecure network
// connections should be allowed when executing Preflight checks for containers
// or operators.
func (env *TestEnvironment) IsPreflightInsecureAllowed() bool {
	return env.params.AllowPreflightInsecure
}

// TestEnvironment.GetDockerConfigFile Retrieves the path to the Docker configuration file
//
// This method accesses the TestEnvironment's parameters to return the location
// of the Docker config used by Preflight checks. It returns a string
// representing the file path, which is then supplied to container and operator
// preflight options for authentication. The function performs no additional
// logic beyond fetching the stored value.
func (env *TestEnvironment) GetDockerConfigFile() string {
	return env.params.PfltDockerconfig
}

// TestEnvironment.GetOfflineDBPath Retrieves the configured file system path for an offline database
//
// This method accesses the TestEnvironment's internal parameters to obtain the
// location of the offline database. It returns a string representing that
// filesystem path, which can be used by other components to locate or access
// the database file. No arguments are required and the value is read directly
// from the environment configuration.
func (env *TestEnvironment) GetOfflineDBPath() string {
	return env.params.OfflineDB
}

// TestEnvironment.GetWorkerCount Returns the number of worker nodes in the environment
//
// This method iterates over all nodes stored in the TestEnvironment, checking
// each one to determine if it is marked as a worker node. It counts how many
// nodes satisfy this condition and returns that integer count. The result
// reflects the current composition of worker nodes within the test setup.
func (env *TestEnvironment) GetWorkerCount() int {
	workerCount := 0
	for _, e := range env.Nodes {
		if e.IsWorkerNode() {
			workerCount++
		}
	}
	return workerCount
}

// TestEnvironment.GetMasterCount Counts control plane nodes in the test environment
//
// This method iterates over all nodes stored in the TestEnvironment, checks
// each node to see if it is a control‑node by examining its labels, and
// tallies them. It returns the total number of master nodes as an integer.
func (env *TestEnvironment) GetMasterCount() int {
	masterCount := 0
	for _, e := range env.Nodes {
		if e.IsControlPlaneNode() {
			masterCount++
		}
	}
	return masterCount
}

// TestEnvironment.IsSNO Checks whether the environment contains a single node
//
// The method inspects the collection of nodes in the test environment and
// determines if exactly one node is present. It returns true when the count
// equals one, indicating a single-node setup; otherwise it returns false.
func (env *TestEnvironment) IsSNO() bool {
	return len(env.Nodes) == 1
}

// getMachineConfig Retrieves a machine configuration by name, using caching
//
// The function first checks an in-memory map for the requested configuration;
// if present it returns it immediately. Otherwise it queries the Kubernetes API
// for the MachineConfig resource, decodes its raw YAML into a Go struct, and
// stores the result for future calls. Errors from fetching or unmarshalling are
// propagated to the caller.
func getMachineConfig(mcName string, machineConfigs map[string]MachineConfig) (MachineConfig, error) {
	client := clientsholder.GetClientsHolder()

	// Check whether we had already downloaded and parsed that machineConfig resource.
	if mc, exists := machineConfigs[mcName]; exists {
		return mc, nil
	}

	nodeMc, err := client.MachineCfg.MachineconfigurationV1().MachineConfigs().Get(context.TODO(), mcName, metav1.GetOptions{})
	if err != nil {
		return MachineConfig{}, err
	}

	mc := MachineConfig{
		MachineConfig: nodeMc,
	}

	err = json.Unmarshal(nodeMc.Spec.Config.Raw, &mc.Config)
	if err != nil {
		return MachineConfig{}, fmt.Errorf("failed to unmarshal mc's Config field, err: %v", err)
	}

	return mc, nil
}

// createNodes Builds a mapping of node names to enriched node structures
//
// The function iterates over supplied node objects, skipping machine
// configuration retrieval for non‑OpenShift clusters and logging warnings in
// that case. For OpenShift nodes it extracts the current MachineConfig
// annotation, fetches or reuses the corresponding config, and attaches it to
// the resulting Node wrapper. The returned map keys each node name to its
// enriched data structure.
func createNodes(nodes []corev1.Node) map[string]Node {
	wrapperNodes := map[string]Node{}

	// machineConfigs is a helper map to avoid download & process the same mc twice.
	machineConfigs := map[string]MachineConfig{}
	for i := range nodes {
		node := &nodes[i]

		if !IsOCPCluster() {
			// Avoid getting Mc info for non ocp clusters.
			wrapperNodes[node.Name] = Node{Data: node}
			log.Warn("Non-OCP cluster detected. MachineConfig retrieval for node %q skipped.", node.Name)
			continue
		}

		// Get Node's machineConfig name
		mcName, exists := node.Annotations["machineconfiguration.openshift.io/currentConfig"]
		if !exists {
			log.Error("Failed to get machineConfig name for node %q", node.Name)
			continue
		}
		log.Info("Node %q - mc name %q", node.Name, mcName)
		mc, err := getMachineConfig(mcName, machineConfigs)
		if err != nil {
			log.Error("Failed to get machineConfig %q, err: %v", mcName, err)
			continue
		}

		wrapperNodes[node.Name] = Node{
			Data: node,
			Mc:   mc,
		}
	}

	return wrapperNodes
}

// TestEnvironment.GetBaremetalNodes Retrieves nodes that use a bare‑metal provider
//
// It iterates over the environment’s node list, selecting those whose
// ProviderID begins with "baremetalhost://". Matching nodes are collected into
// a slice which is returned. The function returns only the filtered set of
// bare‑metal nodes.
func (env *TestEnvironment) GetBaremetalNodes() []Node {
	var baremetalNodes []Node
	for _, node := range env.Nodes {
		if strings.HasPrefix(node.Data.Spec.ProviderID, "baremetalhost://") {
			baremetalNodes = append(baremetalNodes, node)
		}
	}
	return baremetalNodes
}

// GetPreflightResultsDB Transforms runtime preflight test outcomes into a structured result set
//
// The function receives a pointer to the runtime results of preflight checks.
// It iterates over each passed, failed, and errored check, extracting the name,
// description, remediation suggestion, and error message when applicable. For
// every check it constructs a PreflightTest entry and appends it to the
// corresponding slice in a PreflightResultsDB structure. Finally, it returns
// this populated database for use by the container or operator result handling.
func GetPreflightResultsDB(results *plibRuntime.Results) PreflightResultsDB {
	resultsDB := PreflightResultsDB{}
	for _, res := range results.Passed {
		test := PreflightTest{Name: res.Name(), Description: res.Metadata().Description, Remediation: res.Help().Suggestion}
		resultsDB.Passed = append(resultsDB.Passed, test)
	}
	for _, res := range results.Failed {
		test := PreflightTest{Name: res.Name(), Description: res.Metadata().Description, Remediation: res.Help().Suggestion}
		resultsDB.Failed = append(resultsDB.Failed, test)
	}
	for _, res := range results.Errors {
		test := PreflightTest{Name: res.Name(), Description: res.Metadata().Description, Remediation: res.Help().Suggestion, Error: res.Error()}
		resultsDB.Errors = append(resultsDB.Errors, test)
	}

	return resultsDB
}
