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

package claimhelper

import (
	j "encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"

	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/diagnostics"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/labels"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/versions"
)

const (
	claimFilePermissions                 = 0o644
	CNFFeatureValidationJunitXMLFileName = "validation_junit.xml"
	CNFFeatureValidationReportKey        = "cnf-feature-validation"
	// dateTimeFormatDirective is the directive used to format date/time according to ISO 8601.
	DateTimeFormatDirective = "2006-01-02 15:04:05 -0700 MST"

	// States for test cases
	TestStateFailed  = "failed"
	TestStateSkipped = "skipped"
)

// SkippedMessage signals a skipped claim during processing
//
// This struct holds the text of a message that is omitted from normal output
// and any associated metadata. The Text field contains the raw XML character
// data while Messages stores an optional attribute providing additional
// context. It is used by the claim helper to record items that were
// intentionally left out during certificate claim generation.
type SkippedMessage struct {
	Text     string `xml:",chardata"`
	Messages string `xml:"message,attr,omitempty"`
}

// FailureMessage Represents an error message returned by a claim helper operation
//
// The structure holds the error text as well as optional attributes for the
// message and its type. It is used to convey failure information in XML
// responses, with the Text field containing the main content, while Message and
// Type provide metadata that can be omitted if empty.
type FailureMessage struct {
	Text    string `xml:",chardata"`
	Message string `xml:"message,attr,omitempty"`
	Type    string `xml:"type,attr,omitempty"`
}

// TestCase Holds the results of an individual test run
//
// This structure stores metadata and outcome information for a single test
// case, including its name, class context, execution status, duration, and any
// error output. It also provides optional sub-structures to represent skipped
// or failed executions, enabling detailed reporting in XML format.
type TestCase struct {
	Text      string          `xml:",chardata"`
	Name      string          `xml:"name,attr,omitempty"`
	Classname string          `xml:"classname,attr,omitempty"`
	Status    string          `xml:"status,attr,omitempty"`
	Time      string          `xml:"time,attr,omitempty"`
	SystemErr string          `xml:"system-err,omitempty"`
	Skipped   *SkippedMessage `xml:"skipped"`
	Failure   *FailureMessage `xml:"failure"`
}

// Testsuite Represents the results of a test suite execution
//
// This struct holds metadata about a collection of tests, including counts for
// total tests, failures, errors, skipped and disabled cases. It also stores
// timing information, timestamps, and any properties that may be attached to
// the suite. Each individual test case is captured in a slice of TestCase
// structs, allowing detailed inspection of each test's outcome.
type Testsuite struct {
	Text       string `xml:",chardata"`
	Name       string `xml:"name,attr,omitempty"`
	Package    string `xml:"package,attr,omitempty"`
	Tests      string `xml:"tests,attr,omitempty"`
	Disabled   string `xml:"disabled,attr,omitempty"`
	Skipped    string `xml:"skipped,attr,omitempty"`
	Errors     string `xml:"errors,attr,omitempty"`
	Failures   string `xml:"failures,attr,omitempty"`
	Time       string `xml:"time,attr,omitempty"`
	Timestamp  string `xml:"timestamp,attr,omitempty"`
	Properties struct {
		Text     string `xml:",chardata"`
		Property []struct {
			Text  string `xml:",chardata"`
			Name  string `xml:"name,attr,omitempty"`
			Value string `xml:"value,attr,omitempty"`
		} `xml:"property"`
	} `xml:"properties"`
	Testcase []TestCase `xml:"testcase"`
}

// TestSuitesXML Represents an XML report of test suite results
//
// This struct holds attributes such as the total number of tests, failures,
// disabled tests, errors, and elapsed time for a test run. It also contains a
// nested Testsuite element that provides more detailed information about each
// individual test case. The fields are marshaled into XML with corresponding
// attribute tags.
type TestSuitesXML struct {
	XMLName   xml.Name  `xml:"testsuites"`
	Text      string    `xml:",chardata"`
	Tests     string    `xml:"tests,attr,omitempty"`
	Disabled  string    `xml:"disabled,attr,omitempty"`
	Errors    string    `xml:"errors,attr,omitempty"`
	Failures  string    `xml:"failures,attr,omitempty"`
	Time      string    `xml:"time,attr,omitempty"`
	Testsuite Testsuite `xml:"testsuite"`
}

// ClaimBuilder Creates and writes claim reports in various formats
//
// It gathers test results, populates the claim structure with metadata,
// configurations, and node information, then serializes the data to a file. The
// builder can also reset timestamps or output JUnit XML for CI integration.
// Errors during marshaling or file writing are logged as fatal.
type ClaimBuilder struct {
	claimRoot *claim.Root
}

