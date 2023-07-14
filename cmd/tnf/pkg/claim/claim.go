package claim

import (
	"encoding/json"
	"fmt"
	"os"
)

type Cni struct {
	Name    string        "json:\"name\""
	Plugins []interface{} "json:\"plugins\""
}

type TestCaseRawResult struct {
	Name   string `json:"-name"`
	Status string `json:"-status"`
}

type TestCaseResult struct {
	TestID struct {
		ID    string `json:"id"`
		Suite string `json:"suite"`
		Tags  string `json:"tags"`
	}
	Description string `json:"testText"`

	Output        string `json:"CapturedTestOutput"`
	FailureReason string `json:"failureReason"`
	State         string `json:"state"`

	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type Schema struct {
	Claim struct {
		Nodes struct {
			CniPlugins  map[string][]Cni       `json:"cniPlugins"`
			NodesHwInfo map[string]interface{} `json:"nodesHwInfo"`
			CsiDriver   interface{}            `json:"csiDriver"`
		} `json:"nodes"`

		RawResults struct {
			Cnfcertificationtest struct {
				Testsuites struct {
					Testsuite struct {
						Testcase []TestCaseRawResult `json:"testcase"`
					} `json:"testsuite"`
				} `json:"testsuites"`
			} `json:"cnf-certification-test"`
		} `json:"rawResults"`

		Results map[string][]TestCaseResult `json:"results"`
	} `json:"claim"`
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
