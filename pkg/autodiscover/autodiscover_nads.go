package autodiscover

import (
	"context"

	nadClient "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getNetworkAttachmentDefinitions Retrieves all network attachment definitions from specified namespaces
//
// The function iterates over a list of namespace names, querying each for its
// NetworkAttachmentDefinition resources via the CNCF networking client. It
// collects any found items into a single slice, handling missing namespaces
// gracefully by ignoring not‑found errors. The resulting slice and an are
// returned to the caller.
func getNetworkAttachmentDefinitions(client *clientsholder.ClientsHolder, namespaces []string) ([]nadClient.NetworkAttachmentDefinition, error) {
	var nadList []nadClient.NetworkAttachmentDefinition

	for _, ns := range namespaces {
		nad, err := client.CNCFNetworkingClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil && !kerrors.IsNotFound(err) {
			return nil, err
		}

		// Append the list of networkAttachmentDefinitions to the nadList slice
		nadList = append(nadList, nad.Items...)
	}

	return nadList, nil
}
