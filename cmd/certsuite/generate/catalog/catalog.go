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

	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/arrayhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"

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
// Decompose claim.Identifier urls like http://redhat-best-practices-for-k8s.com/testcases/SuiteName/TestName
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
		log.Error("Error creating artifact, failed to add preflight tests to catalog: %v", err)
		return
	}
	ctx := artifacts.ContextWithWriter(context.TODO(), artifactsWriter)
	optsOperator := []plibOperator.Option{}
	optsContainer := []plibContainer.Option{}
	checkOperator := plibOperator.NewCheck(dummy, dummy, []byte(""), optsOperator...)
	checkContainer := plibContainer.NewCheck(dummy, optsContainer...)
	_, checksOperator, err := checkOperator.List(ctx)
	if err != nil {
		log.Error("Error getting preflight operator tests: %v", err)
	}
	_, checksContainer, err := checkContainer.List(ctx)
	if err != nil {
		log.Error("Error getting preflight container tests: %v", err)
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
		"Test cases may run once, or several times against several targets. The Red Hat Best Practices Test Suite for Kubernetes includes " +
		"a number of normative and informative tests to ensure that workloads follow best practices. " +
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
			// GetTestIDAndLabels function.
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

			// Skip the test if it has the "waiting-for-release" tag.
			if strings.Contains(tags, "waiting-for-release") {
				continue
			}

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
		log.Error("could not Marshall classification, err=%s", err)
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
	headerStr :=
		"<!-- markdownlint-disable line-length no-bare-urls blanks-around-lists ul-indent blanks-around-headings no-trailing-spaces -->\n" +
			"# Red Hat Best Practices Test Suite for Kubernetes catalog\n\n"
	introStr :=
		"The catalog for the Red Hat Best Practices Test Suite for Kubernetes contains a list of test cases " +
			"aiming at testing best practices in various areas. Test suites are defined in 10 areas : `platform-alteration`, `access-control`, `affiliated-certification`, " +
			"`lifecycle`, `manageability`,`networking`, `observability`, `operator`, and `performance.`" +
			"\n\nDepending on the workload type, not all tests are required to pass to satisfy best practice requirements. The scenario section" +
			" indicates which tests are mandatory or optional depending on the scenario. The following workload types / scenarios are defined: `Telco`, `Non-Telco`, `Far-Edge`, `Extended`.\n\n"

	return headerStr + introStr
}

func outputSccCategories() (sccCategories string) {
	sccCategories = "\n## Security Context Categories\n"

	intro := "\nSecurity context categories referred here are applicable to the [access control test case](#access-control-security-context).\n\n"

	firstCat := "### 1st Category\n" +
		"Default SCC for all users if namespace does not use service mesh.\n\n" +
		"Workloads under this category should: \n" +
		" - Use default CNI (OVN) network interface\n" +
		" - Not request NET_ADMIN or NET_RAW for advanced networking functions\n\n"

	secondCat := "### 2nd Category\n" +
		"For workloads which utilize Service Mesh sidecars for mTLS or load balancing. These workloads must utilize an alternative SCC “restricted-no-uid0” to workaround a service mesh UID limitation. " +
		"Workloads under this category should not run as root (UID0).\n\n"

	thirdCat := "### 3rd Category\n" +
		"For workloads with advanced networking functions/requirements (e.g. CAP_NET_RAW, CAP_NET_ADMIN, may run as root).\n\n" +
		"For example:\n" +
		"  - Manipulate the low-level protocol flags, such as the 802.1p priority, VLAN tag, DSCP value, etc.\n" +
		"  - Manipulate the interface IP addresses or the routing table or the firewall rules on-the-fly.\n" +
		"  - Process Ethernet packets\n" +
		"Workloads under this category may\n" +
		"  - Use Macvlan interface to sending and receiving Ethernet packets\n" +
		"  - Request CAP_NET_RAW for creating raw sockets\n" +
		"  - Request CAP_NET_ADMIN for\n" +
		"    - Modify the interface IP address on-the-fly\n" +
		"    - Manipulating the routing table on-the-fly\n" +
		"    - Manipulating firewall rules on-the-fly\n" +
		"    - Setting packet DSCP value\n\n"

	fourthCat := "### 4th Category\n" +
		"For workloads handling user plane traffic or latency-sensitive payloads at line rate, such as load balancing, routing, deep packet inspection etc. " +
		"Workloads under this category may also need to process the packets at a lower level.\n\n" +
		"These workloads shall \n" +
		"  - Use SR-IOV interfaces \n" +
		"  - Fully or partially bypassing kernel networking stack with userspace networking technologies," +
		"such as DPDK, F-stack, VPP, OpenFastPath, etc. A userspace networking stack not only improves" +
		"the performance but also reduces the need for CAP_NET_ADMIN and CAP_NET_RAW.\n" +
		"CAP_IPC_LOCK is mandatory for allocating hugepage memory, hence shall be granted to DPDK applications. If the workload is latency-sensitive and needs a real-time kernel, CAP_SYS_NICE would be required.\n"

	return sccCategories + intro + firstCat + secondCat + thirdCat + fourthCat
}

// runGenerateMarkdownCmd generates a markdown test catalog.
func runGenerateMarkdownCmd(_ *cobra.Command, _ []string) error {
	// prints intro
	intro := outputIntro()
	// process the test cases
	tcs, summaryRaw := outputTestCases()
	// create summary
	summary := summaryToMD(summaryRaw)

	sccCategories := outputSccCategories()
	fmt.Fprintf(os.Stdout, "%s", intro+summary+tcs+sccCategories)

	return nil
}

// Execute executes the "catalog" CLI.
func NewCommand() *cobra.Command {
	generateCmd.AddCommand(markdownGenerateCmd)

	generateCmd.AddCommand(markdownGenerateClassification)
	return generateCmd
}
