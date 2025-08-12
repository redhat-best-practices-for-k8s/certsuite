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

// SkippedMessage holds information about a message that was ignored
// during claim handling. The Messages field contains a concise
// identifier or brief description of the skipped item, while the Text
// field provides a more detailed explanation or context for why it
// was omitted. This struct is used to aggregate and report all
// messages that were intentionally left out from further processing.
type SkippedMessage struct {
	Text     string `xml:",chardata"`
	Messages string `xml:"message,attr,omitempty"`
}

// FailureMessage represents an error message that can be attached to a claim.
//
// It contains a human readable Message, optional detailed Text, and a Type indicating the kind of failure. The fields are exported so they can be marshalled to JSON or other formats when communicating claim status.
type FailureMessage struct {
	Text    string `xml:",chardata"`
	Message string `xml:"message,attr,omitempty"`
	Type    string `xml:"type,attr,omitempty"`
}

// TestCase represents the result of a single test execution.
//
// It holds metadata such as the test name, classname and status,
// along with optional failure or skip messages, system error output,
// test output text, and the duration of the test.
// The fields are used to serialize test results into standard reporting formats.
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

// Testsuite represents the result of a test suite execution.
//
// It contains metadata such as the suite name, package, and timestamps,
// along with aggregated counts for tests, failures, errors, skips, and disabled cases.
// The Text field holds any human‑readable output from the suite run.
// Properties provides additional key/value data, while Testcase is a slice of individual test case results.
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

// TestSuitesXML represents the XML structure of a JUnit test suite report.
//
// It contains summary attributes such as the total number of tests, failures,
// and errors, along with a nested Testsuite element that holds individual
// test case details. The struct is marshalled to XML using the standard
// encoding/xml package, with XMLName specifying the root element name.
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

// ClaimBuilder creates, resets, and exports claim data.
//
// It holds a reference to the root of a claim tree and provides methods
// to build the claim from a test environment, reset its state, and
// generate JUnit XML output for reporting. The builder orchestrates
// time stamping, formatting, and writing of claim artifacts, handling
// errors internally and logging informational messages during each step.
type ClaimBuilder struct {
	claimRoot *claim.Root
}

// NewClaimBuilder creates a ClaimBuilder configured for the given test environment.
//
// It reads environment variables, initializes claim roots, marshals and unmarshals
// configuration data, generates nodes, and fetches version information from
// various components. The function returns a pointer to the constructed ClaimBuilder
// and an error if any step fails.
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

// Build creates a closure that produces the claim report for a given test run.
//
// It records the current UTC time using the standard RFC3339 format,
// retrieves reconciled test results, marshals them into JSON or
// XML based on the feature flag, and writes the output to a file.
// The returned function accepts the test name as its argument
// and performs the write operation, logging progress with Info.
func (c *ClaimBuilder) Build(outputFile string) {
	endTime := time.Now()

	c.claimRoot.Claim.Metadata.EndTime = endTime.UTC().Format(DateTimeFormatDirective)
	c.claimRoot.Claim.Results = checksdb.GetReconciledResults()

	// Marshal the claim and output to file
	payload := MarshalClaimOutput(c.claimRoot)
	WriteClaimOutput(outputFile, payload)

	log.Info("Claim file created at %s", outputFile)
}

// populateXMLFromClaim converts a Claim into a TestSuitesXML representation suitable for JUnit output.
//
// It takes a Claim, the start time of the test run, and the end time,
// producing a structured XML object that includes test case results,
// durations, and overall status information. The returned TestSuitesXML
// contains suite names, timestamps in UTC, and aggregated metrics such as
// total tests, failures, and skipped counts derived from the Claim data.
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

// ToJUnitXML generates a JUnit XML report from a ClaimBuilder instance.
//
// It accepts a filename, a start time, and an end time.
// The method creates the XML representation of the claim data,
// writes it to the specified file with proper permissions,
// logs success or failure using the logger, and returns
// a no‑op function that can be called by the caller when the
// report is no longer needed.
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

// Reset clears the ClaimBuilder and sets default values.
//
// It resets all internal fields of the builder to their zero or
// default state, then initializes the claim with a new timestamp
// in UTC using the standard format. After calling this method,
// the builder is ready for constructing a fresh claim.
func (c *ClaimBuilder) Reset() {
	c.claimRoot.Claim.Metadata.StartTime = time.Now().UTC().Format(DateTimeFormatDirective)
}

