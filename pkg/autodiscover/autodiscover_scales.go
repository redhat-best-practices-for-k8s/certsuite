package autodiscover

import (
	"context"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	scalingv1 "k8s.io/api/autoscaling/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ScaleObject represents a scalable custom resource
//
// This structure holds the scale subresource of a custom resource, along with
// its group‑resource identity. It is used to read or modify the replica count
// for that resource via the Kubernetes scaling API.
type ScaleObject struct {
	Scale               *scalingv1.Scale
	GroupResourceSchema schema.GroupResource
}

// GetScaleCrUnderTest Retrieves scalable custom resources across specified namespaces
//
// It iterates over a list of CustomResourceDefinitions, filtering for
// namespace-scoped and having a scale subresource. For each qualifying CRD it
// lists the custom resources in the provided namespaces using a dynamic client,
// then gathers their scale objects. The result is a slice of ScaleObject
// containing scaling information for each found resource.
func GetScaleCrUnderTest(namespaces []string, crds []*apiextv1.CustomResourceDefinition) []ScaleObject {
	dynamicClient := clientsholder.GetClientsHolder().DynamicClient

	var scaleObjects []ScaleObject
	for _, crd := range crds {
		if crd.Spec.Scope != apiextv1.NamespaceScoped {
			log.Warn("Target CRD %q is cluster-wide scoped. Skipping search of scale objects.", crd.Name)
			continue
		}

		for i := range crd.Spec.Versions {
			crdVersion := crd.Spec.Versions[i]
			gvr := schema.GroupVersionResource{
				Group:    crd.Spec.Group,
				Version:  crdVersion.Name,
				Resource: crd.Spec.Names.Plural,
			}

			// Filter out non-scalable CRDs.
			if crdVersion.Subresources == nil || crdVersion.Subresources.Scale == nil {
				log.Info("Target CRD %q is not scalable. Skipping search of scalable CRs.", crd.Name)
				continue
			}

			log.Debug("Looking for Scalable CRs of CRD %q (api version %q, group %q, plural %q) in target namespaces.",
				crd.Name, crdVersion.Name, crd.Spec.Group, crd.Spec.Names.Plural)

			for _, ns := range namespaces {
				crs, err := dynamicClient.Resource(gvr).Namespace(ns).List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					log.Fatal("Error getting CRs of CRD %q in namespace %q, err: %v", crd.Name, ns, err)
				}

				if len(crs.Items) > 0 {
					scaleObjects = append(scaleObjects, getCrScaleObjects(crs.Items, crd)...)
				} else {
					log.Warn("No CRs of CRD %q found in the target namespaces.", crd.Name)
				}
			}
		}
	}

	return scaleObjects
}

// getCrScaleObjects Retrieves scaling information for custom resources
//
// This function iterates over a list of unstructured custom resources, querying
// the Kubernetes scaling API to obtain each resource's scale subresource. It
// constructs a group-resource schema from the CRD metadata and appends each
// retrieved ScaleObject to a slice. Errors during retrieval are logged fatally,
// ensuring only successfully fetched scales are returned.
func getCrScaleObjects(crs []unstructured.Unstructured, crd *apiextv1.CustomResourceDefinition) []ScaleObject {
	var scaleObjects []ScaleObject
	clients := clientsholder.GetClientsHolder()
	for _, cr := range crs {
		groupResourceSchema := schema.GroupResource{
			Group:    crd.Spec.Group,
			Resource: crd.Spec.Names.Plural,
		}

		name := cr.GetName()
		namespace := cr.GetNamespace()
		crScale, err := clients.ScalingClient.Scales(namespace).Get(context.TODO(), groupResourceSchema, name, metav1.GetOptions{})
		if err != nil {
			log.Fatal("Error while getting the scale of CR=%s (CRD=%s) in namespace %s: %v", name, crd.Name, namespace, err)
		}

		scaleObjects = append(scaleObjects, ScaleObject{Scale: crScale, GroupResourceSchema: groupResourceSchema})
	}
	return scaleObjects
}
