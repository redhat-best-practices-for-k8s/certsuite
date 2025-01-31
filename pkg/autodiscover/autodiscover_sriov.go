package autodiscover

import (
	"context"

	sriovNetworkOp "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getSriovNetworks(client *clientsholder.ClientsHolder, namespaces []string) (sriovNetworks []sriovNetworkOp.SriovNetwork, err error) {
	var sriovNetworkList []sriovNetworkOp.SriovNetwork

	for _, ns := range namespaces {
		snl, err := client.SriovNetworkingClient.SriovNetworks(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil && !kerrors.IsNotFound(err) {
			return nil, err
		}

		// Append the list of sriovNetworks to the sriovNetworks slice
		sriovNetworkList = append(sriovNetworkList, snl.Items...)
	}
	return sriovNetworkList, nil
}

func getSriovNetworkNodePolicies(client *clientsholder.ClientsHolder, namespaces []string) (sriovNetworkNodePolicies []sriovNetworkOp.SriovNetworkNodePolicy, err error) {
	var sriovNetworkNodePolicyList []sriovNetworkOp.SriovNetworkNodePolicy

	for _, ns := range namespaces {
		snnp, err := client.SriovNetworkingClient.SriovNetworkNodePolicies(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil && !kerrors.IsNotFound(err) {
			return nil, err
		}

		// Append the list of sriovNetworkNodePolicies to the sriovNetworkNodePolicies slice
		sriovNetworkNodePolicyList = append(sriovNetworkNodePolicyList, snnp.Items...)
	}
	return sriovNetworkNodePolicyList, nil
}
