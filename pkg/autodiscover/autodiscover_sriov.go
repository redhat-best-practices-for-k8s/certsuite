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

// getSriovNetworks retrieves SR‑IOV network resources from a cluster.
//
// It accepts a ClientsHolder to access the Kubernetes API and a slice of namespace names.
// For each namespace it lists SriovNetwork and SriovNetworkNodePolicy objects,
// collecting them into a single slice of unstructured.Unstructured values.
// The function returns that slice along with any error encountered during listing,
// including handling of not‑found resources gracefully.
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

// getSriovNetworkNodePolicies retrieves all SriovNetworkNodePolicy objects from the provided client set and namespaces.
//
// It accepts a ClientsHolder containing Kubernetes clients and a slice of namespace names.
// For each namespace it lists resources defined by SriovNetworkNodePolicyGVR.
// If a namespace is not found, the error is ignored and processing continues.
// The function returns a slice of unstructured.Unstructured objects representing
// the policies that were successfully retrieved, or an error if any other issue occurs.
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
