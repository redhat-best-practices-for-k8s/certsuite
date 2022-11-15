package securitycontextcontainer

// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/
// api we used as a reference

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	corev1 "k8s.io/api/core/v1"
)

type OkNok int

const (
	OKNOK = iota
	NOK   // 0
	OK    // 1
)

const (
	OKString  = "true"
	NOKString = "false"
)

// print the strings
func (okNok OkNok) String() string {
	switch okNok {
	case OK:
		return OKString
	case NOK:
		return NOKString
	}
	return "false"
}

type ContainerSCC struct {
	HostDirVolumePluginPresent      OkNok // 0 or 1 - 0 is false 1 - true
	HostIPC                         OkNok
	HostNetwork                     OkNok
	HostPID                         OkNok
	HostPorts                       OkNok
	PrivilegeEscalation             OkNok // this can be true or false
	PrivilegedContainer             OkNok
	RunAsUserPresent                OkNok // thes filed that checking if the value is present
	ReadOnlyRootFilesystem          OkNok
	RunAsNonRoot                    OkNok
	FsGroupPresent                  OkNok
	SeLinuxContextPresent           OkNok
	CapabilitiesCategory            CategoryID
	RequiredDropCapabilitiesPresent OkNok
	AllVolumeAllowed                OkNok
}

type CategoryID int

const (
	Undefined CategoryID = iota
	CategoryID1
	CategoryID1NoUID0
	CategoryID2
	CategoryID3
	CategoryID4
)

type PodListcategory struct {
	Containername string
	Podname       string
	NameSpace     string
	Category      CategoryID
}

var (
	requiredDropCapabilities = []string{"MKNOD", "SETUID", "SETGID", "KILL"}
	dropAll                  = []string{"ALL"}
	category2AddCapabilities = []string{"NET_ADMIN, NET_RAW"}
	category3AddCapabilities = []string{"NET_ADMIN, NET_RAW, IPC_LOCK"}
	Category1                = ContainerSCC{
		NOK,         // HostDirVolumePluginPresent
		NOK,         // HostIPC
		NOK,         // HostNetwork
		NOK,         // HostPID
		NOK,         // HostPorts
		OK,          // PrivilegeEscalation
		NOK,         // PrivilegedContainer
		OK,          // RunAsUserPresent
		NOK,         // ReadOnlyRootFilesystem
		NOK,         // RunAsNonRoot
		OK,          // FsGroupPresent
		OK,          // SeLinuxContextPresent
		CategoryID1, // Capabilities
		OK,          // RequiredDropCapabilitiesPresent
		OK}          // AllVolumeAllowed

	Category1NoUID0 = ContainerSCC{
		NOK,         // HostDirVolumePluginPresent
		NOK,         // HostIPC
		NOK,         // HostNetwork
		NOK,         // HostPID
		NOK,         // HostPorts
		OK,          // PrivilegeEscalation
		NOK,         // PrivilegedContainer
		OK,          // RunAsUserPresent
		NOK,         // ReadOnlyRootFilesystem
		OK,          // RunAsNonRoot
		OK,          // FsGroupPresent
		OK,          // SeLinuxContextPresent
		CategoryID1, // Capabilities
		OK,          // RequiredDropCapabilitiesPresent
		OK}          // AllVolumeAllowed

	Category2 = ContainerSCC{
		NOK,         // HostDirVolumePluginPresent
		NOK,         // HostIPC
		NOK,         // HostNetwork
		NOK,         // HostPID
		NOK,         // HostPorts
		OK,          // PrivilegeEscalation
		NOK,         // PrivilegedContainer
		OK,          // RunAsUserPresent
		NOK,         // ReadOnlyRootFilesystem
		OK,          // RunAsNonRoot
		OK,          // FsGroupPresent
		OK,          // SeLinuxContextPresent
		CategoryID2, // Capabilities
		OK,          // RequiredDropCapabilitiesPresent
		OK}          // AllVolumeAllowed

	Category3 = ContainerSCC{
		NOK,         // HostDirVolumePluginPresent
		NOK,         // HostIPC
		NOK,         // HostNetwork
		NOK,         // HostPID
		NOK,         // HostPorts
		OK,          // PrivilegeEscalation
		NOK,         // PrivilegedContainer
		OK,          // RunAsUserPresent
		NOK,         // ReadOnlyRootFilesystem
		OK,          // RunAsNonRoot
		OK,          // FsGroupPresent
		OK,          // SeLinuxContextPresent
		CategoryID3, // Capabilities
		OK,          // RequiredDropCapabilitiesPresent
		OK}          // AllVolumeAllowed
)

