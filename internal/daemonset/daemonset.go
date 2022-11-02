package daemonset

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"

	k8sPriviledgedDs "github.com/test-network-function/privileged-daemonset"
)

const (
	timeout           = 5 * time.Minute
	containerName     = "container-00"
	tnfPartnerRepoDef = "quay.io/testnetworkfunction"
	supportImageDef   = "debug-partner:latest"
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

// Deploy daemon set on repo partner
func DeployPartnerTestDaemonset() error {
	imageWithVersion := buildImageWithVersion()
	oc := clientsholder.GetClientsHolder()
	k8sPriviledgedDs.SetDaemonSetClient(oc.K8sClient)

	matchLabels := make(map[string]string)
	matchLabels["name"] = provider.DaemonSetName
	matchLabels["test-network-function.com/app"] = provider.DaemonSetName

	_, err := k8sPriviledgedDs.CreateDaemonSet(provider.DaemonSetName, provider.DaemonSetNamespace, containerName, imageWithVersion, matchLabels, timeout)
	if err != nil {
		logrus.Errorf("Error deploying partner daemonset %s", err)
		return err
	}
	return nil
}
