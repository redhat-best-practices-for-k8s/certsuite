package preflight

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
)

var (
	ResultsFilePathFlag      string
	OperatorListFilePathFlag string
	DebugFlag                bool

	enrichPreflightResultsCommand = &cobra.Command{
		Use:   "preflight",
		Short: "enriches preflight results with classification information",
		Long: `
Enriches preflight results with classification information.
The tool needs two input CSVs:
- a list of "telco" operators to identify which operators are "telco"
- a list of TNF results per operator and per test case as a base result file
The test case catalog, classification table and telco operator list are used to do the following:
- the catalog is used to add a remediation column to each test case
- the classification is used to determine whether a test case is mandatory or optional. The determination is based on the test name and operator type (telco/non-telco)
- add a telco/non-telco column based on the telco operator list

In addition a debug mode prints the following debug information:
- telco manual operator map: this is the map manually created to account for naming discrepancy between the telco operator list and the results list.
Maps telco operator list name to result operator list name
- telco operators map: This is list of all telco operators
- telco-unused: this is the list of telco operators not seen in the results file (it should be empty)
- non-telco: this is the list of non-telco operators in the test results. Could potentially contain operators from the telco-unused list that need
to have a new mapped entry.
	`,
		Example: `
Normal command:
./tnf enrich preflight -r operator-results.csv -o operator-list.csv
Debug command:
./tnf enrich preflight -r operator-results.csv -o operator-list.csv -d
`,
		RunE: enrichPreflight,
	}
)

func NewCommand() *cobra.Command {
	enrichPreflightResultsCommand.Flags().StringVarP(&ResultsFilePathFlag, "results-csv", "r", "",
		"Required: existing result CSV file.",
	)

	err := enrichPreflightResultsCommand.MarkFlagRequired("results-csv")
	if err != nil {
		log.Fatalf("Failed to mark result file path as required parameter: %v", err)
		return nil
	}

	enrichPreflightResultsCommand.Flags().StringVarP(&OperatorListFilePathFlag, "telco-operators-csv", "o", "",
		"Required: telco operator list CSV file.",
	)

	err = enrichPreflightResultsCommand.MarkFlagRequired("telco-operators-csv")
	if err != nil {
		log.Fatalf("Failed to mark telco-operators CSV file path as required parameter: %v", err)
		return nil
	}

	enrichPreflightResultsCommand.Flags().BoolVarP(&DebugFlag, "debug", "d", false,
		"Optional: Displays debug info.",
	)

	return enrichPreflightResultsCommand
}

const (
	expectedResultsCsvFields      = 6
	expectedOperatorListCsvFields = 24
	skipFirst1Lines               = 1
	skipFirst2Lines               = 2
)

func enrichPreflight(_ *cobra.Command, _ []string) error {
	// set log output to stderr
	log.SetOutput(os.Stderr)

	// loads the records from the results CSV file
	resultsCsv, err := loadCsv(ResultsFilePathFlag, expectedResultsCsvFields, skipFirst1Lines)
	if err != nil {
		log.Fatalf("Failed to load results CSV (%s): %v", ResultsFilePathFlag, err)
		return nil
	}

	// loads the records from the operator list CSV file
	operatorsCsv, err := loadCsv(OperatorListFilePathFlag, expectedOperatorListCsvFields, skipFirst2Lines)
	if err != nil {
		log.Fatalf("Failed to load results CSV (%s): %v", OperatorListFilePathFlag, err)
		return nil
	}

	// creates the telco operator map to identify which operator names are "telco"
	telcoOperatorMap := buildTelcoOperatorList(operatorsCsv)

	// builds a catalog map indexed by test ID
	catalogMap := buildCatalogByID()

	// add a column indicating if the operator is "telco"
	resultsCsv = addTelcoColumn(resultsCsv, telcoOperatorMap)

	// add a column indicating if passing the test is mandatory for this operator
	// depending of whether it is "telco" or "non-telco"
	resultsCsv = addOptionalMandatoryColumn(resultsCsv)

	// Add a remediation column corresponding to the test case
	resultsCsv = addRemediationColumn(resultsCsv, catalogMap)

	// only display enriched CSV if not in debug mode
	if !DebugFlag {
		writer := csv.NewWriter(os.Stdout)
		err = writer.WriteAll(resultsCsv)
		if err != nil {
			log.Fatalf("Failed to write results CSV to screen, err: %s", err)
			return nil
		}
		// flushes buffer to screen
		writer.Flush()
		// Check for any writing errors
		if err := writer.Error(); err != nil {
			panic(err)
		}
	}
	return nil
}

