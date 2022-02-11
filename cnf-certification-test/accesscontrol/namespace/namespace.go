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

package namespace

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/ocpclient"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TestCrsNamespaces finds the list of the input CRDs (crds parameter) instances (CRs) and verify that they are only in namespaces provided as input.
// The list of CRs not belonging to the namespaces passed as input is returned as invalid
func TestCrsNamespaces(crds []*apiextv1beta1.CustomResourceDefinition, configNamespaces []string) (invalidCrs map[string]map[string][]string, err error) {
	// Initialize the top level map
	if invalidCrs == nil {
		invalidCrs = make(map[string]map[string][]string)
	}
	for _, crd := range crds {
		crNamespaces, err := getCrsPerNamespaces(crd)
		if err != nil {
			return invalidCrs, fmt.Errorf("failed to get CRs for CRD %s - Error: %v", crd.Name, err)
		}
		for namespace, crNames := range crNamespaces {
			if !stringhelper.StringInSlice(configNamespaces, namespace, false) {
				logrus.Tracef("CRD: %s (kind:%s/ plural:%s) has CRs %v deployed in namespace (%s) not in configured namespaces %v",
					crd.Name, crd.Spec.Names.Kind, crd.Spec.Names.Plural, crNames, namespace, configNamespaces)
				// Initialize this map dimenension before use
				if invalidCrs[crd.Name] == nil {
					invalidCrs[crd.Name] = make(map[string][]string)
				}
				invalidCrs[crd.Name][namespace] = append(invalidCrs[crd.Name][namespace], crNames...)
			}
		}
	}
	return invalidCrs, nil
}

// getCrsPerNamespaces gets the list of CRs instantiated in the cluster per namespace.
// Returns a map indexed by namespace and data is a list of CR names
func getCrsPerNamespaces(aCrd *apiextv1beta1.CustomResourceDefinition) (crdNamespaces map[string][]string, err error) {
	oc := ocpclient.NewOcpClient()
	gvr := schema.GroupVersionResource{
		Group:    aCrd.Spec.Group,
		Version:  aCrd.Spec.Version,
		Resource: aCrd.Spec.Names.Plural,
	}
	crs, err := oc.DynamicClient.Resource(gvr).List(context.Background(), v1.ListOptions{})
	if err != nil {
		logrus.Errorf("error getting %s: %v\n", aCrd.Name, err)
		return crdNamespaces, err
	}
	crdNamespaces = make(map[string][]string)
	for _, cr := range crs.Items {
		name := cr.Object["metadata"].(map[string]interface{})["name"]
		namespace := cr.Object["metadata"].(map[string]interface{})["namespace"]
		var namespaceStr, nameStr string
		if namespace == nil {
			namespaceStr = ""
		} else {
			namespaceStr = fmt.Sprintf("%s", namespace)
		}
		if name == nil {
			nameStr = ""
		} else {
			nameStr = fmt.Sprintf("%s", name)
		}
		crdNamespaces[namespaceStr] = append(crdNamespaces[namespaceStr], nameStr)
	}
	return crdNamespaces, nil
}
