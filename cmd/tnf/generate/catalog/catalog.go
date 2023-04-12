// Copyright (C) 2020-2023 Red Hat, Inc.
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

package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"

	"github.com/spf13/cobra"
)

const (

	// introMDFilename is the name of the file that contains the introductory text for CATALOG.md.
	introMDFilename = "INTRO.md"

	// tccFilename is the name of the file that contains the test case catalog section introductory text for CATALOG.md.
	tccFilename = "TEST_CASE_CATALOG.md"
)

var (
	// introMDFile is the path to the file that contains the test case catalog section introductory text for CATALOG.md.
	introMDFile = path.Join(mdDirectory, introMDFilename)

	// mdDirectory is the path to the directory of files that contain static text for CATALOG.md.
	mdDirectory = path.Join("cmd", "tnf", "generate", "catalog")

	// tccFile is the path to the file that contains the test case catalog section introductory text for CATALOG.md.
	tccFile = path.Join(mdDirectory, tccFilename)

	// generateCmd is the root of the "catalog generate" CLI program.
	generateCmd = &cobra.Command{
		Use:   "catalog",
		Short: "Generates the test catalog",
	}

	markdownGenerateClassification = &cobra.Command{
		Use:   "javascript",
		Short: "Generates java script file for classification",
		RunE:  generateJS,
	}

	// markdownGenerateCmd is used to generate a markdown formatted catalog to stdout.
	markdownGenerateCmd = &cobra.Command{
		Use:   "markdown",
		Short: "Generates the test catalog in markdown format.",
		RunE:  runGenerateMarkdownCmd,
	}
)

type catalogElement struct {
	testName   string
	identifier claim.Identifier // {url and version}
}

// emitTextFromFile is a utility method to stream file contents to stdout.  This allows more natural specification of
// the non-dynamic aspects of CATALOG.md.
func emitTextFromFile(filename string) error {
	text, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	fmt.Print(string(text))
	return nil
}

// createPrintableCatalogFromIdentifiers creates an structured catalogue.
// Decompose claim.Identifier urls like http://test-network-function.com/testcases/SuiteName/TestName
// to get SuiteNames and TestNames and build a "more printable" catalogue in the way of:
//
//	{
//	    suiteNameA: [
//						{testName, identifier{url, version}},
//						{testName2, identifier{url, version}}
//	               ]
//	    suiteNameB: [
//						{testName3, identifier{url, version}},
//						{testName4, identifier{url, version}}
//	               ]
//	}
func createPrintableCatalogFromIdentifiers(keys []claim.Identifier) map[string][]catalogElement {
	catalog := make(map[string][]catalogElement)
	// we need the list of suite's names
	for _, i := range keys {
		catalog[i.Suite] = append(catalog[i.Suite], catalogElement{
			testName:   i.Id,
			identifier: i,
		})
	}
	return catalog
}

func getSuitesFromIdentifiers(keys []claim.Identifier) []string {
	var suites []string
	for _, i := range keys {
		suites = append(suites, i.Suite)
	}
	return Unique(suites)
}

func Unique(slice []string) []string {
	// create a map with all the values as key
	uniqMap := make(map[string]struct{})
	for _, v := range slice {
		uniqMap[v] = struct{}{}
	}

	// turn the map keys into a slice
	uniqSlice := make([]string, 0, len(uniqMap))
	for v := range uniqMap {
		uniqSlice = append(uniqSlice, v)
	}
	return uniqSlice
}

func ScenarioIDToText(id string) (text string) {
	switch id {
	case identifiers.FarEdge:
		text = "Far-Edge"
	case identifiers.Telco:
		text = "Telco"
	case identifiers.NonTelco:
		text = "Non-Telco"
	case identifiers.Extended:
		text = "Extended"
	default:
		text = "Unknown Scenario"
	}
	return text
}

