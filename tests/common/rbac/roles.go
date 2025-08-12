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

// RoleRule represents a single rule within a Kubernetes role.
//
// It contains the resource that the rule applies to and the verb that specifies the action allowed on that resource.
// The Resource field is of type RoleResource, which defines group, version, and plural name of the target resource.
// The Verb field is a string such as "get", "list", "create", or "*" for all verbs.
type RoleRule struct {
	Resource RoleResource
	Verb     string
}

// RoleResource represents a Kubernetes RBAC resource.
//
// It contains the group and name of a role or clusterrole that can be referenced
// when defining permissions for subjects within the system. The Group field
// specifies the API group (e.g., rbac.authorization.k8s.io) and the Name field
// identifies the specific role. This struct is used to construct bindings
// and queries against RBAC objects.
type RoleResource struct {
	Group, Name string
}

// CrdResource represents a custom resource definition's basic identity.
//
// CrdResource holds the identifying fields of a CRD.
// It includes the API group, plural and singular names,
// as well as any short names that can be used to refer
// to the resource in Kubernetes manifests and commands.
type CrdResource struct {
	Group, SingularName, PluralName string
	ShortNames                      []string
}

// GetCrdResources converts a list of apiextv1.CustomResourceDefinition structs into CrdResource slices.
//
// It takes a slice of CustomResourceDefinition pointers, iterates over them,
// and returns a slice of CrdResource objects representing the resources defined by each CRD.
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

// GetAllRules retrieves all rules defined by a role.
//
// It accepts a pointer to an rbacv1.Role and returns a slice of RoleRule objects
// representing each rule in the role's policy.
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

// isResourceInRoleRule checks whether a CRD resource matches a role rule by comparing its group and plural name.
//
// It takes a CrdResource and a RoleRule, compares the group and plural fields of both,
// and returns true if they match or false otherwise. The function does not modify its inputs.
func isResourceInRoleRule(crd CrdResource, roleRule RoleRule) bool {
	// remove subresources to keep only resource (plural) name
	ruleResourcePluralName := strings.Split(roleRule.Resource.Name, "/")[0]

	return crd.Group == roleRule.Resource.Group && crd.PluralName == ruleResourcePluralName
}

// FilterRulesNonMatchingResources filters RoleRules based on whether they match any CrdResource in the resourceList.
//
// It receives a slice of RoleRule and a slice of CrdResource.
// The function examines each rule to determine if it matches at least one
// CRD resource using isResourceInRoleRule. Rules that do not match any
// resource are collected into the result slice, which is returned.
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

// SliceDifference returns the RoleRule elements that exist in the first slice but not in the second.
//
// It compares two slices of RoleRule and builds a new slice containing only those
// elements from s1 that are absent in s2. The result is a slice of RoleRule
// representing the difference between the inputs.
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