// NewClaimBuilder Creates a claim builder from test environment
//
// The function accepts a test environment, marshals its configuration into
// JSON, unmarshals it back into a map, and populates a new claim root with
// configurations, node information, and version data. It handles unit test mode
// by skipping marshalling steps. The resulting ClaimBuilder contains the fully
// prepared claim structure for later serialization.
func NewClaimBuilder(env *provider.TestEnvironment) (*ClaimBuilder, error) {
	if os.Getenv("UNIT_TEST") == "true" {
		return &ClaimBuilder{
			claimRoot: CreateClaimRoot(),
		}, nil
	}

	log.Debug("Creating claim file builder.")
	configurations, err := MarshalConfigurations(env)
	if err != nil {
		return nil, fmt.Errorf("configuration node missing because of: %v", err)
	}

	claimConfigurations := map[string]interface{}{}
	UnmarshalConfigurations(configurations, claimConfigurations)

	root := CreateClaimRoot()

	root.Claim.Configurations = claimConfigurations
	root.Claim.Nodes = GenerateNodes()
	root.Claim.Versions = &claim.Versions{
		CertSuite:          versions.GitDisplayRelease,
		CertSuiteGitCommit: versions.GitCommit,
		OcClient:           diagnostics.GetVersionOcClient(),
		Ocp:                diagnostics.GetVersionOcp(),
		K8s:                diagnostics.GetVersionK8s(),
		ClaimFormat:        versions.ClaimFormatVersion,
	}

	return &ClaimBuilder{
		claimRoot: root,
	}, nil
}

// ClaimBuilder.Build generates a claim file with results and timestamps
//
// This method records the current time as the claim's end time, retrieves
// reconciled test results from the database, marshals the complete claim
// structure into JSON, writes that data to the specified output file, and logs
// the creation location. It relies on helper functions for marshalling and file
// writing and uses UTC formatting for consistency.
func (c *ClaimBuilder) Build(outputFile string) {
	endTime := time.Now()

	c.claimRoot.Claim.Metadata.EndTime = endTime.UTC().Format(DateTimeFormatDirective)
	c.claimRoot.Claim.Results = checksdb.GetReconciledResults()

	// Marshal the claim and output to file
	payload := MarshalClaimOutput(c.claimRoot)
	WriteClaimOutput(outputFile, payload)

	log.Info("Claim file created at %s", outputFile)
}

// populateXMLFromClaim Builds a JUnit XML representation of claim test results
//
// The function collects all test IDs from the claim, counts failures and skips,
// and constructs a TestSuitesXML structure with aggregated suite metrics. It
// iterates over sorted test IDs to create individual TestCase entries,
// calculating each case's duration and attaching skipped or failure messages as
// needed. The resulting XML object is returned for marshaling into a file.
//
//nolint:funlen
func populateXMLFromClaim(c claim.Claim, startTime, endTime time.Time) TestSuitesXML {
	const (
		TestSuiteName = "CNF Certification Test Suite"
	)

	// Collector all of the Test IDs
	allTestIDs := []string{}
	for testID := range c.Results {
		allTestIDs = append(allTestIDs, c.Results[testID].TestID.Id)
	}

	// Sort the test IDs
	sort.Strings(allTestIDs)

	xmlOutput := TestSuitesXML{}
	// <testsuites>
	xmlOutput.Tests = strconv.Itoa(len(c.Results))

	// Count all of the failed tests in the suite
	failedTests := 0
	for testID := range c.Results {
		if c.Results[testID].State == TestStateFailed {
			failedTests++
		}
	}

	// Count all of the skipped tests in the suite
	skippedTests := 0
	for testID := range c.Results {
		if c.Results[testID].State == TestStateSkipped {
			skippedTests++
		}
	}

	xmlOutput.Failures = strconv.Itoa(failedTests)
	xmlOutput.Disabled = strconv.Itoa(skippedTests)
	xmlOutput.Errors = strconv.Itoa(0)
	xmlOutput.Time = strconv.FormatFloat(endTime.Sub(startTime).Seconds(), 'f', 5, 64)

	// <testsuite>
	xmlOutput.Testsuite.Name = TestSuiteName
	xmlOutput.Testsuite.Tests = strconv.Itoa(len(c.Results))
	// Counters for failed and skipped tests
	xmlOutput.Testsuite.Failures = strconv.Itoa(failedTests)
	xmlOutput.Testsuite.Skipped = strconv.Itoa(skippedTests)
	xmlOutput.Testsuite.Errors = strconv.Itoa(0)

	xmlOutput.Testsuite.Time = strconv.FormatFloat(endTime.Sub(startTime).Seconds(), 'f', 5, 64)
	xmlOutput.Testsuite.Timestamp = time.Now().UTC().Format(DateTimeFormatDirective)

	// <properties>

	// <testcase>
	// Loop through all of the sorted test IDs
	for _, testID := range allTestIDs {
		testCase := TestCase{}
		testCase.Name = testID
		testCase.Classname = TestSuiteName
		testCase.Status = c.Results[testID].State

		// Clean the time strings to remove the " m=" suffix
		start, err := time.Parse(DateTimeFormatDirective, strings.Split(c.Results[testID].StartTime, " m=")[0])
		if err != nil {
			log.Error("Failed to parse start time: %v", err)
		}
		end, err := time.Parse(DateTimeFormatDirective, strings.Split(c.Results[testID].EndTime, " m=")[0])
		if err != nil {
			log.Error("Failed to parse end time: %v", err)
		}

		// Calculate the duration of the test case
		difference := end.Sub(start)
		testCase.Time = strconv.FormatFloat(difference.Seconds(), 'f', 10, 64)

		// Populate the skipped message if the test case was skipped
		if testCase.Status == TestStateSkipped {
			testCase.Skipped = &SkippedMessage{}
			testCase.Skipped.Text = c.Results[testID].SkipReason
		} else {
			testCase.Skipped = nil
		}

		// Populate the failure message if the test case failed
		if testCase.Status == TestStateFailed {
			testCase.Failure = &FailureMessage{}
			testCase.Failure.Text = c.Results[testID].CheckDetails
		} else {
			testCase.Failure = nil
		}

		// Append the test case to the test suite
		xmlOutput.Testsuite.Testcase = append(xmlOutput.Testsuite.Testcase, testCase)
	}

	return xmlOutput
}

