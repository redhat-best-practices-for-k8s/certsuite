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

package configuration

// CertifiedContainerRequestInfo contains all certified images request info
type CertifiedContainerRequestInfo struct {
	// Name is the name of the `operator bundle package name` or `image-version` that you want to check if exists in the RedHat catalog
	Name string `yaml:"name" json:"name"`

	// Repository is the name of the repository `rhel8` of the container
	// This is valid for container only and required field
	Repository string `yaml:"repository" json:"repository"`
}

// CertifiedOperatorRequestInfo contains all certified operator request info
type CertifiedOperatorRequestInfo struct {

	// Name is the name of the `operator bundle package name` that you want to check if exists in the RedHat catalog
	Name string `yaml:"name" json:"name"`

	// Organization as understood by the operator publisher , e.g. `redhat-marketplace`
	Organization string `yaml:"organization" json:"organization"`
}

// Label ns/name/value for resource lookup
type Label struct {
	Prefix string `yaml:"prefix" json:"prefix"`
	Name   string `yaml:"name" json:"name"`
	Value  string `yaml:"value" json:"value"`
}

// Namespace struct defines namespace properties
type Namespace struct {
	Name string `yaml:"name" json:"name"`
}

// CrdFilter defines a CustomResourceDefinition config filter.
type CrdFilter struct {
	NameSuffix string `yaml:"nameSuffix" json:"nameSuffix"`
	// labels []Label
}

// TestConfiguration provides test related configuration
type TestConfiguration struct {
	// targetNameSpaces to be used in
	TargetNameSpaces []Namespace `yaml:"targetNameSpaces" json:"targetNameSpaces"`
	// Custom Pod labels for discovering containers/pods under test
	TargetPodLabels []Label `yaml:"targetPodLabels,omitempty" json:"targetPodLabels,omitempty"`
	// CertifiedContainerInfo is the list of container images to be checked for certification status.
	CertifiedContainerInfo []CertifiedContainerRequestInfo `yaml:"certifiedcontainerinfo,omitempty" json:"certifiedcontainerinfo,omitempty"`
	// CertifiedOperatorInfo is list of operator bundle names that are queried for certification status.
	CertifiedOperatorInfo []CertifiedOperatorRequestInfo `yaml:"certifiedoperatorinfo,omitempty" json:"certifiedoperatorinfo,omitempty"`
	// CRDs section.
	CrdFilters []CrdFilter `yaml:"targetCrdFilters" json:"targetCrdFilters"`
}

type TestParameters struct {
	Home              string `envconfig:"home"`
	Kubeconfig        string `envconfig:"kubeconfig"`
	ConfigurationPath string `split_words:"true" default:"tnf_config.yml"`
	NonIntrusiveOnly  bool   `split_words:"true"`
	LogLevel          string `default:"debug" split_words:"true"`
}
