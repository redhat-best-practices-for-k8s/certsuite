package securitycontextcontainer

import (
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

type ContainerSCC struct {
	HostDirVolumePlugin    bool
	HostIPC                bool
	HostNetwork            bool
	HostPID                bool
	HostPorts              bool
	PrivilegeEscalation    bool // this can be true or false
	PrivilegedContainer    bool
	RunAsUser              bool
	ReadOnlyRootFilesystem bool
	RunAsNonRoot           bool
	FsGroup                bool
	SeLinuxContext         bool
	Capabilities           string
	HaveDropCapabilities   bool
	AllVolumeAllowed       bool
}

var (
	requiredDropCapabilities = []string{"MKNOD", "SETUID", "SETGID", "KILL"}
	dropAll                  = []string{"ALL"}
	category2AddCapabilities = []string{"NET_ADMIN, NET_RAW"}
	category3AddCapabilities = []string{"NET_ADMIN, NET_RAW, IPC_LOCK"}
	Category1                = ContainerSCC{false,
		false,
		false,
		false,
		false,
		true,
		false,
		true,
		false,
		true,
		true,
		true,
		category1,
		true,
		true}

	Category1NoUID0 = ContainerSCC{false,
		false,
		false,
		false,
		false,
		true,
		false,
		false,
		false,
		false,
		true,
		true,
		category1,
		true,
		true}

	Category2 = ContainerSCC{false,
		false,
		false,
		false,
		false,
		true,
		false,
		true,
		false,
		true,
		true,
		true,
		"category2",
		true,
		true}

	Category3 = ContainerSCC{false,
		false,
		false,
		false,
		false,
		true,
		false,
		true,
		false,
		true,
		true,
		true,
		"category3",
		true,
		true}
)

func GetContainerSCC(cut *corev1.Container, containerSCC ContainerSCC) ContainerSCC {
	containerSCC.HostPorts = false
	for _, aPort := range cut.Ports {
		if aPort.HostPort != 0 {
			containerSCC.HostPorts = true
			break
		}
	}
	containerSCC = updateCapabilities(cut, containerSCC)
	if cut.SecurityContext != nil && cut.SecurityContext.AllowPrivilegeEscalation != nil {
		logrus.Info("PrivilegeEscalation is true")
		containerSCC.PrivilegeEscalation = true
	}
	containerSCC.PrivilegedContainer = false
	if cut.SecurityContext != nil && cut.SecurityContext.Privileged != nil {
		containerSCC.PrivilegedContainer = *(cut.SecurityContext.Privileged)
	}
	if cut.SecurityContext != nil && cut.SecurityContext.RunAsUser != nil {
		logrus.Info("RunAsUser is ", cut.SecurityContext.RunAsUser)
		containerSCC.RunAsUser = true
	}
	containerSCC.ReadOnlyRootFilesystem = false
	if cut.SecurityContext != nil && cut.SecurityContext.ReadOnlyRootFilesystem != nil {
		containerSCC.ReadOnlyRootFilesystem = *cut.SecurityContext.ReadOnlyRootFilesystem
	}
	containerSCC.RunAsNonRoot = false
	if cut.SecurityContext != nil && cut.SecurityContext.RunAsNonRoot != nil {
		logrus.Info("RunAsNonRoot is ", *(cut.SecurityContext.RunAsNonRoot))
		containerSCC.RunAsNonRoot = *cut.SecurityContext.RunAsNonRoot
	}
	if cut.SecurityContext != nil && cut.SecurityContext.SELinuxOptions != nil {
		containerSCC.SeLinuxContext = true
	}
	return containerSCC
}

func updateCapabilities(cut *corev1.Container, containerSCC ContainerSCC) ContainerSCC {
	containerSCC.HaveDropCapabilities = false
	if cut.SecurityContext != nil && cut.SecurityContext.Capabilities != nil {
		var sliceDropCapabilities []string
		for _, ncc := range cut.SecurityContext.Capabilities.Drop {
			sliceDropCapabilities = append(sliceDropCapabilities, string(ncc))
		}
		logrus.Info("cut.SecurityContext.Capabilities.Drop", cut.SecurityContext.Capabilities.Drop)
		sort.Strings(sliceDropCapabilities)
		logrus.Info("sliceDropCapabilities", sliceDropCapabilities)

		sort.Strings(requiredDropCapabilities)
		if reflect.DeepEqual(sliceDropCapabilities, requiredDropCapabilities) || reflect.DeepEqual(sliceDropCapabilities, dropAll) {
			containerSCC.HaveDropCapabilities = true
		}

		if checkContainCategory(cut.SecurityContext.Capabilities.Add, category2AddCapabilities) {
			logrus.Info("category is category2")
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
	return containerSCC
}

func AllVolumeAllowed(volumes []corev1.Volume) bool {
	countVolume := 0
	for j := 0; j < len(volumes); j++ {
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
	return countVolume == len(volumes)
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

const (
	CategoryID1String       = "CategoryID1"
	CategoryID1NoUID0String = "CategoryID1NoUID0"
	CategoryID2String       = "CategoryID2"
	CategoryID3String       = "CategoryID3"
	CategoryID4String       = "OtherTypes"
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

//nolint:funlen
func CheckCategory(containers []corev1.Container, containerSCC ContainerSCC, podName, nameSpace string) []PodListcategory {
	var ContainerList []PodListcategory
	var categoryinfo PodListcategory
	for j := 0; j < len(containers); j++ {
		cut := &(containers[j])
		percontainerSCC := GetContainerSCC(cut, containerSCC)
		tnf.ClaimFilePrintf("percontainerSCC is ", percontainerSCC)
		// after building the containerSCC need to check to which category it is
		switch percontainerSCC {
		case Category1:
			categoryinfo = PodListcategory{
				Containername: cut.Name,
				Podname:       podName,
				NameSpace:     nameSpace,
				Category:      CategoryID1,
			}
			logrus.Info("Category1")
		case Category1NoUID0:
			categoryinfo = PodListcategory{
				Containername: cut.Name,
				Podname:       podName,
				NameSpace:     nameSpace,
				Category:      CategoryID1NoUID0,
			}
			logrus.Info("its Category1-no-uid0")
		case Category2:
			categoryinfo = PodListcategory{
				Containername: cut.Name,
				Podname:       podName,
				NameSpace:     nameSpace,
				Category:      CategoryID2,
			}
			logrus.Info("its Category2")
		case Category3:
			categoryinfo = PodListcategory{
				Containername: cut.Name,
				Podname:       podName,
				NameSpace:     nameSpace,
				Category:      CategoryID3,
			}
			logrus.Info("its Category3")
		default:
			categoryinfo = PodListcategory{
				Containername: cut.Name,
				Podname:       podName,
				NameSpace:     nameSpace,
				Category:      CategoryID4,
			}

			logrus.Info("no one from the categories")
		}
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
	containerSCC.HostIPC = pod.Spec.HostIPC
	containerSCC.HostNetwork = pod.Spec.HostNetwork
	containerSCC.HostPID = pod.Spec.HostPID
	if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.SELinuxOptions != nil {
		logrus.Info("SELinuxOptions is ", pod.Spec.SecurityContext.SELinuxOptions)
		logrus.Info("SELinuxOptions is true")
		containerSCC.SeLinuxContext = true
	} else {
		logrus.Info("SELinuxOptions is false")
		containerSCC.SeLinuxContext = false
	}
	containerSCC.AllVolumeAllowed = AllVolumeAllowed(pod.Spec.Volumes)
	if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.RunAsUser != nil {
		logrus.Info("Spec.SecurityContext.RunAsUser is true")
		containerSCC.RunAsUser = true
	} else {
		logrus.Info("Spec.SecurityContext.RunAsUser is false")
		containerSCC.RunAsUser = false
	}
	if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.FSGroup != nil {
		logrus.Info("FsGroupis true")
		containerSCC.FsGroup = true
	} else {
		logrus.Info("FsGroupis false")
		containerSCC.FsGroup = false
	}
	return CheckCategory(pod.Spec.Containers, containerSCC, pod.Name, pod.Namespace)
}
