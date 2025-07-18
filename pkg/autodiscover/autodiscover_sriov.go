package autodiscover

import (
	"context"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SriovNetworkGVR defines the GroupVersionResource for SriovNetwork
var SriovNetworkGVR = schema.GroupVersionResource{
	Group:    "sriovnetwork.openshift.io",
	Version:  "v1",
	Resource: "sriovnetworks",
}

// SriovNetworkNodePolicyGVR defines the GroupVersionResource for SriovNetworkNodePolicy
var SriovNetworkNodePolicyGVR = schema.GroupVersionResource{
	Group:    "sriovnetwork.openshift.io",
	Version:  "v1",
	Resource: "sriovnetworknodepolicies",
}

func getSriovNetworks(client *clientsholder.ClientsHolder, namespaces []string) (sriovNetworks []unstructured.Unstructured, err error) {
	// Check for nil client or DynamicClient to prevent panic
	if client == nil || client.DynamicClient == nil {
		return []unstructured.Unstructured{}, nil
	}

	var sriovNetworkList []unstructured.Unstructured

	for _, ns := range namespaces {
		snl, err := client.DynamicClient.Resource(SriovNetworkGVR).Namespace(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil && !kerrors.IsNotFound(err) {
			return nil, err
		}

		// Append the list of sriovNetworks to the sriovNetworks slice
		if snl != nil {
			sriovNetworkList = append(sriovNetworkList, snl.Items...)
		}
	}
	return sriovNetworkList, nil
}

func getSriovNetworkNodePolicies(client *clientsholder.ClientsHolder, namespaces []string) (sriovNetworkNodePolicies []unstructured.Unstructured, err error) {
	// Check for nil client or DynamicClient to prevent panic
	if client == nil || client.DynamicClient == nil {
		return []unstructured.Unstructured{}, nil
	}

	var sriovNetworkNodePolicyList []unstructured.Unstructured

	for _, ns := range namespaces {
		snnp, err := client.DynamicClient.Resource(SriovNetworkNodePolicyGVR).Namespace(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil && !kerrors.IsNotFound(err) {
			return nil, err
		}

		// Append the list of sriovNetworkNodePolicies to the sriovNetworkNodePolicies slice
		if snnp != nil {
			sriovNetworkNodePolicyList = append(sriovNetworkNodePolicyList, snnp.Items...)
		}
	}
	return sriovNetworkNodePolicyList, nil
}
