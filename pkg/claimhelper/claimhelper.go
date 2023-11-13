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
	"fmt"
	"path/filepath"

	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/test-network-function/cnf-certification-test/pkg/claim"
	"github.com/test-network-function/cnf-certification-test/pkg/diagnostics"
	"github.com/test-network-function/cnf-certification-test/pkg/junit"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

const (
	claimFilePermissions                 = 0o644
	CNFFeatureValidationJunitXMLFileName = "validation_junit.xml"
	CNFFeatureValidationReportKey        = "cnf-feature-validation"
	// dateTimeFormatDirective is the directive used to format date/time according to ISO 8601.
	DateTimeFormatDirective = "2006-01-02T15:04:05+00:00"
)

// MarshalConfigurations creates a byte stream representation of the test configurations.  In the event of an error,
// this method fatally fails.
func MarshalConfigurations() (configurations []byte, err error) {
	config := provider.GetTestEnvironment()
	configurations, err = j.Marshal(config)
	if err != nil {
		log.Errorf("error converting configurations to JSON: %v", err)
		return configurations, err
	}
	return configurations, nil
}

// UnmarshalConfigurations creates a map from configurations byte stream.  In the event of an error, this method fatally
// fails.
func UnmarshalConfigurations(configurations []byte, claimConfigurations map[string]interface{}) {
	err := j.Unmarshal(configurations, &claimConfigurations)
	if err != nil {
		log.Fatalf("error unmarshalling configurations: %v", err)
	}
}

// UnmarshalClaim unmarshals the claim file
func UnmarshalClaim(claimFile []byte, claimRoot *claim.Root) {
	err := j.Unmarshal(claimFile, &claimRoot)
	if err != nil {
		log.Fatalf("error unmarshalling claim file: %v", err)
	}
}

// ReadClaimFile writes the output payload to the claim file.  In the event of an error, this method fatally fails.
func ReadClaimFile(claimFileName string) (data []byte, err error) {
	data, err = os.ReadFile(claimFileName)
	if err != nil {
		log.Errorf("ReadFile failed with err: %v", err)
	}
	path, err := os.Getwd()
	if err != nil {
		log.Errorf("Getwd failed with err: %v", err)
	}
	log.Infof("Reading claim file at path: %s", path)
	return data, nil
}

// GetConfigurationFromClaimFile retrieves configuration details from claim file
func GetConfigurationFromClaimFile(claimFileName string) (env *provider.TestEnvironment, err error) {
	data, err := ReadClaimFile(claimFileName)
	if err != nil {
		log.Errorf("ReadClaimFile failed with err: %v", err)
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
		log.Fatalf("Failed to generate the claim: %v", err)
	}
	return payload
}

// WriteClaimOutput writes the output payload to the claim file.  In the event of an error, this method fatally fails.
func WriteClaimOutput(claimOutputFile string, payload []byte) {
	err := os.WriteFile(claimOutputFile, payload, claimFilePermissions)
	if err != nil {
		log.Fatalf("Error writing claim data:\n%s", string(payload))
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
		log.Fatalf("error reading JUnit XML file into JSON: %v", err)
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
