// Copyright (C) 2020-2023 Red Hat, Inc.
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
	"os"
	"time"

	"errors"
	"fmt"
	"strings"

	"encoding/json"

	mcv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/autodiscover"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	k8sPriviledgedDs "github.com/test-network-function/privileged-daemonset"
	"helm.sh/helm/v3/pkg/release"
	scalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AffinityRequiredKey              = "AffinityRequired"
	containerName                    = "container-00"
	DaemonSetNamespace               = "default"
	DaemonSetName                    = "tnf-debug"
	debugPodsTimeout                 = 5 * time.Minute
	CniNetworksStatusKey             = "k8s.v1.cni.cncf.io/networks-status"
	skipConnectivityTestsLabel       = "test-network-function.com/skip_connectivity_tests"
	skipMultusConnectivityTestsLabel = "test-network-function.com/skip_multus_connectivity_tests"
	rhcosName                        = "Red Hat Enterprise Linux CoreOS"
	rhelName                         = "Red Hat Enterprise Linux"
	tnfPartnerRepoDef                = "quay.io/testnetworkfunction"
	supportImageDef                  = "debug-partner:latest"
)

// Node's roles labels. Node is role R if it has **any** of the labels of each list.
// Master's role label "master" is deprecated since k8s 1.20.
var (
	WorkerLabels      = []string{"node-role.kubernetes.io/worker"}
	MasterLabels      = []string{"node-role.kubernetes.io/master", "node-role.kubernetes.io/control-plane"}
	rhcosRelativePath = "%s/platform/operatingsystem/files/rhcos_version_map"
)

type TestEnvironment struct { // rename this with testTarget
	Namespaces     []string `json:"testNamespaces"`
	AbnormalEvents []*Event

	// Pod Groupings
	Pods      []*Pod                 `json:"testPods"`
	DebugPods map[string]*corev1.Pod // map from nodename to debugPod
	AllPods   []*Pod                 `json:"AllPods"`

	// Deployment Groupings
	Deployments []*Deployment `json:"testDeployments"`

	// StatefulSet Groupings
	StatefulSets []*StatefulSet `json:"testStatefulSets"`

	// Note: Containers is a filtered list of objects based on a block list of disallowed container names.
	Containers             []*Container `json:"testContainers"`
	Operators              []*Operator  `json:"testOperators"`
	AllOperators           []*Operator  `json:"AllOperators"`
	AllOperatorsSummary    []string     `json:"AllOperatorsSummary"`
	PersistentVolumes      []corev1.PersistentVolume
	PersistentVolumeClaims []corev1.PersistentVolumeClaim

	Config    configuration.TestConfiguration
	variables configuration.TestParameters
	Crds      []*apiextv1.CustomResourceDefinition `json:"testCrds"`

	HorizontalScaler       map[string]*scalingv1.HorizontalPodAutoscaler `json:"testHorizontalScaler"`
	Services               []*corev1.Service                             `json:"testServices"`
	Nodes                  map[string]Node                               `json:"-"`
	K8sVersion             string                                        `json:"-"`
	OpenshiftVersion       string                                        `json:"-"`
	OCPStatus              string                                        `json:"-"`
	HelmChartReleases      []*release.Release                            `json:"testHelmChartReleases"`
	ResourceQuotas         []corev1.ResourceQuota
	PodDisruptionBudgets   []policyv1.PodDisruptionBudget
	NetworkPolicies        []networkingv1.NetworkPolicy
	AllInstallPlans        []*olmv1Alpha.InstallPlan   `json:"-"`
	AllCatalogSources      []*olmv1Alpha.CatalogSource `json:"-"`
	IstioServiceMesh       bool
	ValidProtocolNames     []string
	DaemonsetFailedToSpawn bool
	StorageClassList       []storagev1.StorageClass
}

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

type cniNetworkInterface struct {
	Name       string                 `json:"name"`
	Interface  string                 `json:"interface"`
	IPs        []string               `json:"ips"`
	Default    bool                   `json:"default"`
	DNS        map[string]interface{} `json:"dns"`
	DeviceInfo deviceInfo             `json:"device-info"`
}

