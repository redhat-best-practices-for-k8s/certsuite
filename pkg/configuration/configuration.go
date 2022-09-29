// Copyright (C) 2020-2023 Red Hat, Inc.
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

type SkipHelmChartList struct {
	// Name is the name of the `operator bundle package name` or `image-version` that you want to check if exists in the RedHat catalog
	Name string `yaml:"name" json:"name"`
}

// CertifiedOperatorRequestInfo contains all certified operator request info
type CertifiedOperatorRequestInfo struct {

	// Name is the name of the `operator bundle package name` that you want to check if exists in the RedHat catalog
	Name string `yaml:"name" json:"name"`

	// Organization as understood by the operator publisher, e.g. `redhat-marketplace`
	Organization string `yaml:"organization" json:"organization"`
}

// AcceptedKernelTaintsInfo contains all certified operator request info
type AcceptedKernelTaintsInfo struct {

	// Accepted modules that cause taints that we want to supply to the test suite
	Module string `yaml:"module" json:"module"`
}

// SkipScalingTestDeploymentsInfo contains a list of names of deployments that should be skipped by the scaling tests to prevent issues
type SkipScalingTestDeploymentsInfo struct {

	// Deployment name and namespace that can be skipped by the scaling tests
	Name      string `yaml:"name" json:"name"`
	Namespace string `yaml:"namespace" json:"namespace"`
}

// SkipScalingTestStatefulSetsInfo contains a list of names of statefulsets that should be skipped by the scaling tests to prevent issues
type SkipScalingTestStatefulSetsInfo struct {

	// StatefulSet name and namespace that can be skipped by the scaling tests
	Name      string `yaml:"name" json:"name"`
	Namespace string `yaml:"namespace" json:"namespace"`
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
}

// Tag and Digest should not be populated at the same time. Digest takes precedence if both are populated
type ContainerImageIdentifier struct {
	// Name is the name of the image that you want to check if exists in the RedHat catalog
	Name string `yaml:"name" json:"name"`

	// Repository is the name of the repository `rhel8` of the container
	// This is valid for container only and required field
	Repository string `yaml:"repository" json:"repository"`

	// Tag is the optional image tag. "latest" is implied if not specified
	Tag string `yaml:"tag" json:"tag"`

	// Digest is the image digest following the "@" in a URL, e.g. image@sha256:45b23dee08af5e43a7fea6c4cf9c25ccf269ee113168c19722f87876677c5cb2
	Digest string `yaml:"digest" json:"digest"`
}

// TestConfiguration provides test related configuration
type TestConfiguration struct {
	// targetNameSpaces to be used in
	TargetNameSpaces []Namespace `yaml:"targetNameSpaces" json:"targetNameSpaces"`
	// Custom Pod labels for discovering containers/pods under test
	TargetPodLabels []Label `yaml:"targetPodLabels,omitempty" json:"targetPodLabels,omitempty"`
	// CertifiedContainerInfo is the list of container images to be checked for certification status.
	CertifiedContainerInfo []ContainerImageIdentifier `yaml:"certifiedcontainerinfo,omitempty" json:"certifiedcontainerinfo,omitempty"`
	// CertifiedOperatorInfo is list of operator bundle names that are queried for certification status.
	CertifiedOperatorInfo []CertifiedOperatorRequestInfo `yaml:"certifiedoperatorinfo,omitempty" json:"certifiedoperatorinfo,omitempty"`
	// CRDs section.
	CrdFilters []CrdFilter `yaml:"targetCrdFilters" json:"targetCrdFilters"`
	// AcceptedKernelTaints
	AcceptedKernelTaints []AcceptedKernelTaintsInfo `yaml:"acceptedKernelTaints,omitempty" json:"acceptedKernelTaints,omitempty"`
	SkipHelmChartList    []SkipHelmChartList        `yaml:"skipHelmChartList" json:"skipHelmChartList"`
	// SkipScalingTestDeploymentNames
	SkipScalingTestDeployments []SkipScalingTestDeploymentsInfo `yaml:"skipScalingTestDeployments,omitempty" json:"skipScalingTestDeployments,omitempty"`
	// SkipScalingTestStatefulSetNames
	SkipScalingTestStatefulSets []SkipScalingTestStatefulSetsInfo `yaml:"skipScalingTestStatefulSets,omitempty" json:"skipScalingTestStatefulSets,omitempty"`
	// CheckDiscoveredContainerCertificationStatus controls whether the container certification test will validate images used by autodiscovered containers, in addition to the configured image list
	CheckDiscoveredContainerCertificationStatus bool     `yaml:"checkDiscoveredContainerCertificationStatus" json:"checkDiscoveredContainerCertificationStatus"`
	ValidProtocolNames                          []string `yaml:"validProtocolNames" json:"validProtocolNames"`
	ServicesIgnoreList                          []string `yaml:"servicesignorelist" json:"servicesignorelist"`
}

type TestParameters struct {
	Home                   string `envconfig:"home"`
	Kubeconfig             string `envconfig:"kubeconfig"`
	ConfigurationPath      string `split_words:"true" default:"tnf_config.yml"`
	NonIntrusiveOnly       bool   `split_words:"true"`
	LogLevel               string `default:"debug" split_words:"true"`
	OfflineDB              string `split_words:"true"`
	AllowPreflightInsecure bool   `split_words:"true"`
}
