package securitycontextcontainer

import (
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
	Capabilities           string
}

var (
	Catagory1 = ContainerSCC{false,
		false,
		false,
		false,
		false,
		true,
		false,
		true,
		false,
		false,
		""}

	Catagory2 = ContainerSCC{false,
		false,
		false,
		false,
		false,
		true,
		false,
		false,
		false,
		true,
		""}

	Catagory3 = ContainerSCC{false,
		false,
		false,
		false,
		false,
		true,
		false,
		true,
		false,
		false,
		"NET_ADMIN, NET_RAW"}
	Catagory4 = ContainerSCC{false,
		false,
		false,
		false,
		false,
		true,
		false,
		true,
		false,
		false,
		"IPC_LOCK, NET_ADMIN, NET_RAW"}
)

func GetContainerSCC(cut *v1.Container, containerSCC ContainerSCC) ContainerSCC {
	const istioProxyContainerUID = 1337
	containerSCC.HostPorts = false
	for _, aPort := range cut.Ports {
		if aPort.HostPort != 0 {
			containerSCC.HostPorts = true
			break
		}
	}
	if cut.SecurityContext != nil && cut.SecurityContext.AllowPrivilegeEscalation != nil {
		if *(cut.SecurityContext.AllowPrivilegeEscalation) {
			containerSCC.PrivilegedContainer = true
		} else {
			containerSCC.PrivilegedContainer = false
		}
	}
	if cut.SecurityContext != nil && cut.SecurityContext.Capabilities != nil {
		containerSCC.Capabilities = cut.SecurityContext.Capabilities.String()
	} else {
		containerSCC.Capabilities = ""
	}
	if cut.SecurityContext != nil && cut.SecurityContext.RunAsUser != nil && *cut.SecurityContext.RunAsUser == int64(istioProxyContainerUID) {
		containerSCC.RunAsUser = true
	} else {
		containerSCC.RunAsUser = false
	}
	if cut.SecurityContext != nil && cut.SecurityContext.ReadOnlyRootFilesystem != nil {
		containerSCC.ReadOnlyRootFilesystem = *cut.SecurityContext.ReadOnlyRootFilesystem
	}
	if cut.SecurityContext != nil && cut.SecurityContext.RunAsNonRoot != nil {
		containerSCC.RunAsNonRoot = *cut.SecurityContext.RunAsNonRoot
	}
	return containerSCC
}
