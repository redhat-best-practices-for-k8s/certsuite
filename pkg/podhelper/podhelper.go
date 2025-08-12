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

// TopOwner represents the highest-level owner of a pod in the ownership chain.
//
// It contains the API version, kind, name, and namespace of that owner.
// The structure is used by helper functions to map pods to their primary
// controlling resources such as ReplicaSets, Deployments, or StatefulSets.
type TopOwner struct {
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
}

// GetPodTopOwner returns a map of top owners for each pod and an error if any.
//
// It takes a namespace string and a slice of OwnerReference objects, follows the
// ownership chain to determine the highest-level owner for each pod, and builds
// a mapping from pod names to their corresponding TopOwner structures. The
// function may return an error if client retrieval or reference resolution
// fails.
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

// followOwnerReferences recursively walks the ownership chain of a Kubernetes object to identify its top-level owners.
//
// It accepts a list of API resource lists, a dynamic client interface,
// a map tracking known top owners, a namespace string, and a slice of owner references.
// The function traverses each owner reference, fetching the corresponding
// resource via the dynamic client, and continues following any nested
// owner references until it reaches objects without further owners.
// The resulting top-level owners are recorded in the provided map. It returns an error if
// any API lookup fails or if the ownership chain cannot be fully resolved.
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

// searchAPIResource retrieves a specific API resource by kind and apiVersion from a list of APIResourceList.
//
// It iterates over the provided slices of metav1.APIResourceList, searching for an
// entry whose Kind matches the given kind and whose GroupVersion matches the
// supplied apiVersion. If found, it returns a pointer to that metav1.APIResource
// and a nil error. If no matching resource is found, it returns nil and an
// error describing the failure.
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
