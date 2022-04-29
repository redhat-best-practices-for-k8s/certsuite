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
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"

	"context"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	clientsholder "github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getClusterCrdNames returns a list of crd names found in the cluster.
func getClusterCrdNames() (crdList []*apiextv1.CustomResourceDefinition, err error) {
	oc := clientsholder.GetClientsHolder()
	crds, err := oc.APIExtClient.ApiextensionsV1().CustomResourceDefinitions().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorln("error when listing crds")
		return crdList, err
	}
	for idx := range crds.Items {
		crdList = append(crdList, &crds.Items[idx])
	}
	return crdList, nil
}

// FindTestCrdNames gets a list of CRD names based on configured groups.
func FindTestCrdNames(crdFilters []configuration.CrdFilter) (targetCrds []*apiextv1.CustomResourceDefinition) {
	clusterCrds, err := getClusterCrdNames()
	if err != nil {
		logrus.Errorf("Unable to get cluster CRD.")
		return []*apiextv1.CustomResourceDefinition{}
	}
	for _, crd := range clusterCrds {
		for _, crdFilter := range crdFilters {
			if strings.HasSuffix(crd.Name, crdFilter.NameSuffix) {
				targetCrds = append(targetCrds, crd)
				break
			}
		}
	}
	return targetCrds
}
