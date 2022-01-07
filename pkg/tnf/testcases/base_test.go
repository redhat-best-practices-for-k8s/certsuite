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

package testcases_test

import (
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf/testcases"
	"gopkg.in/yaml.v2"
)

var (
	file          *os.File
	err           error
	testConfigure testcases.ConfiguredTestCase
)

const (
	filePerm         = 0644
	testTempFile     = "testconfigure.yml"
	InValidData      = "INVALID_DATA"
	InValidKey       = "INVALID_KEY"
	cnfFilePath      = "./files/cnf"
	operatorFilePath = "./files/operator"
	name             = "testpod"
	invalidFilePath  = "./invalid"
	inValidFile      = "dummy.yaml"
	allowAll         = `.+`
)

func setup() {
	configuredTest := testcases.ConfiguredTest{}
	configuredTest.Name = "PRIVILEGED_POD"
	configuredTest.Tests = []string{"HOST_NETWORK_CHECK", "HOST_PORT_CHECK", "HOST_IPC_CHECK"}
	testConfigure.CnfTest = append(testConfigure.CnfTest, configuredTest)

	file, err = os.CreateTemp(".", testTempFile)
	if err != nil {
		log.Fatal(err)
	}
	bytes, _ := yaml.Marshal(testConfigure)
	err = os.WriteFile(file.Name(), bytes, filePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func teardown() {
	if file != nil {
		os.Remove(file.Name())
	}
}

func TestLoadCNFPodPrivilegedPodTestCaseSpecs(t *testing.T) {
	testCase, err := testcases.LoadCnfTestCaseSpecs(testcases.PrivilegedPod)
	assert.Nil(t, err)
	assert.NotNil(t, testCase)
}

func TestLoadInvalidTestCaseSpecs(t *testing.T) {
	testCase, err := testcases.LoadCnfTestCaseSpecs(InValidData)
	assert.NotNil(t, err)
	assert.Nil(t, testCase)
}

func TestLoadCNFPodPrivilegedRolesTestCaseSpecs(t *testing.T) {
	testCase, err := testcases.LoadCnfTestCaseSpecs(testcases.PrivilegedRoles)
	assert.Nil(t, err)
	assert.NotNil(t, testCase)
}

func TestLoadCNFPodGatherFactsTestCaseSpecs(t *testing.T) {
	testCase, err := testcases.LoadCnfTestCaseSpecs(testcases.GatherFacts)
	assert.Nil(t, err)
	assert.NotNil(t, testCase)
}

func TestLoadOperatorOperatorStatusTestCaseSpecs(t *testing.T) {
	testCase, err := testcases.LoadOperatorTestCaseSpecs(testcases.OperatorStatus)
	assert.Nil(t, err)
	assert.NotNil(t, testCase)
}

func TestLoadCNFTestCaseSpecsFromFile(t *testing.T) {
	testCase, err := testcases.LoadTestCaseSpecsFromFile(testcases.PrivilegedRoles, cnfFilePath, testcases.Cnf)
	assert.Nil(t, err)
	assert.NotNil(t, testCase)
}

func TestLoadOperatorTestCaseSpecsFromFile(t *testing.T) {
	testCase, err := testcases.LoadTestCaseSpecsFromFile(testcases.OperatorStatus, operatorFilePath, testcases.Operator)
	assert.Nil(t, err)
	assert.NotNil(t, testCase)
}

func TestLoadInvalidPathCNFTestCaseSpecsFromFile(t *testing.T) {
	testCase, err := testcases.LoadTestCaseSpecsFromFile(testcases.PrivilegedRoles, invalidFilePath, testcases.Cnf)
	assert.NotNil(t, err)
	assert.Nil(t, testCase)
	_, err = testcases.LoadTestCaseSpecsFromFile("ReadMeTxt", cnfFilePath, testcases.Cnf)
	assert.NotNil(t, err)
	assert.Nil(t, testCase)
}

func TestBaseTestCase_CNFExpectedStatusFn(t *testing.T) {
	var facts = testcases.PodFact{}
	facts.Name = name
	facts.ServiceAccount = "TEST_SERVICE_ACCOUNT_NAME"
	testCase, err := testcases.LoadTestCaseSpecsFromFile(testcases.PrivilegedRoles, cnfFilePath, testcases.Cnf)
	assert.Nil(t, err)
	assert.NotNil(t, testCase)
	testCase.TestCase[0].ExpectedStatusFn(facts.ServiceAccount, testcases.ServiceAccountFn)
	assert.Equal(t, facts.ServiceAccount, testCase.TestCase[0].ExpectedStatus[0])
}

func TestConfiguredTest_Operator_RenderTestCaseSpec(t *testing.T) {
	var c = testcases.ConfiguredTest{}
	c.Name = "OPERATOR_STATUS"
	c.Tests = []string{"CSV_INSTALLED", "CSV_SCC"}
	b, err := c.RenderTestCaseSpec(testcases.Operator, testcases.OperatorStatus)
	assert.Nil(t, err)
	assert.NotNil(t, b)
	assert.Equal(t, "CSV_INSTALLED", b.TestCase[0].Name)
	assert.Equal(t, "CSV_SCC", b.TestCase[1].Name)

	c.Name = "PRIVILEGED_POD"
	c.Tests = []string{"HOST_NETWORK_CHECK"}
	b, err = c.RenderTestCaseSpec(testcases.Cnf, testcases.PrivilegedPod)
	assert.Nil(t, err)
	assert.NotNil(t, b)
	assert.Equal(t, "HOST_NETWORK_CHECK", b.TestCase[0].Name)

	b, err = c.RenderTestCaseSpec(testcases.Cnf, InValidKey)
	assert.NotNil(t, err)
	assert.Nil(t, b)
}

func TestConfiguredTest_CNF_RenderTestCaseSpec(t *testing.T) {
	var c = testcases.ConfiguredTest{}
	c.Name = "PRIVILEGED_POD"
	c.Tests = []string{"HOST_NETWORK_CHECK"}
	b, err := c.RenderTestCaseSpec(testcases.Cnf, testcases.PrivilegedPod)
	assert.Nil(t, err)
	assert.NotNil(t, b)
	assert.Equal(t, "HOST_NETWORK_CHECK", b.TestCase[0].Name)
}

func TestGetOutRegExp(t *testing.T) {
	assert.Equal(t, allowAll, testcases.GetOutRegExp(testcases.AllowAll))
	assert.Equal(t, InValidKey, testcases.GetOutRegExp(InValidKey))
}

func TestLoadConfiguredTestFile(t *testing.T) {
	setup()
	defer (teardown)()
	b, e := testcases.LoadConfiguredTestFile(file.Name())
	assert.Nil(t, e)
	assert.NotNil(t, b)

	b, e = testcases.LoadConfiguredTestFile(inValidFile)
	assert.NotNil(t, e)
	assert.Nil(t, b)
}

func TestContainsConfiguredTest(t *testing.T) {
	setup()
	defer (teardown)()
	b, e := testcases.LoadConfiguredTestFile(file.Name())
	assert.Nil(t, e)
	assert.NotNil(t, b)

	c := testcases.ContainsConfiguredTest(b.CnfTest, "PRIVILEGED_POD")
	assert.NotNil(t, c.Name)
	assert.Equal(t, "PRIVILEGED_POD", c.Name)
	// Invalid Key
	c = testcases.ContainsConfiguredTest(b.CnfTest, "PRIVILEGED_POD_INVALID")
	assert.NotNil(t, c.Name)
	assert.Equal(t, reflect.DeepEqual(c, testcases.ConfiguredTest{}), true)
}

func TestContainsConfiguredTestForInvalidKey(t *testing.T) {
	setup()
	defer (teardown)()
	b, e := testcases.LoadConfiguredTestFile(file.Name())
	assert.Nil(t, e)
	assert.NotNil(t, b)
	// Invalid Key
	c := testcases.ContainsConfiguredTest(b.CnfTest, "PRIVILEGED_POD_INVALID")
	assert.NotNil(t, c.Name)
	assert.Equal(t, reflect.DeepEqual(c, testcases.ConfiguredTest{}), true)
}
