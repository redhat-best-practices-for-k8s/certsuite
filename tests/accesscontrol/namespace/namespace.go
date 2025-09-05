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

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/stringhelper"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TestCrsNamespaces identifies custom resources outside allowed namespaces
//
// The function examines each provided CRD, gathers all its instances across the
// cluster, and checks whether their namespaces match a given list of permitted
// namespaces. For any instance found in an unauthorized namespace, it records
// the CRD name, namespace, and resource names in a nested map. The resulting
// map is returned along with any error that occurred during retrieval;
// otherwise nil indicates success.
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

// getCrsPerNamespaces Retrieves custom resources per namespace
//
// This function queries the Kubernetes cluster for all instances of a given
// CustomResourceDefinition across its versions, organizing them into a map
// keyed by namespace with lists of resource names as values. It uses a dynamic
// client from a shared holder to perform list operations and logs debug
// information during the search. If any listing operation fails, an error is
// returned along with a nil or partially filled map.
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
			log.Error("Error getting %s: %v\n", aCrd.Name, err)
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

// GetInvalidCRsNum Counts the number of custom resources that are not in allowed namespaces
//
// The function walks through a nested map where each CRD maps to namespaces and
// then to lists of CR names, logging an error for every invalid entry it finds.
// It tallies these occurrences into an integer which is returned as the total
// count of invalid custom resources.
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
