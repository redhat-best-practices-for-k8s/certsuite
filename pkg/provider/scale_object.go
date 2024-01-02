package provider

import (
	"fmt"

	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/autodiscover"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/scale"

	scalingv1 "k8s.io/api/autoscaling/v1"
)

type CrScale struct {
	*scalingv1.Scale
}

func (crScale CrScale) IsScaleObjectReady() bool {
	replicas := (crScale.Spec.Replicas)
	log.Info("replicas is %d status replica is %d", replicas, crScale.Status.Replicas)
	return crScale.Status.Replicas == replicas
}

func (crScale CrScale) ToString() string {
	return fmt.Sprintf("cr: %s ns: %s",
		crScale.Name,
		crScale.Namespace,
	)
}
func GetUpdatedCrObject(sg scale.ScalesGetter, namespace, name string, groupResourceSchema schema.GroupResource) (*CrScale, error) {
	result, err := autodiscover.FindCrObjectByNameByNamespace(sg, namespace, name, groupResourceSchema)
	return &CrScale{
		result,
	}, err
}