// loads records from a CSV
func loadCsv(csvPath string, expectedFieldsNumber, skip int) (recordsList [][]string, err error) {
	// Open the CSV file
	file, err := os.Open(csvPath)
	if err != nil {
		return recordsList, fmt.Errorf("error opening csv file: %s, err:%s", csvPath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	for i := 0; i < skip; i++ {
		_, err = reader.Read()
		if err != nil {
			return recordsList, fmt.Errorf("error skipping %d lines of csv file: %s, err:%s", skip, csvPath, err)
		}
	}

	for {
		// Read a single record from the CSV
		record, err := reader.Read()
		if err == io.EOF {
			// Reached the end of the file
			break
		}
		// skip records with unexpected fields count
		if len(record) != expectedFieldsNumber {
			log.Printf("[WARNING] unexpected number of fields got: %d, expected %d, record: %v", len(record), expectedFieldsNumber, record)
			continue
		}
		// skip records in error
		if err != nil {
			log.Printf("[WARNING] error decoding csv record: %s", err)
			continue
		}

		// store record
		recordsList = append(recordsList, record)
	}
	return recordsList, nil
}

const (
	testIDIndex        = 1
	testIsTelcoIndex   = 6
	testMandatoryIndex = 7
)

// Map mapping operator names in the telco operator list file and the results file, if they are different
var telcoManualOperatorMap = map[string]string{
	"gatekeeper-operator":           "gatekeeper-operator-product",
	"amq-broker-operator":           "amq-broker-rhel8",
	"prometheus-operator":           "rhods-prometheus-operator",
	"ACM":                           "advanced-cluster-management",
	"cert-manager":                  "openshift-cert-manager-operator",
	"servicemesh-operator":          "servicemeshoperator",
	"netcool-integrations-operator": "multicluster-engine",
}

// builds the list of telco operators
func buildTelcoOperatorList(recordsList [][]string) (telcoOperatorMap map[string]bool) {
	const (
		operatorNameListIndex       = 0
		operatorIsTestedByPreflight = 2
	)
	telcoOperatorMap = map[string]bool{}
	for _, record := range recordsList {
		operatorName := strings.TrimSpace(record[operatorNameListIndex])
		if manualMappedName, ok := telcoManualOperatorMap[operatorName]; ok {
			operatorName = manualMappedName
		}
		if operatorName != "" && strings.ToLower(strings.TrimSpace(record[operatorIsTestedByPreflight])) == "yes" {
			telcoOperatorMap[operatorName] = true
		}
	}
	if DebugFlag {
		fmt.Printf("Telco manual operator map:\n%v\n---\nTelco operators map:\n%v\n", telcoManualOperatorMap, telcoOperatorMap)
	}
	return telcoOperatorMap
}

// adds a "telco" column to CSV records based on the passed list of telco operators, used to identify telco operators
// based on operator name
func addTelcoColumn(recordsList [][]string, telcoOperatorMap map[string]bool) (outputRecordList [][]string) {
	const (
		operatorNameResultsIndex = 0
	)
	nonTelcoOperators := map[string]bool{}
	for _, record := range recordsList {
		operatorName := record[operatorNameResultsIndex]

		if _, ok := telcoOperatorMap[operatorName]; ok {
			record = append(record, identifiers.Telco)
			outputRecordList = append(outputRecordList, record)
			telcoOperatorMap[operatorName] = false
		} else {
			record = append(record, identifiers.NonTelco)
			outputRecordList = append(outputRecordList, record)
			nonTelcoOperators[operatorName] = true
		}
	}
	if DebugFlag {
		fmt.Printf("---\nTelco-unused:\n")
		for name, unused := range telcoOperatorMap {
			if unused {
				fmt.Println(name)
			}
		}
		fmt.Printf("---\nNon-Telco:\n")
		for name := range nonTelcoOperators {
			fmt.Println(name)
		}
	}
	return outputRecordList
}

// adds a mandatory/optional column to indicate if a test for a given operator needs to pass or not.
// the catalog classification table makes this determination
func addOptionalMandatoryColumn(recordsList [][]string) (outputRecordList [][]string) {
	for _, record := range recordsList {
		testID := record[testIDIndex]
		isTelco := record[testIsTelcoIndex] == "true"
		classificationForTestID := identifiers.Classification[testID]
		if isTelco {
			if classificationForTestID[identifiers.Telco] == identifiers.Mandatory {
				record = append(record, identifiers.Mandatory)
			} else {
				record = append(record, identifiers.Optional)
			}
		} else {
			if classificationForTestID[identifiers.NonTelco] == identifiers.Mandatory {
				record = append(record, identifiers.Mandatory)
			} else {
				record = append(record, identifiers.Optional)
			}
		}
		outputRecordList = append(outputRecordList, record)
	}
	return outputRecordList
}

// builds a catalog map indexed by test case ID
func buildCatalogByID() (catalogMap map[string]claim.TestCaseDescription) {
	catalogMap = make(map[string]claim.TestCaseDescription)

	for index := range identifiers.Catalog {
		catalogMap[index.Id] = identifiers.Catalog[index]
	}
	return catalogMap
}

// adds a column with the remediation for this test case, pulled from the passed catalog map
func addRemediationColumn(recordsList [][]string, catalogMap map[string]claim.TestCaseDescription) (outputRecordList [][]string) {
	for _, record := range recordsList {
		testID := record[testIDIndex]
		remediation := catalogMap[testID].Remediation
		record = append(record, remediation)
		outputRecordList = append(outputRecordList, record)
	}
	return outputRecordList
}
