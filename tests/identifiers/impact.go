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

const (
	PreflightAllImageRefsInRelatedImagesImpact                  = `Missing or incorrect image references in related images can cause deployment failures and broken operator functionality.`
	PreflightBasedOnUbiImpact                                   = `Non-UBI base images may lack security updates, enterprise support, and compliance certifications required for production use.`
	PreflightBundleImageRefsAreCertifiedImpact                  = `Uncertified bundle image references can introduce security vulnerabilities and compatibility issues in production deployments.`
	PreflightDeployableByOLMImpact                              = `Operators not deployable by OLM cannot be properly managed, updated, or integrated into OpenShift lifecycle management.`
	PreflightFollowsRestrictedNetworkEnablementGuidelinesImpact = `Non-compliance with restricted network guidelines can prevent deployment in air-gapped environments and violate security policies.`
	PreflightHasLicenseImpact                                   = `Missing license information can create legal compliance issues and prevent proper software asset management.`
	PreflightHasModifiedFilesImpact                             = `Modified files in containers can introduce security vulnerabilities, create inconsistent behavior, and violate immutable infrastructure principles.`
	PreflightHasNoProhibitedLabelsImpact                        = `Misuse of Red Hat trademarks in name, vendor, or maintainer labels creates legal and compliance risks that can block certification and publication.`
	PreflightHasNoProhibitedPackagesImpact                      = `Prohibited packages can introduce security vulnerabilities, licensing issues, and compliance violations.`
	PreflightHasProhibitedContainerNameImpact                   = `Prohibited container names can cause conflicts with system components and violate naming conventions.`
	PreflightHasRequiredLabelImpact                             = `Missing required labels prevent proper metadata management and can cause deployment and management issues.`
	PreflightHasUniqueTagImpact                                 = `Non-unique tags can cause version conflicts and deployment inconsistencies, making rollbacks and troubleshooting difficult.`
	PreflightLayerCountAcceptableImpact                         = `Excessive image layers can cause poor performance, increased storage usage, and longer deployment times.`
	PreflightRequiredAnnotationsImpact                          = `Missing required annotations can prevent proper operator lifecycle management and cause deployment failures.`
	PreflightRunAsNonRootImpact                                 = `Running containers as root increases the blast radius of security vulnerabilities and can lead to full host compromise if containers are breached.`
	PreflightScorecardBasicSpecCheckImpact                      = `Failing basic scorecard checks indicates fundamental operator implementation issues that can cause runtime failures.`
	PreflightScorecardOlmSuiteCheckImpact                       = `Failing OLM suite checks indicates operator lifecycle management issues that can prevent proper installation and updates.`
	PreflightSecurityContextConstraintsInCSVImpact              = `Incorrect SCC definitions in CSV can cause security policy violations and deployment failures.`
	PreflightValidateOperatorBundleImpact                       = `Invalid operator bundles can cause deployment failures, update issues, and operational instability.`
)

// ImpactMap maps test IDs to their impact statements.
// Non-preflight impact statements are in the checks library (CheckInfo.ImpactStatement).
var ImpactMap = map[string]string{
	"preflight-AllImageRefsInRelatedImages":                  PreflightAllImageRefsInRelatedImagesImpact,
	"preflight-BasedOnUbi":                                   PreflightBasedOnUbiImpact,
	"preflight-BundleImageRefsAreCertified":                  PreflightBundleImageRefsAreCertifiedImpact,
	"preflight-DeployableByOLM":                              PreflightDeployableByOLMImpact,
	"preflight-FollowsRestrictedNetworkEnablementGuidelines": PreflightFollowsRestrictedNetworkEnablementGuidelinesImpact,
	"preflight-HasLicense":                                   PreflightHasLicenseImpact,
	"preflight-HasModifiedFiles":                             PreflightHasModifiedFilesImpact,
	"preflight-HasNoProhibitedLabels":                        PreflightHasNoProhibitedLabelsImpact,
	"preflight-HasNoProhibitedPackages":                      PreflightHasNoProhibitedPackagesImpact,
	"preflight-HasProhibitedContainerName":                   PreflightHasProhibitedContainerNameImpact,
	"preflight-HasRequiredLabel":                             PreflightHasRequiredLabelImpact,
	"preflight-HasUniqueTag":                                 PreflightHasUniqueTagImpact,
	"preflight-LayerCountAcceptable":                         PreflightLayerCountAcceptableImpact,
	"preflight-RequiredAnnotations":                          PreflightRequiredAnnotationsImpact,
	"preflight-RunAsNonRoot":                                 PreflightRunAsNonRootImpact,
	"preflight-ScorecardBasicSpecCheck":                      PreflightScorecardBasicSpecCheckImpact,
	"preflight-ScorecardOlmSuiteCheck":                       PreflightScorecardOlmSuiteCheckImpact,
	"preflight-SecurityContextConstraintsInCSV":              PreflightSecurityContextConstraintsInCSVImpact,
	"preflight-ValidateOperatorBundle":                       PreflightValidateOperatorBundleImpact,
}
