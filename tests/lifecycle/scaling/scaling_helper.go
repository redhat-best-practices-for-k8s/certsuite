package scaling

import (
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiv1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	scalingv1 "k8s.io/api/autoscaling/v1"
)

// GetResourceHPA retrieves an HPA from a list based on identifiers.
//
// It iterates over the provided slice of HorizontalPodAutoscalers and
// returns the first one that matches the given namespace, name, and resource type.
// If no matching HPA is found, it returns nil.
func GetResourceHPA(hpaList []*scalingv1.HorizontalPodAutoscaler, name, namespace, kind string) *scalingv1.HorizontalPodAutoscaler {
	for _, hpa := range hpaList {
		if hpa.Spec.ScaleTargetRef.Kind == kind && hpa.Spec.ScaleTargetRef.Name == name && hpa.Namespace == namespace {
			return hpa
		}
	}
	return nil
}

// IsManaged checks whether a given deployment is managed by CertSuite.
//
// It takes the name of a deployment and a slice of managed statefulset
// definitions, and returns true if the deployment matches any entry in the
// list. The function does not modify its inputs.
func IsManaged(podSetName string, managedPodSet []configuration.ManagedDeploymentsStatefulsets) bool {
	for _, ps := range managedPodSet {
		if ps.Name == podSetName {
			return true
		}
	}
	return false
}

// CheckOwnerReference determines if any of the provided OwnerReferences match
// the given set of CRD filters.
//
// It iterates over the list of owner references and compares each reference's
// API version, kind, and name against the supplied filter rules. The function
// returns true if a match is found for at least one owner reference; otherwise,
// it returns false. This helper is used to validate ownership relationships
// during scaling tests.
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
