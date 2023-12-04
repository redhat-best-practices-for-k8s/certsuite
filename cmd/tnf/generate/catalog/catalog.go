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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/redhat-openshift-ecosystem/openshift-preflight/artifacts"
	plibContainer "github.com/redhat-openshift-ecosystem/openshift-preflight/container"
	plibOperator "github.com/redhat-openshift-ecosystem/openshift-preflight/operator"
	"github.com/sirupsen/logrus"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/pkg/arrayhelper"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"

	"github.com/spf13/cobra"
)

var (

	// generateCmd is the root of the "catalog generate" CLI program.
	generateCmd = &cobra.Command{
		Use:   "catalog",
		Short: "Generates the test catalog.",
	}

	markdownGenerateClassification = &cobra.Command{
		Use:   "javascript",
		Short: "Generates java script file for classification.",
		RunE:  generateJS,
	}

	// markdownGenerateCmd is used to generate a markdown formatted catalog to stdout.
	markdownGenerateCmd = &cobra.Command{
		Use:   "markdown",
		Short: "Generates the test catalog in markdown format.",
		RunE:  runGenerateMarkdownCmd,
	}
)

type Entry struct {
	testName   string
	identifier claim.Identifier // {url and version}
}

type catalogSummary struct {
	totalSuites     int
	totalTests      int
	testsPerSuite   map[string]int
	testPerScenario map[string]map[string]int
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
func CreatePrintableCatalogFromIdentifiers(keys []claim.Identifier) map[string][]Entry {
	catalog := make(map[string][]Entry)
	// we need the list of suite's names
	for _, i := range keys {
		catalog[i.Suite] = append(catalog[i.Suite], Entry{
			testName:   i.Id,
			identifier: i,
		})
	}
	return catalog
}

func GetSuitesFromIdentifiers(keys []claim.Identifier) []string {
	var suites []string
	for _, i := range keys {
		suites = append(suites, i.Suite)
	}
	return arrayhelper.Unique(suites)
}

func scenarioIDToText(id string) (text string) {
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

func addPreflightTestsToCatalog() {
	const dummy = "dummy"
	// Create artifacts handler
	artifactsWriter, err := artifacts.NewMapWriter()
	if err != nil {
		logrus.Errorf("error creating artifact, failed to add preflight tests to catalog")
		return
	}
	ctx := artifacts.ContextWithWriter(context.TODO(), artifactsWriter)
	optsOperator := []plibOperator.Option{}
	optsContainer := []plibContainer.Option{}
	checkOperator := plibOperator.NewCheck(dummy, dummy, []byte(""), optsOperator...)
	checkContainer := plibContainer.NewCheck(dummy, optsContainer...)
	_, checksOperator, err := checkOperator.List(ctx)
	if err != nil {
		logrus.Errorf("error getting preflight operator tests.")
	}
	_, checksContainer, err := checkContainer.List(ctx)
	if err != nil {
		logrus.Errorf("error getting preflight container tests.")
	}

	allChecks := checksOperator
	allChecks = append(allChecks, checksContainer...)

	for _, c := range allChecks {
		_ = identifiers.AddCatalogEntry(
			c.Name(),
			common.PreflightTestKey,
			c.Metadata().Description,
			c.Help().Suggestion,
			identifiers.NoDocumentedProcess,
			identifiers.NoDocLink,
			true,
			map[string]string{
				identifiers.FarEdge:  identifiers.Optional,
				identifiers.Telco:    identifiers.Optional,
				identifiers.NonTelco: identifiers.Optional,
				identifiers.Extended: identifiers.Optional,
			},
			identifiers.TagCommon)
	}
}

// outputTestCases outputs the Markdown representation for test cases from the catalog to stdout.
func outputTestCases() (outString string, summary catalogSummary) { //nolint:funlen
	// Adds Preflight tests to catalog
	addPreflightTestsToCatalog()

	// Building a separate data structure to store the key order for the map
	keys := make([]claim.Identifier, 0, len(identifiers.Catalog))
	for k := range identifiers.Catalog {
		keys = append(keys, k)
	}

	// Sorting the map by identifier ID
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Id < keys[j].Id
	})

	catalog := CreatePrintableCatalogFromIdentifiers(keys)
	if catalog == nil {
		return
	}
	// we need the list of suite's names
	suites := GetSuitesFromIdentifiers(keys)

	// Sort the list of suite names
	sort.Strings(suites)

	// Iterating the map by test and suite names
	outString = "## Test Case list\n\n" +
		"Test Cases are the specifications used to perform a meaningful test. " +
		"Test cases may run once, or several times against several targets. CNF Certification includes " +
		"a number of normative and informative tests to ensure CNFs follow best practices. " +
		"Here is the list of available Test Cases:\n"

	summary.testPerScenario = make(map[string]map[string]int)
	summary.testsPerSuite = make(map[string]int)
	summary.totalSuites = len(suites)
	for _, suite := range suites {
		outString += fmt.Sprintf("\n### %s\n", suite)
		for _, k := range catalog[suite] {
			summary.testsPerSuite[suite]++
			summary.totalTests++
			// Add the suite to the comma separate list of tags shown.  The tags are also modified in the:
			// GetGinkgoTestIDAndLabels function for usage by Ginkgo.
			tags := strings.ReplaceAll(identifiers.Catalog[k.identifier].Tags, "\n", " ") + "," + k.identifier.Suite

			keys := make([]string, 0, len(identifiers.Catalog[k.identifier].CategoryClassification))

			for scenario := range identifiers.Catalog[k.identifier].CategoryClassification {
				keys = append(keys, scenario)
				_, ok := summary.testPerScenario[scenarioIDToText(scenario)]
				if !ok {
					child := make(map[string]int)
					summary.testPerScenario[scenarioIDToText(scenario)] = child
				}
				switch scenario {
				case identifiers.NonTelco:
					tag := identifiers.TagCommon
					if identifiers.Catalog[k.identifier].Tags == tag {
						summary.testPerScenario[scenarioIDToText(scenario)][identifiers.Catalog[k.identifier].CategoryClassification[scenario]]++
					}
				default:
					tag := strings.ToLower(scenario)
					if strings.Contains(identifiers.Catalog[k.identifier].Tags, tag) {
						summary.testPerScenario[scenarioIDToText(scenario)][identifiers.Catalog[k.identifier].CategoryClassification[scenario]]++
					}
				}
			}
			sort.Strings(keys)
			classificationString := "|**Scenario**|**Optional/Mandatory**|\n"
			for _, j := range keys {
				classificationString += "|" + scenarioIDToText(j) + "|" + identifiers.Catalog[k.identifier].CategoryClassification[j] + "|\n"
			}

			// Every paragraph starts with a new line.

			outString += fmt.Sprintf("\n#### %s\n\n", k.testName)
			outString += "Property|Description\n"
			outString += "---|---\n"
			outString += fmt.Sprintf("Unique ID|%s\n", k.identifier.Id)
			outString += fmt.Sprintf("Description|%s\n", strings.ReplaceAll(identifiers.Catalog[k.identifier].Description, "\n", " "))
			outString += fmt.Sprintf("Suggested Remediation|%s\n", strings.ReplaceAll(identifiers.Catalog[k.identifier].Remediation, "\n", " "))
			outString += fmt.Sprintf("Best Practice Reference|%s\n", strings.ReplaceAll(identifiers.Catalog[k.identifier].BestPracticeReference, "\n", " "))
			outString += fmt.Sprintf("Exception Process|%s\n", strings.ReplaceAll(identifiers.Catalog[k.identifier].ExceptionProcess, "\n", " "))
			outString += fmt.Sprintf("Tags|%s\n", tags)
			outString += classificationString
		}
	}

	return outString, summary
}

