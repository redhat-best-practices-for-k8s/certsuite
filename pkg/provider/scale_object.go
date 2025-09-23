package provider

import (
	"fmt"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/autodiscover"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/scale"

	scalingv1 "k8s.io/api/autoscaling/v1"
)

// CrScale Wraps a scale object with status tracking
//
// This type extends the base scaling API object by embedding its fields and
// providing helper methods to inspect readiness and generate a concise string
// representation. The embedded struct contains both specification and current
// status, allowing direct access to replica counts and other properties.
type CrScale struct {
	*scalingv1.Scale
}

// CrScale.IsScaleObjectReady Checks whether the scale object has reached the desired replica count
//
// The function compares the desired number of replicas defined in the
// specification with the current replica count reported in the status. It logs
// both values for debugging purposes. The result is a boolean indicating if the
// actual count matches the requested count.
func (crScale CrScale) IsScaleObjectReady() bool {
	replicas := (crScale.Spec.Replicas)
	log.Info("replicas is %d status replica is %d", replicas, crScale.Status.Replicas)
	return crScale.Status.Replicas == replicas
}

// CrScale.ToString Formats the CrScale object into a readable string
//
// This method returns a single string that contains both the name and namespace
// of the CrScale instance. It uses formatting to combine the two fields with
// clear labels, producing output like "cr: <name> ns: <namespace>". The
// function requires no arguments and yields a straightforward textual
// representation for logging or display purposes.
func (crScale CrScale) ToString() string {
	return fmt.Sprintf("cr: %s ns: %s",
		crScale.Name,
		crScale.Namespace,
	)
}

// GetUpdatedCrObject Retrieves a scaled custom resource and wraps it for further use
//
// This function calls the discovery helper to fetch a custom resource by name
// within a namespace, using the provided scale getter and group-resource
// schema. It then packages the returned scaling object into a CrScale
// structure, returning that along with any error encountered during retrieval.
func GetUpdatedCrObject(sg scale.ScalesGetter, namespace, name string, groupResourceSchema schema.GroupResource) (*CrScale, error) {
	result, err := autodiscover.FindCrObjectByNameByNamespace(sg, namespace, name, groupResourceSchema)
	return &CrScale{
		result,
	}, err
}
