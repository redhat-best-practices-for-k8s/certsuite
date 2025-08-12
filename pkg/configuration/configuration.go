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

package configuration

import "time"

const (
	defaultProbeDaemonSetNamespace = "cnf-suite"
)

// SkipHelmChartList holds the name of a Helm chart that should be skipped during processing.
//
// The Name field specifies the identifier of the Helm chart to exclude.
// This structure is used by configuration logic to filter out unwanted
// charts from installation or validation steps.
type SkipHelmChartList struct {
	// Name is the name of the `operator bundle package name` or `image-version` that you want to check if exists in the RedHat catalog
	Name string `yaml:"name" json:"name"`
}

// AcceptedKernelTaintsInfo holds all certified operator request information.
//
// It contains the module name that specifies which kernel taint handling
// logic should be accepted by the system. This struct is used to capture
// configuration data required during certification checks.
type AcceptedKernelTaintsInfo struct {

	// Accepted modules that cause taints that we want to supply to the test suite
	Module string `yaml:"module" json:"module"`
}

// SkipScalingTestDeploymentsInfo holds deployment identifiers that should be excluded from scaling tests to avoid conflicts.
//
// It contains a Name field specifying the deployment name and a Namespace field indicating where the deployment resides.
// Deployments listed in this structure are skipped during scaling test runs, ensuring those resources remain untouched.
type SkipScalingTestDeploymentsInfo struct {

	// Deployment name and namespace that can be skipped by the scaling tests
	Name      string `yaml:"name" json:"name"`
	Namespace string `yaml:"namespace" json:"namespace"`
}

// SkipScalingTestStatefulSetsInfo specifies statefulsets that should be excluded from scaling tests to avoid problems.
//
// It holds the name and namespace of a StatefulSet that must not be subjected to scaling operations during test runs. This prevents unintended side effects on critical workloads.
type SkipScalingTestStatefulSetsInfo struct {

	// StatefulSet name and namespace that can be skipped by the scaling tests
	Name      string `yaml:"name" json:"name"`
	Namespace string `yaml:"namespace" json:"namespace"`
}

// Namespace represents a Kubernetes namespace configuration.
//
// It holds the name of the namespace to be used in various operations.
// The struct is intended to be instantiated with a string value for Name,
// which specifies the target namespace for resource creation, deletion,
// or inspection within the application.
type Namespace struct {
	Name string `yaml:"name" json:"name"`
}

// CrdFilter defines a CustomResourceDefinition config filter.
//
// It filters CRDs based on their names and scalability flag.
// NameSuffix specifies a suffix that matching CRD names must have.
// Scalable indicates whether the filtered CRDs are expected to be scalable resources.
type CrdFilter struct {
	NameSuffix string `yaml:"nameSuffix" json:"nameSuffix"`
	Scalable   bool   `yaml:"scalable" json:"scalable"`
}

// ManagedDeploymentsStatefulsets represents the configuration for a managed stateful set deployment.
//
// It stores the name of a StatefulSet that is managed by the system.
// The struct can be extended with additional fields to capture
// further deployment properties such as namespace, labels, or replicas.
type ManagedDeploymentsStatefulsets struct {
	Name string `yaml:"name" json:"name"`
}

// ConnectAPIConfig holds configuration values for accessing the Red Hat Connect API.
//
// It contains the API key used for authentication, the base URL of the service,
// the project identifier, and optional proxy settings including the proxy URL
// and port. These fields are populated from environment variables or a
// configuration file and are passed to clients that need to make requests
// to the Connect API.
type ConnectAPIConfig struct {
	// APIKey is the API key for the Red Hat Connect
	APIKey string `yaml:"apiKey" json:"apiKey"`
	// ProjectID is the project ID for the Red Hat Connect
	ProjectID string `yaml:"projectID" json:"projectID"`
	// BaseURL is the base URL for the Red Hat Connect API
	BaseURL string `yaml:"baseURL" json:"baseURL"`
	// ProxyURL is the proxy URL for the Red Hat Connect API
	ProxyURL string `yaml:"proxyURL" json:"proxyURL"`
	// ProxyPort is the proxy port for the Red Hat Connect API
	ProxyPort string `yaml:"proxyPort" json:"proxyPort"`
}

