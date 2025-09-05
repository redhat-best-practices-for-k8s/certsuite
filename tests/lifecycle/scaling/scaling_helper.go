package scaling

import (
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiv1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	scalingv1 "k8s.io/api/autoscaling/v1"
)

// GetResourceHPA Finds an HPA matching a resource name, namespace, and kind
//
// The function iterates over a list of HorizontalPodAutoscaler objects,
// checking each one's scale target reference for the specified kind, name, and
// namespace. If a match is found, that HPA is returned; otherwise the function
// returns nil to indicate no suitable HPA exists.
func GetResourceHPA(hpaList []*scalingv1.HorizontalPodAutoscaler, name, namespace, kind string) *scalingv1.HorizontalPodAutoscaler {
	for _, hpa := range hpaList {
		if hpa.Spec.ScaleTargetRef.Kind == kind && hpa.Spec.ScaleTargetRef.Name == name && hpa.Namespace == namespace {
			return hpa
		}
	}
	return nil
}

// IsManaged Checks if a deployment or stateful set is listed as managed
//
// The function iterates over the provided slice of managed pod sets, comparing
// each name with the supplied pod set name. If a match is found it returns
// true, indicating that the object should be considered under management for
// scaling tests. Otherwise, it returns false.
func IsManaged(podSetName string, managedPodSet []configuration.ManagedDeploymentsStatefulsets) bool {
	for _, ps := range managedPodSet {
		if ps.Name == podSetName {
			return true
		}
	}
	return false
}

// CheckOwnerReference Determines if owner references match scalable CRD filters
//
// The function iterates over each OwnerReference of a resource, comparing its
// kind to the kinds defined in available CustomResourceDefinitions. For
// matching kinds it checks whether the CRD name ends with any configured
// suffix; if so, it returns the corresponding scalability flag from that
// filter. If no match is found, it returns false.
func CheckOwnerReference(ownerReference []apiv1.OwnerReference, crdFilter []configuration.CrdFilter, crds []*apiextv1.CustomResourceDefinition) bool {
	for _, owner := range ownerReference {
		for _, aCrd := range crds {
			if aCrd.Spec.Names.Kind == owner.Kind {
				for _, crdF := range crdFilter {
					if strings.HasSuffix(aCrd.Name, crdF.NameSuffix) {
						return crdF.Scalable
					}
				}
			}
		}
	}
	return false
}
