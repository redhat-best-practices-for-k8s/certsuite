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

// ScaleObject represents a Kubernetes resource that supports scaling operations.
//
// It contains the GroupResource schema identifying the resource type and a pointer to its Scale subresource,
// which holds the current size and desired replicas.
// This structure is used by autodiscover functions to gather scaling information from custom resources.
type ScaleObject struct {
	Scale               *scalingv1.Scale
	GroupResourceSchema schema.GroupResource
}

// GetScaleCrUnderTest returns a list of scale objects that are under test.
//
// It examines the provided CustomResourceDefinition objects and
// extracts any scale-related resources defined within them.
// The function uses a client holder to query the cluster,
// logs progress with Info, Debug, Warn, and Fatal as needed,
// and aggregates results into a slice of ScaleObject.
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

// getCrScaleObjects returns ScaleObject instances for a set of custom resources.
//
// It takes a slice of unstructured objects representing custom resource
// instances and the corresponding CustomResourceDefinition.
// For each object it retrieves the scale subresource via the Kubernetes API,
// constructs a ScaleObject with namespace, name, and scale information,
// and appends it to the result slice. If any error occurs during retrieval,
// the function logs a fatal message and stops execution. The returned
// slice contains all successfully retrieved ScaleObjects.
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
