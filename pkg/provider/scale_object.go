package provider

import (
	"fmt"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/autodiscover"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/scale"

	scalingv1 "k8s.io/api/autoscaling/v1"
)

// CrScale represents a custom resource scaling object.
//
// It embeds the scalingv1.Scale type and provides helper methods to
// determine readiness and produce a string representation of the
// underlying object. The embedded fields expose all properties of
// the original Scale resource, allowing direct access to its spec,
// status, metadata, etc. Use IsScaleObjectReady to check if the scale
// has reached a ready state, and ToString for debugging or logging.
type CrScale struct {
	*scalingv1.Scale
}

// IsScaleObjectReady reports whether the scale object is ready.
//
// It checks internal state of CrScale and logs status via Info.
// Returns true if the scale object's readiness conditions are satisfied, otherwise false.
func (crScale CrScale) IsScaleObjectReady() bool {
	replicas := (crScale.Spec.Replicas)
	log.Info("replicas is %d status replica is %d", replicas, crScale.Status.Replicas)
	return crScale.Status.Replicas == replicas
}

// ToString returns a human-readable representation of the CrScale.
//
// It formats the fields of the CrScale struct into a string using fmt.Sprintf
// and returns that string. The output can be used for logging or debugging to
// inspect the current scale configuration.
func (crScale CrScale) ToString() string {
	return fmt.Sprintf("cr: %s ns: %s",
		crScale.Name,
		crScale.Namespace,
	)
}

// GetUpdatedCrObject retrieves the current scale object for a custom resource and returns an updated CrScale representation.
//
// It accepts a ScalesGetter to query the cluster, the namespace and name of the target CR,
// and the GroupResource identifying its kind. The function first locates the CR by calling
// FindCrObjectByNameByNamespace, then constructs a *CrScale based on the current state.
// If any step fails, an error is returned; otherwise the updated CrScale pointer is returned.
func GetUpdatedCrObject(sg scale.ScalesGetter, namespace, name string, groupResourceSchema schema.GroupResource) (*CrScale, error) {
	result, err := autodiscover.FindCrObjectByNameByNamespace(sg, namespace, name, groupResourceSchema)
	return &CrScale{
		result,
	}, err
}
