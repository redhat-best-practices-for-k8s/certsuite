package scaling

import (
	"strings"

	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	apiv1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	scalingv1 "k8s.io/api/autoscaling/v1"
)

func GetResourceHPA(hpaList []*scalingv1.HorizontalPodAutoscaler, name, namespace, kind string) *scalingv1.HorizontalPodAutoscaler {
	for _, hpa := range hpaList {
		if hpa.Spec.ScaleTargetRef.Kind == kind && hpa.Spec.ScaleTargetRef.Name == name && hpa.Namespace == namespace {
			return hpa
		}
	}
	return nil
}
func IsManaged(podSetName string, managedPodSet []configuration.ManagedDeploymentsStatefulsets) bool {
	for _, ps := range managedPodSet {
		if ps.Name == podSetName {
			return true
		}
	}
	return false
}
func CheckOwnerReference(ownerReference []apiv1.OwnerReference, scalable []configuration.CrdFilter) bool {
	for _, owner := range ownerReference {
		for _, crd := range scalable {
			if strings.Contains(crd.NameSuffix, owner.Name) {
				return crd.Scalable
			}
		}
	}

	return false
}