// outputTestCases outputs the Markdown representation for test cases from the catalog to stdout.
func outputTestCases() { //nolint:funlen
	// Building a separate data structure to store the key order for the map
	keys := make([]claim.Identifier, 0, len(identifiers.Catalog))
	for k := range identifiers.Catalog {
		keys = append(keys, k)
	}

	// Sorting the map by identifier ID
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Id < keys[j].Id
	})

	catalog := createPrintableCatalogFromIdentifiers(keys)
	if catalog == nil {
		return
	}
	// we need the list of suite's names
	suites := getSuitesFromIdentifiers(keys)

	// Sort the list of suite names
	sort.Strings(suites)

	// Iterating the map by test and suite names
	for _, suite := range suites {
		fmt.Fprintf(os.Stdout, "\n### %s\n", suite)
		for _, k := range catalog[suite] {
			// Add the suite to the comma separate list of tags shown.  The tags are also modified in the:
			// GetGinkgoTestIDAndLabels function for usage by Ginkgo.
			tags := strings.ReplaceAll(identifiers.Catalog[k.identifier].Tags, "\n", " ") + "," + k.identifier.Suite

			keys := make([]string, 0, len(identifiers.Catalog[k.identifier].CategoryClassification))

			for j := range identifiers.Catalog[k.identifier].CategoryClassification {
				keys = append(keys, j)
			}
			sort.Strings(keys)
			classificationString := "|**Scenario**|**Optional/Mandatory**|\n"
			for _, j := range keys {
				classificationString += "|" + ScenarioIDToText(j) + "|" + identifiers.Catalog[k.identifier].CategoryClassification[j] + "|\n"
			}

			// Every paragraph starts with a new line.
			fmt.Fprintf(os.Stdout, "\n#### %s\n\n", k.testName)
			fmt.Println("Property|Description")
			fmt.Println("---|---")
			fmt.Fprintf(os.Stdout, "Unique ID|%s\n", k.identifier.Id)
			fmt.Fprintf(os.Stdout, "Description|%s\n", strings.ReplaceAll(identifiers.Catalog[k.identifier].Description, "\n", " "))
			fmt.Fprintf(os.Stdout, "Result Type|%s\n", identifiers.Catalog[k.identifier].Type)
			fmt.Fprintf(os.Stdout, "Suggested Remediation|%s\n", strings.ReplaceAll(identifiers.Catalog[k.identifier].Remediation, "\n", " "))
			fmt.Fprintf(os.Stdout, "Best Practice Reference|%s\n", strings.ReplaceAll(identifiers.Catalog[k.identifier].BestPracticeReference, "\n", " "))
			fmt.Fprintf(os.Stdout, "Exception Process|%s\n", strings.ReplaceAll(identifiers.Catalog[k.identifier].ExceptionProcess, "\n", " "))
			fmt.Fprintf(os.Stdout, "Tags|%s\n", tags)
			fmt.Fprintf(os.Stdout, "%s", classificationString)
		}
	}
}

func outputJS() {
	out, err := json.MarshalIndent(identifiers.Classification, "", "  ")
	if err != nil {
		logrus.Errorf("could not Marshall classification, err=%s", err)
		return
	}
	fmt.Printf("classification=  %s ", out)
}
func generateJS(_ *cobra.Command, _ []string) error {
	// process the test cases
	outputJS()

	return nil
}

// runGenerateMarkdownCmd generates a markdown test catalog.
func runGenerateMarkdownCmd(_ *cobra.Command, _ []string) error {
	// static introductory generation
	if err := emitTextFromFile(introMDFile); err != nil {
		return err
	}
	if err := emitTextFromFile(tccFile); err != nil {
		return err
	}

	// process the test cases
	outputTestCases()

	return nil
}

/*func NewCommandclassification() *cobra.Command {
	generateClassification.AddCommand(markdownGenerateClassification)
	return generateClassification
}*/

// Execute executes the "catalog" CLI.
func NewCommand() *cobra.Command {
	generateCmd.AddCommand(markdownGenerateCmd)

	generateCmd.AddCommand(markdownGenerateClassification)
	return generateCmd
}
