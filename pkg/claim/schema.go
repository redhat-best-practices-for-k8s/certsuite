// Copyright (C) 2023 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later
// version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied
// warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along with this program; if not, write to the Free
// Software Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package claim

import (
	"bytes"
	"encoding/json"
	"errors"
)

// CatalogInfo test specific information from the catalog
type CatalogInfo struct {

	// Link to the best practice document supporting this test case
	BestPracticeReference string `json:"bestPracticeReference"`

	// The test description.
	Description string `json:"description"`

	// Indicates the exception process if defined
	ExceptionProcess string `json:"exceptionProcess"`

	// steps required to fix a failing test case
	Remediation string `json:"remediation"`
}

// CategoryClassification categoryClassification is the classification for a single test case.
type CategoryClassification struct {

	// indicates whether this test case is mandatory or optional in the Extended scenario
	Extended string `json:"Extended,omitempty"`

	// indicates whether this test case is mandatory or optional in the FarEdge scenario
	FarEdge string `json:"FarEdge,omitempty"`

	// indicates whether this test case is mandatory or optional in the NonTelco scenario
	NonTelco string `json:"NonTelco,omitempty"`

	// indicates whether this test case is mandatory or optional in the Telco scenario
	Telco string `json:"Telco,omitempty"`
}

// Claim
type Claim struct {

	// Tests within test-network-function often require configuration.  For example, the generic test suite requires listing all CNF containers.
	// This information is used to derive per-container IP address information, which is then used as input to the connectivity test suite.
	// Test suites within test-network-function may use multiple configurations, but each with a unique name.
	Configurations map[string]interface{} `json:"configurations"`
	Metadata       *Metadata              `json:"metadata"`

	// An OpenShift cluster is composed of an arbitrary number of Nodes used for platform and application services.
	// Since a claim must be reproducible, a variety of per-Node information must be collected and stored in the claim.
	// Node names are unique within a given OpenShift cluster.
	Nodes map[string]interface{} `json:"nodes"`

	// The test-network-function test results.  Results are a JSON representation of the JUnit output.
	RawResults map[string]interface{} `json:"rawResults"`

	// The results for each unique test case.
	Results  map[string]interface{} `json:"results,omitempty"`
	Versions *Versions              `json:"versions"`
}

// Identifier identifier is a per testcase unique identifier.
type Identifier struct {

	// id stores a unique id for the testcase.
	Id string `json:"id"` //nolint:revive, stylecheck

	// suite stores the test suite name for the testcase.
	Suite string `json:"suite"`

	// tags stores the different tags applied to a test.
	Tags string `json:"tags,omitempty"`
}

// Metadata
type Metadata struct {

	// The UTC end time of a claim evaluation.  This is recorded when the test-network-function test suite completes.
	EndTime string `json:"endTime"`

	// The UTC start time of a claim evaluation.  This is recorded when the test-network-function test suite is invoked.
	StartTime string `json:"startTime"`
}

// Result result is the result of running a testcase.
type Result struct {

	// Ginkgo writer output during the test run.
	CapturedTestOutput string `json:"capturedTestOutput"`

	// Test detailed information from catalog
	CatalogInfo *CatalogInfo `json:"catalogInfo"`

	// Category classification for the test
	CategoryClassification *CategoryClassification `json:"categoryClassification"`

	// The duration of the test in nanoseconds.
	Duration int `json:"duration"`

	// The end time of the test.
	EndTime string `json:"endTime,omitempty"`

	// The content of the line where the failure happened
	FailureLineContent string `json:"failureLineContent"`

	// The Filename and line number where the failure happened
	FailureLocation string `json:"failureLocation"`

	// Describes the test failure in detail.
	FailureReason string `json:"failureReason"`

	// The start time of the test.
	StartTime string `json:"startTime"`

	// The test result state: INVALID SPEC STATE, pending,skipped,passed,failed,aborted,panicked,interrupted
	State string `json:"state"`

	// The test identifier
	TestID *Identifier `json:"testID"`
}

// Root A test-network-function claim is an attestation of the tests performed, the results and the various configurations.  Since a claim must be reproducible, it also includes an overview of the systems under test and their physical configurations.
type Root struct {
	Claim *Claim `json:"claim"`
}

