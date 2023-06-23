// Copyright (C) 2022 Red Hat, Inc.
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

type RoleRule struct {
	Resource RoleResource
	Verb     string
}
type RoleResource struct {
	Group, Name string
}

type CrdResource struct {
	Group, SingularName, PluralName string
	ShortNames                      []string
}

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

// Checks the resource name in the role against plural name
func isResourceInRoleRule(crd CrdResource, roleRule RoleRule) bool {
	// remove subresources to keep only resource (plural) name
	ruleResourcePluralName := strings.Split(roleRule.Resource.Name, "/")[0]

	return crd.Group == roleRule.Resource.Group && crd.PluralName == ruleResourcePluralName
}

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
