package securitycontextcontainer

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	corev1 "k8s.io/api/core/v1"
)

const (
	category1        = "category1"
	haveRequiredDrop = "haveRequiredDrop"
)

type OkNok int

const (
	OKNOK = iota
	OK
	NOK
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
	HostDirVolumePlugin    OkNok // 0 or 1 - 0 is false 1 - true
	HostIPC                OkNok
	HostNetwork            OkNok
	HostPID                OkNok
	HostPorts              OkNok
	PrivilegeEscalation    OkNok // this can be true or false
	PrivilegedContainer    OkNok
	RunAsUser              OkNok
	ReadOnlyRootFilesystem OkNok
	RunAsNonRoot           OkNok
	FsGroup                OkNok
	SeLinuxContext         OkNok
	Capabilities           string
	HaveDropCapabilities   OkNok
	AllVolumeAllowed       OkNok
}

var (
	requiredDropCapabilities = []string{"MKNOD", "SETUID", "SETGID", "KILL"}
	dropAll                  = []string{"ALL"}
	category2AddCapabilities = []string{"NET_ADMIN, NET_RAW"}
	category3AddCapabilities = []string{"NET_ADMIN, NET_RAW, IPC_LOCK"}
	Category1                = ContainerSCC{NOK,
		NOK,
		NOK,
		NOK,
		NOK,
		OK,
		NOK,
		OK,
		NOK,
		OK,
		OK,
		OK,
		category1,
		OK,
		OK}

	Category1NoUID0 = ContainerSCC{NOK,
		NOK,
		NOK,
		NOK,
		NOK,
		OK,
		NOK,
		NOK,
		NOK,
		NOK,
		OK,
		OK,
		category1,
		OK,
		OK}

	Category2 = ContainerSCC{NOK,
		NOK,
		NOK,
		NOK,
		NOK,
		OK,
		NOK,
		OK,
		NOK,
		OK,
		OK,
		OK,
		"category2",
		OK,
		OK}

	Category3 = ContainerSCC{NOK,
		NOK,
		NOK,
		NOK,
		NOK,
		OK,
		NOK,
		OK,
		NOK,
		OK,
		OK,
		OK,
		"category3",
		OK,
		OK}
)

func GetContainerSCC(cut *provider.Container, containerSCC ContainerSCC) ContainerSCC {
	containerSCC.HostPorts = NOK
	for _, aPort := range cut.Ports {
		if aPort.HostPort != 0 {
			containerSCC.HostPorts = OK
			break
		}
	}
	updateCapabilities(cut, &containerSCC)
	containerSCC.PrivilegeEscalation = NOK
	if cut.SecurityContext != nil && cut.SecurityContext.AllowPrivilegeEscalation != nil {
		containerSCC.PrivilegeEscalation = OK
	}
	containerSCC.PrivilegedContainer = NOK
	if cut.SecurityContext != nil && cut.SecurityContext.Privileged != nil && *(cut.SecurityContext.Privileged) {
		containerSCC.PrivilegedContainer = OK
	}
	containerSCC.RunAsUser = NOK
	if cut.SecurityContext != nil && cut.SecurityContext.RunAsUser != nil {
		containerSCC.RunAsUser = OK
	}
	containerSCC.ReadOnlyRootFilesystem = NOK
	if cut.SecurityContext != nil && cut.SecurityContext.ReadOnlyRootFilesystem != nil && *cut.SecurityContext.ReadOnlyRootFilesystem {
		containerSCC.ReadOnlyRootFilesystem = OK
	}
	containerSCC.RunAsNonRoot = NOK
	if cut.SecurityContext != nil && cut.SecurityContext.RunAsNonRoot != nil && *cut.SecurityContext.RunAsNonRoot {
		containerSCC.RunAsNonRoot = OK
	}
	containerSCC.SeLinuxContext = NOK
	if cut.SecurityContext != nil && cut.SecurityContext.SELinuxOptions != nil {
		containerSCC.SeLinuxContext = OK
	}
	return containerSCC
}

