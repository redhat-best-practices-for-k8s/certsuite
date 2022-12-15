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
	Claim1    string
	Claim2    string

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
	claimCompareFiles = &cobra.Command{
		Use:   "compare",
		Short: "Compare 2 \"claim\" file",
		RunE:  claimCompare,
	}
)

const (
	claimFilePermissions = 0o644
)

type claimFileStruct struct {
	Claim struct {
		Nodes struct {
			CniPlugins  map[string]cnistruct   `json:"cniPlugins"`
			NodesHwInfo map[string]interface{} `json:"nodesHwInfo"`
			CsiDriver   interface{}            `json:"csiDriver"`
		} `json:"nodes"`

		RawResults struct {
			Cnfcertificationtest struct {
				Testsuites struct {
					Testsuite struct {
						Testcase []testCase `json:"testcase"`
					} `json:"testsuite"`
				} `json:"testsuites"`
			} `json:"cnf-certification-test"`
		} `json:"rawResults"`
	} `json:"claim"`
}
type cnistruct []struct {
	Name    string        "json:\"name\""
	Plugins []interface{} "json:\"plugins\""
}

func claimCompare(cmd *cobra.Command, args []string) error {
	claimFileTextPtr := Claim1

	claimFileTextPtr2 := Claim2
	err := claimCompareFilesfunc(claimFileTextPtr, claimFileTextPtr2)
	if err != nil {
		log.Fatalf("Error rclaimCompareFilesfunc :%v", err)
	}
	return nil
}

type testCase struct {
	Name   string `json:"-name"`
	Status string `json:"-status"`
}

func claimCompareFilesfunc(claim1, claim2 string) error {
	// readfiles
	calimdata1, err := os.ReadFile(claim1)
	if err != nil {
		log.Fatalf("Error reading claim1 file:%v", err)
	}
	calimdata2, err2 := os.ReadFile(claim2)
	if err != nil {
		log.Fatalf("Error reading claim2 file :%v", err2)
	}
	// unmarshal the files
	claimFile1Data, err := unmarshalClaimFile(calimdata1)
	if err != nil {
		log.Fatalf("Error in unmarshal claim1 file  :%v", err)
		return err
	}
	claimFile2Data, err := unmarshalClaimFile(calimdata2)
	if err != nil {
		log.Fatalf("Error in unmarshal cliam2 file  :%v", err)
		return err
	}
	// compares function
	compare2cni(claimFile1Data.Claim.Nodes.CniPlugins, claimFile2Data.Claim.Nodes.CniPlugins)
	compare2Hwinfo(claimFile1Data.Claim.Nodes.NodesHwInfo, claimFile2Data.Claim.Nodes.NodesHwInfo)

	slist, r, r2 := compare2TestCaseResults(claimFile1Data.Claim.RawResults.Cnfcertificationtest.Testsuites.Testsuite.Testcase,
		claimFile2Data.Claim.RawResults.Cnfcertificationtest.Testsuites.Testsuite.Testcase)
	log.Info("claim1 and claim2 has diff RawResults ", slist)
	log.Info("test name that claim1 has but claim 2 dont has", r)
	log.Info("test name that claim2 has but claim 1 dont has", r2)

	return nil
}

func unmarshalClaimFile(calimdata []byte) (claimFileStruct, error) {
	var claimDataResult claimFileStruct

	errclaimDataResult := json.Unmarshal(calimdata, &claimDataResult)
	if errclaimDataResult != nil {
		log.Fatalf("Error in unmarshal the claim file :%v", errclaimDataResult)
		return claimDataResult, errclaimDataResult
	}
	// csi
	return claimDataResult, nil
}

func compare2Hwinfo(hwinfo1, hwinfo2 map[string]interface{}) {
	var nodesIn1, nodesIn2 []string

	for key := range hwinfo1 {
		nodesIn1 = append(nodesIn1, key)
	}
	for key := range hwinfo2 {
		nodesIn2 = append(nodesIn2, key)
	}
	missIn1, missIn2 := missing(nodesIn2, nodesIn1)
	fmt.Println("nodes2 and nodes diffs ", missIn1, missIn2)
}

