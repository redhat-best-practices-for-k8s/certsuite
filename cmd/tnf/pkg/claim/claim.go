package claim

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	corev1 "k8s.io/api/core/v1"
)

const (
	supportedClaimFormatVersion = "v0.1.0"
)

const (
	TestCaseResultPassed  = "passed"
	TestCaseResultSkipped = "skipped"
	TestCaseResultFailed  = "failed"
)

type TestCaseRawResult struct {
	Name   string `json:"-name"`
	Status string `json:"-status"`
}

type TestCaseID struct {
	ID    string `json:"id"`
	Suite string `json:"suite"`
	Tags  string `json:"tags"`
}

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
	FailureReason          string            `json:"failureReason"`
	StartTime              string            `json:"startTime"`
	State                  string            `json:"state"`
	TestID                 struct {
		ID    string `json:"id"`
		Suite string `json:"suite"`
		Tags  string `json:"tags"`
	} `json:"testID"`
}

// Maps a test suite name to a list of TestCaseResult
type TestSuiteResults map[string][]TestCaseResult

type Nodes struct {
	NodesSummary map[string]*corev1.Node `json:"nodeSummary"`
	CniNetworks  interface{}             `json:"cniPlugins"`
	NodesHwInfo  interface{}             `json:"nodesHwInfo"`
	CsiDriver    interface{}             `json:"csiDriver"`
}

type Schema struct {
	Claim struct {
		Nodes Nodes `json:"nodes"`

		RawResults struct {
			Cnfcertificationtest struct {
				Testsuites struct {
					Testsuite struct {
						Testcase []TestCaseRawResult `json:"testcase"`
					} `json:"testsuite"`
				} `json:"testsuites"`
			} `json:"cnf-certification-test"`
		} `json:"rawResults"`

		Results  TestSuiteResults `json:"results"`
		Versions struct {
			ClaimFormat string `json:"claimFormat"`
		} `json:"versions"`
	} `json:"claim"`
}

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
