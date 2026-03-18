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
	"github.com/redhat-best-practices-for-k8s/oct/pkg/certdb"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

// octValidatorAdapter wraps oct's CertificationStatusValidator to implement
// checks.CertificationValidator with simple string parameters.
type octValidatorAdapter struct {
	inner certdb.CertificationStatusValidator
}

func (a *octValidatorAdapter) IsContainerCertified(registry, repository, tag, digest string) bool {
	return a.inner.IsContainerCertified(registry, repository, tag, digest)
}

func (a *octValidatorAdapter) IsOperatorCertified(csvName, ocpVersion string) bool {
	return a.inner.IsOperatorCertified(csvName, ocpVersion)
}

func (a *octValidatorAdapter) IsHelmChartCertified(chartName, chartVersion, kubeVersion string) bool {
	rel := &release.Release{
		Name: chartName,
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{Version: chartVersion},
		},
	}
	return a.inner.IsHelmChartCertified(rel, kubeVersion)
}

// NewCertValidator creates a checks.CertificationValidator backed by oct's
// CertificationStatusValidator. Returns nil if the validator cannot be created.
func NewCertValidator(offlineDBPath string) *octValidatorAdapter {
	validator, err := certdb.GetValidator(offlineDBPath)
	if err != nil {
		return nil
	}
	return &octValidatorAdapter{inner: validator}
}
