package autodiscover

import (
	"context"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/ocplite"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// ClusterOperatorGVR references OpenShift ClusterOperator resources
var ClusterOperatorGVR = schema.GroupVersionResource{
	Group:    "config.openshift.io",
	Version:  "v1",
	Resource: "clusteroperators",
}

func findClusterOperators(client dynamic.Interface) ([]ocplite.ClusterOperator, error) {
	list, err := client.Resource(ClusterOperatorGVR).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Debug("ClusterOperator CR not found in the cluster")
			return nil, nil
		}
		return nil, err
	}

	var result []ocplite.ClusterOperator
	for i := range list.Items {
		u := list.Items[i]
		co := ocplite.ClusterOperator{}
		co.Name = u.GetName()
		if status, ok := u.Object["status"].(map[string]interface{}); ok {
			if conds, ok := status["conditions"].([]interface{}); ok {
				for _, c := range conds {
					if m, ok := c.(map[string]interface{}); ok {
						co.Status.Conditions = append(co.Status.Conditions, ocplite.ClusterOperatorStatusCondition{
							Type:   toString(m["type"]),
							Status: toString(m["status"]),
						})
					}
				}
			}
			if vers, ok := status["versions"].([]interface{}); ok {
				for _, v := range vers {
					if m, ok := v.(map[string]interface{}); ok {
						co.Status.Versions = append(co.Status.Versions, ocplite.OperandVersion{
							Name:    toString(m["name"]),
							Version: toString(m["version"]),
						})
					}
				}
			}
		}
		result = append(result, co)
	}
	return result, nil
}

func toString(v interface{}) string {
	s, _ := v.(string)
	return s
}