// Versions
type Versions struct {

	// The claim file format version.
	ClaimFormat string `json:"claimFormat"`

	// The Kubernetes release version.
	K8s string `json:"k8s,omitempty"`

	// The oc client release version.
	OcClient string `json:"ocClient,omitempty"`

	// OCP cluster release version.
	Ocp string `json:"ocp,omitempty"`

	// The test-network-function (tnf) release version.
	Tnf string `json:"tnf"`

	// The test-network-function (tnf) Git Commit.
	TnfGitCommit string `json:"tnfGitCommit,omitempty"`
}

func (strct *CatalogInfo) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.WriteString("{")
	// "BestPracticeReference" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "bestPracticeReference" field
	buf.WriteString("\"bestPracticeReference\": ")
	tmp, err := json.Marshal(strct.BestPracticeReference)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "Description" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "description" field
	buf.WriteString(",")
	buf.WriteString("\"description\": ")
	tmp, err = json.Marshal(strct.Description)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "ExceptionProcess" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "exceptionProcess" field
	buf.WriteString(",")
	buf.WriteString("\"exceptionProcess\": ")
	tmp, err = json.Marshal(strct.ExceptionProcess)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "Remediation" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "remediation" field
	buf.WriteString(",")
	buf.WriteString("\"remediation\": ")
	tmp, err = json.Marshal(strct.Remediation)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)

	buf.WriteString("}")
	rv := buf.Bytes()
	return rv, nil
}

func (strct *CatalogInfo) UnmarshalJSON(b []byte) error {
	bestPracticeReferenceReceived := false
	descriptionReceived := false
	exceptionProcessReceived := false
	remediationReceived := false
	var jsonMap map[string]json.RawMessage
	if err := json.Unmarshal(b, &jsonMap); err != nil {
		return err
	}
	// parse all the defined properties
	for k, v := range jsonMap {
		switch k {
		case "bestPracticeReference":
			if err := json.Unmarshal([]byte(v), &strct.BestPracticeReference); err != nil {
				return err
			}
			bestPracticeReferenceReceived = true
		case "description":
			if err := json.Unmarshal([]byte(v), &strct.Description); err != nil {
				return err
			}
			descriptionReceived = true
		case "exceptionProcess":
			if err := json.Unmarshal([]byte(v), &strct.ExceptionProcess); err != nil {
				return err
			}
			exceptionProcessReceived = true
		case "remediation":
			if err := json.Unmarshal([]byte(v), &strct.Remediation); err != nil {
				return err
			}
			remediationReceived = true
		default:
			return errors.New("additional property not allowed: \"" + k + "\"") //nolint:goconst
		}
	}
	// check if bestPracticeReference (a required property) was received
	if !bestPracticeReferenceReceived {
		return errors.New("\"bestPracticeReference\" is required but was not present")
	}
	// check if description (a required property) was received
	if !descriptionReceived {
		return errors.New("\"description\" is required but was not present")
	}
	// check if exceptionProcess (a required property) was received
	if !exceptionProcessReceived {
		return errors.New("\"exceptionProcess\" is required but was not present")
	}
	// check if remediation (a required property) was received
	if !remediationReceived {
		return errors.New("\"remediation\" is required but was not present")
	}
	return nil
}

func (strct *CategoryClassification) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.WriteString("{")
	// Marshal the "Extended" field
	buf.WriteString("\"Extended\": ")
	tmp, err := json.Marshal(strct.Extended)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// Marshal the "FarEdge" field
	buf.WriteString(",")
	buf.WriteString("\"FarEdge\": ")
	tmp, err = json.Marshal(strct.FarEdge)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// Marshal the "NonTelco" field
	buf.WriteString(",")
	buf.WriteString("\"NonTelco\": ")
	tmp, err = json.Marshal(strct.NonTelco)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// Marshal the "Telco" field
	buf.WriteString(",")
	buf.WriteString("\"Telco\": ")
	tmp, err = json.Marshal(strct.Telco)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)

	buf.WriteString("}")
	rv := buf.Bytes()
	return rv, nil
}

