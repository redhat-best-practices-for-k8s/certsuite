package claim

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	corev1 "k8s.io/api/core/v1"
)

const (
	supportedClaimFormatVersion = "v0.0.2"
)

const (
	TestCaseResultPassed  = "passed"
	TestCaseResultSkipped = "skipped"
	TestCaseResultFailed  = "failed"
)

type CNIPlugin map[string]interface{}

type Cni struct {
}

type CNINetwork struct {
	Name         string      `json:"name"`
	CNIVersion   string      `json:"cniVersion"`
	DisableCheck bool        `json:"disableCheck"`
	Plugins      []CNIPlugin `json:"plugins"`
}

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
	TestID      TestCaseID `json:"TestID"`
	Description string     `json:"testText"`

	Output        string `json:"CapturedTestOutput"`
	FailureReason string `json:"failureReason"`
	State         string `json:"state"`

	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

// Maps a test suite name to a list of TestCaseResult
type TestSuiteResults map[string][]TestCaseResult

type Nodes struct {
	NodesSummary map[string]corev1.Node  `json:"nodeSummary"`
	CniNetworks  map[string][]CNINetwork `json:"cniPlugins"`
	NodesHwInfo  map[string]interface{}  `json:"nodesHwInfo"`
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

func CheckClaimVersion(version string) error {
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
