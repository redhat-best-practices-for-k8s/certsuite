package compare

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/configurations"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/nodes"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/testcases"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/versions"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/spf13/cobra"
)

const longHelp = `Compares sections of both claim files and the differences are shown in a table per section.
This tool can be helpful when the result of some test cases is different between two (consecutive) runs, as it shows
configuration differences in both the Workload Cert Suite config and the cluster nodes that could be the root cause for
some of the test cases results discrepancy.

All the compared sections, except the test cases results are compared blindly, traversing the whole json tree and
substrees to get a list of all the fields and their values. Three tables are shown:
 - Differences: same fields with different values.
 - Fields in claim 1 only: json fields in claim file 1 that don't exist in claim 2.
 - Fields in claim 2 only: json fields in claim file 2 that don't exist in claim 1.

Let's say one of the nodes of the claim.json file contains this struct:
{
	"field1": "value1",
	"field2": {
		"field3": "value2",
		"field4": {
			"field5": "value3",
			"field6": "value4"
		}
	}
}

When parsing that json struct fields, it will produce a list of fields like this:
/field1=value1
/field2/field3=value2
/field2/field4/field5=value3
/field2/field4/field6=finalvalue2

Once this list of field's path+value strings has been obtained from both claim files,
it is compared in order to find the differences or the fields that only exist on each file.

This is a fake example of a node "clus0-0" whose first CNI (index 0) has a different cniVersion
and the ipMask flag of its first plugin (also index 0) has changed to false in the second run.
Also, the plugin has another "newFakeFlag" config flag in claim 2 that didn't exist in clam file 1.

...
CNIs: Differences
FIELD                           CLAIM 1      CLAIM 2
/clus0-0/0/cniVersion           1.0.0        1.0.1
/clus0-1/0/plugins/0/ipMasq     true         false

CNIs: Only in CLAIM 1
<none>

CNIs: Only in CLAIM 2
/clus0-1/0/plugins/0/newFakeFlag=true
...

 Currently, the following sections are compared, in this order:
 - claim.versions
 - claim.Results
 - claim.configurations.Config
 - claim.nodes.cniPlugins
 - claim.nodes.csiDriver
 - claim.nodes.nodesHwInfo
 - claim.nodes.nodeSummary
`

var (
	Claim1FilePathFlag string
	Claim2FilePathFlag string

	claimCompareFiles = &cobra.Command{
		Use:     "compare",
		Short:   "Compare two claim files.",
		Long:    longHelp,
		Example: "claim compare -1 claim1.json -2 claim2.json",
		RunE:    claimCompare,
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
		log.Error("Failed to mark flag claim1 as required: %v", err)
		return nil
	}
	err = claimCompareFiles.MarkFlagRequired("claim2")
	if err != nil {
		log.Error("Failed to mark flag claim2 as required: %v", err)
		return nil
	}

	return claimCompareFiles
}

func claimCompare(_ *cobra.Command, _ []string) error {
	err := claimCompareFilesfunc(Claim1FilePathFlag, Claim2FilePathFlag)
	if err != nil {
		log.Fatal("Error comparing claim files: %v", err)
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

	// Compare claim versions.
	versionsDiff := versions.Compare(&claimFile1Data.Claim.Versions, &claimFile2Data.Claim.Versions)
	fmt.Println(versionsDiff)

	// Show test cases results summary and differences.
	tcsDiffReport := testcases.GetDiffReport(claimFile1Data.Claim.Results, claimFile2Data.Claim.Results)
	fmt.Println(tcsDiffReport)

	// Show Workload Certification Suite configuration differences.
	claim1Configurations := &claimFile1Data.Claim.Configurations
	claim2Configurations := &claimFile2Data.Claim.Configurations
	configurationsDiffReport := configurations.GetDiffReport(claim1Configurations, claim2Configurations)
	fmt.Println(configurationsDiffReport)

	// Show the cluster differences.
	nodesDiff := nodes.GetDiffReport(&claimFile1Data.Claim.Nodes, &claimFile2Data.Claim.Nodes)
	fmt.Print(nodesDiff)

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
