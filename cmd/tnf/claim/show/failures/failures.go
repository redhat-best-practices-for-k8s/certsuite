package failures

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/pkg/claim"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
)

var (
	claimFilePathFlag string
	testSuitesFlag    string
	outputFormatFlag  string

	showFailuresCommand = &cobra.Command{
		Use:   "failures",
		Short: "Shows failed test cases from a claim file.",
		Long: `Parses a claim.json file and shows a report containing only the failed test cases per test suite.
For each failed test case, shows every non compliant object in a readable way in order to help users to understand the
failure reasons. Using the flag "--output json", the program will print a json representation of those failed test cases.
A comma separated list of test suites can be provided with the flag "--testsuites "testSuite1,testSuite2", so the
output will only print failed test cases from those test suites only.
`,
		Example: `./tnf claim show failures --claim path/to/claim.json
Test Suite: access-control
  Test Case: access-control-sys-admin-capability-check
    Description: Ensures that containers do not use SYS_ADMIN capability
    Failure reasons:
       1 - Type: Container, Reason: Non compliant capability detected in container
           Namespace: tnf, Pod Name: test-887998557-8gwwm, Container Name: test, SCC Capability: SYS_ADMIN
       2 - Type: Container, Reason: Non compliant capability detected in container
           Namespace: tnf, Pod Name: test-887998557-pr2w5, Container Name: test, SCC Capability: SYS_ADMIN
  Test Case: access-control-security-context
    Description: Checks the security context matches one of the 4 categories
    Failure reasons:
       1 - Type: ContainerCategory, Reason: container category is NOT category 1 or category NoUID0
           Namespace: tnf, Pod Name: jack-6f88b5bfb4-q5cw6, Container Name: jack, Category: CategoryID4(anything not matching lower category)
       2 - ...
       ...
Test Suite: lifecycle
  Test Case: ...
    ...

./tnf claim show failures --claim path/to/claim.json --o json
{
	"testSuites": [
	  {
		"name": "access-control",
		"failures": [
		  {
			"name": "access-control-sys-admin-capability-check",
			"description": "Ensures that containers do not use SYS_ADMIN capability",
			"nonCompliantObjects": [
			  {
				"type": "Container",
				"reason": "Non compliant capability detected in container",
				"spec": {
				  "Namespace": "tnf",
				  "Pod Name": "test-887998557-8gwwm",
				  "Container Name": "test",
				  "SCC Capability": "SYS_ADMIN"
				}
			  },
			  {
				"type": "Container",
				"reason": "Non compliant capability detected in container",
				"spec": {
				  "Namespace": "tnf",
				  "Pod Name": "test-887998557-pr2w5",
				  "Container Name": "test",
				  "SCC Capability": "SYS_ADMIN"
				}
			  }
			]
		  },
		]
	  },
	  {
		"name" : "lifecycle",
		"failures": [
			...
			...
		]
	  },
	]
  }
`,
		RunE: showFailures,
	}
)

const (
	outputFormatText    = "text"
	outputFormatJSON    = "json"
	outputFarmatInvalid = "invalid"
)

var availableOutputFormats = []string{
	outputFormatText, outputFormatJSON,
}

func NewCommand() *cobra.Command {
	showFailuresCommand.Flags().StringVarP(&claimFilePathFlag, "claim", "c", "",
		"Required: Existing claim file path.",
	)

	err := showFailuresCommand.MarkFlagRequired("claim")
	if err != nil {
		log.Fatalf("Failed to mark claim file path as required parameter: %v", err)
		return nil
	}

	// This command accepts a (optional) list of comma separated suite to filter the
	// output. Only the failures from those test suites will be printed.
	showFailuresCommand.Flags().StringVarP(&testSuitesFlag, "testsuites", "s", "",
		"Optional: comma separated list of test suites names whose failures will be shown.",
	)

	// The format of the output can be changed. Default is plain text, but it can also print
	// it in json format.
	showFailuresCommand.Flags().StringVarP(&outputFormatFlag, "output", "o", outputFormatText,
		fmt.Sprintf("Optional: output format. Available formats: %v", availableOutputFormats),
	)

	return showFailuresCommand
}

// Parses the comma separated list to create a helper map, whose
// keys are the test suite names.
func parseTargetTestSuitesFlag() map[string]bool {
	if testSuitesFlag == "" {
		return nil
	}

	targetTestSuites := map[string]bool{}
	for _, testSuite := range strings.Split(testSuitesFlag, ",") {
		targetTestSuites[strings.TrimSpace(testSuite)] = true
	}

	return targetTestSuites
}

// Parses the output format flag. Returns error if the format
// does not appear in the list "availableOutputFormats".
func parseOutputFormatFlag() (string, error) {
	for _, outputFormat := range availableOutputFormats {
		if outputFormat == outputFormatFlag {
			return outputFormat, nil
		}
	}

	return "", fmt.Errorf("invalid output format flag %q - available formats: %v", outputFormatFlag, availableOutputFormats)
}

