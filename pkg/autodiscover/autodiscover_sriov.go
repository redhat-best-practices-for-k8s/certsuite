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

// getSriovNetworks Retrieves all SR‑IOV network resources from the specified namespaces
//
// The function iterates over each namespace, using a dynamic client to list
// objects of the SR‑IOV Network type. It skips namespaces where the resource
// is not found and aggregates the items into a single slice. If the client or
// its DynamicClient is nil, it safely returns an empty result without error.
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

// getSriovNetworkNodePolicies Collects SR-IOV network node policies from specified namespaces
//
// The function iterates over each provided namespace, querying the dynamic
// client for SR‑IOV network node policy resources. It aggregates all found
// items into a single slice, handling missing clients or non‑existent
// resources gracefully by returning an empty list instead of panicking. Errors
// unrelated to a resource not being found are propagated back to the caller.
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
