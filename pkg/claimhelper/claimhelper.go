// Copyright (C) 2020-2022 Red Hat, Inc.
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

package claimhelper

import (
	j "encoding/json"
	"encoding/xml"
	"fmt"
	"path/filepath"
	"strconv"

	"os"
	"time"

	"github.com/test-network-function/cnf-certification-test/internal/log"

	"github.com/test-network-function/cnf-certification-test/pkg/checksdb"
	"github.com/test-network-function/cnf-certification-test/pkg/diagnostics"
	"github.com/test-network-function/cnf-certification-test/pkg/junit"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/versions"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
)

const (
	claimFilePermissions                 = 0o644
	CNFFeatureValidationJunitXMLFileName = "validation_junit.xml"
	CNFFeatureValidationReportKey        = "cnf-feature-validation"
	// dateTimeFormatDirective is the directive used to format date/time according to ISO 8601.
	DateTimeFormatDirective = "2006-01-02T15:04:05+00:00"

	// States for test cases
	TestStateFailed  = "failed"
	TestStateSkipped = "skipped"
)

type SkippedMessage struct {
	Text     string `xml:",chardata"`
	Messages string `xml:"message,attr,omitempty"`
}

type FailureMessage struct {
	Text    string `xml:",chardata"`
	Message string `xml:"message,attr,omitempty"`
	Type    string `xml:"type,attr,omitempty"`
}

type TestCase struct {
	Text      string         `xml:",chardata"`
	Name      string         `xml:"name,attr"`
	Classname string         `xml:"classname,attr"`
	Status    string         `xml:"status,attr"`
	Time      string         `xml:"time,attr"`
	SystemErr string         `xml:"system-err,omitempty"`
	Skipped   SkippedMessage `xml:"skipped,omitempty"`
	Failure   FailureMessage `xml:"failure,omitempty"`
}

type Testsuite struct {
	Text       string `xml:",chardata"`
	Name       string `xml:"name,attr"`
	Package    string `xml:"package,attr"`
	Tests      string `xml:"tests,attr"`
	Disabled   string `xml:"disabled,attr"`
	Skipped    string `xml:"skipped,attr,omitempty"`
	Errors     string `xml:"errors,attr,omitempty"`
	Failures   string `xml:"failures,attr,omitempty"`
	Time       string `xml:"time,attr"`
	Timestamp  string `xml:"timestamp,attr"`
	Properties struct {
		Text     string `xml:",chardata"`
		Property []struct {
			Text  string `xml:",chardata"`
			Name  string `xml:"name,attr"`
			Value string `xml:"value,attr"`
		} `xml:"property"`
	} `xml:"properties"`
	Testcase []TestCase `xml:"testcase"`
}

type TestSuitesXML struct {
	XMLName   xml.Name  `xml:"testsuites"`
	Text      string    `xml:",chardata"`
	Tests     string    `xml:"tests,attr"`
	Disabled  string    `xml:"disabled,attr"`
	Errors    string    `xml:"errors,attr,omitempty"`
	Failures  string    `xml:"failures,attr,omitempty"`
	Time      string    `xml:"time,attr"`
	Testsuite Testsuite `xml:"testsuite"`
}

type ClaimBuilder struct {
	claimRoot *claim.Root
}

func NewClaimBuilder() (*ClaimBuilder, error) {
	log.Debug("Creating claim file builder.")
	configurations, err := MarshalConfigurations()
	if err != nil {
		return nil, fmt.Errorf("configuration node missing because of: %v", err)
	}

	claimConfigurations := map[string]interface{}{}
	UnmarshalConfigurations(configurations, claimConfigurations)

	root := CreateClaimRoot()

	root.Claim.Configurations = claimConfigurations
	root.Claim.Nodes = GenerateNodes()
	root.Claim.Versions = &claim.Versions{
		Tnf:          versions.GitDisplayRelease,
		TnfGitCommit: versions.GitCommit,
		OcClient:     diagnostics.GetVersionOcClient(),
		Ocp:          diagnostics.GetVersionOcp(),
		K8s:          diagnostics.GetVersionK8s(),
		ClaimFormat:  versions.ClaimFormatVersion,
	}

	return &ClaimBuilder{
		claimRoot: root,
	}, nil
}

func (c *ClaimBuilder) Build(outputFile string) {
	endTime := time.Now()

	c.claimRoot.Claim.Metadata.EndTime = endTime.UTC().Format(DateTimeFormatDirective)
	c.claimRoot.Claim.Results = checksdb.GetReconciledResults()

	// Marshal the claim and output to file
	payload := MarshalClaimOutput(c.claimRoot)
	WriteClaimOutput(outputFile, payload)

	log.Info("Claim file created at %s", outputFile)
}