// print the strings
func (category PodListcategory) String() string {
	returnString := fmt.Sprintf("Containername: %s Podname: %s NameSpace: %s Category: %s \n ",
		category.Containername, category.Podname, category.NameSpace, category.Category)
	return returnString
}

const (
	CategoryID1String       = "CategoryID1(limited access granted automatically)"
	CategoryID1NoUID0String = "CategoryID1NoUID0(automatically granted, basic rights with mesh networks)"
	CategoryID2String       = "CategoryID2(advanced networking (vlan tag, dscp, priority))"
	CategoryID3String       = "CategoryID3(SRIOV and DPDK)"
	CategoryID4String       = "CategoryID4(anything not matching lower category)"
)

// print the strings
func (category CategoryID) String() string {
	switch category {
	case CategoryID1:
		return CategoryID1String
	case CategoryID1NoUID0:
		return CategoryID1NoUID0String
	case CategoryID2:
		return CategoryID2String
	case CategoryID3:
		return CategoryID3String
	case CategoryID4:
		return CategoryID4String
	case Undefined:
		return CategoryID4String
	}
	return CategoryID4String
}

//nolint:gocritic,gocyclo
func GetContainerSCC(cut *provider.Container, containerSCC ContainerSCC) ContainerSCC {
	containerSCC.HostPorts = NOK
	for _, aPort := range cut.Ports {
		if aPort.HostPort != 0 {
			containerSCC.HostPorts = OK
			break
		}
	}
	updateCapabilitiesFromContainer(cut, &containerSCC)
	containerSCC.PrivilegeEscalation = NOK
	if cut.SecurityContext != nil && cut.SecurityContext.AllowPrivilegeEscalation != nil {
		containerSCC.PrivilegeEscalation = OK
	}
	containerSCC.PrivilegedContainer = NOK
	if cut.SecurityContext != nil && cut.SecurityContext.Privileged != nil && *(cut.SecurityContext.Privileged) {
		containerSCC.PrivilegedContainer = OK
	}
	if cut.SecurityContext != nil && cut.SecurityContext.RunAsUser != nil {
		containerSCC.RunAsUserPresent = OK
	}
	containerSCC.ReadOnlyRootFilesystem = NOK
	if cut.SecurityContext != nil && cut.SecurityContext.ReadOnlyRootFilesystem != nil && *cut.SecurityContext.ReadOnlyRootFilesystem {
		containerSCC.ReadOnlyRootFilesystem = OK
	}
	containerSCC.RunAsNonRoot = NOK
	if cut.SecurityContext != nil && cut.SecurityContext.RunAsNonRoot != nil && *cut.SecurityContext.RunAsNonRoot {
		containerSCC.RunAsNonRoot = OK
	}
	if cut.SecurityContext != nil && cut.SecurityContext.SELinuxOptions != nil {
		containerSCC.SeLinuxContextPresent = OK
	}
	return containerSCC
}

func updateCapabilitiesFromContainer(cut *provider.Container, containerSCC *ContainerSCC) {
	containerSCC.RequiredDropCapabilitiesPresent = NOK
	if cut.SecurityContext != nil && cut.SecurityContext.Capabilities != nil {
		var sliceDropCapabilities []string
		for _, ncc := range cut.SecurityContext.Capabilities.Drop {
			sliceDropCapabilities = append(sliceDropCapabilities, string(ncc))
		}
		sort.Strings(sliceDropCapabilities)

		sort.Strings(requiredDropCapabilities)
		if subslice(requiredDropCapabilities, sliceDropCapabilities) || reflect.DeepEqual(sliceDropCapabilities, dropAll) {
			containerSCC.RequiredDropCapabilitiesPresent = OK
		}
		if len(cut.SecurityContext.Capabilities.Add) == 0 { // check if the len=0 this mean that is cat1
			containerSCC.CapabilitiesCategory = CategoryID1
		} else if checkContainCategory(cut.SecurityContext.Capabilities.Add, category2AddCapabilities) {
			containerSCC.CapabilitiesCategory = CategoryID2
		} else {
			if checkContainCategory(cut.SecurityContext.Capabilities.Add, category3AddCapabilities) {
				containerSCC.CapabilitiesCategory = CategoryID3
			} else {
				containerSCC.CapabilitiesCategory = CategoryID4
			}
		}
	} else {
		containerSCC.CapabilitiesCategory = CategoryID1
	}
}
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func subslice(s1, s2 []string) bool {
	if len(s1) > len(s2) {
		return false
	}
	for _, e := range s1 {
		if !contains(s2, e) {
			return false
		}
	}
	return true
}
func AllVolumeAllowed(volumes []corev1.Volume) (r1, r2 OkNok) {
	countVolume := 0
	var value OkNok
	value = NOK
	for j := 0; j < len(volumes); j++ {
		if volumes[j].HostPath != nil {
			value = OK
		}
		if volumes[j].ConfigMap != nil {
			countVolume++
		}
		if volumes[j].DownwardAPI != nil {
			countVolume++
		}
		if volumes[j].EmptyDir != nil {
			countVolume++
		}
		if volumes[j].PersistentVolumeClaim != nil {
			countVolume++
		}
		if volumes[j].Projected != nil {
			countVolume++
		}
		if volumes[j].Secret != nil {
			countVolume++
		}
	}
	if countVolume == len(volumes) {
		return OK, value
	}
	return NOK, value
}

