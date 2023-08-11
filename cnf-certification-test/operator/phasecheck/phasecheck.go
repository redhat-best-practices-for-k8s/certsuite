// Copyright (C) 2020-2023 Red Hat, Inc.
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
	"fmt"
	"time"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	timeout = 5 * time.Minute
)

func waitOperatorReady(csv *v1alpha1.ClusterServiceVersion) bool {
	oc := clientsholder.GetClientsHolder()
	start := time.Now()
	for time.Since(start) < timeout {
		freshCsv, err := oc.OlmClient.OperatorsV1alpha1().ClusterServiceVersions(csv.Namespace).Get(context.TODO(), csv.Name, metav1.GetOptions{})
		if err != nil {
			errMsg := fmt.Sprintf("could not get csv %s, err: %v", provider.CsvToString(freshCsv), err)
			logrus.Errorf(errMsg)
			tnf.ClaimFilePrintf(errMsg)
			return false
		}

		// update old csv and check status again
		*csv = *freshCsv
		if csv.Status.Phase == v1alpha1.CSVPhaseSucceeded {
			tnf.ClaimFilePrintf("%s is ready", provider.CsvToString(csv))
			return true
		}

		tnf.ClaimFilePrintf("Waiting for %s to be in Succeeded phase: %s", provider.CsvToString(freshCsv), csv.Status.Phase)
		time.Sleep(time.Second)
	}
	if time.Since(start) > timeout {
		errMsg := fmt.Sprintf("timeout waiting for csv %s to be ready", provider.CsvToString(csv))
		logrus.Errorf(errMsg)
		tnf.ClaimFilePrintf(errMsg)
	}

	return false
}

func IsOperatorPhaseSucceeded(csv *v1alpha1.ClusterServiceVersion) bool {
	logrus.Tracef("Checking status phase for csv %s (ns %s). Phase: %v", csv.Name, csv.Namespace, csv.Status.Phase)
	switch csv.Status.Phase {
	case v1alpha1.CSVPhaseSucceeded:
		return true
	case v1alpha1.CSVPhaseFailed:
		fallthrough
	case v1alpha1.CSVPhaseUnknown:
		return false
	// Operator is not ready, but we need to take into account that its pods
	// could have been deleted by some of the lifecycle test cases, so they
	// could be restarting. Let's give it some time before declaring it failed.
	case v1alpha1.CSVPhasePending:
		fallthrough
	case v1alpha1.CSVPhaseInstalling:
		fallthrough
	case v1alpha1.CSVPhaseInstallReady:
		fallthrough
	case v1alpha1.CSVPhaseDeleting:
		fallthrough
	case v1alpha1.CSVPhaseReplacing:
		return waitOperatorReady(csv)
	default:
		return false
	}
}