func summaryToMD(aSummary catalogSummary) (out string) {
	const tableHeader = "|---|---|\n"
	out += "## Test cases summary\n\n"
	out += fmt.Sprintf("### Total test cases: %d\n\n", aSummary.totalTests)
	out += fmt.Sprintf("### Total suites: %d\n\n", aSummary.totalSuites)
	out += "|Suite|Tests per suite|\n"
	out += tableHeader

	keys := make([]string, 0, len(aSummary.testsPerSuite))

	for j := range aSummary.testsPerSuite {
		keys = append(keys, j)
	}
	sort.Strings(keys)
	for _, suite := range keys {
		out += fmt.Sprintf("|%s|%d|\n", suite, aSummary.testsPerSuite[suite])
	}
	out += "\n"

	keys = make([]string, 0, len(aSummary.testPerScenario))

	for j := range aSummary.testPerScenario {
		keys = append(keys, j)
	}

	sort.Strings(keys)

	for _, scenario := range keys {
		out += fmt.Sprintf("### %s specific tests only: %d\n\n", scenario, aSummary.testPerScenario[scenario][identifiers.Mandatory]+aSummary.testPerScenario[scenario][identifiers.Optional])
		out += "|Mandatory|Optional|\n"
		out += tableHeader
		out += fmt.Sprintf("|%d|%d|\n", aSummary.testPerScenario[scenario][identifiers.Mandatory], aSummary.testPerScenario[scenario][identifiers.Optional])
		out += "\n"
	}
	return out
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

func outputIntro() (out string) {
	return "<!-- markdownlint-disable line-length no-bare-urls -->\n" +
		"# cnf-certification-test catalog\n\n" +
		"The catalog for cnf-certification-test contains a list of test cases " +
		"aiming at testing CNF best practices in various areas. Test suites are defined in 10 areas : `platform-alteration`, `access-control`, `affiliated-certification`, " +
		"`lifecycle`, `manageability`,`networking`, `observability`, `operator`, and `performance.`" +
		"\n\nDepending on the CNF type, not all tests are required to pass to satisfy best practice requirements. The scenario section" +
		" indicates which tests are mandatory or optional depending on the scenario. The following CNF types / scenarios are defined: `Telco`, `Non-Telco`, `Far-Edge`, `Extended`.\n\n"
}

// runGenerateMarkdownCmd generates a markdown test catalog.
func runGenerateMarkdownCmd(_ *cobra.Command, _ []string) error {
	// prints intro
	intro := outputIntro()
	// process the test cases
	tcs, summaryRaw := outputTestCases()
	// create summary
	summary := summaryToMD(summaryRaw)
	fmt.Fprintf(os.Stdout, "%s", intro+summary+tcs)

	return nil
}

// Execute executes the "catalog" CLI.
func NewCommand() *cobra.Command {
	generateCmd.AddCommand(markdownGenerateCmd)

	generateCmd.AddCommand(markdownGenerateClassification)
	return generateCmd
}