func (strct *CategoryClassification) UnmarshalJSON(b []byte) error {
	var jsonMap map[string]json.RawMessage
	if err := json.Unmarshal(b, &jsonMap); err != nil {
		return err
	}
	// parse all the defined properties
	for k, v := range jsonMap {
		switch k {
		case "Extended":
			if err := json.Unmarshal([]byte(v), &strct.Extended); err != nil {
				return err
			}
		case "FarEdge":
			if err := json.Unmarshal([]byte(v), &strct.FarEdge); err != nil {
				return err
			}
		case "NonTelco":
			if err := json.Unmarshal([]byte(v), &strct.NonTelco); err != nil {
				return err
			}
		case "Telco":
			if err := json.Unmarshal([]byte(v), &strct.Telco); err != nil {
				return err
			}
		default:
			return errors.New("additional property not allowed: \"" + k + "\"")
		}
	}
	return nil
}

//nolint:funlen
func (strct *Claim) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.WriteString("{")
	// "Configurations" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "configurations" field
	buf.WriteString("\"configurations\": ")
	tmp, err := json.Marshal(strct.Configurations)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "Metadata" field is required
	if strct.Metadata == nil {
		return nil, errors.New("metadata is a required field")
	}
	// Marshal the "metadata" field
	buf.WriteString(",")
	buf.WriteString("\"metadata\": ")
	tmp, err = json.Marshal(strct.Metadata)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "Nodes" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "nodes" field
	buf.WriteString(",")
	buf.WriteString("\"nodes\": ")
	tmp, err = json.Marshal(strct.Nodes)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "RawResults" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "rawResults" field
	buf.WriteString(",")
	buf.WriteString("\"rawResults\": ")
	tmp, err = json.Marshal(strct.RawResults)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// Marshal the "results" field
	buf.WriteString(",")
	buf.WriteString("\"results\": ")
	tmp, err = json.Marshal(strct.Results)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "Versions" field is required
	if strct.Versions == nil {
		return nil, errors.New("versions is a required field")
	}
	// Marshal the "versions" field
	buf.WriteString(",")
	buf.WriteString("\"versions\": ")
	tmp, err = json.Marshal(strct.Versions)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)

	buf.WriteString("}")
	rv := buf.Bytes()
	return rv, nil
}

//nolint:funlen
func (strct *Claim) UnmarshalJSON(b []byte) error {
	configurationsReceived := false
	metadataReceived := false
	nodesReceived := false
	rawResultsReceived := false
	versionsReceived := false
	var jsonMap map[string]json.RawMessage
	if err := json.Unmarshal(b, &jsonMap); err != nil {
		return err
	}
	// parse all the defined properties
	for k, v := range jsonMap {
		switch k {
		case "configurations":
			if err := json.Unmarshal([]byte(v), &strct.Configurations); err != nil {
				return err
			}
			configurationsReceived = true
		case "metadata":
			if err := json.Unmarshal([]byte(v), &strct.Metadata); err != nil {
				return err
			}
			metadataReceived = true
		case "nodes":
			if err := json.Unmarshal([]byte(v), &strct.Nodes); err != nil {
				return err
			}
			nodesReceived = true
		case "rawResults":
			if err := json.Unmarshal([]byte(v), &strct.RawResults); err != nil {
				return err
			}
			rawResultsReceived = true
		case "results":
			if err := json.Unmarshal([]byte(v), &strct.Results); err != nil {
				return err
			}
		case "versions":
			if err := json.Unmarshal([]byte(v), &strct.Versions); err != nil {
				return err
			}
			versionsReceived = true
		default:
			return errors.New("additional property not allowed: \"" + k + "\"")
		}
	}
	// check if configurations (a required property) was received
	if !configurationsReceived {
		return errors.New("\"configurations\" is required but was not present")
	}
	// check if metadata (a required property) was received
	if !metadataReceived {
		return errors.New("\"metadata\" is required but was not present")
	}
	// check if nodes (a required property) was received
	if !nodesReceived {
		return errors.New("\"nodes\" is required but was not present")
	}
	// check if rawResults (a required property) was received
	if !rawResultsReceived {
		return errors.New("\"rawResults\" is required but was not present")
	}
	// check if versions (a required property) was received
	if !versionsReceived {
		return errors.New("\"versions\" is required but was not present")
	}
	return nil
}

func (strct *Identifier) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.WriteString("{")
	// "Id" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "id" field
	buf.WriteString("\"id\": ")
	tmp, err := json.Marshal(strct.Id)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "Suite" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "suite" field
	buf.WriteString(",")
	buf.WriteString("\"suite\": ")
	tmp, err = json.Marshal(strct.Suite)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// Marshal the "tags" field
	buf.WriteString(",")
	buf.WriteString("\"tags\": ")
	tmp, err = json.Marshal(strct.Tags)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)

	buf.WriteString("}")
	rv := buf.Bytes()
	return rv, nil
}