// ClaimBuilder.ToJUnitXML Generate a JUnit XML file from claim data
//
// This method builds a structured JUnit XML representation of the current claim
// results, marshals it into indented XML, and writes it to the specified file
// path with appropriate permissions. It logs progress and aborts execution if
// marshalling or file writing fails.
func (c *ClaimBuilder) ToJUnitXML(outputFile string, startTime, endTime time.Time) {
	// Create the JUnit XML file from the claim output.
	xmlOutput := populateXMLFromClaim(*c.claimRoot.Claim, startTime, endTime)

	// Write the JUnit XML file.
	payload, err := xml.MarshalIndent(xmlOutput, "", "  ")
	if err != nil {
		log.Fatal("Failed to generate the xml: %v", err)
	}

	log.Info("Writing JUnit XML file: %s", outputFile)
	err = os.WriteFile(outputFile, payload, claimFilePermissions)
	if err != nil {
		log.Fatal("Failed to write the xml file")
	}
}

// ClaimBuilder.Reset Updates the claim's start timestamp
//
// The method assigns the current UTC time, formatted with the predefined
// directive, to the Claim.Metadata.StartTime field of the builder. It performs
// this operation in place and does not return a value.
func (c *ClaimBuilder) Reset() {
	c.claimRoot.Claim.Metadata.StartTime = time.Now().UTC().Format(DateTimeFormatDirective)
}

// MarshalConfigurations Converts test environment data into JSON bytes
//
// This routine accepts a pointer to the test configuration structure, falls
// back to a default instance if nil, and marshals it into a JSON byte slice.
// Errors during marshalling are logged as errors and returned for callers to
// handle. The function returns the resulting byte slice along with any error
// encountered.
func MarshalConfigurations(env *provider.TestEnvironment) (configurations []byte, err error) {
	config := env
	if config == nil {
		*config = provider.GetTestEnvironment()
	}
	configurations, err = j.Marshal(&config)
	if err != nil {
		log.Error("Error converting configurations to JSON: %v", err)
		return configurations, err
	}
	return configurations, nil
}

// UnmarshalConfigurations converts a JSON byte stream into a map of configurations
//
// The function takes raw configuration data as a byte slice and decodes it into
// a provided map using the standard JSON unmarshaler. If decoding fails, it
// logs a fatal error and terminates the program. The resulting map is populated
// with key/value pairs representing configuration settings.
func UnmarshalConfigurations(configurations []byte, claimConfigurations map[string]interface{}) {
	err := j.Unmarshal(configurations, &claimConfigurations)
	if err != nil {
		log.Fatal("error unmarshalling configurations: %v", err)
	}
}

