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

// SkipHelmChartList Specifies a Helm chart to exclude from catalog checks
//
// This structure holds the identifier for an operator bundle package or image
// version that should be omitted when verifying existence against the RedHat
// catalog. The Name field contains the exact name used in the catalog lookup.
// When populated, the system will skip any validation or processing related to
// this chart.
type SkipHelmChartList struct {
	// Name is the name of the `operator bundle package name` or `image-version` that you want to check if exists in the RedHat catalog
	Name string `yaml:"name" json:"name"`
}

// AcceptedKernelTaintsInfo stores information about kernel module taints used in tests
//
// This structure holds the name of a kernel module that, when loaded, causes
// specific taints on nodes. The module field is used by the test suite to
// identify which taints should be accepted during certification testing. It
// facilitates configuration of test environments that require certain kernel
// behavior.
type AcceptedKernelTaintsInfo struct {

	// Accepted modules that cause taints that we want to supply to the test suite
	Module string `yaml:"module" json:"module"`
}

// SkipScalingTestDeploymentsInfo Lists deployments excluded from scaling tests
//
// This structure stores a deployment's name and namespace that should be
// ignored during scaling test runs. By including these entries in the
// configuration, the testing framework bypasses any checks or actions that
// could interfere with or corrupt the selected deployments.
type SkipScalingTestDeploymentsInfo struct {

	// Deployment name and namespace that can be skipped by the scaling tests
	Name      string `yaml:"name" json:"name"`
	Namespace string `yaml:"namespace" json:"namespace"`
}

// SkipScalingTestStatefulSetsInfo Specifies statefulsets excluded from scaling tests
//
// This structure holds the name and namespace of a StatefulSet that should be
// ignored during scaling test runs to avoid potential failures or conflicts. By
// listing such StatefulSets, the testing framework can bypass them while still
// evaluating other components.
type SkipScalingTestStatefulSetsInfo struct {

	// StatefulSet name and namespace that can be skipped by the scaling tests
	Name      string `yaml:"name" json:"name"`
	Namespace string `yaml:"namespace" json:"namespace"`
}

// Namespace Represents a Kubernetes namespace configuration
//
// This structure holds information about a single namespace, primarily its name
// used for identification in the cluster. The name is serialized to YAML or
// JSON under the key "name". It serves as a basic unit for configuring
// namespace-specific settings within the application.
type Namespace struct {
	Name string `yaml:"name" json:"name"`
}

// CrdFilter filters CustomResourceDefinitions by name suffix and scaling capability
//
// This structure holds criteria for selecting CRDs from a configuration. The
// NameSuffix field specifies a string that must appear at the end of a CRD’s
// name to be considered a match. The Scalable boolean indicates whether only
// scalable CRDs should be included in the filtered set.
type CrdFilter struct {
	NameSuffix string `yaml:"nameSuffix" json:"nameSuffix"`
	Scalable   bool   `yaml:"scalable" json:"scalable"`
}

// ManagedDeploymentsStatefulsets Represents the identifier of a StatefulSet in a managed deployment
//
// This structure stores the name of a Kubernetes StatefulSet that should be
// tracked or controlled by the system. It is used as part of configuration
// data, typically loaded from YAML or JSON files, to specify which stateful
// sets are relevant for monitoring or management tasks.
type ManagedDeploymentsStatefulsets struct {
	Name string `yaml:"name" json:"name"`
}

// ConnectAPIConfig configuration holder for accessing the Red Hat Connect API
//
// It stores the credentials, project identifier, endpoint address, and optional
// proxy settings required to communicate with the Red Hat Connect service. Each
// field is mapped to YAML and JSON keys so it can be loaded from configuration
// files or environment variables.
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

// TestConfiguration holds configuration values used during test execution
//
// This struct groups settings that control which namespaces, pods, operators,
// and CRDs are considered in a test run. It also contains parameters for the
// collector application and connection to an external API. The fields support
// filtering, skipping certain resources, and specifying accepted kernel taints
// or protocol names.
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

// TestParameters holds configuration settings for test execution
//
// This structure contains a collection of fields that control how tests are
// run, including resource limits, image repositories, API connection details,
// and output options. It also flags whether to include non-running pods, enable
// data collection or XML creation, and sets timeouts and log levels for the
// test environment.
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
	// AllowNonRunning determines whether autodiscovery includes non-Running pods
	AllowNonRunning bool
}