func (strct *Identifier) UnmarshalJSON(b []byte) error {
	idReceived := false
	suiteReceived := false
	var jsonMap map[string]json.RawMessage
	if err := json.Unmarshal(b, &jsonMap); err != nil {
		return err
	}
	// parse all the defined properties
	for k, v := range jsonMap {
		switch k {
		case "id":
			if err := json.Unmarshal([]byte(v), &strct.Id); err != nil {
				return err
			}
			idReceived = true
		case "suite":
			if err := json.Unmarshal([]byte(v), &strct.Suite); err != nil {
				return err
			}
			suiteReceived = true
		case "tags":
			if err := json.Unmarshal([]byte(v), &strct.Tags); err != nil {
				return err
			}
		default:
			return errors.New("additional property not allowed: \"" + k + "\"")
		}
	}
	// check if id (a required property) was received
	if !idReceived {
		return errors.New("\"id\" is required but was not present")
	}
	// check if suite (a required property) was received
	if !suiteReceived {
		return errors.New("\"suite\" is required but was not present")
	}
	return nil
}

func (strct *Metadata) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.WriteString("{")
	// "EndTime" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "endTime" field
	buf.WriteString("\"endTime\": ")
	tmp, err := json.Marshal(strct.EndTime)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "StartTime" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "startTime" field
	buf.WriteString(",")
	buf.WriteString("\"startTime\": ")
	tmp, err = json.Marshal(strct.StartTime)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)

	buf.WriteString("}")
	rv := buf.Bytes()
	return rv, nil
}

func (strct *Metadata) UnmarshalJSON(b []byte) error {
	endTimeReceived := false
	startTimeReceived := false
	var jsonMap map[string]json.RawMessage
	if err := json.Unmarshal(b, &jsonMap); err != nil {
		return err
	}
	// parse all the defined properties
	for k, v := range jsonMap {
		switch k {
		case "endTime": //nolint:goconst
			if err := json.Unmarshal([]byte(v), &strct.EndTime); err != nil {
				return err
			}
			endTimeReceived = true
		case "startTime": //nolint:goconst
			if err := json.Unmarshal([]byte(v), &strct.StartTime); err != nil {
				return err
			}
			startTimeReceived = true
		default:
			return errors.New("additional property not allowed: \"" + k + "\"")
		}
	}
	// check if endTime (a required property) was received
	if !endTimeReceived {
		return errors.New("\"endTime\" is required but was not present")
	}
	// check if startTime (a required property) was received
	if !startTimeReceived {
		return errors.New("\"startTime\" is required but was not present")
	}
	return nil
}

//nolint:funlen
func (strct *Result) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.WriteString("{")
	// "CapturedTestOutput" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "capturedTestOutput" field
	buf.WriteString("\"capturedTestOutput\": ")
	tmp, err := json.Marshal(strct.CapturedTestOutput)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "CatalogInfo" field is required
	if strct.CatalogInfo == nil {
		return nil, errors.New("catalogInfo is a required field")
	}
	// Marshal the "catalogInfo" field
	buf.WriteString(",")
	buf.WriteString("\"catalogInfo\": ")
	tmp, err = json.Marshal(strct.CatalogInfo)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "CategoryClassification" field is required
	if strct.CategoryClassification == nil {
		return nil, errors.New("categoryClassification is a required field")
	}
	// Marshal the "categoryClassification" field
	buf.WriteString(",")
	buf.WriteString("\"categoryClassification\": ")
	tmp, err = json.Marshal(strct.CategoryClassification)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "Duration" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "duration" field
	buf.WriteString(",")
	buf.WriteString("\"duration\": ")
	tmp, err = json.Marshal(strct.Duration)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// Marshal the "endTime" field
	buf.WriteString(",")
	buf.WriteString("\"endTime\": ")
	tmp, err = json.Marshal(strct.EndTime)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "FailureLineContent" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "failureLineContent" field
	buf.WriteString(",")
	buf.WriteString("\"failureLineContent\": ")
	tmp, err = json.Marshal(strct.FailureLineContent)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "FailureLocation" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "failureLocation" field
	buf.WriteString(",")
	buf.WriteString("\"failureLocation\": ")
	tmp, err = json.Marshal(strct.FailureLocation)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "FailureReason" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "failureReason" field
	buf.WriteString(",")
	buf.WriteString("\"failureReason\": ")
	tmp, err = json.Marshal(strct.FailureReason)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "StartTime" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "startTime" field
	buf.WriteString(",")
	buf.WriteString("\"startTime\": ")
	tmp, err = json.Marshal(strct.StartTime)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "State" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "state" field
	buf.WriteString(",")
	buf.WriteString("\"state\": ")
	tmp, err = json.Marshal(strct.State)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "TestID" field is required
	if strct.TestID == nil {
		return nil, errors.New("testID is a required field")
	}
	// Marshal the "testID" field
	buf.WriteString(",")
	buf.WriteString("\"testID\": ")
	tmp, err = json.Marshal(strct.TestID)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)

	buf.WriteString("}")
	rv := buf.Bytes()
	return rv, nil
}