func updateCapabilities(cut *provider.Container, containerSCC *ContainerSCC) {
	containerSCC.HaveDropCapabilities = NOK
	if cut.SecurityContext != nil && cut.SecurityContext.Capabilities != nil {
		var sliceDropCapabilities []string
		for _, ncc := range cut.SecurityContext.Capabilities.Drop {
			sliceDropCapabilities = append(sliceDropCapabilities, string(ncc))
		}
		sort.Strings(sliceDropCapabilities)

		sort.Strings(requiredDropCapabilities)
		if reflect.DeepEqual(sliceDropCapabilities, requiredDropCapabilities) || reflect.DeepEqual(sliceDropCapabilities, dropAll) {
			containerSCC.HaveDropCapabilities = OK
		}

		if checkContainCategory(cut.SecurityContext.Capabilities.Add, category2AddCapabilities) {
			containerSCC.Capabilities = "category2"
		} else {
			if checkContainCategory(cut.SecurityContext.Capabilities.Add, category3AddCapabilities) {
				containerSCC.Capabilities = "category3"
			} else {
				if len(cut.SecurityContext.Capabilities.Add) > 0 {
					containerSCC.Capabilities = "category5"
				} else {
					logrus.Info("category is category1")
					containerSCC.Capabilities = category1
				}
			}
		}
	} else {
		containerSCC.Capabilities = category1
	}
}

