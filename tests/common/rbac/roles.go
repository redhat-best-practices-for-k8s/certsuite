// Copyright (C) 2022-2024 Red Hat, Inc.
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

package rbac

import (
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// RoleRule Represents a single permission within a role
//
// This structure pairs an API resource, identified by its group and name, with
// a verb that defines the action allowed on that resource. It is used
// throughout the package to flatten complex Role objects into individual rules
// for easier comparison and filtering. Each instance encapsulates one specific
// permission granted by a Kubernetes RBAC Role.
type RoleRule struct {
	Resource RoleResource
	Verb     string
}

// RoleResource Represents an RBAC resource with its API group and kind
//
// This structure holds the API group and the resource name used in Role or
// ClusterRole rules. It allows code to identify which Kubernetes resource a
// rule applies to, such as "apps" for deployments or "core" for pods. The
// fields are simple strings that can be populated from YAML manifests or
// constructed programmatically.
type RoleResource struct {
	Group, Name string
}

// CrdResource Represents a custom resource definition's identity within RBAC
//
// This struct holds the group and names of a CRD, including singular, plural,
// and short forms. It is used to match resources against role rules when
// determining permissions for custom resources.
type CrdResource struct {
	Group, SingularName, PluralName string
	ShortNames                      []string
}

// GetCrdResources Converts CRD definitions into a slice of resource descriptors
//
// This function iterates over each CustomResourceDefinition provided,
// extracting the group, singular name, plural name, and short names from its
// specification. For every CRD it creates a CrdResource struct populated with
// these fields and appends it to a list. The resulting slice is returned for
// use in permission checks or reporting.
func GetCrdResources(crds []*apiextv1.CustomResourceDefinition) (resourceList []CrdResource) {
	for _, crd := range crds {
		var aResource CrdResource
		aResource.Group = crd.Spec.Group
		aResource.SingularName = crd.Spec.Names.Singular
		aResource.PluralName = crd.Spec.Names.Plural
		aResource.ShortNames = crd.Spec.Names.ShortNames
		resourceList = append(resourceList, aResource)
	}
	return resourceList
}

// GetAllRules Collects every rule from a role into individual entries
//
// The function iterates over each rule in the supplied role, expanding its API
// groups, resources, and verbs into separate RoleRule objects. Each combination
// of group, resource name, and verb is appended to a slice, which is returned.
// The resulting list can be used for detailed policy analysis or filtering.
func GetAllRules(aRole *rbacv1.Role) (ruleList []RoleRule) {
	for _, aRule := range aRole.Rules {
		for _, aGroup := range aRule.APIGroups {
			for _, aResource := range aRule.Resources {
				for _, aVerb := range aRule.Verbs {
					var aRoleRule RoleRule
					aRoleRule.Resource.Group = aGroup
					aRoleRule.Resource.Name = aResource
					aRoleRule.Verb = aVerb
					ruleList = append(ruleList, aRoleRule)
				}
			}
		}
	}
	return ruleList
}

// isResourceInRoleRule Determines if a CRD matches a role rule by group and resource name
//
// The function receives a custom resource definition and a role rule, extracts
// the base resource name from the rule by removing any subresource part, and
// then compares the API group and plural name of the CRD to those of the rule.
// If both match exactly, it returns true; otherwise it returns false.
func isResourceInRoleRule(crd CrdResource, roleRule RoleRule) bool {
	// remove subresources to keep only resource (plural) name
	ruleResourcePluralName := strings.Split(roleRule.Resource.Name, "/")[0]

	return crd.Group == roleRule.Resource.Group && crd.PluralName == ruleResourcePluralName
}

// FilterRulesNonMatchingResources Separates role rules into those that match CRD resources
//
// This routine examines each rule against a list of CRD resources, collecting
// any rule whose resource group and plural name align with a CRD. Rules that do
// not find a match are returned separately by computing the difference from the
// original list. The output consists of two slices: one for matching rules and
// one for non‑matching ones.
func FilterRulesNonMatchingResources(ruleList []RoleRule, resourceList []CrdResource) (matching, nonMatching []RoleRule) {
	for _, aRule := range ruleList {
		for _, aResource := range resourceList {
			if isResourceInRoleRule(aResource, aRule) {
				matching = append(matching, aRule)
			}
		}
	}
	nonMatching = SliceDifference(ruleList, matching)
	return matching, nonMatching
}

// SliceDifference identifies RoleRule entries present in one slice but absent from another
//
// The function takes two slices of RoleRule values and returns a new slice
// containing elements that exist in the first slice but not in the second. It
// swaps the slices if the second is longer to reduce comparisons, then iterates
// through each element of the larger slice, checking for equality against all
// elements of the other slice. Matching items are omitted; non‑matching ones
// are appended to the result, which is returned.
func SliceDifference(s1, s2 []RoleRule) (diff []RoleRule) {
	var temp []RoleRule
	if len(s2) > len(s1) {
		temp = s1
		s1 = s2
		s2 = temp
	}
	for _, v1 := range s1 {
		missing := true
		for _, v2 := range s2 {
			if v1 == v2 {
				missing = false
				break
			}
		}
		if missing {
			diff = append(diff, v1)
		}
	}
	return diff
}