//nolint:funlen,gocyclo
func (strct *Result) UnmarshalJSON(b []byte) error {
	capturedTestOutputReceived := false
	catalogInfoReceived := false
	categoryClassificationReceived := false
	durationReceived := false
	failureLineContentReceived := false
	failureLocationReceived := false
	failureReasonReceived := false
	startTimeReceived := false
	stateReceived := false
	testIDReceived := false
	var jsonMap map[string]json.RawMessage
	if err := json.Unmarshal(b, &jsonMap); err != nil {
		return err
	}
	// parse all the defined properties
	for k, v := range jsonMap {
		switch k {
		case "capturedTestOutput":
			if err := json.Unmarshal([]byte(v), &strct.CapturedTestOutput); err != nil {
				return err
			}
			capturedTestOutputReceived = true
		case "catalogInfo":
			if err := json.Unmarshal([]byte(v), &strct.CatalogInfo); err != nil {
				return err
			}
			catalogInfoReceived = true
		case "categoryClassification":
			if err := json.Unmarshal([]byte(v), &strct.CategoryClassification); err != nil {
				return err
			}
			categoryClassificationReceived = true
		case "duration":
			if err := json.Unmarshal([]byte(v), &strct.Duration); err != nil {
				return err
			}
			durationReceived = true
		case "endTime":
			if err := json.Unmarshal([]byte(v), &strct.EndTime); err != nil {
				return err
			}
		case "failureLineContent":
			if err := json.Unmarshal([]byte(v), &strct.FailureLineContent); err != nil {
				return err
			}
			failureLineContentReceived = true
		case "failureLocation":
			if err := json.Unmarshal([]byte(v), &strct.FailureLocation); err != nil {
				return err
			}
			failureLocationReceived = true
		case "failureReason":
			if err := json.Unmarshal([]byte(v), &strct.FailureReason); err != nil {
				return err
			}
			failureReasonReceived = true
		case "startTime":
			if err := json.Unmarshal([]byte(v), &strct.StartTime); err != nil {
				return err
			}
			startTimeReceived = true
		case "state":
			if err := json.Unmarshal([]byte(v), &strct.State); err != nil {
				return err
			}
			stateReceived = true
		case "testID":
			if err := json.Unmarshal([]byte(v), &strct.TestID); err != nil {
				return err
			}
			testIDReceived = true
		default:
			return errors.New("additional property not allowed: \"" + k + "\"")
		}
	}
	// check if capturedTestOutput (a required property) was received
	if !capturedTestOutputReceived {
		return errors.New("\"capturedTestOutput\" is required but was not present")
	}
	// check if catalogInfo (a required property) was received
	if !catalogInfoReceived {
		return errors.New("\"catalogInfo\" is required but was not present")
	}
	// check if categoryClassification (a required property) was received
	if !categoryClassificationReceived {
		return errors.New("\"categoryClassification\" is required but was not present")
	}
	// check if duration (a required property) was received
	if !durationReceived {
		return errors.New("\"duration\" is required but was not present")
	}
	// check if failureLineContent (a required property) was received
	if !failureLineContentReceived {
		return errors.New("\"failureLineContent\" is required but was not present")
	}
	// check if failureLocation (a required property) was received
	if !failureLocationReceived {
		return errors.New("\"failureLocation\" is required but was not present")
	}
	// check if failureReason (a required property) was received
	if !failureReasonReceived {
		return errors.New("\"failureReason\" is required but was not present")
	}
	// check if startTime (a required property) was received
	if !startTimeReceived {
		return errors.New("\"startTime\" is required but was not present")
	}
	// check if state (a required property) was received
	if !stateReceived {
		return errors.New("\"state\" is required but was not present")
	}
	// check if testID (a required property) was received
	if !testIDReceived {
		return errors.New("\"testID\" is required but was not present")
	}
	return nil
}

