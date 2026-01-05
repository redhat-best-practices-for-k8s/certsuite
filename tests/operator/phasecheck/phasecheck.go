// Copyright (C) 2020-2026 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package phasecheck

import (
	"context"
	"time"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	timeout = 5 * time.Minute
)

func WaitOperatorReady(csv *v1alpha1.ClusterServiceVersion) bool {
	oc := clientsholder.GetClientsHolder()
	start := time.Now()
	for time.Since(start) < timeout {
		if isOperatorPhaseSucceeded(csv) {
			log.Debug("%s is ready", provider.CsvToString(csv))
			return true
		} else if isOperatorPhaseFailedOrUnknown(csv) {
			log.Debug("%s failed to be ready, status=%s", provider.CsvToString(csv), csv.Status.Phase)
			return false
		}

		// Operator is not ready, but we need to take into account that its pods
		// could have been deleted by some of the lifecycle test cases, so they
		// could be restarting. Let's give it some time before declaring it failed.
		log.Debug("Waiting for %s to be in Succeeded phase: %s", provider.CsvToString(csv), csv.Status.Phase)
		time.Sleep(time.Second)

		freshCsv, err := oc.OlmClient.OperatorsV1alpha1().ClusterServiceVersions(csv.Namespace).Get(context.TODO(), csv.Name, metav1.GetOptions{})
		if err != nil {
			log.Error("could not get csv %s, err: %v", provider.CsvToString(freshCsv), err)
			return false
		}

		// update old csv and check status again
		*csv = *freshCsv
	}
	if time.Since(start) > timeout {
		log.Error("timeout waiting for csv %s to be ready", provider.CsvToString(csv))
	}

	return false
}

func isOperatorPhaseSucceeded(csv *v1alpha1.ClusterServiceVersion) bool {
	log.Debug("Checking succeeded status phase for csv %s (ns %s). Phase: %v", csv.Name, csv.Namespace, csv.Status.Phase)
	return csv.Status.Phase == v1alpha1.CSVPhaseSucceeded
}

func isOperatorPhaseFailedOrUnknown(csv *v1alpha1.ClusterServiceVersion) bool {
	log.Debug("Checking failed status phase for csv %s (ns %s). Phase: %v", csv.Name, csv.Namespace, csv.Status.Phase)
	return csv.Status.Phase == v1alpha1.CSVPhaseFailed ||
		csv.Status.Phase == v1alpha1.CSVPhaseUnknown
}
