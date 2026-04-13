// Copyright (C) 2020-2026 Red Hat, Inc.
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

package checksadapter

import (
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/checks"
)

// CheckInfoToClaimIdentifier converts a checks library CheckInfo to a claim.Identifier.
func CheckInfoToClaimIdentifier(info *checks.CheckInfo) claim.Identifier {
	return claim.Identifier{
		Id:    info.Name,
		Suite: info.Category,
		Tags:  strings.Join(info.Tags, ","),
	}
}

// CheckInfoToTestCaseDescription converts a checks library CheckInfo to a claim.TestCaseDescription.
func CheckInfoToTestCaseDescription(info *checks.CheckInfo) claim.TestCaseDescription {
	id := CheckInfoToClaimIdentifier(info)
	return claim.TestCaseDescription{
		Identifier:             id,
		Description:            info.Description,
		Remediation:            info.Remediation,
		BestPracticeReference:  info.BestPracticeReference,
		ExceptionProcess:       info.ExceptionProcess,
		Tags:                   strings.Join(info.Tags, ","),
		Qe:                     info.Qe,
		CategoryClassification: info.CategoryClassification,
	}
}

// GetCheckIDAndLabels looks up a check by name in the checks library registry and returns
// the check ID and label tags. If the check is not found, it returns the name itself as
// both the ID and the sole tag.
func GetCheckIDAndLabels(name string) (testID string, tags []string) {
	info, ok := checks.ByName(name)
	if !ok {
		return name, []string{name}
	}
	tags = make([]string, len(info.Tags))
	copy(tags, info.Tags)
	tags = append(tags, info.Name, info.Category)
	return info.Name, tags
}
