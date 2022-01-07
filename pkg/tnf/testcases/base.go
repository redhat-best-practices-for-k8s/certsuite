// Copyright (C) 2020-2021 Red Hat, Inc.
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

package testcases

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf/testcases/data/cnf"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf/testcases/data/operator"
	"gopkg.in/yaml.v2"
)

// StatusFunctionType type for holding function name for expected status in the test template
type StatusFunctionType string

const (
	// ServiceAccountFn  function name to be called to replace expected status
	ServiceAccountFn StatusFunctionType = "FN_SERVICE_ACCOUNT_NAME"
	// ConfiguredTestFile the name for the default "container" test list file
	ConfiguredTestFile = "testconfigure.yml"
)

// TestResultType Defines Test Result Type
type TestResultType string

const (
	// StringType the return result is of type string
	StringType TestResultType = "string"
	// ArrayType the return result is of type array
	ArrayType TestResultType = "array"
	// IntType the return result is of type int
	IntType TestResultType = "int"
)

// TestAction Action performed on the test result
type TestAction string

const (
	// Allow if the test result matches the expected result then allow
	Allow TestAction = "allow"
	// Deny if the test result matches the expected result then deny
	Deny TestAction = "deny"
)

// TestExpectedType defines what type is the value set for expected type
type TestExpectedType string

const (
	// RegEx  the result type will not be modified for regex
	RegEx TestExpectedType = "regex"
	// Function the expected result will be overridden by the function name defined under ExpectedType
	Function TestExpectedType = "function"
	// String  the result type will not be modified for string
	String TestExpectedType = "string"
)

// TestSpecType defines  spec type is of type Operator or CNF
type TestSpecType string

const (
	// Operator to load operator spec from config
	Operator TestSpecType = "operator"
	// Cnf to load CNF spec from config
	Cnf TestSpecType = "CNF"
)
const (
	// GatherFacts is name of the test case template for  gathering pod facts
	GatherFacts = "GATHER_FACTS_POD"
	// PrivilegedPod is name of the test case template for running pod privilege tests
	PrivilegedPod = "PRIVILEGED_POD"
	// PrivilegedRoles is name of the test case template for running cluster roles and permission tests
	PrivilegedRoles = "PRIVILEGED_ROLE"
	// OperatorStatus checks if csv for a given operator is installed
	OperatorStatus = "OPERATOR_STATUS"
)

// PodFactType type to hold container fact types
type PodFactType string

const (
	// ServiceAccountName - for k8s service account name
	ServiceAccountName PodFactType = "SERVICE_ACCOUNT_NAME"
	// Name for pod name
	Name PodFactType = "NAME"
	// NameSpace for pod namespace
	NameSpace PodFactType = "NAMESPACE"
	// ClusterRole for cluster roles
	ClusterRole PodFactType = "CLUSTER_ROLE"
	// ContainerCount for count of containers in the pod
	ContainerCount PodFactType = "CONTAINER_COUNT"
)

// PodFact struct to store pod facts
type PodFact struct {
	// Name of the pod under test
	Name string
	// Namespace of the pod under test
	Namespace string
	// ServiceAccount name used by the pod
	ServiceAccount string
	// ContainerCount is the count of containers inside the pod
	ContainerCount int
	// Exists if the pod is found in the cluster
	Exists bool
}

// PodTestTemplateDataMap  is map of available json data test case templates
var PodTestTemplateDataMap = map[string]string{
	GatherFacts:     cnf.GatherPodFactsJSON,
	PrivilegedPod:   cnf.PrivilegedPodJSON,
	PrivilegedRoles: cnf.RolesJSON,
}

// OperatorTestTemplateDataMap  is map of available json data test case templates
var OperatorTestTemplateDataMap = map[string]string{
	OperatorStatus: operator.OperatorJSON,
}

// CnfTestTemplateFileMap is map of configured test case filenames
var CnfTestTemplateFileMap = map[string]string{
	PrivilegedPod:   "privilegedpod.yml",
	PrivilegedRoles: "privilegedroles.yml",
	"ReadMeTxt":     "readme.txt",
}

// OperatorTestTemplateFileMap is map of configured test case filenames
var OperatorTestTemplateFileMap = map[string]string{
	OperatorStatus: "operatorstatus.yml",
	"ReadMeTxt":    "readme.txt",
}

// RegExType holds regex constant name.
type RegExType string