//nolint:funlen
func populateXMLFromClaim(c claim.Claim, startTime, endTime time.Time) TestSuitesXML {
	const (
		TestSuiteName = "CNF Certification Test Suite"
	)
	xmlOutput := TestSuitesXML{}
	// <testsuites>
	xmlOutput.Tests = strconv.Itoa(len(c.Results))

	// Count all of the failed tests in the suite
	failedTests := 0
	for _, result := range c.Results {
		typedResult := result.(claim.Result)
		if typedResult.State == TestStateFailed {
			failedTests++
		}
	}

	// Count all of the skipped tests in the suite
	skippedTests := 0
	for _, result := range c.Results {
		typedResult := result.(claim.Result)
		if typedResult.State == TestStateSkipped {
			skippedTests++
		}
	}

	xmlOutput.Failures = strconv.Itoa(failedTests)
	xmlOutput.Disabled = strconv.Itoa(skippedTests)
	xmlOutput.Errors = strconv.Itoa(0)
	xmlOutput.Time = strconv.FormatFloat(endTime.Sub(startTime).Seconds(), 'f', 3, 64)

	// <testsuite>
	xmlOutput.Testsuite.Name = TestSuiteName
	xmlOutput.Testsuite.Tests = strconv.Itoa(len(c.Results))
	// Counters for failed and skipped tests
	xmlOutput.Testsuite.Failures = strconv.Itoa(failedTests)
	xmlOutput.Testsuite.Skipped = strconv.Itoa(skippedTests)
	xmlOutput.Testsuite.Errors = strconv.Itoa(0)

	xmlOutput.Testsuite.Time = strconv.FormatFloat(endTime.Sub(startTime).Seconds(), 'f', 3, 64)
	xmlOutput.Testsuite.Timestamp = time.Now().UTC().Format(DateTimeFormatDirective)

	// <properties>

	// <testcase>
	for testID, result := range c.Results {
		// Type the result
		typedResult := result.(claim.Result)

		testCase := TestCase{}
		testCase.Name = testID
		testCase.Classname = TestSuiteName
		testCase.Status = typedResult.State
		testCase.Time = strconv.FormatFloat(float64(typedResult.Duration), 'f', 3, 64)

		// Populate the skipped message if the test case was skipped
		if testCase.Status == TestStateSkipped {
			testCase.Skipped.Text = typedResult.SkipReason
		}

		// Populate the failure message if the test case failed
		if testCase.Status == TestStateFailed {
			testCase.Failure.Text = typedResult.CheckDetails
		}

		// Append the test case to the test suite
		xmlOutput.Testsuite.Testcase = append(xmlOutput.Testsuite.Testcase, testCase)
	}

	return xmlOutput
}

func (c *ClaimBuilder) ToJUnitXML(outputFile string, startTime, endTime time.Time) {
	// Create the JUnit XML file from the claim output.
	xmlOutput := populateXMLFromClaim(*c.claimRoot.Claim, startTime, endTime)

	// Write the JUnit XML file.
	payload, err := xml.MarshalIndent(xmlOutput, "", "  ")
	if err != nil {
		log.Error("Failed to generate the xml: %v", err)
		os.Exit(1)
	}

	log.Info("Writing JUnit XML file: %s", outputFile)
	err = os.WriteFile(outputFile, payload, claimFilePermissions)
	if err != nil {
		log.Error("Failed to write the xml file")
		os.Exit(1)
	}
}

func (c *ClaimBuilder) Reset() {
	c.claimRoot.Claim.Metadata.StartTime = time.Now().UTC().Format(DateTimeFormatDirective)
}

// MarshalConfigurations creates a byte stream representation of the test configurations.  In the event of an error,
// this method fatally fails.
func MarshalConfigurations() (configurations []byte, err error) {
	config := provider.GetTestEnvironment()
	configurations, err = j.Marshal(config)
	if err != nil {
		log.Error("error converting configurations to JSON: %v", err)
		return configurations, err
	}
	return configurations, nil
}

// UnmarshalConfigurations creates a map from configurations byte stream.  In the event of an error, this method fatally
// fails.
func UnmarshalConfigurations(configurations []byte, claimConfigurations map[string]interface{}) {
	err := j.Unmarshal(configurations, &claimConfigurations)
	if err != nil {
		log.Error("error unmarshalling configurations: %v", err)
		os.Exit(1)
	}
}