// MarshalConfigurations creates a byte stream representation of the test configurations.
//
// It takes a TestEnvironment pointer, marshals its data into JSON,
// and returns the resulting byte slice along with an error if the
// operation fails. If marshalling encounters an error, the function
// will terminate the program by calling Error on the logger.
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

// UnmarshalConfigurations creates a map from configurations byte stream.
// It takes the configuration data as a byte slice and populates the provided map with the parsed values. If parsing fails, it logs a fatal error and terminates the program.
func UnmarshalConfigurations(configurations []byte, claimConfigurations map[string]interface{}) {
	err := j.Unmarshal(configurations, &claimConfigurations)
	if err != nil {
		log.Fatal("error unmarshalling configurations: %v", err)
	}
}

// UnmarshalClaim unmarshals a claim file into a Root structure.
//
// It accepts the raw bytes of a claim file and a pointer to a claim.Root
// where the decoded data will be stored. The function uses json.Unmarshal
// internally; if unmarshalling fails, it logs a fatal error and terminates
// execution. On success, the provided root object is populated with the
// contents of the input bytes.
func UnmarshalClaim(claimFile []byte, claimRoot *claim.Root) {
	err := j.Unmarshal(claimFile, &claimRoot)
	if err != nil {
		log.Fatal("error unmarshalling claim file: %v", err)
	}
}

// ReadClaimFile reads the contents of a claim file.
//
// It takes a single argument which is the file path to read.
// The function returns the raw bytes from the file and an error if one occurred.
// On success the byte slice contains the file data; on failure an empty slice and the error are returned.  
// Logging is performed for informational and error events.
func ReadClaimFile(claimFileName string) (data []byte, err error) {
	data, err = os.ReadFile(claimFileName)
	if err != nil {
		log.Error("ReadFile failed with err: %v", err)
	}
	log.Info("Reading claim file at path: %s", claimFileName)
	return data, nil
}

// GetConfigurationFromClaimFile retrieves configuration details from a claim file.
//
// It reads the specified claim file, parses its contents into a Claim structure,
// extracts the TestEnvironment configuration, and returns it.
// The function returns a pointer to provider.TestEnvironment and an error if any step fails.
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

// MarshalClaimOutput serializes a claim into JSON.
//
// It accepts a pointer to a claim.Root and returns the indented JSON
// representation as a byte slice. If serialization fails, it logs the
// error and terminates the program with a fatal exit.
func MarshalClaimOutput(claimRoot *claim.Root) []byte {
	payload, err := j.MarshalIndent(claimRoot, "", "  ")
	if err != nil {
		log.Fatal("Failed to generate the claim: %v", err)
	}
	return payload
}

// WriteClaimOutput writes the output payload to a claim file.
//
// It accepts the path of the claim file and the payload as a byte slice,
// writes the data to disk with appropriate permissions, logs success,
// and aborts the program on any error.
func WriteClaimOutput(claimOutputFile string, payload []byte) {
	log.Info("Writing claim data to %s", claimOutputFile)
	err := os.WriteFile(claimOutputFile, payload, claimFilePermissions)
	if err != nil {
		log.Fatal("Error writing claim data:\n%s", string(payload))
	}
}

// GenerateNodes builds a map of node information used in claims.
//
// It gathers data from the cluster such as node JSON, CNI plugins,
// hardware info and CSI drivers by calling helper functions.
// The returned map has string keys and interface{} values representing
// various node attributes that are later included in claim files.
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

// CreateClaimRoot creates a claim root object based on the model defined in the certsuite-claim repository.
//
// It constructs a new claim.Root with default values and sets its creation timestamp to the current UTC time.
// The function returns a pointer to the newly created Root instance, ready for further population or serialization.
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

// SanitizeClaimFile removes invalid labels from a claim file and writes a cleaned version.
//
// It reads the input claim file, unmarshals its contents, evaluates label expressions,
// deletes any labels that do not match the expected pattern, and writes the sanitized
// output to the specified destination path. The function returns the path of the
// written file and an error if any step fails.
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