const (
	// AllowEmpty Allows the result to match empty string
	AllowEmpty RegExType = "ALLOW_EMPTY"
	// AllowAll Allows the result to match non empty string
	AllowAll RegExType = "allowAll"
	// EmptyNullFalse Allows the result to match either empty,null or false string
	EmptyNullFalse RegExType = "EMPTY_NULL_FALSE"
	// NullFalse Allows the result to match either null or false string
	NullFalse RegExType = "NULL_FALSE"
	// True Allows the result to match `true` string
	True RegExType = "TRUE"
	// Null Allows the result to match `null` string
	Null RegExType = "NULL"
	// Zero Allows the result to match 0 number
	Zero RegExType = "ZERO"
	// NonZeroNumber Allows the result to match non 0 number
	NonZeroNumber RegExType = "NON_ZERO_NUMBER"
	// Error Allows the result to match error string
	Error RegExType = "ERROR"
	// Digit Allows the result to match any number
	Digit RegExType = "DIGIT"
)

// outRegExp types of available regular expression to parse matched result
var outRegExp = map[RegExType]string{
	AllowEmpty:     `(.*?)`,
	AllowAll:       `.+`,
	EmptyNullFalse: `^\b(null|false)\b$|^$`,
	NullFalse:      `^\b(null|false)\b$`,
	True:           `^\b(true)\b$`,
	Null:           `^\b(null)\b$`,
	Zero:           `0`,
	NonZeroNumber:  `^(0*([1-9]\d*)|null)$`,
	Error:          "error",
	Digit:          `\d`,
}

// BaseTestCaseConfigSpec slcie of test configurations template
type BaseTestCaseConfigSpec struct {
	// TestCase, Is the list of test cases that available along with their test steps
	TestCase []BaseTestCase `yaml:"testcase" json:"testcase"`
}

// BaseTestCase spec of available test template
type BaseTestCase struct {
	// Name, Is the test case step name
	Name string `yaml:"name" json:"name"`
	// SkipTest, Is set to true by default. This is overridden in the test configuration file by defining test cases
	SkipTest bool `yaml:"skiptest" json:"skiptest"`
	// Loop, Indicates whether the testing resource has multiple objects to iterate and the number indicates how many iterable objects.
	// This value is set by gather facts test case
	Loop int `yaml:"loop" json:"loop"`
	// Command, Is the actual command  to be executed on the testing subject
	Command string `yaml:"command" json:"command"`
	// ExpectedType, Is to identify what type of data is set in ExpectedStatus(function or string)
	ExpectedType TestExpectedType `yaml:"expectedtype" json:"expectedtype"`
	// ExpectedStatus, Is a list of strings that are expected to receive from the command execution
	ExpectedStatus []string `yaml:"expectedstatus" json:"expectedstatus"`
	// ResultType, Is the type of result that is expected from the execution of the command (String,Array, Int)
	ResultType TestResultType `default:"string" yaml:"resulttype" json:"resulttype"`
	// Action, Defines the type of action to be taken on the result (Allow or Deny)
	Action TestAction `yaml:"action" json:"action"`
}

// ConfiguredTestCase this loads the contents of testconfigured.yml file
type ConfiguredTestCase struct {
	// CnfTest, Is the list of configured cnf test's that will be executed on the test subject
	CnfTest []ConfiguredTest `yaml:"cnftest"`
	// OperatorTest, Is the list of configured operator test that will be executed on the test subject
	OperatorTest []ConfiguredTest `yaml:"operatortest"`
}

// ConfiguredTest list all the test that are configured to run
type ConfiguredTest struct {
	// Name of the configured tests.
	Name string `yaml:"name"`
	// Tests is a list of test steps under each Test case.
	Tests []string `yaml:"tests"`
}

// LoadConfiguredTestFile loads configured test cases to struct
func LoadConfiguredTestFile(filepath string) (c *ConfiguredTestCase, err error) {
	yamlFile, err := os.ReadFile(filepath)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(yamlFile, &c)
	return
}

// RenderTestCaseSpec applies configured test case to template
func (c *ConfiguredTest) RenderTestCaseSpec(testSpecType TestSpecType, testName string) (b *BaseTestCaseConfigSpec, err error) {
	if testSpecType == Operator {
		b, err = LoadOperatorTestCaseSpecs(testName)
	} else {
		b, err = LoadCnfTestCaseSpecs(testName)
	}
	if err != nil {
		return
	}
	for _, elem := range c.Tests { // for each name you have data
		for i, e := range b.TestCase {
			if e.Name == elem {
				b.TestCase[i].SkipTest = false
			}
		}
	}
	return
}

