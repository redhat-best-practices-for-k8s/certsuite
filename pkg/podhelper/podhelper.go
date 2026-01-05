// Copyright (C) 2024-2026 Red Hat, Inc.
package podhelper

import (
	"context"
	"fmt"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Structure to describe a top owner of a pod
type TopOwner struct {
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
}

// Get the list of top owners of pods
func GetPodTopOwner(podNamespace string, podOwnerReferences []metav1.OwnerReference) (topOwners map[string]TopOwner, err error) {
	topOwners = make(map[string]TopOwner)
	err = followOwnerReferences(
		clientsholder.GetClientsHolder().GroupResources,
		clientsholder.GetClientsHolder().DynamicClient,
		topOwners,
		podNamespace,
		podOwnerReferences)
	if err != nil {
		return topOwners, fmt.Errorf("could not get top owners, err: %v", err)
	}
	return topOwners, nil
}

// Recursively follow the ownership tree to find the top owners
func followOwnerReferences(resourceList []*metav1.APIResourceList, dynamicClient dynamic.Interface, topOwners map[string]TopOwner, namespace string, ownerRefs []metav1.OwnerReference) (err error) {
	for _, ownerRef := range ownerRefs {
		apiResource, err := searchAPIResource(ownerRef.Kind, ownerRef.APIVersion, resourceList)
		if err != nil {
			return fmt.Errorf("error searching APIResource for owner reference %v: %v", ownerRef, err)
		}

		gv, err := schema.ParseGroupVersion(ownerRef.APIVersion)
		if err != nil {
			return fmt.Errorf("failed to parse apiVersion %q: %v", ownerRef.APIVersion, err)
		}

		gvr := schema.GroupVersionResource{
			Group:    gv.Group,
			Version:  gv.Version,
			Resource: apiResource.Name,
		}

		// If the owner reference is a non-namespaced resource (like Node), we need to change the namespace to empty string.
		if !apiResource.Namespaced {
			namespace = ""
		}

		// Get the owner resource, but don't care if it's not found: it might happen for ocp jobs that are constantly
		// spawned and removed after completion.
		resource, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), ownerRef.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return fmt.Errorf("could not get object indicated by owner references %+v (gvr=%+v): %v", ownerRef, gvr, err)
		}

		// Get owner references of the unstructured object
		ownerReferences := resource.GetOwnerReferences()
		// if no owner references, we have reached the top record it
		if len(ownerReferences) == 0 {
			topOwners[ownerRef.Name] = TopOwner{APIVersion: ownerRef.APIVersion, Kind: ownerRef.Kind, Name: ownerRef.Name, Namespace: namespace}
			continue
		}

		err = followOwnerReferences(resourceList, dynamicClient, topOwners, namespace, ownerReferences)
		if err != nil {
			return err
		}
	}

	return nil
}

// searchAPIResource is a helper func that returns the metav1.APIResource pointer of the resource by kind and apiVersion.
// from a metav1.APIResourceList.
func searchAPIResource(kind, apiVersion string, apis []*metav1.APIResourceList) (*metav1.APIResource, error) {
	for _, api := range apis {
		if api.GroupVersion != apiVersion {
			continue
		}

		for i := range api.APIResources {
			apiResource := &api.APIResources[i]

			if kind == apiResource.Kind {
				return apiResource, nil
			}
		}
	}

	return nil, fmt.Errorf("apiResource not found for kind=%v and APIVersion=%v", kind, apiVersion)
}