// TestConfiguration provides test related configuration.
//
// It holds settings that control how tests are executed, including which
// Kubernetes resources to target, labels for operators and pods under test,
// collector endpoint details, and lists of items to skip or ignore during
// testing. The fields include information on kernel taints, managed
// deployments, protocol names, and namespace targets. This struct is used by
// the configuration package to load and supply runtime settings for test
// execution.
type TestConfiguration struct {
	// targetNameSpaces to be used in
	TargetNameSpaces []Namespace `yaml:"targetNameSpaces,omitempty" json:"targetNameSpaces,omitempty"`
	// labels identifying pods under test
	PodsUnderTestLabels []string `yaml:"podsUnderTestLabels,omitempty" json:"podsUnderTestLabels,omitempty"`
	// labels identifying operators unde test
	OperatorsUnderTestLabels []string `yaml:"operatorsUnderTestLabels,omitempty" json:"operatorsUnderTestLabels,omitempty"`
	// CRDs section.
	CrdFilters          []CrdFilter                      `yaml:"targetCrdFilters,omitempty" json:"targetCrdFilters,omitempty"`
	ManagedDeployments  []ManagedDeploymentsStatefulsets `yaml:"managedDeployments,omitempty" json:"managedDeployments,omitempty"`
	ManagedStatefulsets []ManagedDeploymentsStatefulsets `yaml:"managedStatefulsets,omitempty" json:"managedStatefulsets,omitempty"`

	// AcceptedKernelTaints
	AcceptedKernelTaints []AcceptedKernelTaintsInfo `yaml:"acceptedKernelTaints,omitempty" json:"acceptedKernelTaints,omitempty"`
	SkipHelmChartList    []SkipHelmChartList        `yaml:"skipHelmChartList,omitempty" json:"skipHelmChartList,omitempty"`
	// SkipScalingTestDeploymentNames
	SkipScalingTestDeployments []SkipScalingTestDeploymentsInfo `yaml:"skipScalingTestDeployments,omitempty" json:"skipScalingTestDeployments,omitempty"`
	// SkipScalingTestStatefulSetNames
	SkipScalingTestStatefulSets []SkipScalingTestStatefulSetsInfo `yaml:"skipScalingTestStatefulSets,omitempty" json:"skipScalingTestStatefulSets,omitempty"`
	ValidProtocolNames          []string                          `yaml:"validProtocolNames,omitempty" json:"validProtocolNames,omitempty"`
	ServicesIgnoreList          []string                          `yaml:"servicesignorelist,omitempty" json:"servicesignorelist,omitempty"`
	ProbeDaemonSetNamespace     string                            `yaml:"probeDaemonSetNamespace,omitempty" json:"probeDaemonSetNamespace,omitempty"`
	// Collector's parameters
	ExecutedBy           string `yaml:"executedBy,omitempty" json:"executedBy,omitempty"`
	PartnerName          string `yaml:"partnerName,omitempty" json:"partnerName,omitempty"`
	CollectorAppPassword string `yaml:"collectorAppPassword,omitempty" json:"collectorAppPassword,omitempty"`
	CollectorAppEndpoint string `yaml:"collectorAppEndpoint,omitempty" json:"collectorAppEndpoint,omitempty"`
	// ConnectAPIConfig contains the configuration for the Red Hat Connect API
	ConnectAPIConfig ConnectAPIConfig `yaml:"connectAPIConfig,omitempty" json:"connectAPIConfig,omitempty"`
}

// TestParameters holds configuration values used by CertSuite to control test execution and output generation.
//
// It contains fields that specify image repositories, API endpoints, resource limits, logging levels, data collection flags, and various file paths. The structure is populated from command-line options or configuration files and passed to the testing framework to configure runtime behavior, such as enabling preflight checks, setting Kubernetes client settings, and defining output directories.
type TestParameters struct {
	Kubeconfig                    string
	ConfigFile                    string
	PfltDockerconfig              string
	OutputDir                     string
	LabelsFilter                  string
	LogLevel                      string
	OfflineDB                     string
	DaemonsetCPUReq               string
	DaemonsetCPULim               string
	DaemonsetMemReq               string
	DaemonsetMemLim               string
	SanitizeClaim                 bool
	CertSuiteImageRepo            string
	CertSuiteProbeImage           string
	Intrusive                     bool
	AllowPreflightInsecure        bool
	IncludeWebFilesInOutputFolder bool
	OmitArtifactsZipFile          bool
	EnableDataCollection          bool
	EnableXMLCreation             bool
	ServerMode                    bool
	Timeout                       time.Duration
	ConnectAPIKey                 string
	ConnectProjectID              string
	ConnectAPIBaseURL             string
	ConnectAPIProxyURL            string
	ConnectAPIProxyPort           string
}
