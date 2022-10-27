package securitycontextcontainer

import (
	"reflect"
	"sort"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

type ContainerSCC struct {
	HostDirVolumePlugin    bool
	HostIPC                bool
	HostNetwork            bool
	HostPID                bool
	HostPorts              bool
	PrivilegeEscalation    bool
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
	Allowvolumes             = []string{"configMap, downwardAPI, emptyDir,persistentVolumeClaim,projected,secret"}
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
		"",
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
		"",
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
		"catagory3",
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
		"catagory4",
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
		containerSCC.PrivilegeEscalation = *(cut.SecurityContext.AllowPrivilegeEscalation)
	}
	if cut.SecurityContext != nil && cut.SecurityContext.Privileged != nil {
		containerSCC.PrivilegedContainer = *(cut.SecurityContext.Privileged)
	}

	if cut.SecurityContext != nil && cut.SecurityContext.RunAsUser != nil {
		containerSCC.RunAsUser = true
	}
	if cut.SecurityContext != nil && cut.SecurityContext.ReadOnlyRootFilesystem != nil {
		containerSCC.ReadOnlyRootFilesystem = *cut.SecurityContext.ReadOnlyRootFilesystem
	}
	if cut.SecurityContext != nil && cut.SecurityContext.RunAsNonRoot != nil {
		containerSCC.RunAsNonRoot = *cut.SecurityContext.RunAsNonRoot
	}
	if cut.SecurityContext != nil && cut.SecurityContext.SELinuxOptions != nil {
		containerSCC.SeLinuxContext = true
	} else {
		containerSCC.SeLinuxContext = false
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
		sort.Strings(sliceDropCapabilities)
		sort.Strings(requiredDropCapabilities)
		containerSCC.HaveDropCapabilities = reflect.DeepEqual(sliceDropCapabilities, requiredDropCapabilities)
		contain := true

		for _, ncc := range cut.SecurityContext.Capabilities.Add {
			if !contains(category3AddCapabilities, string(ncc)) {
				contain = false
			}
		}
		if contain {
			containerSCC.Capabilities = "catagory3"
		} else {
			contain = true
			for _, ncc := range cut.SecurityContext.Capabilities.Add {
				if !contains(category4AddCapabilities, string(ncc)) {
					contain = false
				}
			}
			if contain {
				containerSCC.Capabilities = "catagory4"
			} else {
				containerSCC.Capabilities = "catagory5"
			}
		}
	} else {
		containerSCC.Capabilities = ""
	}
	return containerSCC
}
func contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
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

func Checkcategory(containers []v1.Container, containerSCC ContainerSCC) []string {
	var badCcontainer []string
	for j := 0; j < len(containers); j++ {
		cut := &(containers[j])
		percontainerSCC := GetContainerSCC(cut, containerSCC)
		// after building the containerSCC need to check to which category it is
		switch percontainerSCC {
		case Category1:
			logrus.Info("is ok")
		case Category2:
			logrus.Info("is ok")
		case Category3:
			logrus.Info("is ok")
		case Category4:
			logrus.Info("is ok")
		default:
			badCcontainer = append(badCcontainer, cut.Name)
		}
	}
	return badCcontainer
}
