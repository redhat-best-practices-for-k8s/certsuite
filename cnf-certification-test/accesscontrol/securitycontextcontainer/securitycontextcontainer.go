package securitycontextcontainer

import (
	"reflect"
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	v1 "k8s.io/api/core/v1"
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
	requiredDropCapabilities = []string{"KILL", "MKNOD", "SETUID", "SETGID"}
	category3AddCapabilities = []string{"NET_ADMIN, NET_RAW"}
	category4AddCapabilities = []string{"NET_ADMIN, NET_RAW, IPC_LOCK"}
	Category1                = ContainerSCC{false,
		false,
		false,
		false,
		false,
		true,
		false,
		true,
		false,
		false,
		true,
		true,
		"category1,2",
		true,
		true}

	Category2 = ContainerSCC{false,
		false,
		false,
		false,
		false,
		true,
		false,
		false,
		false,
		true,
		true,
		true,
		"category1,2",
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
		false,
		true,
		true,
		"category3",
		true,
		true}
	Category4 = ContainerSCC{false,
		false,
		false,
		false,
		false,
		true,
		false,
		true,
		false,
		false,
		true,
		true,
		"category4",
		true,
		true}
)

func GetContainerSCC(cut *v1.Container, containerSCC ContainerSCC) ContainerSCC {
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

func updateCapabilities(cut *v1.Container, containerSCC ContainerSCC) ContainerSCC {
	containerSCC.HaveDropCapabilities = false
	if cut.SecurityContext != nil && cut.SecurityContext.Capabilities != nil {
		var sliceDropCapabilities []string
		for _, ncc := range cut.SecurityContext.Capabilities.Drop {
			sliceDropCapabilities = append(sliceDropCapabilities, string(ncc))
		}
		logrus.Info("cut.SecurityContext.Capabilities.Drop", cut.SecurityContext.Capabilities.Drop)
		sort.Strings(sliceDropCapabilities)
		sort.Strings(requiredDropCapabilities)
		containerSCC.HaveDropCapabilities = reflect.DeepEqual(sliceDropCapabilities, requiredDropCapabilities)
		if checkContainCateegory(cut.SecurityContext.Capabilities.Add, category3AddCapabilities) {
			logrus.Info("category is category3")
			containerSCC.Capabilities = "category3"
		} else {
			if checkContainCateegory(cut.SecurityContext.Capabilities.Add, category4AddCapabilities) {
				containerSCC.Capabilities = "category4"
			} else {
				if len(cut.SecurityContext.Capabilities.Add) > 0 {
					containerSCC.Capabilities = "category5"
				} else {
					logrus.Info("category is category1,2")
					containerSCC.Capabilities = "category1,2"
				}
			}
		}
	} else {
		containerSCC.Capabilities = "category1,2"
	}
	return containerSCC
}

func AllVolumeAllowed(volumes []v1.Volume) bool {
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

func CheckCategory(containers []v1.Container, containerSCC ContainerSCC) []string {
	var badCcontainer []string
	for j := 0; j < len(containers); j++ {
		cut := &(containers[j])
		percontainerSCC := GetContainerSCC(cut, containerSCC)
		tnf.ClaimFilePrintf("percontainerSCC is ", percontainerSCC)

		// after building the containerSCC need to check to which category it is
		switch percontainerSCC {
		case Category1:
			logrus.Info("is ok")
		case Category2:
			badCcontainer = append(badCcontainer, cut.Name)
			logrus.Info("its Category2")
		case Category3:
			badCcontainer = append(badCcontainer, cut.Name)
			logrus.Info("its Category3")
		case Category4:
			badCcontainer = append(badCcontainer, cut.Name)
			logrus.Info("its Category4")
		default:
			logrus.Info("no one from the categories")
			badCcontainer = append(badCcontainer, cut.Name)
		}
	}
	return badCcontainer
}
func checkContainCateegory(addCapability []v1.Capability, categoryAddCapabilities []string) bool {
	for _, ncc := range addCapability {
		return stringhelper.StringInSlice(categoryAddCapabilities, string(ncc), true)
	}
	return len(addCapability) > 0
}

func CheckPod(pod *provider.Pod) []string {
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
	return CheckCategory(pod.Spec.Containers, containerSCC)
}
