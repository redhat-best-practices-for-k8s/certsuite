package failures

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/spf13/cobra"
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
		Example: `./certsuite claim show failures --claim path/to/claim.json
Test Suite: access-control
  Test Case: access-control-sys-admin-capability-check
    Description: Ensures that containers do not use SYS_ADMIN capability
    Failure reasons:
       1 - Type: Container, Reason: Non compliant capability detected in container
           Namespace: certsuite, Pod Name: test-887998557-8gwwm, Container Name: test, SCC Capability: SYS_ADMIN
       2 - Type: Container, Reason: Non compliant capability detected in container
           Namespace: certsuite, Pod Name: test-887998557-pr2w5, Container Name: test, SCC Capability: SYS_ADMIN
  Test Case: access-control-security-context
    Description: Checks the security context matches one of the 4 categories
    Failure reasons:
       1 - Type: ContainerCategory, Reason: container category is NOT category 1 or category NoUID0
           Namespace: certsuite, Pod Name: jack-6f88b5bfb4-q5cw6, Container Name: jack, Category: CategoryID4(anything not matching lower category)
       2 - ...
       ...
Test Suite: lifecycle
  Test Case: ...
    ...

./certsuite claim show failures --claim path/to/claim.json --o json
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
				  "Namespace": "certsuite",
				  "Pod Name": "test-887998557-8gwwm",
				  "Container Name": "test",
				  "SCC Capability": "SYS_ADMIN"
				}
			  },
			  {
				"type": "Container",
				"reason": "Non compliant capability detected in container",
				"spec": {
				  "Namespace": "certsuite",
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

// NewCommand Creates a command to display claim failures
//
// The function builds a Cobra command that requires a path to an existing claim
// file and optionally accepts a comma‑separated list of test suites to filter
// the output. It also allows specifying the output format, defaulting to plain
// text but supporting JSON. Errors during flag configuration are logged
// fatally, after which the command is returned for registration.
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

// parseTargetTestSuitesFlag Creates a map of test suite names from the flag input
//
// This function checks if the global test suites flag is empty; if so, it
// returns nil. Otherwise, it splits the comma-separated string into individual
// suite names, trims whitespace from each, and stores them as keys in a boolean
// map set to true. The resulting map is used elsewhere to quickly determine
// whether a given test suite should be processed.
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

// parseOutputFormatFlag Validates the output format flag
//
// It checks whether the user-specified format matches one of the supported
// formats listed in "availableOutputFormats". If a match is found, it returns
// that format string with no error; otherwise it returns an empty string and an
// error explaining the invalid value and listing the allowed options.
func parseOutputFormatFlag() (string, error) {
	for _, outputFormat := range availableOutputFormats {
		if outputFormat == outputFormatFlag {
			return outputFormat, nil
		}
	}

	return "", fmt.Errorf("invalid output format flag %q - available formats: %v", outputFormatFlag, availableOutputFormats)
}

// getNonCompliantObjectsFromFailureReason parses a test case failure payload into non‑compliant objects
//
// The function receives the JSON string that represents a test case’s check
// details, decodes it to extract compliant and non‑compliant report objects,
// and then builds a slice of NonCompliantObject structures. It returns the
// constructed list along with an error if the payload cannot be decoded. The
// output includes each object's type, reason, and any additional specification
// fields.
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

// printFailuresText Prints a plain text summary of failed test suites and cases
//
// The function iterates over each test suite, outputting its name and then
// details for every failing test case. For each case it shows the name,
// description, and either a single failure reason or a list of non‑compliant
// objects with type, reason, and spec fields. The information is formatted
// using printf statements to produce a readable report.
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

// printFailuresJSON Outputs failures as indented JSON
//
// The function receives a slice of failure objects, wraps them in a struct with
// a field named "testSuites", marshals this structure to pretty‑printed JSON,
// and prints the result. If marshalling fails it logs a fatal error and exits.
// The output is written to standard output as a single line containing the JSON
// string.
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

// getFailedTestCasesByTestSuite generates a list of failing test suites from parsed claim data
//
// The function iterates over test suite results, filtering by the target suites
// if specified. For each failed test case it extracts details, attempts to
// parse non‑compliant objects, and records either the parsed objects or the
// raw failure reason on error. It returns a slice of structures that represent
// only those test suites containing at least one failing test case.
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

// showFailures Displays failed test cases from a claim file
//
// The function reads the claim file, validates its format version, groups
// results by test suite, filters for failures, and outputs them either in JSON
// or plain text based on a flag. It returns an error if parsing or validation
// fails.
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