// Parses the claim's test case's checkDetails field and creates a list
// of NonCompliantObject's.
func getNonCompliantObjectsFromFailureReason(checkDetails string) ([]NonCompliantObject, error) {
	objects := struct {
		Compliant    []testhelper.ReportObject `json:"CompliantObjectsOut"`
		NonCompliant []testhelper.ReportObject `json:"NonCompliantObjectsOut"`
	}{}

	err := json.Unmarshal([]byte(checkDetails), &objects)
	if err != nil {
		return nil, fmt.Errorf("failed to decode checkDetails %s: %v", checkDetails, err)
	}

	// Now let's create a list of our NonCompliantObject-type items.
	nonCompliantObjects := []NonCompliantObject{}
	for _, object := range objects.NonCompliant {
		outputObject := NonCompliantObject{Type: object.ObjectType, Reason: object.ObjectFieldsValues[0]}
		for i := 1; i < len(object.ObjectFieldsKeys); i++ {
			outputObject.Spec.AddField(object.ObjectFieldsKeys[i], object.ObjectFieldsValues[i])
		}

		nonCompliantObjects = append(nonCompliantObjects, outputObject)
	}

	return nonCompliantObjects, nil
}

// Prints the failures in plain text.
func printFailuresText(testSuites []FailedTestSuite) {
	for _, ts := range testSuites {
		fmt.Printf("Test Suite: %s\n", ts.TestSuiteName)
		for _, tc := range ts.FailingTestCases {
			fmt.Printf("  Test Case: %s\n", tc.TestCaseName)
			fmt.Printf("    Description: %s\n", tc.TestCaseDescription)

			// In case this tc was not using report objects, just print the failure reason string.
			if len(tc.NonCompliantObjects) == 0 {
				fmt.Printf("    Failure reason: %s\n", tc.CheckDetails)
				continue
			}

			fmt.Printf("    Failure reasons:\n")
			for i := range tc.NonCompliantObjects {
				nonCompliantObject := tc.NonCompliantObjects[i]
				fmt.Printf("      %2d - Type: %s, Reason: %s\n", i+1, nonCompliantObject.Type, nonCompliantObject.Reason)
				fmt.Printf("           ")
				for i := range nonCompliantObject.Spec.Fields {
					if i != 0 {
						fmt.Printf(", ")
					}
					field := nonCompliantObject.Spec.Fields[i]
					fmt.Printf("%s: %s", field.Key, field.Value)
				}
				fmt.Printf("\n")
			}
		}
	}
}

// Prints the failures in json format.
func printFailuresJSON(testSuites []FailedTestSuite) {
	type ClaimFailures struct {
		Failures []FailedTestSuite `json:"testSuites"`
	}

	claimFailures := ClaimFailures{Failures: testSuites}
	bytes, err := json.MarshalIndent(claimFailures, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal failures: %v", err)
	}

	fmt.Printf("%s\n", string(bytes))
}

// Creates a list of FailingTestSuite from the results parsed from a claim file. The parsed
// results in claimResultsByTestSuite var maps a test suite name to a list of TestCaseResult,
// which are processed to create the list of FailingTestSuite, filtering out those test suites
// that don't exist in the targetTestSuites map.
func getFailedTestCasesByTestSuite(claimResultsByTestSuite map[string][]*claim.TestCaseResult, targetTestSuites map[string]bool) []FailedTestSuite {
	testSuites := []FailedTestSuite{}
	for testSuite := range claimResultsByTestSuite {
		if targetTestSuites != nil && !targetTestSuites[testSuite] {
			continue
		}

		failedTcs := []FailedTestCase{}
		for _, tc := range claimResultsByTestSuite[testSuite] {
			if tc.State != "failed" {
				continue
			}

			failingTc := FailedTestCase{
				TestCaseName:        tc.TestID.ID,
				TestCaseDescription: tc.CatalogInfo.Description,
			}

			nonCompliantObjects, err := getNonCompliantObjectsFromFailureReason(tc.CheckDetails)
			if err != nil {
				// This means the test case doesn't use the report objects yet. Just use the raw failure reason instead.
				// Also, send the error into stderr, so it can be filtered out with "2>/errors.txt" or "2>/dev/null".
				fmt.Fprintf(os.Stderr, "Failed to parse non compliant objects from test case %s (test suite %s): %v", tc.TestID.ID, testSuite, err)
				failingTc.CheckDetails = tc.CheckDetails
			} else {
				failingTc.NonCompliantObjects = nonCompliantObjects
			}

			failedTcs = append(failedTcs, failingTc)
		}

		if len(failedTcs) > 0 {
			testSuites = append(testSuites, FailedTestSuite{
				TestSuiteName:    testSuite,
				FailingTestCases: failedTcs,
			})
		}
	}

	return testSuites
}

// Main function for the `show failures` subcommand.
func showFailures(_ *cobra.Command, _ []string) error {
	outputFormat, err := parseOutputFormatFlag()
	if err != nil {
		return err
	}

	// Parse the claim file into the claim scheme.
	claimScheme, err := claim.Parse(claimFilePathFlag)
	if err != nil {
		return fmt.Errorf("failed to parse claim file %s: %v", claimFilePathFlag, err)
	}

	// Check claim format version
	err = claim.CheckVersion(claimScheme.Claim.Versions.ClaimFormat)
	if err != nil {
		return err
	}

	// Order test case results by test suite, using a helper map.
	resultsByTestSuite := map[string][]*claim.TestCaseResult{}
	for id := range claimScheme.Claim.Results {
		tcResult := claimScheme.Claim.Results[id]
		resultsByTestSuite[tcResult.TestID.Suite] = append(resultsByTestSuite[tcResult.TestID.Suite], &tcResult)
	}

	targetTestSuites := parseTargetTestSuitesFlag()
	// From the target test suites, get their failed test cases and put them in
	// our custom types.
	testSuites := getFailedTestCasesByTestSuite(resultsByTestSuite, targetTestSuites)

	switch outputFormat {
	case outputFormatJSON:
		printFailuresJSON(testSuites)
	default:
		printFailuresText(testSuites)
	}

	return nil
}
