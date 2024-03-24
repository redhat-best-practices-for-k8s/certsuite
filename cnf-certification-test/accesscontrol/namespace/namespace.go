// Copyright (C) 2020-2024 Red Hat, Inc.
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

	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TestCrsNamespaces finds the list of the input CRDs (crds parameter) instances (CRs) and verify that they are only in namespaces provided as input.
// Returns :
//   - map[string]map[string][]string : The list of CRs not belonging to the namespaces passed as input is returned as invalid.
//   - error : if exist error.
func TestCrsNamespaces(crds []*apiextv1.CustomResourceDefinition, configNamespaces []string, logger *log.Logger) (invalidCrs map[string]map[string][]string, err error) {
	// Initialize the top level map
	invalidCrs = make(map[string]map[string][]string)
	for _, crd := range crds {
		crNamespaces, err := getCrsPerNamespaces(crd)
		if err != nil {
			return invalidCrs, fmt.Errorf("failed to get CRs for CRD %s - Error: %v", crd.Name, err)
		}
		for namespace, crNames := range crNamespaces {
			if !stringhelper.StringInSlice(configNamespaces, namespace, false) {
				logger.Error("CRD: %q (kind:%q/ plural:%q) has CRs %v deployed in namespace %q not in configured namespaces %v",
					crd.Name, crd.Spec.Names.Kind, crd.Spec.Names.Plural, crNames, namespace, configNamespaces)
				// Initialize this map dimension before use
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
// Returns :
//   - map[string][]string : a map indexed by namespace and data is a list of CR names.
//   - error : if exist error.
func getCrsPerNamespaces(aCrd *apiextv1.CustomResourceDefinition) (crdNamespaces map[string][]string, err error) {
	oc := clientsholder.GetClientsHolder()
	for _, version := range aCrd.Spec.Versions {
		gvr := schema.GroupVersionResource{
			Group:    aCrd.Spec.Group,
			Version:  version.Name,
			Resource: aCrd.Spec.Names.Plural,
		}
		log.Debug("Looking for CRs from CRD: %s api version:%s group:%s plural:%s", aCrd.Name, version.Name, aCrd.Spec.Group, aCrd.Spec.Names.Plural)
		crs, err := oc.DynamicClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error("error getting %s: %v\n", aCrd.Name, err)
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
	}
	return crdNamespaces, nil
}

// GetInvalidCRDsNum returns the number of invalid CRs in the map.
// Return:
//   - int : number of invalid CRs in the map.
func GetInvalidCRsNum(invalidCrs map[string]map[string][]string, logger *log.Logger) int {
	var invalidCrsNum int
	for crdName, namespaces := range invalidCrs {
		for namespace, crNames := range namespaces {
			for _, crName := range crNames {
				logger.Error("crName=%q namespace=%q is invalid (crd=%q)", crName, namespace, crdName)
				invalidCrsNum++
			}
		}
	}
	return invalidCrsNum
}