func (strct *Root) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.WriteString("{")
	// "Claim" field is required
	if strct.Claim == nil {
		return nil, errors.New("claim is a required field")
	}
	// Marshal the "claim" field
	buf.WriteString("\"claim\": ")
	tmp, err := json.Marshal(strct.Claim)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)

	buf.WriteString("}")
	rv := buf.Bytes()
	return rv, nil
}

func (strct *Root) UnmarshalJSON(b []byte) error {
	claimReceived := false
	var jsonMap map[string]json.RawMessage
	if err := json.Unmarshal(b, &jsonMap); err != nil {
		return err
	}
	// parse all the defined properties
	for k, v := range jsonMap {
		switch k {
		case "claim":
			if err := json.Unmarshal([]byte(v), &strct.Claim); err != nil {
				return err
			}
			claimReceived = true
		default:
			return errors.New("additional property not allowed: \"" + k + "\"")
		}
	}
	// check if claim (a required property) was received
	if !claimReceived {
		return errors.New("\"claim\" is required but was not present")
	}
	return nil
}

func (strct *Versions) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.WriteString("{")
	// "ClaimFormat" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "claimFormat" field
	buf.WriteString("\"claimFormat\": ")
	tmp, err := json.Marshal(strct.ClaimFormat)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// Marshal the "k8s" field
	buf.WriteString(",")
	buf.WriteString("\"k8s\": ")
	tmp, err = json.Marshal(strct.K8s)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// Marshal the "ocClient" field
	buf.WriteString(",")
	buf.WriteString("\"ocClient\": ")
	tmp, err = json.Marshal(strct.OcClient)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// Marshal the "ocp" field
	buf.WriteString(",")
	buf.WriteString("\"ocp\": ")
	tmp, err = json.Marshal(strct.Ocp)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// "Tnf" field is required
	// only required object types supported for marshal checking (for now)
	// Marshal the "tnf" field
	buf.WriteString(",")
	buf.WriteString("\"tnf\": ")
	tmp, err = json.Marshal(strct.Tnf)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)
	// Marshal the "tnfGitCommit" field
	buf.WriteString(",")
	buf.WriteString("\"tnfGitCommit\": ")
	tmp, err = json.Marshal(strct.TnfGitCommit)
	if err != nil {
		return nil, err
	}
	buf.Write(tmp)

	buf.WriteString("}")
	rv := buf.Bytes()
	return rv, nil
}

func (strct *Versions) UnmarshalJSON(b []byte) error {
	claimFormatReceived := false
	tnfReceived := false
	var jsonMap map[string]json.RawMessage
	if err := json.Unmarshal(b, &jsonMap); err != nil {
		return err
	}
	// parse all the defined properties
	for k, v := range jsonMap {
		switch k {
		case "claimFormat":
			if err := json.Unmarshal([]byte(v), &strct.ClaimFormat); err != nil {
				return err
			}
			claimFormatReceived = true
		case "k8s":
			if err := json.Unmarshal([]byte(v), &strct.K8s); err != nil {
				return err
			}
		case "ocClient":
			if err := json.Unmarshal([]byte(v), &strct.OcClient); err != nil {
				return err
			}
		case "ocp":
			if err := json.Unmarshal([]byte(v), &strct.Ocp); err != nil {
				return err
			}
		case "tnf":
			if err := json.Unmarshal([]byte(v), &strct.Tnf); err != nil {
				return err
			}
			tnfReceived = true
		case "tnfGitCommit":
			if err := json.Unmarshal([]byte(v), &strct.TnfGitCommit); err != nil {
				return err
			}
		default:
			return errors.New("additional property not allowed: \"" + k + "\"")
		}
	}
	// check if claimFormat (a required property) was received
	if !claimFormatReceived {
		return errors.New("\"claimFormat\" is required but was not present")
	}
	// check if tnf (a required property) was received
	if !tnfReceived {
		return errors.New("\"tnf\" is required but was not present")
	}
	return nil
}