//nolint:gocritic
func checkContainerCategory(containers []corev1.Container, containerSCC ContainerSCC, podName, nameSpace string) []PodListcategory {
	var ContainerList []PodListcategory
	var categoryinfo PodListcategory
	for j := 0; j < len(containers); j++ {
		cut := &provider.Container{Podname: podName, Namespace: nameSpace, Container: &containers[j]}
		percontainerSCC := GetContainerSCC(cut, containerSCC)
		tnf.ClaimFilePrintf("containerSCC %s is %+v", cut, percontainerSCC)
		// after building the containerSCC need to check to which category it is
		categoryinfo = PodListcategory{
			Containername: cut.Name,
			Podname:       podName,
			NameSpace:     nameSpace,
		}
		if compareCategory(&Category1, &percontainerSCC, CategoryID1) {
			tnf.ClaimFilePrintf("Testing if pod belongs to category1 ")
			categoryinfo.Category = CategoryID1
		} else if compareCategory(&Category1NoUID0, &percontainerSCC, CategoryID1NoUID0) {
			tnf.ClaimFilePrintf("Testing if pod belongs to category1NoUID0 ")
			categoryinfo.Category = CategoryID1NoUID0
		} else if compareCategory(&Category2, &percontainerSCC, CategoryID2) {
			categoryinfo.Category = CategoryID2
		} else if compareCategory(&Category3, &percontainerSCC, CategoryID3) {
			categoryinfo.Category = CategoryID3
		} else {
			categoryinfo.Category = CategoryID4
		}
		// after building the containerSCC need to check to which category it is
		ContainerList = append(ContainerList, categoryinfo)
	}
	return ContainerList
}

func checkContainCategory(addCapability []corev1.Capability, referenceCategoryAddCapabilities []string) bool {
	for _, ncc := range addCapability {
		if !stringhelper.StringInSlice(referenceCategoryAddCapabilities, string(ncc), true) {
			return false
		}
	}
	return true
}

func CheckPod(pod *provider.Pod) []PodListcategory {
	var containerSCC ContainerSCC
	containerSCC.HostIPC = NOK
	if pod.Spec.HostIPC {
		containerSCC.HostIPC = OK
	}
	containerSCC.HostNetwork = NOK
	if pod.Spec.HostNetwork {
		containerSCC.HostNetwork = OK
	}
	containerSCC.HostPID = NOK
	if pod.Spec.HostPID {
		containerSCC.HostPID = OK
	}
	containerSCC.SeLinuxContextPresent = NOK
	if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.SELinuxOptions != nil {
		containerSCC.SeLinuxContextPresent = OK
	}
	containerSCC.AllVolumeAllowed, containerSCC.HostDirVolumePluginPresent = AllVolumeAllowed(pod.Spec.Volumes)
	if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.RunAsUser != nil {
		containerSCC.RunAsUserPresent = OK
	} else {
		containerSCC.RunAsUserPresent = NOK
	}
	if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.FSGroup != nil {
		containerSCC.FsGroupPresent = OK
	} else {
		containerSCC.FsGroupPresent = NOK
	}
	return checkContainerCategory(pod.Spec.Containers, containerSCC, pod.Name, pod.Namespace)
}

