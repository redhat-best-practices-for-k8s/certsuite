package csvtelco

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/pkg/claim"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	claimschema "github.com/test-network-function/test-network-function-claim/pkg/claim"
)

var (
	claimFilePathFlag   string
	CNFNameFlag         string
	CNFListFilePathFlag string

	CSVDumpCommand = &cobra.Command{
		Use:   "csv-telco",
		Short: "Dumps claim file as CSV with telco/non-telco classification",
		Long: `
Dumps claim results as CSV with classification information.
Required inputs:
- a list of "telco" CNF names to identify which CNF names are "telco"
- a claim file
- the name of the CNF to which the claim file belongs to
The test case catalog, classification table and telco CNF list are used to do the following:
- the catalog is used to add a remediation column to each test case in the generated CSV
- the classification is used to determine whether a test case is mandatory or optional (and add a column). The determination is based on the test name and CNF type (telco/non-telco)
- add a telco/non-telco column based on the telco CNF list
	`,
		Example: `
./tnf claim show csv-telco -c claim.yaml -o elasticsearch-operator -t example-telco.txt > claim.csv
`,
		RunE: dumpCsvTelco,
	}
)

func NewCommand() *cobra.Command {
	CSVDumpCommand.Flags().StringVarP(&claimFilePathFlag, "claim-file", "c", "",
		"Required: claim file.",
	)

	err := CSVDumpCommand.MarkFlagRequired("claim-file")
	if err != nil {
		log.Fatalf("Failed to mark claim file path as required parameter: %v", err)
		return nil
	}

	CSVDumpCommand.Flags().StringVarP(&CNFNameFlag, "cnf-name", "o", "",
		"Required: CNF name.",
	)

	err = CSVDumpCommand.MarkFlagRequired("cnf-name")
	if err != nil {
		log.Fatalf("Failed to mark CNF name as required parameter: %v", err)
		return nil
	}

	CSVDumpCommand.Flags().StringVarP(&CNFListFilePathFlag, "telco-list", "t", "",
		"Required: telco CNF list (text file).",
	)

	err = CSVDumpCommand.MarkFlagRequired("telco-list")
	if err != nil {
		log.Fatalf("Failed to mark telco CNF list as required parameter: %v", err)
		return nil
	}

	return CSVDumpCommand
}

func dumpCsvTelco(_ *cobra.Command, _ []string) error {
	// set log output to stderr
	log.SetOutput(os.Stderr)

	// Parse the claim file into the claim scheme.
	claimScheme, err := claim.Parse(claimFilePathFlag)
	if err != nil {
		return fmt.Errorf("failed to parse claim file %s: %v", claimFilePathFlag, err)
	}

	// loads the list of telco CNFs as a map
	telcoCNFMap, err := loadTelcoCNFMap(CNFListFilePathFlag)
	if err != nil {
		log.Fatalf("Failed to load telco CNF list (%s): %v", CNFListFilePathFlag, err)
		return nil
	}

	// builds a catalog map indexed by test ID
	catalogMap := buildCatalogByID()

	// builds CSV file
	resultsCsv := buildCSV(claimScheme, telcoCNFMap, catalogMap)

	// initializes CSV writer
	writer := csv.NewWriter(os.Stdout)

	// writes all CSV records
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

	return nil
}

// dumps claim file in CSV format. Interprets the results in the Telco / non-telco scenarios
// adds remediation, mandatory/optional, telco/non-telco to the claim data
func buildCSV(claimScheme *claim.Schema, telcoCNFMap map[string]bool, catalogMap map[string]claimschema.TestCaseDescription) (resultsCSVRecords [][]string) {
	// add header
	resultsCSVRecords = append(resultsCSVRecords, []string{
		"CNFName", "testID", "Suite",
		"Description", "State",
		"StartTime", "EndTime",
		"FailureReason", "Output",
		"Remediation", "Telco/Non-Telco",
		"Mandatory/Optional",
	})

	for testID := range claimScheme.Claim.Results {
		// get classification map for current test case
		classificationForTestID := identifiers.Classification[testID]
		// initialize record
		record := []string{}
		// creates and appends new CSV record
		record = append(record,
			CNFNameFlag,
			testID,
			claimScheme.Claim.Results[testID][0].TestID.Suite,
			claimScheme.Claim.Results[testID][0].Description,
			claimScheme.Claim.Results[testID][0].State,
			claimScheme.Claim.Results[testID][0].StartTime,
			claimScheme.Claim.Results[testID][0].EndTime,
			claimScheme.Claim.Results[testID][0].FailureReason,
			claimScheme.Claim.Results[testID][0].Output,
			catalogMap[testID].Remediation,
		)
		// if the name of this CNF matches a name in the telco CNF list, all the results are tagged as "telco"
		if _, ok := telcoCNFMap[CNFNameFlag]; ok {
			record = append(record, identifiers.Telco)
			// indicate if the testcase is mandatory or optional in the telco scenario
			if classificationForTestID[identifiers.Telco] == identifiers.Mandatory {
				record = append(record, identifiers.Mandatory)
			} else {
				record = append(record, identifiers.Optional)
			}
		} else {
			record = append(record, identifiers.NonTelco)
			// indicate if the testcase is mandatory or optional in the non-telco scenario
			if classificationForTestID[identifiers.NonTelco] == identifiers.Mandatory {
				record = append(record, identifiers.Mandatory)
			} else {
				record = append(record, identifiers.Optional)
			}
		}
		resultsCSVRecords = append(resultsCSVRecords, record)
	}
	return resultsCSVRecords
}

// loads records from a CSV
func loadTelcoCNFMap(csvPath string) (telcoCNFMap map[string]bool, err error) {
	// Open the CSV file
	file, err := os.Open(csvPath)
	if err != nil {
		return telcoCNFMap, fmt.Errorf("error opening text file: %s, err:%s", csvPath, err)
	}
	defer file.Close()
	// initialize map
	telcoCNFMap = make(map[string]bool)

	// Create a new scanner
	scanner := bufio.NewScanner(file)

	// read the file
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		telcoCNFMap[line] = true
	}
	return telcoCNFMap, nil
}

// builds a catalog map indexed by test case ID
func buildCatalogByID() (catalogMap map[string]claimschema.TestCaseDescription) {
	catalogMap = make(map[string]claimschema.TestCaseDescription)

	for index := range identifiers.Catalog {
		catalogMap[index.Id] = identifiers.Catalog[index]
	}
	return catalogMap
}