// UnmarshalClaim unmarshals the claim file
func UnmarshalClaim(claimFile []byte, claimRoot *claim.Root) {
	err := j.Unmarshal(claimFile, &claimRoot)
	if err != nil {
		log.Error("error unmarshalling claim file: %v", err)
		os.Exit(1)
	}
}

// ReadClaimFile writes the output payload to the claim file.  In the event of an error, this method fatally fails.
func ReadClaimFile(claimFileName string) (data []byte, err error) {
	data, err = os.ReadFile(claimFileName)
	if err != nil {
		log.Error("ReadFile failed with err: %v", err)
	}
	path, err := os.Getwd()
	if err != nil {
		log.Error("Getwd failed with err: %v", err)
	}
	log.Info("Reading claim file at path: %s", path)
	return data, nil
}

// GetConfigurationFromClaimFile retrieves configuration details from claim file
func GetConfigurationFromClaimFile(claimFileName string) (env *provider.TestEnvironment, err error) {
	data, err := ReadClaimFile(claimFileName)
	if err != nil {
		log.Error("ReadClaimFile failed with err: %v", err)
		return env, err
	}
	var aRoot claim.Root
	fmt.Printf("%s", data)
	UnmarshalClaim(data, &aRoot)
	configJSON, err := j.Marshal(aRoot.Claim.Configurations)
	if err != nil {
		return nil, fmt.Errorf("cannot convert config to json")
	}
	err = j.Unmarshal(configJSON, &env)
	return env, err
}

// MarshalClaimOutput is a helper function to serialize a claim as JSON for output.  In the event of an error, this
// method fatally fails.
func MarshalClaimOutput(claimRoot *claim.Root) []byte {
	payload, err := j.MarshalIndent(claimRoot, "", "  ")
	if err != nil {
		log.Error("Failed to generate the claim: %v", err)
		os.Exit(1)
	}
	return payload
}

// WriteClaimOutput writes the output payload to the claim file.  In the event of an error, this method fatally fails.
func WriteClaimOutput(claimOutputFile string, payload []byte) {
	err := os.WriteFile(claimOutputFile, payload, claimFilePermissions)
	if err != nil {
		log.Error("Error writing claim data:\n%s", string(payload))
		os.Exit(1)
	}
}

func GenerateNodes() map[string]interface{} {
	const (
		nodeSummaryField = "nodeSummary"
		cniPluginsField  = "cniPlugins"
		nodesHwInfo      = "nodesHwInfo"
		csiDriverInfo    = "csiDriver"
	)
	nodes := map[string]interface{}{}
	nodes[nodeSummaryField] = diagnostics.GetNodeJSON()  // add node summary
	nodes[cniPluginsField] = diagnostics.GetCniPlugins() // add cni plugins
	nodes[nodesHwInfo] = diagnostics.GetHwInfoAllNodes() // add nodes hardware information
	nodes[csiDriverInfo] = diagnostics.GetCsiDriver()    // add csi drivers info
	return nodes
}

// CreateClaimRoot creates the claim based on the model created in
// https://github.com/test-network-function/cnf-certification-test-claim.
func CreateClaimRoot() *claim.Root {
	// Initialize the claim with the start time.
	startTime := time.Now()
	return &claim.Root{
		Claim: &claim.Claim{
			Metadata: &claim.Metadata{
				StartTime: startTime.UTC().Format(DateTimeFormatDirective),
			},
		},
	}
}

// LoadJUnitXMLIntoMap converts junitFilename's XML-formatted JUnit test results into a Go map, and adds the result to
// the result Map.
func LoadJUnitXMLIntoMap(result map[string]interface{}, junitFilename, key string) {
	var err error
	if key == "" {
		var extension = filepath.Ext(junitFilename)
		key = junitFilename[0 : len(junitFilename)-len(extension)]
	}
	result[key], err = junit.ExportJUnitAsMap(junitFilename)
	if err != nil {
		log.Error("error reading JUnit XML file into JSON: %v", err)
		os.Exit(1)
	}
}

// AppendCNFFeatureValidationReportResults is a helper method to add the results of running the cnf-features-deploy
// test suite to the claim file.
func AppendCNFFeatureValidationReportResults(junitPath *string, junitMap map[string]interface{}) {
	cnfFeaturesDeployJUnitFile := filepath.Join(*junitPath, CNFFeatureValidationJunitXMLFileName)
	if _, err := os.Stat(cnfFeaturesDeployJUnitFile); err == nil {
		LoadJUnitXMLIntoMap(junitMap, cnfFeaturesDeployJUnitFile, CNFFeatureValidationReportKey)
	}
}
