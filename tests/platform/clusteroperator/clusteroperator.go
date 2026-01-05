// Copyright (C) 2025-2026 Red Hat, Inc.
package clusteroperator

import (
	configv1 "github.com/openshift/api/config/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

func IsClusterOperatorAvailable(co *configv1.ClusterOperator) bool {
	// Loop through the conditions, looking for the 'Available' state.
	for _, condition := range co.Status.Conditions {
		if condition.Type == configv1.OperatorAvailable {
			log.Info("ClusterOperator %q is in an 'Available' state", co.Name)
			return true
		}
	}

	log.Info("ClusterOperator %q is not in an 'Available' state", co.Name)
	return false
}