//nolint:gocyclo,funlen
func compareCategory(refCategory, containerSCC *ContainerSCC, id CategoryID) bool {
	result := true
	tnf.ClaimFilePrintf("Testing if pod belongs to category %s", &id)
	// AllVolumeAllowed reports whether the volumes in the container are compliant to the SCC (same volume list for all SCCs).
	// True means that all volumes declared in the pod are allowed in the SCC.
	// False means that at least one volume is disallowed
	if refCategory.AllVolumeAllowed == containerSCC.AllVolumeAllowed {
		tnf.ClaimFilePrintf("AllVolumeAllowed = %s - OK", containerSCC.AllVolumeAllowed)
	} else {
		result = false
		tnf.ClaimFilePrintf("AllVolumeAllowed = %s but expected >=<=%s -  NOK", containerSCC.AllVolumeAllowed, refCategory.AllVolumeAllowed)
	}
	// RunAsUserPresent reports whether the RunAsUser Field is set to something other than nil as requested by All SCC categories.
	// True means that the RunAsUser Field is set.
	// False means that it is not set (nil)
	// The runAsUser range can be specified in the SCC itself. If not, it is specified in the namespace, see
	// https://docs.openshift.com/container-platform/4.11/authentication/managing-security-context-constraints.html#security-context-constraints-pre-allocated-values_configuring-internal-oauth
	// runAsUser:
	// type: MustRunAsRange
	// uidRangeMin: 1000
	// uidRangeMax: 2000
	if refCategory.RunAsUserPresent == containerSCC.RunAsUserPresent {
		tnf.ClaimFilePrintf("RunAsUserPresent = %s - OK", containerSCC.RunAsUserPresent)
	} else {
		tnf.ClaimFilePrintf("RunAsUserPresent = %s but expected  %s - NOK", containerSCC.RunAsUserPresent, refCategory.RunAsUserPresent)
		result = false
	}
	// RunAsNonRoot is true if the RunAsNonRoot field is set to true, false otherwise.
	// if setting a range including the roor UID 0 ( for instance 0-2000), then this option can disallow it.
	if refCategory.RunAsNonRoot >= containerSCC.RunAsNonRoot {
		tnf.ClaimFilePrintf("RunAsNonRoot = %s - OK", containerSCC.RunAsNonRoot)
	} else {
		tnf.ClaimFilePrintf("RunAsNonRoot = %s but expected  %s - NOK", containerSCC.RunAsNonRoot, refCategory.RunAsNonRoot)
		result = false
	}
	// FsGroupPresent reports whether the FsGroup Field is set to something other than nil as requested by All SCC categories.
	// True means that the FsGroup Field is set.
	// False means that it is not set (nil)
	// The FSGroup range can be specified in the SCC itself. If not, it is specified in the namespace, see
	// https://docs.openshift.com/container-platform/4.11/authentication/managing-security-context-constraints.html#security-context-constraints-pre-allocated-values_configuring-internal-oauth
	// fsGroup:
	// type: MustRunAs
	// ranges:
	// - min: 1000900000
	// max: 1000900010
	if refCategory.FsGroupPresent == containerSCC.FsGroupPresent {
		tnf.ClaimFilePrintf("FsGroupPresent  = %s - OK", containerSCC.FsGroupPresent)
	} else {
		tnf.ClaimFilePrintf("FsGroupPresent  = %s but expected  %s - NOK", containerSCC.FsGroupPresent, refCategory.FsGroupPresent)
		result = false
	}
	// RequiredDropCapabilitiesPresent is true if the drop DropCapabilities field has at least the set of required drop capabilities ( same required set for all categories ).
	// False means that some required DropCapabilities are missing.
	if refCategory.RequiredDropCapabilitiesPresent == containerSCC.RequiredDropCapabilitiesPresent {
		tnf.ClaimFilePrintf("DropCapabilities list - OK")
	} else {
		tnf.ClaimFilePrintf("RequiredDropCapabilitiesPresent = %s but expected  %s - NOK", containerSCC.RequiredDropCapabilitiesPresent, refCategory.RequiredDropCapabilitiesPresent)
		tnf.ClaimFilePrintf("its didnt have all the required (MKNOD, SETUID, SETGID, KILL)/(ALL) drop value ")
		result = false
	}
	// HostDirVolumePluginPresent is true if a hostpath volume is configured, false otherwise.
	// It is a deprecated field and is derived from the volume list currently configured in the container.
	// see https://docs.openshift.com/container-platform/3.11/admin_guide/manage_scc.html#use-the-hostpath-volume-plugin
	if refCategory.HostDirVolumePluginPresent == containerSCC.HostDirVolumePluginPresent {
		tnf.ClaimFilePrintf("HostDirVolumePluginPresent = %s - OK", containerSCC.HostDirVolumePluginPresent)
	} else {
		tnf.ClaimFilePrintf("HostDirVolumePluginPresent = %s but expected  %s - NOK", containerSCC.HostDirVolumePluginPresent, refCategory.HostDirVolumePluginPresent)
		result = false
	}
	// HostIPC is true if the HostIPC field is set to true, false otherwise.
	if refCategory.HostIPC >= containerSCC.HostIPC {
		tnf.ClaimFilePrintf("HostIPC = %s - OK", containerSCC.HostIPC)
	} else {
		result = false
		tnf.ClaimFilePrintf("HostIPC = %s but expected <= %s - NOK", containerSCC.HostIPC, refCategory.HostIPC)
	}
	// HostNetwork is true if the HostNetwork field is set to true, false otherwise.
	if refCategory.HostNetwork >= containerSCC.HostNetwork {
		tnf.ClaimFilePrintf("HostNetwork = %s - OK", containerSCC.HostNetwork)
	} else {
		result = false
		tnf.ClaimFilePrintf("HostNetwork = %s but expected <= %s - NOK", containerSCC.HostNetwork, refCategory.HostNetwork)
	}
	// HostPID is true if the HostPID field is set to true, false otherwise.
	if refCategory.HostPID >= containerSCC.HostPID {
		tnf.ClaimFilePrintf("HostPID = %s - OK", containerSCC.HostPID)
	} else {
		result = false
		tnf.ClaimFilePrintf("HostPID = %s but expected <= %s - NOK", containerSCC.HostPID, refCategory.HostPID)
	}
	// HostPorts is true if the HostPorts field is set to true, false otherwise.
	if refCategory.HostPorts >= containerSCC.HostPorts {
		tnf.ClaimFilePrintf("HostPorts = %s - OK", containerSCC.HostPorts)
	} else {
		result = false
		tnf.ClaimFilePrintf("HostPorts = %s but expected <= %s - NOK", containerSCC.HostPorts, refCategory.HostPorts)
	}
	// PrivilegeEscalation is true if the PrivilegeEscalation field is set to true, false otherwise.
	if refCategory.PrivilegeEscalation >= containerSCC.PrivilegeEscalation {
		tnf.ClaimFilePrintf("HostNetwork = %s - OK", containerSCC.HostNetwork)
	} else {
		result = false
		tnf.ClaimFilePrintf("PrivilegeEscalation = %s but expected <= %s - NOK", containerSCC.PrivilegeEscalation, refCategory.PrivilegeEscalation)
	}
	// PrivilegedContainer is true if the PrivilegedContainer field is set to true, false otherwise.
	if refCategory.PrivilegedContainer >= containerSCC.PrivilegedContainer {
		tnf.ClaimFilePrintf("PrivilegedContainer = %s - OK", containerSCC.PrivilegedContainer)
	} else {
		result = false
		tnf.ClaimFilePrintf("PrivilegedContainer = %s but expected <= %s - NOK", containerSCC.PrivilegedContainer, refCategory.PrivilegedContainer)
	}
	// ReadOnlyRootFilesystem is true if the ReadOnlyRootFilesystem field is set to true, false otherwise.
	if refCategory.ReadOnlyRootFilesystem >= containerSCC.ReadOnlyRootFilesystem {
		tnf.ClaimFilePrintf("ReadOnlyRootFilesystem = %s - OK", containerSCC.ReadOnlyRootFilesystem)
	} else {
		result = false
		tnf.ClaimFilePrintf("ReadOnlyRootFilesystem = %s but expected <= %s - NOK", containerSCC.ReadOnlyRootFilesystem, refCategory.ReadOnlyRootFilesystem)
	}
	// SeLinuxContextPresent is true if the SeLinuxContext field is present and set to a value (e.g. not nil), false otherwise.
	// An SELinuxContext strategy of MustRunAs with no level set. Admission looks for the openshift.io/sa.scc.mcs annotation to populate the level.
	if refCategory.SeLinuxContextPresent == containerSCC.SeLinuxContextPresent {
		tnf.ClaimFilePrintf("SeLinuxContextPresent  is not nil - OK")
	} else {
		result = false
		tnf.ClaimFilePrintf("SeLinuxContextPresent  = %s but expected  %s expected to be non nil - NOK", containerSCC.SeLinuxContextPresent, refCategory.SeLinuxContextPresent)
	}
	// CapabilitiesCategory indicates the lowest SCC category to which the list of capabilities.add in the container can be mapped to.
	if refCategory.CapabilitiesCategory != containerSCC.CapabilitiesCategory {
		result = false
		tnf.ClaimFilePrintf("CapabilitiesCategory = %s but expected  %s - NOK", containerSCC.CapabilitiesCategory, refCategory.CapabilitiesCategory)
	} else {
		tnf.ClaimFilePrintf("CapabilitiesCategory  list is as expected %s - OK", containerSCC.CapabilitiesCategory)
	}
	return result
}
