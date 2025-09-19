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

// NewCommand Creates a command for exporting claim data to CSV
//
// This function configures a command with required flags for the claim file
// path, CNF name, and CNF type mapping file, as well as an optional flag to
// include a header row. It marks each required flag, handling any errors by
// logging a fatal message. The configured command is then returned for use in
// the CLI.
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

// dumpCsv Exports claim results to CSV format
//
// This function parses a claim file, validates its version, loads CNF type
// mappings, builds a catalog map, and then constructs CSV records for each test
// result. It writes the assembled data to standard output using a CSV writer,
// handling any errors that occur during parsing or writing. The function
// returns nil on success or an error describing what failed.
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

// buildCSV Creates CSV rows from claim data with remediation, CNF type, and optional header
//
// It iterates over each test result in the claim schema, building a record that
// includes operator versions, test identifiers, suite names, descriptions,
// states, timestamps, skip reasons, check details, captured output, remediation
// actions, CNF type, and mandatory/optional status. If a header flag is set, a
// header row is added first. The function returns a slice of string slices
// ready for CSV writing.
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

// loadCNFTypeMap Loads a mapping of CNF names to their types
//
// This routine opens the specified file, reads its contents, and unmarshals the
// data into a string-to-string map that associates each CNF name with its
// corresponding type. If any step fails—opening, reading, or decoding—the
// function returns an error describing the issue; otherwise it supplies the
// populated map.
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

// buildCatalogByID Creates a map of test case descriptions keyed by ID
//
// It initializes an empty mapping, then iterates over the global catalog
// collection, inserting each entry into the map using its identifier as the
// key. The resulting map is returned for quick lookup of test cases by their
// unique IDs.
func buildCatalogByID() (catalogMap map[string]claimschema.TestCaseDescription) {
	catalogMap = make(map[string]claimschema.TestCaseDescription)

	for index := range identifiers.Catalog {
		catalogMap[index.Id] = identifiers.Catalog[index]
	}
	return catalogMap
}