func AllVolumeAllowed(volumes []corev1.Volume) (OkNok, OkNok) {
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

// print the strings
func (category PodListcategory) String() string {
	returnString := fmt.Sprintf("Containername: %s Podname: %s NameSpace: %s Category: %s \n ",
		category.Containername, category.Podname, category.Podname, category.Category)
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

// //nolint:gocritic
func CheckCategory(containers []corev1.Container, containerSCC ContainerSCC, podName, nameSpace string) []PodListcategory {
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
		if compareCategory(Category1, percontainerSCC, CategoryID1) {
			tnf.ClaimFilePrintf("Testing if pod belongs to category1 ")
			categoryinfo.Category = CategoryID1
		} else if compareCategory(Category1NoUID0, percontainerSCC, CategoryID1NoUID0) {
			tnf.ClaimFilePrintf("Testing if pod belongs to category1NoUID0 ")
			categoryinfo.Category = CategoryID1NoUID0
		} else if compareCategory(Category2, percontainerSCC, CategoryID2) {
			categoryinfo.Category = CategoryID2
		} else if compareCategory(Category3, percontainerSCC, CategoryID3) {
			categoryinfo.Category = CategoryID3
		} else {
			categoryinfo.Category = CategoryID4
		}
		// after building the containerSCC need to check to which category it is
		ContainerList = append(ContainerList, categoryinfo)
	}
	return ContainerList
}

func checkContainCategory(addCapability []corev1.Capability, categoryAddCapabilities []string) bool {
	for _, ncc := range addCapability {
		return stringhelper.StringInSlice(categoryAddCapabilities, string(ncc), true)
	}
	return len(addCapability) > 0
}

func CheckPod(pod *provider.Pod) []PodListcategory {
	var containerSCC ContainerSCC
	//containerSCC.HostDirVolumePlugin = NOK
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
	if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.SELinuxOptions != nil {
		containerSCC.SeLinuxContext = OK
	} else {
		containerSCC.SeLinuxContext = NOK
	}
	containerSCC.AllVolumeAllowed, containerSCC.HostDirVolumePlugin = AllVolumeAllowed(pod.Spec.Volumes)
	if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.RunAsUser != nil {
		containerSCC.RunAsUser = OK
	} else {
		containerSCC.RunAsUser = NOK
	}
	if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.FSGroup != nil {
		containerSCC.FsGroup = OK
	} else {
		containerSCC.FsGroup = NOK
	}
	return CheckCategory(pod.Spec.Containers, containerSCC, pod.Name, pod.Namespace)
}

//nolint:funlen
func compareCategory(refCategory, containerSCC ContainerSCC, id CategoryID) bool {
	result := true
	tnf.ClaimFilePrintf("Testing if pod belongs to category %s", &id)
	if refCategory.AllVolumeAllowed < containerSCC.AllVolumeAllowed {
		tnf.ClaimFilePrintf("AllVolumeAllowed = %s but expected <= %s -  NOK", containerSCC.AllVolumeAllowed, refCategory.AllVolumeAllowed)
		result = false
	} else {
		tnf.ClaimFilePrintf("AllVolumeAllowed = %s - OK", containerSCC.AllVolumeAllowed)
	}
	if refCategory.FsGroup < containerSCC.FsGroup {
		tnf.ClaimFilePrintf("FsGroup = %s but expected <= %s - NOK", containerSCC.FsGroup, refCategory.FsGroup)
		result = false
	} else {
		tnf.ClaimFilePrintf("FsGroup = %s - OK", containerSCC.FsGroup)
	}
	if refCategory.HaveDropCapabilities < containerSCC.HaveDropCapabilities {
		tnf.ClaimFilePrintf("HaveDropCapabilities = %s but expected <= %s - NOK", containerSCC.HaveDropCapabilities, refCategory.HaveDropCapabilities)
		tnf.ClaimFilePrintf("its didnt have all the required (MKNOD, SETUID, SETGID, KILL)/(ALL) drop value ")
		result = false
	} else {
		tnf.ClaimFilePrintf("DropCapabilities list - OK")
	}
	if refCategory.HostDirVolumePlugin < containerSCC.HostDirVolumePlugin {
		tnf.ClaimFilePrintf("HostDirVolumePlugin = %s but expected <= %s - NOK", containerSCC.HostDirVolumePlugin, refCategory.HostDirVolumePlugin)
		result = false
	} else {
		tnf.ClaimFilePrintf("HostDirVolumePlugin = %s - OK", containerSCC.HostDirVolumePlugin)
	}
	if refCategory.HostIPC < containerSCC.HostIPC {
		result = false
		tnf.ClaimFilePrintf("HostIPC = %s but expected <= %s - NOK", containerSCC.HostIPC, refCategory.HostIPC)
	} else {
		tnf.ClaimFilePrintf("HostIPC = %s - OK", containerSCC.HostIPC)
	}
	if refCategory.HostNetwork < containerSCC.HostNetwork {
		result = false
		tnf.ClaimFilePrintf("HostNetwork = %s but expected <= %s - NOK", containerSCC.HostNetwork, refCategory.HostNetwork)
	} else {
		tnf.ClaimFilePrintf("HostNetwork = %s - OK", containerSCC.HostNetwork)
	}
	if refCategory.HostPID < containerSCC.HostPID {
		result = false
		tnf.ClaimFilePrintf("HostPID = %s but expected <= %s - NOK", containerSCC.HostPID, refCategory.HostPID)
	} else {
		tnf.ClaimFilePrintf("HostPID = %s - OK", containerSCC.HostPID)
	}
	if refCategory.HostPorts < containerSCC.HostPorts {
		result = false
		tnf.ClaimFilePrintf("HostPorts = %s but expected <= %s - NOK", containerSCC.HostPorts, refCategory.HostPorts)
	} else {
		tnf.ClaimFilePrintf("HostPorts = %s - OK", containerSCC.HostPorts)
	}
	if refCategory.PrivilegeEscalation < containerSCC.PrivilegeEscalation {
		//result = false
		tnf.ClaimFilePrintf("PrivilegeEscalation = %s but expected <= %s - NOK", containerSCC.PrivilegeEscalation, refCategory.PrivilegeEscalation)
	} else {
		tnf.ClaimFilePrintf("HostNetwork = %s - OK", containerSCC.HostNetwork)
	}
	if refCategory.PrivilegedContainer < containerSCC.PrivilegedContainer {
		result = false
		tnf.ClaimFilePrintf("PrivilegedContainer = %s but expected <= %s - NOK", containerSCC.PrivilegedContainer, refCategory.PrivilegedContainer)
	} else {
		tnf.ClaimFilePrintf("PrivilegedContainer = %s - OK", containerSCC.PrivilegedContainer)
	}
	if refCategory.ReadOnlyRootFilesystem < containerSCC.ReadOnlyRootFilesystem {
		result = false
		tnf.ClaimFilePrintf("ReadOnlyRootFilesystem = %s but expected <= %s - NOK", containerSCC.ReadOnlyRootFilesystem, refCategory.ReadOnlyRootFilesystem)
	} else {
		tnf.ClaimFilePrintf("ReadOnlyRootFilesystem = %s - OK", containerSCC.ReadOnlyRootFilesystem)
	}
	if refCategory.SeLinuxContext < containerSCC.SeLinuxContext {
		result = false
		tnf.ClaimFilePrintf("SeLinuxContext = %s but expected <= %s - NOK", containerSCC.SeLinuxContext, refCategory.SeLinuxContext)
		tnf.ClaimFilePrintf("SeLinuxContext expected to be non nil")
	} else {
		tnf.ClaimFilePrintf("SeLinuxContext is not nil - OK")
	}
	if refCategory.Capabilities < containerSCC.Capabilities {
		result = false
		tnf.ClaimFilePrintf("Capabilities = %s but expected <= %s - NOK", containerSCC.Capabilities, refCategory.Capabilities)
	}
	return result
}
