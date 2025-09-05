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

// TopOwner represents the highest-level resource owning a pod
//
// The structure holds identifying information about a pod's ultimate owner,
// including its API version, kind, name, and namespace. It is used by helper
// functions to map pods back to the root resource that created them. The fields
// are all strings and can be populated from Kubernetes object metadata.
type TopOwner struct {
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
}

// GetPodTopOwner Finds the highest-level owners of a pod
//
// This function starts with the namespace and owner references of a pod, then
// walks through each reference to resolve the actual resource objects via
// dynamic client calls. It recursively follows owner chains until it reaches
// resources without further owners, recording those as top owners in a map
// keyed by name. The result is returned along with any errors encountered
// during resolution.
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

// followOwnerReferences traverses owner references to discover top‑level resources
//
// The routine walks the chain of OwnerReference objects for a given Kubernetes
// resource, querying each referenced object until it reaches those without
// further owners. It records these highest-level owners in a map keyed by name,
// storing API version, kind, and namespace information. Errors during lookup or
// parsing are returned to allow callers to handle missing or malformed
// references.
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

// searchAPIResource Finds an API resource by kind and version
//
// The function iterates through a list of APIResourceList objects, matching the
// supplied group-version string to each list's GroupVersion field. Within each
// matching list it scans the contained resources for one whose Kind equals the
// provided kind value. If found, it returns a pointer to that resource;
// otherwise it reports an error indicating no match was located.
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
