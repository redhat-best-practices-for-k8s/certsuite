package autodiscover

import (
	"context"

	configv1 "github.com/openshift/api/config/v1"
	clientconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// findClusterOperators Retrieves all ClusterOperator resources from the cluster
//
// The function calls the client to list ClusterOperator objects, handling
// errors that may occur during the request. If the API returns a not‑found
// error, it logs a debug message and returns nil without error. On success, it
// returns a slice of the retrieved items.
func findClusterOperators(client clientconfigv1.ClusterOperatorInterface) ([]configv1.ClusterOperator, error) {
	clusterOperators, err := client.List(context.TODO(), metav1.ListOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, err
	}

	if k8serrors.IsNotFound(err) {
		log.Debug("ClusterOperator CR not found in the cluster")
		return nil, nil
	}

	return clusterOperators.Items, nil
}
