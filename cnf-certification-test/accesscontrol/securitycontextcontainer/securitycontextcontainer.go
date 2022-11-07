package securitycontextcontainer

// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/
// api we used as a reference

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
	HostDirVolumePluginPresent OkNok // 0 or 1 - 0 is false 1 - true
	HostIPC                    OkNok
	HostNetwork                OkNok
	HostPID                    OkNok
	HostPorts                  OkNok
	PrivilegeEscalation        OkNok // this can be true or false
	PrivilegedContainer        OkNok
	RunAsUserPresent           OkNok // thes filed that checking if the value is present
	ReadOnlyRootFilesystem     OkNok
	RunAsNonRoot               OkNok
	FsGroupPresent             OkNok
	SeLinuxContextPresent      OkNok
	Capabilities               string
	HaveDropCapabilities       OkNok
	AllVolumeAllowed           OkNok
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

// nolint[:gocritic,:gocyclo]
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
					containerSCC.Capabilities = "category4"
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

//nolint:gocritic
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

func checkContainCategory(addCapability []corev1.Capability, categoryAddCapabilities []string) bool {
	for _, ncc := range addCapability {
		return stringhelper.StringInSlice(categoryAddCapabilities, string(ncc), true)
	}
	return len(addCapability) > 0
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
	return CheckCategory(pod.Spec.Containers, containerSCC, pod.Name, pod.Namespace)
}