type deviceInfo struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	PCI     pci    `json:"pci"`
}

type pci struct {
	PciAddress string `json:"pci-address"`
}

var (
	env    = TestEnvironment{}
	loaded = false
)

// Build image with version based on environment variables if provided, else use a default value
func buildImageWithVersion() string {
	tnfPartnerRepo := os.Getenv("TNF_PARTNER_REPO")
	if tnfPartnerRepo == "" {
		tnfPartnerRepo = tnfPartnerRepoDef
	}
	supportImage := os.Getenv("SUPPORT_IMAGE")
	if supportImage == "" {
		supportImage = supportImageDef
	}

	return tnfPartnerRepo + "/" + supportImage
}

func deployDaemonSet() error {
	k8sPriviledgedDs.SetDaemonSetClient(clientsholder.GetClientsHolder().K8sClient)
	dsImage := buildImageWithVersion()

	if k8sPriviledgedDs.IsDaemonSetReady(DaemonSetName, DaemonSetNamespace, dsImage) {
		return nil
	}

	matchLabels := make(map[string]string)
	matchLabels["name"] = DaemonSetName
	matchLabels["test-network-function.com/app"] = DaemonSetName
	_, err := k8sPriviledgedDs.CreateDaemonSet(DaemonSetName, DaemonSetNamespace, containerName, dsImage, matchLabels, debugPodsTimeout)
	if err != nil {
		return fmt.Errorf("could not deploy tnf daemonset, err=%v", err)
	}
	err = k8sPriviledgedDs.WaitDaemonsetReady(DaemonSetNamespace, DaemonSetName, debugPodsTimeout)
	if err != nil {
		return fmt.Errorf("timed out waiting for tnf daemonset, err=%v", err)
	}

	return nil
}