func compare2TestCaseResults(testcaseResult1, testcaseResult2 []testCase) (diffResult []testCase, notFoundtestIn1, notFoundtestIn2 []string) {
	var testcaseR1, testcaseR2 []string
	for _, result1 := range testcaseResult1 {
		for _, result2 := range testcaseResult2 {
			if result2.Name == result1.Name {
				if result2.Status != result1.Status {
					diffResult = append(diffResult, result1)
				}
				break
			}
			testcaseR2 = append(testcaseR2, result2.Name)
		}
		testcaseR1 = append(testcaseR1, result1.Name)
	}
	notFoundtestIn1, notFoundtestIn2 = missing(testcaseR1, testcaseR2)
	return diffResult, removeDuplicateValues(notFoundtestIn1), removeDuplicateValues(notFoundtestIn2)
}

// empty struct (0 bytes)
type void struct{}

// missing compares two slices and returns slice of differences
func missing(a, b []string) (diffsAfromB, diffsBfromA []string) {
	// create map with length of the 'a' slice
	ma := make(map[string]void, len(a))
	mb := make(map[string]void, len(b))
	// Convert first slice to map with empty struct (0 bytes)
	for _, ka := range a {
		ma[ka] = void{}
	}
	// Convert first slice to map with empty struct (0 bytes)
	for _, ka := range b {
		mb[ka] = void{}
	}
	// find missing values in b
	for _, kb := range b {
		if _, ok := ma[kb]; !ok {
			diffsAfromB = append(diffsAfromB, kb)
		}
	}
	for _, ka := range a {
		if _, ok := mb[ka]; !ok {
			diffsBfromA = append(diffsBfromA, ka)
		}
	}
	return diffsAfromB, diffsBfromA
}

func removeDuplicateValues(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	// If the key(values of the slice) is not equal
	// to the already present value in new slice (list)
	// then we append it. else we jump on another element.
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func compare2cni(cni1, cni2 map[string]cnistruct) {
	for node, val := range cni1 {
		for node2, val2 := range cni2 {
			if node != node2 {
				continue
			}
			c, s, e := compare2cniHelper(val, val2)
			if len(s) != 0 {
				log.Info("in node ", node2, " cnis found in claim1 but not present in claim2: ", s)
			}
			if len(e) != 0 {
				log.Info("in node ", node2, " cnis found in claim2 but not present in claim1: ", e)
			}
			if len(c) != 0 {
				log.Info("in node ", node2, " cnis present in both claim 1 and 2 but with different plugins: ", c)
			}
			break
		}
	}
}

func compare2cniHelper(cniList1, cniList2 cnistruct) (diffPlugins cnistruct, notFoundNamesIn1, notFoundNamesIn2 []string) {
	var cniList1Name, cniList2Name []string
	for _, plugin1 := range cniList1 {
		cniList1Name = append(cniList1Name, plugin1.Name)
		for _, plugin2 := range cniList2 {
			cniList2Name = append(cniList2Name, plugin2.Name)
			if plugin2.Name == plugin1.Name {
				if plugin2.Plugins != nil {
					if len(plugin2.Plugins) != len(plugin1.Plugins) {
						diffPlugins = append(diffPlugins, plugin1)
					}
				}
				break
			}
		}
	}
	notFoundNamesIn1, notFoundNamesIn2 = missing(cniList2Name, cniList1Name)
	return diffPlugins, removeDuplicateValues(notFoundNamesIn1), removeDuplicateValues(notFoundNamesIn2)
}

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
	claimCompareFiles.Flags().StringVarP(
		&Claim1, "claim1", "1", "",
		"existing claim1 file. (Required) first file to compare",
	)
	claimCompareFiles.Flags().StringVarP(
		&Claim2, "claim2", "2", "",
		"existing claim2 file. (Required) second file to compare with",
	)
	err = claimCompareFiles.MarkFlagRequired("claim1")
	if err != nil {
		return nil
	}
	err = claimCompareFiles.MarkFlagRequired("claim2")
	if err != nil {
		return nil
	}
	addclaim.AddCommand(claimCompareFiles)
	return addclaim
}
