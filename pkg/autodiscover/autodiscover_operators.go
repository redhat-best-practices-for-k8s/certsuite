// Copyright (C) 2020-2021 Red Hat, Inc.
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

package autodiscover

import (
	"context"

	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	clientOlm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func findOperatorsByLabel(olmClient *clientOlm.Clientset, labels []configuration.Label, namespaces []configuration.Namespace) []olmv1Alpha.ClusterServiceVersion {
	csvs := []olmv1Alpha.ClusterServiceVersion{}
	for _, ns := range namespaces {
		logrus.Debugf("Searching CSVs in namespace %s", ns)
		for _, label := range labels {
			logrus.Debugf("Searching CSVs with label %+v", label)
			options := metav1.ListOptions{}
			label := buildLabelQuery(label)
			options.LabelSelector = label
			csvList, err := olmClient.OperatorsV1alpha1().ClusterServiceVersions(ns.Name).List(context.TODO(), options)
			if err != nil {
				logrus.Errorln("error when listing csvs in ns=", ns, " label=", label)
				continue
			}
			csvs = append(csvs, csvList.Items...)
		}
	}

	logrus.Infof("Found %d CSVs:", len(csvs))
	for i := range csvs {
		logrus.Infof(" CSV name: %s (ns: %s)", csvs[i].Name, csvs[i].Namespace)
	}

	return csvs
}
