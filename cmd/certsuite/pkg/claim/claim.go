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

// TestCaseRawResult holds the raw result of a test case execution.
//
// It contains two fields: Name, which is the identifier of the test case, and Status, which represents the outcome of that test case (for example, passed, failed, or skipped). The struct is used to report individual test results before they are aggregated into higher‑level summaries.
type TestCaseRawResult struct {
	Name   string `json:"-name"`
	Status string `json:"-status"`
}

// TestCaseID represents a unique identifier for a test case within the claim package.
//
// It contains an ID string that uniquely identifies the test case, a Suite name indicating
// which test suite it belongs to, and optional Tags providing additional metadata or
// categorization information about the test case. The struct is used by other components
// to reference, look up, or filter test cases based on these fields.
type TestCaseID struct {
	ID    string `json:"id"`
	Suite string `json:"suite"`
	Tags  string `json:"tags"`
}

// TestCaseResult holds the outcome of a single test case execution.
//
// It captures timing information, state, and any failure details,
// along with catalog metadata such as description and remediation.
// The struct also stores the test identifier, category classifications,
// and raw output captured during the run.
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

// Nodes represents node-related information collected during a claim process.
//
// It aggregates various categories of node data, such as networking configuration,
// storage driver details, hardware specifications, and a summary view.
// Each field is an interface{}, allowing flexible underlying types that
// capture the corresponding aspect of a node's state.
type Nodes struct {
	NodesSummary interface{} `json:"nodeSummary"`
	CniNetworks  interface{} `json:"cniPlugins"`
	NodesHwInfo  interface{} `json:"nodesHwInfo"`
	CsiDriver    interface{} `json:"csiDriver"`
}

// TestOperator represents a test operator deployment.
//
// It holds the basic identification fields required to locate and manage a
// test operator within a Kubernetes cluster: Name, Namespace, and Version.
type TestOperator struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Version   string `json:"version"`
}

// Configurations holds configuration data for claim processing.
//
// It contains a slice of abnormal event representations, a generic config interface,
// and a list of test operators to be used during claim operations. The fields are
// intended to be populated from external sources such as YAML or JSON files
// before being passed to the claim logic.
type Configurations struct {
	Config         interface{}    `json:"Config"`
	AbnormalEvents []interface{}  `json:"AbnormalEvents"`
	TestOperators  []TestOperator `json:"testOperators"`
}

// Schema represents the top‑level structure of a claim file used by certsuite.
//
// It contains Claim, which holds configuration data, node information,
// test results, and version metadata. The Schema type is returned by
// Parse when reading and unmarshalling a claim JSON/YAML document.
type Schema struct {
	Claim struct {
		Configurations `json:"configurations"`

		Nodes Nodes `json:"nodes"`

		Results  TestSuiteResults             `json:"results"`
		Versions officialClaimScheme.Versions `json:"versions"`
	} `json:"claim"`
}

// CheckVersion validates the claim format version string.
//
// It parses the supplied version and compares it against the
// supportedClaimFormatVersion constant. If the version is not
// compatible, an error describing the mismatch is returned.
// On success, nil is returned.
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

// Parse reads a claim file from disk and unmarshals its JSON into a Schema.
//
// It takes the path to a claim file as input, reads the file contents,
// decodes the JSON into a Schema struct, and returns the populated
// Schema along with any error that occurs during reading or parsing.
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