// ContainsConfiguredTest checks whether configured test type is found.
func ContainsConfiguredTest(a []ConfiguredTest, testType string) ConfiguredTest {
	for i, n := range a {
		if testType == n.Name {
			return a[i]
		}
	}
	return ConfiguredTest{}
}

// LoadCnfTestCaseSpecs loads base test template data into a struct
func LoadCnfTestCaseSpecs(name string) (*BaseTestCaseConfigSpec, error) {
	var testCaseConfigSpec BaseTestCaseConfigSpec
	err := json.Unmarshal([]byte(PodTestTemplateDataMap[name]), &testCaseConfigSpec)
	if err != nil {
		return nil, err
	}
	return &testCaseConfigSpec, nil
}

// LoadOperatorTestCaseSpecs loads base test template data into a struct
func LoadOperatorTestCaseSpecs(name string) (testCaseConfigSpec *BaseTestCaseConfigSpec, err error) {
	err = json.Unmarshal([]byte(OperatorTestTemplateDataMap[name]), &testCaseConfigSpec)
	return
}

// LoadTestCaseSpecsFromFile loads base test template files into a struct
func LoadTestCaseSpecsFromFile(name, testCaseDir string, testSpecType TestSpecType) (*BaseTestCaseConfigSpec, error) {
	var file *os.File
	var err error
	testCaseConfigSpec := &BaseTestCaseConfigSpec{}
	var testFile string
	if testSpecType == Cnf {
		testFile = testCaseDir + "/" + CnfTestTemplateFileMap[name]
	} else {
		testFile = testCaseDir + "/" + OperatorTestTemplateFileMap[name]
	}
	if file, err = os.Open(testFile); err != nil {
		return nil, err
	}
	defer file.Close()
	// Init new YAML decode
	d := yaml.NewDecoder(file)
	// Start YAML decoding from file
	if err := d.Decode(&testCaseConfigSpec); err != nil {
		return nil, err
	}
	return testCaseConfigSpec, nil
}

// GetOutRegExp check and get available regular expression to parse
func GetOutRegExp(key RegExType) string {
	if val, ok := outRegExp[key]; ok {
		return val
	}
	return string(key)
}

// ExpectedStatusFn checks for expectedStatus function in the test template and replaces with data from container facts
func (b *BaseTestCase) ExpectedStatusFn(val string, fnType StatusFunctionType) {
	for index, expectedStatus := range b.ExpectedStatus {
		if fnType == StatusFunctionType(expectedStatus) {
			b.ReplaceSAasExpectedStatus(index, val)
			break
		}
	}
}

// ReplaceSAasExpectedStatus replaces dynamic expected status defined in test template via function name
func (b *BaseTestCase) ReplaceSAasExpectedStatus(index int, val string) {
	b.ExpectedStatus[index] = val
}

// IsInFocus matches ginkgo focus strings to description key
func IsInFocus(focus []string, desc string) bool {
	matchesFocus := true
	var focusFilter *regexp.Regexp
	if len(focus) > 0 {
		focusFilter = regexp.MustCompile(strings.Join(focus, "|"))
	}
	if focusFilter != nil {
		matchesFocus = focusFilter.MatchString(desc)
	}
	return matchesFocus
}

// GetConfiguredPodTests loads the `configuredTestFile` and extracts
// the names of test groups from it.
func GetConfiguredPodTests() (cnfTests []string) {
	configuredTests, err := LoadConfiguredTestFile(ConfiguredTestFile)
	if err != nil {
		log.Errorf("failed to load %s, continuing with no tests", ConfiguredTestFile)
		return []string{}
	}
	for _, configuredTest := range configuredTests.CnfTest {
		cnfTests = append(cnfTests, configuredTest.Name)
	}
	log.WithField("cnfTests", cnfTests).Infof("got all tests from %s.", ConfiguredTestFile)
	return cnfTests
}

// GetConfiguredOperatorTests loads the `configuredTestFile` and extracts
// the names of test groups from it.
func GetConfiguredOperatorTests() (operatorTests []string) {
	configuredTests, err := LoadConfiguredTestFile(ConfiguredTestFile)
	if err != nil {
		log.Errorf("failed to load %s, continuing with no tests", ConfiguredTestFile)
		return []string{}
	}
	for _, configuredTest := range configuredTests.OperatorTest {
		operatorTests = append(operatorTests, configuredTest.Name)
	}
	log.WithField("operatorTests", operatorTests).Infof("got all tests from %s.", ConfiguredTestFile)
	return operatorTests
}
