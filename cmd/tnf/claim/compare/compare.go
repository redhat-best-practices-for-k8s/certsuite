package compare

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/claim/compare/nodes"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/claim/compare/testcases"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/pkg/claim"
)

var (
	Claim1FilePathFlag string
	Claim2FilePathFlag string

	claimCompareFiles = &cobra.Command{
		Use:   "compare",
		Short: "Compare two claim files.",
		RunE:  claimCompare,
	}
)

func NewCommand() *cobra.Command {
	claimCompareFiles.Flags().StringVarP(
		&Claim1FilePathFlag, "claim1", "1", "",
		"existing claim1 file. (Required) first file to compare",
	)
	claimCompareFiles.Flags().StringVarP(
		&Claim2FilePathFlag, "claim2", "2", "",
		"existing claim2 file. (Required) second file to compare",
	)
	err := claimCompareFiles.MarkFlagRequired("claim1")
	if err != nil {
		log.Errorf("Failed to mark flag claim1 as required: %v", err)
		return nil
	}
	err = claimCompareFiles.MarkFlagRequired("claim2")
	if err != nil {
		log.Errorf("Failed to mark flag claim2 as required: %v", err)
		return nil
	}

	return claimCompareFiles
}

func claimCompare(_ *cobra.Command, _ []string) error {
	err := claimCompareFilesfunc(Claim1FilePathFlag, Claim2FilePathFlag)
	if err != nil {
		log.Fatalf("Error comparing claim files: %v", err)
	}
	return nil
}

func claimCompareFilesfunc(claim1, claim2 string) error {
	// readfiles
	claimdata1, err := os.ReadFile(claim1)
	if err != nil {
		return fmt.Errorf("failed reading claim1 file: %v", err)
	}

	claimdata2, err := os.ReadFile(claim2)
	if err != nil {
		return fmt.Errorf("failed reading claim2 file: %v", err)
	}

	// unmarshal the files
	claimFile1Data, err := unmarshalClaimFile(claimdata1)
	if err != nil {
		return fmt.Errorf("failed to unmarshal claim1 file: %v", err)
	}

	claimFile2Data, err := unmarshalClaimFile(claimdata2)
	if err != nil {
		return fmt.Errorf("failed to unmarshal claim2 file: %v", err)
	}

	// Show test cases results summary and differences.
	tcsDiffReport := testcases.GetDiffReport(claimFile1Data.Claim.Results, claimFile2Data.Claim.Results)
	fmt.Println(&tcsDiffReport)

	// Show the cluster differences.
	nodesDiff := nodes.GetDiffReport(&claimFile1Data.Claim.Nodes, &claimFile2Data.Claim.Nodes)
	fmt.Printf("%s", nodesDiff)

	return nil
}

func unmarshalClaimFile(claimdata []byte) (claim.Schema, error) {
	var claimDataResult claim.Schema
	err := json.Unmarshal(claimdata, &claimDataResult)
	if err != nil {
		return claim.Schema{}, err
	}
	return claimDataResult, nil
}
