package claim

import (
	"fmt"
	"os"
	"path/filepath"

	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/pkg/junit"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
)

var (
	Reportdir string
	Claim     string

	addclaim = &cobra.Command{
		Use:   "claim",
		Short: "The test suite generates a \"claim\" file",
		RunE:  claimUpdate,
	}
	claimAddFile = &cobra.Command{
		Use:   "add",
		Short: "The test suite generates a \"claim\" file",
		RunE:  claimUpdate,
	}
)

const (
	claimFilePermissions = 0644
)

//nolint:funlen
func claimUpdate(cmd *cobra.Command, args []string) error {
	claimFileTextPtr := &Claim
	reportFilesTextPtr := &Reportdir
	fileUpdated := false
	dat, err := os.ReadFile(*claimFileTextPtr)
	if err != nil {
		log.Fatalf("Error reading claim file :%v", err)
	}
	claimRoot := readClaim(&dat)
	junitMap := claimRoot.Claim.RawResults
	items, err := os.ReadDir(*reportFilesTextPtr)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}
	for _, item := range items {
		fileName := item.Name()
		extension := filepath.Ext(fileName)
		reportKeyName := fileName[0 : len(fileName)-len(extension)]
		if _, ok := junitMap[reportKeyName]; ok {
			log.Printf("Skipping: %s already exists in supplied `%s` claim file", reportKeyName, *claimFileTextPtr)
		} else {
			junitMap[reportKeyName], err = junit.ExportJUnitAsMap(fmt.Sprintf("%s/%s", *reportFilesTextPtr, item.Name()))
			if err != nil {
				log.Fatalf("Error reading JUnit XML file into JSON: %v", err)
			}
			fileUpdated = true
		}
	}
	claimRoot.Claim.RawResults = junitMap
	payload, err := json.MarshalIndent(claimRoot, "", "  ")
	if err != nil {
		log.Fatalf("Failed to generate the claim: %v", err)
	}
	err = os.WriteFile(*claimFileTextPtr, payload, claimFilePermissions)
	if err != nil {
		log.Fatalf("Error writing claim data:\n%s", string(payload))
	}
	if fileUpdated {
		log.Printf("Claim file `%s` updated\n", *claimFileTextPtr)
	} else {
		log.Printf("No changes were applied to `%s`\n", *claimFileTextPtr)
	}
	return nil
}

func readClaim(contents *[]byte) *claim.Root {
	var claimRoot claim.Root
	err := json.Unmarshal(*contents, &claimRoot)
	if err != nil {
		log.Fatalf("Error reading claim constents file into type: %v", err)
	}
	return &claimRoot
}

func NewCommand() *cobra.Command {
	claimAddFile.Flags().StringVarP(
		&Reportdir, "reportdir", "r", "",
		"dir of JUnit XML reports. (Required)",
	)

	err := claimAddFile.MarkFlagRequired("reportdir")
	if err != nil {
		return nil
	}

	claimAddFile.Flags().StringVarP(
		&Claim, "claim", "c", "",
		"existing claim file. (Required)",
	)
	err = claimAddFile.MarkFlagRequired("claim")
	if err != nil {
		return nil
	}
	addclaim.AddCommand(claimAddFile)
	return addclaim
}