// nolint:[gocyclo,funlen]
func compareCategory(refCategory, containerSCC *ContainerSCC, id CategoryID) bool {
	result := true
	tnf.ClaimFilePrintf("Testing if pod belongs to category %s", &id)
	if refCategory.AllVolumeAllowed == containerSCC.AllVolumeAllowed {
		tnf.ClaimFilePrintf("AllVolumeAllowed = %s - OK", containerSCC.AllVolumeAllowed)
	} else {
		result = false
		tnf.ClaimFilePrintf("AllVolumeAllowed = %s but expected >=<=%s -  NOK", containerSCC.AllVolumeAllowed, refCategory.AllVolumeAllowed)
	}
	if refCategory.RunAsUserPresent == containerSCC.RunAsUserPresent {
		tnf.ClaimFilePrintf("RunAsUserPresent = %s - OK", containerSCC.RunAsUserPresent)
	} else {
		tnf.ClaimFilePrintf("RunAsUserPresent = %s but expected  %s - NOK", containerSCC.RunAsUserPresent, refCategory.RunAsUserPresent)
		result = false
	}
	if refCategory.RunAsNonRoot >= containerSCC.RunAsNonRoot {
		tnf.ClaimFilePrintf("RunAsNonRoot = %s - OK", containerSCC.RunAsNonRoot)
	} else {
		tnf.ClaimFilePrintf("RunAsNonRoot = %s but expected  %s - NOK", containerSCC.RunAsNonRoot, refCategory.RunAsNonRoot)
		result = false
	}
	if refCategory.FsGroupPresent == containerSCC.FsGroupPresent {
		tnf.ClaimFilePrintf("FsGroupPresent  = %s - OK", containerSCC.FsGroupPresent)
	} else {
		tnf.ClaimFilePrintf("FsGroupPresent  = %s but expected  %s - NOK", containerSCC.FsGroupPresent, refCategory.FsGroupPresent)
		result = false
	}
	if refCategory.HaveDropCapabilities == containerSCC.HaveDropCapabilities {
		tnf.ClaimFilePrintf("DropCapabilities list - OK")
	} else {
		tnf.ClaimFilePrintf("HaveDropCapabilities = %s but expected  %s - NOK", containerSCC.HaveDropCapabilities, refCategory.HaveDropCapabilities)
		tnf.ClaimFilePrintf("its didnt have all the required (MKNOD, SETUID, SETGID, KILL)/(ALL) drop value ")
		result = false
	}
	if refCategory.HostDirVolumePluginPresent == containerSCC.HostDirVolumePluginPresent {
		tnf.ClaimFilePrintf("HostDirVolumePluginPresent = %s - OK", containerSCC.HostDirVolumePluginPresent)
	} else {
		tnf.ClaimFilePrintf("HostDirVolumePluginPresent = %s but expected  %s - NOK", containerSCC.HostDirVolumePluginPresent, refCategory.HostDirVolumePluginPresent)
		result = false
	}
	if refCategory.HostIPC >= containerSCC.HostIPC {
		tnf.ClaimFilePrintf("HostIPC = %s - OK", containerSCC.HostIPC)
	} else {
		result = false
		tnf.ClaimFilePrintf("HostIPC = %s but expected <= %s - NOK", containerSCC.HostIPC, refCategory.HostIPC)
	}
	if refCategory.HostNetwork >= containerSCC.HostNetwork {
		tnf.ClaimFilePrintf("HostNetwork = %s - OK", containerSCC.HostNetwork)
	} else {
		result = false
		tnf.ClaimFilePrintf("HostNetwork = %s but expected <= %s - NOK", containerSCC.HostNetwork, refCategory.HostNetwork)
	}
	if refCategory.HostPID >= containerSCC.HostPID {
		tnf.ClaimFilePrintf("HostPID = %s - OK", containerSCC.HostPID)
	} else {
		result = false
		tnf.ClaimFilePrintf("HostPID = %s but expected <= %s - NOK", containerSCC.HostPID, refCategory.HostPID)
	}
	if refCategory.HostPorts >= containerSCC.HostPorts {
		tnf.ClaimFilePrintf("HostPorts = %s - OK", containerSCC.HostPorts)
	} else {
		result = false
		tnf.ClaimFilePrintf("HostPorts = %s but expected <= %s - NOK", containerSCC.HostPorts, refCategory.HostPorts)
	}
	if refCategory.PrivilegeEscalation >= containerSCC.PrivilegeEscalation {
		tnf.ClaimFilePrintf("HostNetwork = %s - OK", containerSCC.HostNetwork)
	} else {
		result = false
		tnf.ClaimFilePrintf("PrivilegeEscalation = %s but expected <= %s - NOK", containerSCC.PrivilegeEscalation, refCategory.PrivilegeEscalation)
	}
	if refCategory.PrivilegedContainer >= containerSCC.PrivilegedContainer {
		tnf.ClaimFilePrintf("PrivilegedContainer = %s - OK", containerSCC.PrivilegedContainer)
	} else {
		result = false
		tnf.ClaimFilePrintf("PrivilegedContainer = %s but expected <= %s - NOK", containerSCC.PrivilegedContainer, refCategory.PrivilegedContainer)
	}
	if refCategory.ReadOnlyRootFilesystem >= containerSCC.ReadOnlyRootFilesystem {
		tnf.ClaimFilePrintf("ReadOnlyRootFilesystem = %s - OK", containerSCC.ReadOnlyRootFilesystem)
	} else {
		result = false
		tnf.ClaimFilePrintf("ReadOnlyRootFilesystem = %s but expected <= %s - NOK", containerSCC.ReadOnlyRootFilesystem, refCategory.ReadOnlyRootFilesystem)
	}
	if refCategory.SeLinuxContextPresent == containerSCC.SeLinuxContextPresent {
		tnf.ClaimFilePrintf("SeLinuxContextPresent  is not nil - OK")
	} else {
		result = false
		tnf.ClaimFilePrintf("SeLinuxContextPresent  = %s but expected  %s expected to be non nil - NOK", containerSCC.SeLinuxContextPresent, refCategory.SeLinuxContextPresent)
	}
	if refCategory.Capabilities != containerSCC.Capabilities {
		result = false
		tnf.ClaimFilePrintf("Capabilities = %s but expected  %s - NOK", containerSCC.Capabilities, refCategory.Capabilities)
	}
	return result
}
