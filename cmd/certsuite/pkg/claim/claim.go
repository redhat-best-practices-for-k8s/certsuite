package claim

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	officialClaimScheme "github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
)

const (
	supportedClaimFormatVersion = "v0.5.0"
)

const (
	TestCaseResultPassed  = "passed"
	TestCaseResultSkipped = "skipped"
	TestCaseResultFailed  = "failed"
)

// TestCaseRawResult Represents the outcome of a test case
//
// This structure holds the name of a test case along with its status, such as
// passed or failed. The fields are tagged for JSON serialization but omitted
// from output. It is used to aggregate results before further processing.
type TestCaseRawResult struct {
	Name   string `json:"-name"`
	Status string `json:"-status"`
}

// TestCaseID represents a unique identifier for a test case
//
// This struct holds the ID, suite name, and tags of a test case as strings. The
// fields are exported and annotated for JSON serialization with keys "id",
// "suite", and "tags". It is used to track and reference individual test cases
// within the claim package.
type TestCaseID struct {
	ID    string `json:"id"`
	Suite string `json:"suite"`
	Tags  string `json:"tags"`
}

// TestCaseResult Stores the outcome of an individual test case
//
// This structure captures metadata about a single test execution, including its
// identifier, timing, state, and any failure details. It also holds catalog
// information such as best practice references, descriptions, exception
// handling notes, and remediation steps. The fields are organized to support
// serialization for reporting and analysis of test results.
type TestCaseResult struct {
	CapturedTestOutput string `json:"capturedTestOutput"`
	CatalogInfo        struct {
		BestPracticeReference string `json:"bestPracticeReference"`
		Description           string `json:"description"`
		ExceptionProcess      string `json:"exceptionProcess"`
		Remediation           string `json:"remediation"`
	} `json:"catalogInfo"`
	CategoryClassification map[string]string `json:"categoryClassification"`
	Duration               int               `json:"duration"`
	EndTime                string            `json:"endTime"`
	FailureLineContent     string            `json:"failureLineContent"`
	FailureLocation        string            `json:"failureLocation"`
	SkipReason             string            `json:"skipReason"`
	CheckDetails           string            `json:"checkDetails"`
	StartTime              string            `json:"startTime"`
	State                  string            `json:"state"`
	TestID                 struct {
		ID    string `json:"id"`
		Suite string `json:"suite"`
		Tags  string `json:"tags"`
	} `json:"testID"`
}

// Maps a test suite name to a list of TestCaseResult
type TestSuiteResults map[string]TestCaseResult

// Nodes represents information about nodes in a cluster
//
// This struct holds aggregated data for the nodes, including their hardware
// details, network plugin configuration, CSI driver status, and an overall
// summary of node health or capabilities. Each field is defined as an interface
// to allow flexible JSON unmarshalling from various sources.
type Nodes struct {
	NodesSummary interface{} `json:"nodeSummary"`
	CniNetworks  interface{} `json:"cniPlugins"`
	NodesHwInfo  interface{} `json:"nodesHwInfo"`
	CsiDriver    interface{} `json:"csiDriver"`
}

// TestOperator Describes a Kubernetes operator to be tested
//
// This struct holds the basic identifying information for an operator,
// including its name, the namespace it runs in, and its version string. It is
// used by testing utilities to reference specific operator deployments during
// validation or cleanup operations.
type TestOperator struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Version   string `json:"version"`
}

// Configurations Holds test configuration data
//
// This structure stores the overall configuration for a claim test, including
// any custom settings, a list of abnormal events to be monitored, and a
// collection of operators that should run during the test. Each field is
// designed to be marshalled to or from JSON, allowing easy integration with
// external tools or configuration files.
type Configurations struct {
	Config         interface{}    `json:"Config"`
	AbnormalEvents []interface{}  `json:"AbnormalEvents"`
	TestOperators  []TestOperator `json:"testOperators"`
}

// Schema Encapsulates an entire claim record
//
// The structure holds the top‑level claim object which includes configuration
// settings, node information, test suite outcomes, and schema versioning data.
// Each field maps directly to a JSON key in the claim file, allowing easy
// serialization and deserialization of the claim contents.
type Schema struct {
	Claim struct {
		Configurations `json:"configurations"`

		Nodes Nodes `json:"nodes"`

		Results  TestSuiteResults             `json:"results"`
		Versions officialClaimScheme.Versions `json:"versions"`
	} `json:"claim"`
}

// CheckVersion Validates the claim file format version against a supported version
//
// The function parses the supplied version string into a semantic version
// object, then compares it to the predefined supported claim format version. If
// parsing fails or if the two versions do not match exactly, an error is
// returned describing the issue. When the versions are equal, the function
// returns nil indicating success.
func CheckVersion(version string) error {
	claimSemVersion, err := semver.NewVersion(version)
	if err != nil {
		return fmt.Errorf("claim file version %q is not valid: %v", version, err)
	}

	supportedSemVersion, err := semver.NewVersion(supportedClaimFormatVersion)
	if err != nil {
		return fmt.Errorf("supported claim file version v%v is not valid: v%v", supportedClaimFormatVersion, err)
	}

	if claimSemVersion.Compare(supportedSemVersion) != 0 {
		return fmt.Errorf("claim format version v%v is not supported. Supported version is v%v",
			claimSemVersion, supportedSemVersion)
	}

	return nil
}

// Parse Parses a JSON claim file into a structured schema
//
// The function reads the entire contents of the specified file path, handling
// any read errors with an informative message. It then unmarshals the JSON data
// into a Schema object, returning detailed errors if parsing fails. On success
// it returns a pointer to the populated Schema and a nil error.
func Parse(filePath string) (*Schema, error) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failure reading file: %v", err)
	}

	claimFile := Schema{}
	err = json.Unmarshal(fileBytes, &claimFile)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %v", err)
	}

	return &claimFile, nil
}
