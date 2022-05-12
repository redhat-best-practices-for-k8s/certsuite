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

package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"

	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf/identifier"
)

const (
	// backtickOffset is the number of extra characters required to enclose output in backticks.
	backtickOffset = 2

	// introMDFilename is the name of the file that contains the introductory text for CATALOG.md.
	introMDFilename = "INTRO.md"

	// tccFilename is the name of the file that contains the test case catalog section introductory text for CATALOG.md.
	tccFilename = "TEST_CASE_CATALOG.md"

	// tccbbFilename is the name of the file that contains the test case catalog building blocks section introductory
	// text for CATALOG.md.
	tccbbFilename = "TEST_CASE_BUILDING_BLOCKS_CATALOG.md"
)

var (
	// introMDFile is the path to the file that contains the test case catalog section introductory text for CATALOG.md.
	introMDFile = path.Join(mdDirectory, introMDFilename)

	// mdDirectory is the path to the directory of files that contain static text for CATALOG.md.
	mdDirectory = path.Join("cmd", "tnf", "generate", "catalog")

	// tccFile is the path to the file that contains the test case catalog section introductory text for CATALOG.md.
	tccFile = path.Join(mdDirectory, tccFilename)

	// tccbbFile is the path to the file that contains the test case catalog building blocks section introductory text
	// for CATALOG.md
	tccbbFile = path.Join(mdDirectory, tccbbFilename)

	// generateCmd is the root of the "catalog generate" CLI program.
	generateCmd = &cobra.Command{
		Use:   "catalog",
		Short: "Generates the test catalog",
	}

	// jsonGenerateCmd is used to generate a JSON formatted catalog to stdout.
	jsonGenerateCmd = &cobra.Command{
		Use:   "json",
		Short: "Generates the test catalog in JSON format.",
		RunE:  runGenerateJSONCmd,
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
	testLabel  string
	identifier claim.Identifier // {url and version}
}

// cmdJoin is a utility method abstracted from strings.Join which shims in better formatting for markdown files.
func cmdJoin(elems []string, sep string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return "`" + elems[0] + "`"
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		// backtickOffset is used to track the extra length required by enclosing output commands in backticks.
		n += len(elems[i]) + backtickOffset
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString("`" + elems[0] + "`")
	for _, s := range elems[1:] {
		b.WriteString(sep)
		b.WriteString("`" + s + "`")
	}
	return b.String()
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
// {
//     suiteNameA: [
//					{testName, identifier{url, version}},
//					{testName2, identifier{url, version}}
//                ]
//     suiteNameB: [
//					{testName3, identifier{url, version}},
//					{testName4, identifier{url, version}}
//                ]
// }
func createPrintableCatalogFromIdentifiers(keys []claim.Identifier) map[string][]catalogElement {
	catalog := make(map[string][]catalogElement)
	// we need the list of suite's names
	for _, i := range keys {
		suiteTest := identifiers.GetSuiteAndTestFromIdentifier(i)
		if suiteTest == nil {
			fmt.Fprintf(os.Stderr, "Identifier Url not valid\n")
			return nil
		}
		suiteName := suiteTest[0]
		testName := suiteTest[1]
		testLabel := suiteName + "-" + testName
		catalog[suiteName] = append(catalog[suiteName], catalogElement{
			testName:   testName,
			testLabel:  testLabel,
			identifier: i,
		})
	}
	return catalog
}

func getSuitesFromIdentifiers(keys []claim.Identifier) []string {
	var suites []string

	for _, i := range keys {
		suites = append(suites, identifiers.GetSuiteAndTestFromIdentifier(i)[0])
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

// outputTestCases outputs the Markdown representation for test cases from the catalog to stdout.
func outputTestCases() {
	// Building a separate data structure to store the key order for the map
	keys := make([]claim.Identifier, 0, len(identifiers.Catalog))
	for k := range identifiers.Catalog {
		keys = append(keys, k)
	}

	// Sorting the map by identifier URL
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Url < keys[j].Url
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
		fmt.Fprintf(os.Stdout, "\n### %s\n\n", suite)
		for _, k := range catalog[suite] {
			fmt.Fprintf(os.Stdout, "#### %s\n\n", k.testName)
			fmt.Println("Property|Description")
			fmt.Println("---|---")
			fmt.Fprintf(os.Stdout, "Test Case Name|%s\n", k.testName)
			fmt.Fprintf(os.Stdout, "Test Case Label|%s\n", k.testLabel)
			fmt.Fprintf(os.Stdout, "Unique ID|%s\n", k.identifier.Url)
			fmt.Fprintf(os.Stdout, "Version|%s\n", k.identifier.Version)
			fmt.Fprintf(os.Stdout, "Description|%s\n", strings.ReplaceAll(identifiers.Catalog[k.identifier].Description, "\n", " "))
			fmt.Fprintf(os.Stdout, "Result Type|%s\n", identifiers.Catalog[k.identifier].Type)
			fmt.Fprintf(os.Stdout, "Suggested Remediation|%s\n", strings.ReplaceAll(identifiers.Catalog[k.identifier].Remediation, "\n", " "))
			fmt.Fprintf(os.Stdout, "Best Practice Reference|%s\n", strings.ReplaceAll(identifiers.Catalog[k.identifier].BestPracticeReference, "\n", " "))
		}
	}
	fmt.Println()
}

// outputTestCaseBuildingBlocks outputs the Markdown representation for the test case building blocks from the catalog
// to stdout.
func outputTestCaseBuildingBlocks() {
	// Building a separate data structure to store the key order for the map
	keys := make([]string, 0, len(identifier.Catalog))
	for k := range identifier.Catalog {
		keys = append(keys, k)
	}

	// Sorting the map by identifier URL
	sort.Strings(keys)

	// Iterating the map by sorted identifier URL
	for _, k := range keys {
		testName := identifier.GetShortNameFromIdentifier(identifier.Catalog[k].Identifier)
		fmt.Fprintf(os.Stdout, "### %s", testName)
		fmt.Println()
		fmt.Println("Property|Description")
		fmt.Println("---|---")
		fmt.Fprintf(os.Stdout, "Test Name|%s\n", testName)
		fmt.Fprintf(os.Stdout, "Unique ID|%s\n", identifier.Catalog[k].Identifier.URL)
		fmt.Fprintf(os.Stdout, "Version|%s\n", identifier.Catalog[k].Identifier.SemanticVersion)
		fmt.Fprintf(os.Stdout, "Description|%s\n", identifier.Catalog[k].Description)
		fmt.Fprintf(os.Stdout, "Result Type|%s\n", identifier.Catalog[k].Type)
		fmt.Fprintf(os.Stdout, "Intrusive|%t\n", identifier.Catalog[k].IntrusionSettings.ModifiesSystem)
		fmt.Fprintf(os.Stdout, "Modifications Persist After Test|%t\n", identifier.Catalog[k].IntrusionSettings.ModificationIsPersistent)
		fmt.Fprintf(os.Stdout, "Runtime Binaries Required|%s\n", cmdJoin(identifier.Catalog[k].BinaryDependencies, ", "))
		fmt.Println()
	}
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

	// static generation of test case building blocks section introduction.
	if err := emitTextFromFile(tccbbFile); err != nil {
		return err
	}

	// process the test case building blocks
	outputTestCaseBuildingBlocks()
	return nil
}

// runGenerateJSONCmd generates a JSON test catalog.
func runGenerateJSONCmd(_ *cobra.Command, _ []string) error {
	contents, err := json.MarshalIndent(identifier.Catalog, "", "  ")
	if err != nil {
		return err
	}
	fmt.Print(string(contents))
	return nil
}

// Execute executes the "catalog" CLI.
func NewCommand() *cobra.Command {
	generateCmd.AddCommand(jsonGenerateCmd, markdownGenerateCmd)
	return generateCmd
}
