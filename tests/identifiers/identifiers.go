// Copyright (C) 2021-2026 Red Hat, Inc.
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

package identifiers

import (
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
)

const (
	TagCommon    = "common"
	TagExtended  = "extended"
	TagTelco     = "telco"
	TagFarEdge   = "faredge"
	TagPreflight = "preflight"

	FarEdge  = "FarEdge"
	Telco    = "Telco"
	NonTelco = "NonTelco"
	Extended = "Extended"

	Optional  = "Optional"
	Mandatory = "Mandatory"

	NoDocumentedProcess = `There is no documented exception process for this.`
	NoDocLink           = "No Doc Link"
)

type CatalogEntry struct {
	ID                     string
	Suite                  string
	Description            string
	Remediation            string
	Exception              string
	Reference              string
	QE                     bool
	CategoryClassification map[string]string
	Tags                   []string
}

func AddCatalogEntry(entry *CatalogEntry) (aID claim.Identifier) {
	if strings.TrimSpace(entry.Exception) == "" {
		entry.Exception = NoDocumentedProcess
	}
	if strings.TrimSpace(entry.Reference) == "" {
		entry.Reference = "No Reference Document Specified"
	}
	if len(entry.Tags) == 0 {
		entry.Tags = append(entry.Tags, TagCommon)
	}

	tcDescription, aID := claim.BuildTestCaseDescription(entry.ID, entry.Suite, entry.Description, entry.Remediation, entry.Exception, entry.Reference, entry.QE, entry.CategoryClassification, entry.Tags...)
	Catalog[aID] = tcDescription

	return aID
}

// GetTestIDAndLabels transforms a claim.Identifier into a test ID and label set.
// Used by preflight tests which register dynamically at runtime.
func GetTestIDAndLabels(identifier claim.Identifier) (testID string, tags []string) {
	tags = strings.Split(identifier.Tags, ",")
	tags = append(tags, identifier.Id, identifier.Suite)
	TestIDToClaimID[identifier.Id] = identifier
	return identifier.Id, tags
}

var (
	TestIDToClaimID = map[string]claim.Identifier{}
	Catalog         = map[claim.Identifier]claim.TestCaseDescription{}
)