func buildTestEnvironment() { //nolint:funlen
	start := time.Now()
	env = TestEnvironment{}

	// Wait for the debug pods to be ready before the autodiscovery starts.
	if err := deployDaemonSet(); err != nil {
		logrus.Errorf("The TNF daemonset could not be deployed, err=%v", err)
		// Because of this failure, we are only able to run a certain amount of tests that do not rely
		// on the existence of the daemonset debug pods.
		env.DaemonsetFailedToSpawn = true
	}

	data := autodiscover.DoAutoDiscover()
	// OpenshiftVersion needs to be set asap, as other helper functions will use it here.
	env.OpenshiftVersion = data.OpenshiftVersion
	env.Config = data.TestData
	env.Crds = data.Crds
	env.AllInstallPlans = data.AllInstallPlans
	env.AllCatalogSources = data.AllCatalogSources
	env.AllOperators = createOperators(data.AllCsvs, data.AllSubscriptions, data.AllInstallPlans, data.AllCatalogSources, true, false, false)
	env.AllOperatorsSummary = getSummaryAllOperators(env.AllOperators)
	env.Namespaces = data.Namespaces
	env.variables = data.Env
	env.Nodes = createNodes(data.Nodes.Items)
	env.IstioServiceMesh = data.Istio
	env.ValidProtocolNames = append(env.ValidProtocolNames, data.ValidProtocolNames...)
	for i := range data.AbnormalEvents {
		aEvent := NewEvent(&data.AbnormalEvents[i])
		env.AbnormalEvents = append(env.AbnormalEvents, &aEvent)
	}
	pods := data.Pods
	for i := 0; i < len(pods); i++ {
		aNewPod := NewPod(&pods[i])
		env.Pods = append(env.Pods, &aNewPod)
		// Note: 'getPodContainers' is returning a filtered list of Container objects.
		env.Containers = append(env.Containers, getPodContainers(&pods[i], true)...)
	}
	pods = data.AllPods
	for i := 0; i < len(pods); i++ {
		aNewPod := NewPod(&pods[i])
		env.AllPods = append(env.AllPods, &aNewPod)
	}
	env.DebugPods = make(map[string]*corev1.Pod)
	for i := 0; i < len(data.DebugPods); i++ {
		nodeName := data.DebugPods[i].Spec.NodeName
		env.DebugPods[nodeName] = &data.DebugPods[i]
	}

	env.OCPStatus = data.OCPStatus
	env.K8sVersion = data.K8sVersion
	env.ResourceQuotas = data.ResourceQuotaItems
	env.PodDisruptionBudgets = data.PodDisruptionBudgets
	env.PersistentVolumes = data.PersistentVolumes
	env.PersistentVolumeClaims = data.PersistentVolumeClaims
	env.Services = data.Services
	env.NetworkPolicies = data.NetworkPolicies
	for _, nsHelmChartReleases := range data.HelmChartReleases {
		for _, helmChartRelease := range nsHelmChartReleases {
			if !isSkipHelmChart(helmChartRelease.Name, data.TestData.SkipHelmChartList) {
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
	env.HorizontalScaler = data.Hpas
	env.StorageClassList = data.StorageClasses

	operators := createOperators(data.Csvs, data.Subscriptions, data.AllInstallPlans, data.AllCatalogSources, false, false, true)
	env.Operators = operators
	logrus.Infof("Operators found: %d", len(env.Operators))
	for _, pod := range env.Pods {
		isCreatedByDeploymentConfig, err := pod.CreatedByDeploymentConfig()
		if err != nil {
			logrus.Warnf("Pod %s: failed to get parent resource: %v", pod.String(), err)
			continue
		}

		if isCreatedByDeploymentConfig {
			logrus.Warnf("Pod %s has been deployed using a DeploymentConfig, please use Deployment or StatefulSet instead.", pod.String())
		}
	}
	logrus.Infof("Completed the test environment build process in %.2f seconds", time.Since(start).Seconds())
}

func getPodContainers(aPod *corev1.Pod, useIgnoreList bool) (containerList []*Container) {
	for j := 0; j < len(aPod.Spec.Containers); j++ {
		cut := &(aPod.Spec.Containers[j])
		var status corev1.ContainerStatus
		if len(aPod.Status.ContainerStatuses) > 0 {
			status = aPod.Status.ContainerStatuses[j]
		} else {
			logrus.Errorf("%s is not ready, skipping status collection", aPod.String())
		}
		aRuntime, uid := GetRuntimeUID(&status)
		container := Container{Podname: aPod.Name, Namespace: aPod.Namespace,
			NodeName: aPod.Spec.NodeName, Container: cut, Status: status, Runtime: aRuntime, UID: uid,
			ContainerImageIdentifier: buildContainerImageSource(aPod.Spec.Containers[j].Image)}

		// Warn if readiness probe did not succeeded yet.
		if !status.Ready {
			logrus.Warnf("%s is not ready yet.", &container)
		}

		// Warn if container state is not running.
		if state := &status.State; state.Running == nil {
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

			logrus.Warnf("%s is not running (reason: %s, restarts %d): some test cases might fail.",
				&container, reason, status.RestartCount)
		}

		// Build slices of containers based on whether or not we are "ignoring" them or not.
		if useIgnoreList && container.HasIgnoredContainerName() {
			continue
		} else {
			containerList = append(containerList, &container)
		}
	}
	return containerList
}

func isSkipHelmChart(helmName string, skipHelmChartList []configuration.SkipHelmChartList) bool {
	if len(skipHelmChartList) == 0 {
		return false
	}
	for _, helm := range skipHelmChartList {
		if helmName == helm.Name {
			logrus.Infof("Helm chart with name %s was skipped", helmName)
			return true
		}
	}
	return false
}

func GetTestEnvironment() TestEnvironment {
	if !loaded {
		buildTestEnvironment()
		loaded = true
	}
	return env
}

func IsOCPCluster() bool {
	return env.OpenshiftVersion != autodiscover.NonOpenshiftClusterVersion
}

func buildContainerImageSource(url string) configuration.ContainerImageIdentifier {
	source := configuration.ContainerImageIdentifier{}
	urlSegments := strings.Split(url, "/")
	n := len(urlSegments)
	if n > 1 {
		source.Repository = urlSegments[n-2]
	}
	colonIndex := strings.Index(urlSegments[n-1], ":")
	atIndex := strings.Index(urlSegments[n-1], "@")
	if atIndex == -1 {
		if colonIndex == -1 {
			source.Name = urlSegments[n-1]
		} else {
			source.Name = urlSegments[n-1][:colonIndex]
			source.Tag = urlSegments[n-1][colonIndex+1:]
		}
	} else {
		source.Name = urlSegments[n-1][:atIndex]
		source.Digest = urlSegments[n-1][atIndex+1:]
	}
	return source
}

func GetRuntimeUID(cs *corev1.ContainerStatus) (runtime, uid string) {
	split := strings.Split(cs.ContainerID, "://")
	if len(split) > 0 {
		uid = split[len(split)-1]
		runtime = split[0]
	}
	return runtime, uid
}

// GetPodIPsPerNet gets the IPs of a pod.
// CNI annotation "k8s.v1.cni.cncf.io/networks-status".
// Returns (ips, error).
func GetPodIPsPerNet(annotation string) (ips map[string][]string, err error) {
	// This is a map indexed with the network name (network attachment) and
	// listing all the IPs created in this subnet and belonging to the pod namespace
	// The list of ips pr net is parsed from the content of the "k8s.v1.cni.cncf.io/networks-status" annotation.
	ips = make(map[string][]string)

	var cniInfo []cniNetworkInterface
	err = json.Unmarshal([]byte(annotation), &cniInfo)
	if err != nil {
		return nil, errors.New("could not unmarshal network-status annotation")
	}
	// If this is the default interface, skip it as it is tested separately
	// Otherwise add all non default interfaces
	for _, cniInterface := range cniInfo {
		if !cniInterface.Default {
			ips[cniInterface.Name] = cniInterface.IPs
		}
	}
	return ips, nil
}

func GetPciPerPod(annotation string) (pciAddr []string, err error) {
	var cniInfo []cniNetworkInterface
	err = json.Unmarshal([]byte(annotation), &cniInfo)
	if err != nil {
		return nil, errors.New("could not unmarshal network-status annotation")
	}
	for _, cniInterface := range cniInfo {
		if cniInterface.DeviceInfo.PCI.PciAddress != "" {
			pciAddr = append(pciAddr, cniInterface.DeviceInfo.PCI.PciAddress)
		}
	}
	return pciAddr, nil
}

func (env *TestEnvironment) SetNeedsRefresh() {
	loaded = false
}

func (env *TestEnvironment) IsIntrusive() bool {
	return !env.variables.NonIntrusiveOnly
}

func (env *TestEnvironment) GetOfflineDBPath() string {
	return env.variables.OfflineDB
}

func (env *TestEnvironment) GetWorkerCount() int {
	workerCount := 0
	for _, e := range env.Nodes {
		if e.IsWorkerNode() {
			workerCount++
		}
	}
	return workerCount
}

func (env *TestEnvironment) GetMasterCount() int {
	masterCount := 0
	for _, e := range env.Nodes {
		if e.IsMasterNode() {
			masterCount++
		}
	}
	return masterCount
}

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

func createNodes(nodes []corev1.Node) map[string]Node {
	wrapperNodes := map[string]Node{}

	// machineConfigs is a helper map to avoid download & process the same mc twice.
	machineConfigs := map[string]MachineConfig{}
	for i := range nodes {
		node := &nodes[i]

		if !IsOCPCluster() {
			// Avoid getting Mc info for non ocp clusters.
			wrapperNodes[node.Name] = Node{Data: node}
			logrus.Warnf("Non-OCP cluster detected. MachineConfig retrieval for node %s skipped.", node.Name)
			continue
		}

		// Get Node's machineConfig name
		mcName, exists := node.Annotations["machineconfiguration.openshift.io/currentConfig"]
		if !exists {
			logrus.Errorf("Failed to get machineConfig name for node %s", node.Name)
			continue
		}
		logrus.Infof("Node %s - mc name: %s", node.Name, mcName)
		mc, err := getMachineConfig(mcName, machineConfigs)
		if err != nil {
			logrus.Errorf("Failed to get machineConfig %s, err: %v", mcName, err)
			continue
		}

		wrapperNodes[node.Name] = Node{
			Data: node,
			Mc:   mc,
		}
	}

	return wrapperNodes
}