// UnmarshalClaim parses a claim file into a structured root object
//
// This function takes raw bytes of a claim file and a pointer to a Root
// structure, attempting to unmarshal the data using JSON decoding. If
// unmarshalling fails, it logs a fatal error and terminates the program. On
// success, the provided Root instance is populated with the decoded
// information.
func UnmarshalClaim(claimFile []byte, claimRoot *claim.Root) {
	err := j.Unmarshal(claimFile, &claimRoot)
	if err != nil {
		log.Fatal("error unmarshalling claim file: %v", err)
	}
}

// ReadClaimFile Reads the contents of a claim file
//
// The function attempts to read a file at the provided path using standard I/O
// operations. It logs any errors encountered during reading but always returns
// the data slice, even if an error occurs, leaving error handling to the
// caller. A log entry records the file path that was accessed.
func ReadClaimFile(claimFileName string) (data []byte, err error) {
	data, err = os.ReadFile(claimFileName)
	if err != nil {
		log.Error("ReadFile failed with err: %v", err)
	}
	log.Info("Reading claim file at path: %s", claimFileName)
	return data, nil
}

// GetConfigurationFromClaimFile extracts test environment configuration from a claim file
//
// The function reads the specified claim file, unmarshals its JSON contents
// into an intermediate structure, then marshals the embedded configuration
// section back to JSON before decoding it into a TestEnvironment object. It
// returns that object and any error encountered during reading or parsing. The
// process uses logging for read failures and ensures errors propagate to the
// caller.
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

// MarshalClaimOutput Serializes a claim structure into formatted JSON
//
// The function receives a pointer to the root of a claim object and attempts to
// marshal it into indented JSON. If marshalling fails, it logs a fatal error
// and terminates the program. On success, it returns the resulting byte slice
// for further use.
func MarshalClaimOutput(claimRoot *claim.Root) []byte {
	payload, err := j.MarshalIndent(claimRoot, "", "  ")
	if err != nil {
		log.Fatal("Failed to generate the claim: %v", err)
	}
	return payload
}

// WriteClaimOutput Saves claim payload to a file
//
// This routine writes a byte slice containing claim data to the specified path
// using standard file permissions. If the write fails, it logs a fatal error
// and terminates the program. The function provides no return value.
func WriteClaimOutput(claimOutputFile string, payload []byte) {
	log.Info("Writing claim data to %s", claimOutputFile)
	err := os.WriteFile(claimOutputFile, payload, claimFilePermissions)
	if err != nil {
		log.Fatal("Error writing claim data:\n%s", string(payload))
	}
}

// GenerateNodes Collects node information for claim files
//
// This function aggregates several pieces of data about the cluster nodes,
// including a JSON representation of each node, CNI plugin details, hardware
// characteristics, and CSI driver status. It retrieves this information by
// calling diagnostic helpers that query the test environment or Kubernetes API.
// The resulting map is returned for inclusion in claim documents.
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

// CreateClaimRoot Initializes a claim root with current UTC timestamp
//
// The function obtains the present moment, formats it as an ISO‑8601 string
// in UTC, and embeds that value into a new claim structure. It returns a
// pointer to this freshly constructed root object for use by higher‑level
// builders.
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

// SanitizeClaimFile Removes results that do not match a labels filter
//
// The function reads the claim file, unmarshals it into a structured claim
// object, and then iterates over each test result. For every result it
// evaluates the provided label expression against the test’s labels; if the
// evaluation fails, that result is deleted from the claim. After filtering, the
// modified claim is written back to the original file path, which is returned
// along with any error encountered during processing.
func SanitizeClaimFile(claimFileName, labelsFilter string) (string, error) {
	log.Info("Sanitizing claim file %s", claimFileName)
	data, err := ReadClaimFile(claimFileName)
	if err != nil {
		log.Error("ReadClaimFile failed with err: %v", err)
		return "", err
	}
	var aRoot claim.Root
	UnmarshalClaim(data, &aRoot)

	// Remove the results that do not match the labels filter
	for testID := range aRoot.Claim.Results {
		evaluator, err := labels.NewLabelsExprEvaluator(labelsFilter)
		if err != nil {
			log.Error("Failed to create labels expression evaluator: %v", err)
			return "", err
		}

		_, gatheredLabels := identifiers.GetTestIDAndLabels(*aRoot.Claim.Results[testID].TestID)

		if !evaluator.Eval(gatheredLabels) {
			log.Info("Removing test ID: %s from the claim", testID)
			delete(aRoot.Claim.Results, testID)
		}
	}

	WriteClaimOutput(claimFileName, MarshalClaimOutput(&aRoot))
	return claimFileName, nil
}
