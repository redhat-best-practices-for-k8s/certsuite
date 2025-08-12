package csv

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	claimschema "github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
	"github.com/spf13/cobra"
)

var (
	claimFilePathFlag   string
	CNFNameFlag         string
	CNFListFilePathFlag string
	addHeaderFlag       bool

	CSVDumpCommand = &cobra.Command{
		Use:   "csv",
		Short: "Dumps claim file as CSV with pre-applied classification",
		Long: `
Dumps claim results as CSV with classification information.
Required inputs:
- a CNF type mapping: JSON file indicating the CNF type based on CNF name. If a CNF name is not defined, the CNF type is assumed to be non-telco
- a claim file
- the name of the CNF to which the claim file belongs to
The test case catalog, classification table and CNF Type map are used to do the following:
- the catalog is used to add a remediation column to each test case in the generated CSV
- the classification is used to determine whether a test case is mandatory or optional (and add a column). The determination is based on the test name and CNF type.
- add a CNF type column based on the CNF type mapping
	`,
		Example: `
with no column header:
./tnf claim show csv -c claim.yaml -n elasticsearch-operator -t cnf-type.json > claim.csv
with column header:
./tnf claim show csv -c claim.yaml -n elasticsearch-operator -t cnf-type.json -a > claim.csv
`,
		RunE: dumpCsv,
	}
)

// NewCommand creates the CSV output command for certsuite claim show.
//
// It returns a *cobra.Command configured to dump claim data in CSV format.
// The command defines flags for input file, CNF name, CNF list file,
// and an optional header flag. All required flags are marked as such,
// and validation errors cause the program to exit with a fatal message.
func NewCommand() *cobra.Command {
	CSVDumpCommand.Flags().StringVarP(&claimFilePathFlag, "claim-file", "c", "",
		"Required: path to claim file.",
	)

	err := CSVDumpCommand.MarkFlagRequired("claim-file")
	if err != nil {
		log.Fatalf("Failed to mark claim file path as required parameter: %v", err)
		return nil
	}

	CSVDumpCommand.Flags().StringVarP(&CNFNameFlag, "cnf-name", "n", "",
		"Required: CNF name.",
	)

	err = CSVDumpCommand.MarkFlagRequired("cnf-name")
	if err != nil {
		log.Fatalf("Failed to mark CNF name as required parameter: %v", err)
		return nil
	}

	CSVDumpCommand.Flags().StringVarP(&CNFListFilePathFlag, "cnf-type", "t", "",
		"Required: path to JSON file mapping CNF name to CNF type.",
	)

	err = CSVDumpCommand.MarkFlagRequired("cnf-type")
	if err != nil {
		log.Fatalf("Failed to mark CNF type JSON path as required parameter: %v", err)
		return nil
	}

	CSVDumpCommand.Flags().BoolVarP(&addHeaderFlag, "add-header", "a", false,
		"Optional: if present, adds a header to the CSV file",
	)

	return CSVDumpCommand
}

// dumpCsv writes claim data to CSV format.
//
// dumpCsv writes claim data to CSV format.
// It parses command line flags, verifies the API version,
// loads a CNF type map, builds a catalog by ID, and
// generates CSV output using the standard csv package.
// The function returns an error if any step fails.
func dumpCsv(_ *cobra.Command, _ []string) error {
	// set log output to stderr
	log.SetOutput(os.Stderr)

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

	// loads the mapping between CNF name and type
	CNFTypeMap, err := loadCNFTypeMap(CNFListFilePathFlag)
	if err != nil {
		log.Fatalf("Failed to load CNF type map (%s): %v", CNFListFilePathFlag, err)
		return nil
	}

	// builds a catalog map indexed by test ID
	catalogMap := buildCatalogByID()

	// get CNF type
	cnfType := CNFTypeMap[CNFNameFlag]

	// builds CSV file
	resultsCsv := buildCSV(claimScheme, cnfType, catalogMap)

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

// buildCSV creates a CSV representation of claim data with additional metadata.
//
// It takes a pointer to a claim.Schema, the CNF name as a string, and a map of test case descriptions.
// The returned two-dimensional slice contains rows where each row represents a field in the claim,
// enriched with remediation information, mandatory/optional status, and the CNF type.
func buildCSV(claimScheme *claim.Schema, cnfType string, catalogMap map[string]claimschema.TestCaseDescription) (resultsCSVRecords [][]string) {
	if cnfType == "" {
		cnfType = identifiers.NonTelco
	}

	// add header if flag is present (defaults to no header)
	if addHeaderFlag {
		resultsCSVRecords = append(resultsCSVRecords, []string{
			"CNFName", "OperatorVersion", "testID", "Suite",
			"Description", "State",
			"StartTime", "EndTime",
			"SkipReason", "CheckDetails", "Output",
			"Remediation", "CNFType",
			"Mandatory/Optional",
		})
	}

	opVers := ""
	for i, op := range claimScheme.Claim.TestOperators {
		if i == 0 {
			opVers = op.Version
		} else {
			opVers = opVers + ", " + op.Version
		}
	}

	for testID := range claimScheme.Claim.Results {
		// initialize record
		record := []string{}
		// creates and appends new CSV record
		record = append(record,
			CNFNameFlag,
			opVers,
			testID,
			claimScheme.Claim.Results[testID].TestID.Suite,
			claimScheme.Claim.Results[testID].CatalogInfo.Description,
			claimScheme.Claim.Results[testID].State,
			claimScheme.Claim.Results[testID].StartTime,
			claimScheme.Claim.Results[testID].EndTime,
			claimScheme.Claim.Results[testID].SkipReason,
			claimScheme.Claim.Results[testID].CheckDetails,
			claimScheme.Claim.Results[testID].CapturedTestOutput,
			catalogMap[testID].Remediation,
			cnfType, // Append the CNF type
			claimScheme.Claim.Results[testID].CategoryClassification[cnfType],
		)

		resultsCSVRecords = append(resultsCSVRecords, record)
	}
	return resultsCSVRecords
}

// loadCNFTypeMap loads a CSV file containing CNF type mappings and returns them as a map.
//
// It accepts the path to the CSV file, opens it, reads all rows,
// unmarshals each row into a struct, and builds a map from the
// first column value to the second. The function returns the map
// and an error if any step fails, such as opening the file or
// parsing its contents.
func loadCNFTypeMap(path string) (CNFTypeMap map[string]string, err error) { //nolint:gocritic // CNF is a valid acronym
	// Open the CSV file
	file, err := os.Open(path)
	if err != nil {
		return CNFTypeMap, fmt.Errorf("error opening text file: %s, err:%s", path, err)
	}
	defer file.Close()
	// initialize map
	CNFTypeMap = make(map[string]string)

	// read the file
	data, err := io.ReadAll(file)
	if err != nil {
		return CNFTypeMap, fmt.Errorf("error reading JSON file: %s, err:%s", path, err)
	}

	err = json.Unmarshal(data, &CNFTypeMap)
	if err != nil {
		fmt.Println("Error un-marshaling CNF type JSON:", err)
		return
	}

	return CNFTypeMap, nil
}

// buildCatalogByID builds a map of test case descriptions keyed by ID.
//
// It creates an empty map from string to TestCaseDescription and returns it.
// This function is used internally to index catalog entries by their identifier.
func buildCatalogByID() (catalogMap map[string]claimschema.TestCaseDescription) {
	catalogMap = make(map[string]claimschema.TestCaseDescription)

	for index := range identifiers.Catalog {
		catalogMap[index.Id] = identifiers.Catalog[index]
	}
	return catalogMap
}
