// Copyright (C) 2020-2022 Red Hat, Inc.
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
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	timeout = 5 * time.Minute
)

func WaitOperatorReady(csv *v1alpha1.ClusterServiceVersion) v1alpha1.ClusterServiceVersionPhase {
	oc := clientsholder.GetClientsHolder()
	isReady := false
	start := time.Now()
	var freshCsv *v1alpha1.ClusterServiceVersion
	for !isReady && time.Since(start) < timeout {
		var err error
		freshCsv, err = oc.OlmClient.OperatorsV1alpha1().ClusterServiceVersions(csv.Namespace).Get(context.TODO(), csv.Name, metav1.GetOptions{})
		if err != nil {
			logrus.Errorf("error getting %s", provider.CsvToString(freshCsv))
		}
		isReady = isOperatorSucceeded(freshCsv)
		logrus.Debugf("Waiting for %s to be in Succeeded phase: %s", provider.CsvToString(freshCsv), freshCsv.Status.Phase)
		time.Sleep(time.Second)
	}
	if time.Since(start) > timeout {
		logrus.Fatalf("Timeout waiting for %s to be ready", provider.CsvToString(csv))
	}
	if isReady {
		logrus.Infof("%s is ready", provider.CsvToString(csv))
	}
	return freshCsv.Status.Phase
}

func isOperatorSucceeded(csv *v1alpha1.ClusterServiceVersion) (isReady bool) {
	return csv.Status.Phase == v1alpha1.CSVPhaseSucceeded
}
