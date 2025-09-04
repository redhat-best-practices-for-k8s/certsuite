package hugepages

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	corev1 "k8s.io/api/core/v1"
)

const (
	RhelDefaultHugepagesz    = 2048 // kB
	RhelDefaultHugepages     = 0
	HugepagesParam           = "hugepages"
	HugepageszParam          = "hugepagesz"
	DefaultHugepagesz        = "default_hugepagesz"
	KernArgsKeyValueSplitLen = 2
	cmd                      = "for file in `find /host/sys/devices/system/node/ -name nr_hugepages`; do echo $file count:`cat $file` ; done"
	outputRegex              = `node(\d+).*hugepages-(\d+)kB.* count:(\d+)`
	numRegexFields           = 4
)

// countBySize maps a hugepages size to a count=number of hugepages.
type countBySize map[int]int

// hugepagesByNuma maps a numa id to a hpSizeCounts map.
type hugepagesByNuma map[int]countBySize

// hugepagesByNuma.String Produces a formatted string of NUMA node hugepage allocations
//
// It orders the NUMA indices, then for each index lists all page sizes with
// their counts in a human‑readable format. The resulting string contains
// entries like "Numa=0 [Size=2048kB Count=4]" and is returned for debugging or
// logging purposes.
func (numaHps hugepagesByNuma) String() string {
	// Order numa ids/indexes
	numaIndexes := []int{}

	for numaIdx := range numaHps {
		numaIndexes = append(numaIndexes, numaIdx)
	}
	sort.Ints(numaIndexes)

	var sb strings.Builder
	for _, numaIdx := range numaIndexes {
		sizeCounts := numaHps[numaIdx]
		sb.WriteString(fmt.Sprintf("Numa=%d ", numaIdx))
		for size, count := range sizeCounts {
			sb.WriteString(fmt.Sprintf("[Size=%dkB Count=%d] ", size, count))
		}
	}
	return sb.String()
}

// Tester performs validation of node hugepage configuration against MachineConfig settings
//
// It gathers hugepage counts per NUMA from the node, parses MachineConfig
// kernel arguments or systemd units, and compares these values to ensure
// consistency. The Run method selects the appropriate comparison path based on
// whether systemd units are present. A successful run confirms that all
// configured hugepages match between the node and its MachineConfig.
type Tester struct {
	node      *provider.Node
	context   clientsholder.Context
	commander clientsholder.Command

	nodeHugepagesByNuma      hugepagesByNuma
	mcSystemdHugepagesByNuma hugepagesByNuma
}

// hugepageSizeToInt Converts a hugepage size string into an integer kilobyte value
//
// This function takes a size string such as "2M" or "1G", extracts the numeric
// portion and multiplies it by 1024 for megabytes or 1024 squared for
// gigabytes. It returns the resulting value in kilobytes as an int, ignoring
// any errors from parsing. The conversion is used to translate kernel argument
// values into usable integer sizes within the program.
func hugepageSizeToInt(s string) int {
	num, _ := strconv.Atoi(s[:len(s)-1])
	unit := s[len(s)-1]
	switch unit {
	case 'M':
		num *= 1024
	case 'G':
		num *= 1024 * 1024
	}

	return num
}

// NewTester Creates a tester for node hugepage validation
//
// This function initializes a Tester object with the provided node, probe pod,
// and command executor. It sets up the execution context inside the probe
// container and retrieves the node's NUMA hugepages information along with
// machineconfig systemd unit configurations. The resulting Tester is ready to
// run checks against the gathered data.
func NewTester(node *provider.Node, probePod *corev1.Pod, commander clientsholder.Command) (*Tester, error) {
	tester := &Tester{
		node:      node,
		commander: commander,
		context:   clientsholder.NewContext(probePod.Namespace, probePod.Name, probePod.Spec.Containers[0].Name),
	}

	log.Info("Getting node %s numa's hugepages values.", node.Data.Name)
	var err error
	tester.nodeHugepagesByNuma, err = tester.getNodeNumaHugePages()
	if err != nil {
		return nil, fmt.Errorf("unable to get node hugepages, err: %v", err)
	}

	log.Info("Parsing machineconfig's kernelArguments and systemd's hugepages units.")
	tester.mcSystemdHugepagesByNuma, err = getMcSystemdUnitsHugepagesConfig(&tester.node.Mc)
	if err != nil {
		return nil, fmt.Errorf("failed to get MC systemd hugepages config, err: %v", err)
	}

	return tester, nil
}

// Tester.HasMcSystemdHugepagesUnits Indicates whether MachineConfig contains Systemd hugepage unit definitions
//
// The method returns true if the internal map of Systemd hugepages per NUMA
// node has one or more entries, meaning that the machine configuration includes
// explicit hugepage units. It does this by checking the length of the map; a
// non‑zero count signals presence, otherwise it indicates no such units were
// defined.
func (tester *Tester) HasMcSystemdHugepagesUnits() bool {
	return len(tester.mcSystemdHugepagesByNuma) > 0
}

// Tester.Run Runs the hugepage configuration comparison tests
//
// The method checks whether MachineConfig includes systemd unit definitions for
// hugepages. If so, it verifies that the node's hugepage counts match those
// units; otherwise it compares kernel argument values against the node's
// totals. It logs progress and returns an error if any mismatch or test failure
// occurs.
func (tester *Tester) Run() error {
	if tester.HasMcSystemdHugepagesUnits() {
		log.Info("Comparing MachineConfig Systemd hugepages info against node values.")
		if pass, err := tester.TestNodeHugepagesWithMcSystemd(); !pass {
			return fmt.Errorf("failed to compare machineConfig systemd's unit hugepages config with node values, err: %v", err)
		}
	} else {
		log.Info("Comparing MC KernelArguments hugepages info against node values.")
		if pass, err := tester.TestNodeHugepagesWithKernelArgs(); !pass {
			return fmt.Errorf("failed to compare machineConfig KernelArguments with node ones, err: %v", err)
		}
	}
	return nil
}

// Tester.TestNodeHugepagesWithMcSystemd Verifies node hugepage counts match MachineConfig systemd settings
//
// The function walks through each NUMA node’s actual hugepage configuration,
// ensuring that any size or node absent from the MachineConfig has a count of
// zero. It then cross‑checks every entry in the MachineConfig against the
// node’s values, confirming matching sizes and counts for all NUMA indices.
// If any discrepancy is found, it returns false with an explanatory error;
// otherwise it reports success.
func (tester *Tester) TestNodeHugepagesWithMcSystemd() (bool, error) {
	// Iterate through node's actual hugepages to make sure that each node's size that does not exist in the
	// MachineConfig has a value of 0.
	for nodeNumaIdx, nodeCountBySize := range tester.nodeHugepagesByNuma {
		// First, numa index should exist in MC
		mcCountBySize, numaExistsInMc := tester.mcSystemdHugepagesByNuma[nodeNumaIdx]
		if !numaExistsInMc {
			log.Warn("Numa %d does not exist in machine config. All hugepage count for all sizes must be zero.", nodeNumaIdx)
			for _, count := range nodeCountBySize {
				if count != 0 {
					return false, fmt.Errorf("node's numa %d hugepages config does not exist in node's machineconfig", nodeNumaIdx)
				}
			}
		}

		// Second, all sizes must exist in mc. If it does not exist (e.g. default 2MB size), its count should be 0.
		for nodeSize, nodeCount := range nodeCountBySize {
			if _, sizeExistsInMc := mcCountBySize[nodeSize]; !sizeExistsInMc && nodeCount != 0 {
				return false, fmt.Errorf("node's numa %d hugepages size=%d does not appear in MC, but the count is not zero (%d)",
					nodeNumaIdx, nodeSize, nodeCount)
			}
		}
	}

	// Now, iterate through mc's numas and make sure they exist and have the same sizes and values in the node.
	for mcNumaIdx, mcCountBySize := range tester.mcSystemdHugepagesByNuma {
		nodeCountBySize, numaExistsInNode := tester.nodeHugepagesByNuma[mcNumaIdx]
		// First, numa index should exist in the node
		if !numaExistsInNode {
			return false, fmt.Errorf("node does not have numa id %d found in the machine config", mcNumaIdx)
		}

		// For this numa, iterate through each of the mc's hugepages sizes and compare with node ones.
		for mcSize, mcCount := range mcCountBySize {
			nodeCount, nodeSizeExistsInNode := nodeCountBySize[mcSize]
			if !nodeSizeExistsInNode {
				return false, fmt.Errorf("node's numa id %d does not have size %d found in the machine config",
					mcNumaIdx, mcSize)
			}

			if nodeCount != mcCount {
				return false, fmt.Errorf("mc numa=%d, hugepages count:%d, size:%d does not match node ones=%d",
					mcNumaIdx, mcCount, mcSize, nodeCount)
			}
		}
	}

	return true, nil
}

// Tester.TestNodeHugepagesWithKernelArgs Validates node hugepage counts against kernel argument configuration
//
// The method retrieves the hugepage sizes and counts specified in a machine's
// kernel arguments, then checks that each size present on the node appears in
// those arguments with non‑zero counts. It aggregates node counts per size
// across all NUMA nodes and compares them to the expected totals from the
// kernel arguments, returning an error if any mismatch occurs. On success it
// logs matching sizes and returns true without error.
func (tester *Tester) TestNodeHugepagesWithKernelArgs() (bool, error) {
	kernelArgsHpCountBySize, _ := getMcHugepagesFromMcKernelArguments(&tester.node.Mc)

	// First, check that all the actual hp sizes across all numas exist in the kernelArguments.
	for nodeNumaIdx, nodeCountBySize := range tester.nodeHugepagesByNuma {
		for nodeSize, nodeCount := range nodeCountBySize {
			if _, sizeExistsInKernelArgs := kernelArgsHpCountBySize[nodeSize]; !sizeExistsInKernelArgs && nodeCount != 0 {
				return false, fmt.Errorf("node's numa %d hugepages size=%d does not appear in kernelArgs, but the count is not zero (%d)",
					nodeNumaIdx, nodeSize, nodeCount)
			}
		}
	}

	// kernelArguments don't have numa info, so we'll add up all numa's hp count
	// for the same size and it should match the values in the kernelArgs.
	for kernelSize, kernelCount := range kernelArgsHpCountBySize {
		total := 0
		for numaIdx, numaCountBySize := range tester.nodeHugepagesByNuma {
			nodeCount, sizeExistsInNode := numaCountBySize[kernelSize]
			if !sizeExistsInNode {
				return false, fmt.Errorf("node's numa %d has no hugepages of kernelArgs' size %d", numaIdx, kernelSize)
			}
			total += nodeCount
		}

		if total == kernelCount {
			log.Info("kernelArguments' hugepages count:%d, size:%d match total node ones for that size.", kernelCount, kernelSize)
		} else {
			return false, fmt.Errorf("total hugepages of size %d will not match (node count=%d, expected=%d)", kernelSize, total, kernelCount)
		}
	}

	return true, nil
}

// Tester.getNodeNumaHugePages Retrieves the node's current hugepage configuration
//
// This method runs a command inside the probe pod to read
// /sys/devices/system/node files, parses each line for NUMA node number, page
// size, and count, and aggregates them into a map keyed by node. It returns the
// populated map or an error if execution fails or output cannot be parsed. The
// result is used to compare against desired hugepage settings.
func (tester *Tester) getNodeNumaHugePages() (hugepages hugepagesByNuma, err error) {
	// This command must run inside the node, so we'll need the node's context to run commands inside the probe daemonset pod.
	stdout, stderr, err := tester.commander.ExecCommandContainer(tester.context, cmd)
	log.Debug("getNodeNumaHugePages stdout: %s, stderr: %s", stdout, stderr)
	if err != nil {
		return hugepagesByNuma{}, err
	}
	if stderr != "" {
		return hugepagesByNuma{}, errors.New(stderr)
	}

	hugepages = hugepagesByNuma{}
	r := regexp.MustCompile(outputRegex)
	for _, line := range strings.Split(stdout, "\n") {
		if line == "" {
			continue
		}

		values := r.FindStringSubmatch(line)
		if len(values) != numRegexFields {
			return hugepagesByNuma{}, fmt.Errorf("failed to parse node's numa hugepages output line:%s (stdout: %s)", line, stdout)
		}

		numaNode, _ := strconv.Atoi(values[1])
		hpSize, _ := strconv.Atoi(values[2])
		hpCount, _ := strconv.Atoi(values[3])

		if sizeCounts, exists := hugepages[numaNode]; exists {
			sizeCounts[hpSize] = hpCount
		} else {
			hugepages[numaNode] = countBySize{hpSize: hpCount}
		}
	}

	log.Info("Node %s hugepages: %s", tester.node.Data.Name, hugepages)
	return hugepages, nil
}

// getMcSystemdUnitsHugepagesConfig extracts hugepage configuration from machineconfig systemd units
//
// This function scans the systemd unit files in a machine configuration for
// entries that define hugepage allocations. It parses each matching unit’s
// contents to capture the number, size, and NUMA node of the hugepages,
// organizing them into a nested map keyed by node and page size. The resulting
// structure is returned along with any parsing errors encountered.
func getMcSystemdUnitsHugepagesConfig(mc *provider.MachineConfig) (hugepages hugepagesByNuma, err error) {
	const UnitContentsRegexMatchLen = 4
	hugepages = hugepagesByNuma{}

	r := regexp.MustCompile(`(?ms)HUGEPAGES_COUNT=(\d+).*HUGEPAGES_SIZE=(\d+).*NUMA_NODE=(\d+)`)
	for _, unit := range mc.Config.Systemd.Units {
		unit.Name = strings.Trim(unit.Name, "\"")
		if !strings.Contains(unit.Name, "hugepages-allocation") {
			continue
		}
		log.Info("Systemd Unit with hugepages info -> name: %s, contents: %s", unit.Name, unit.Contents)
		unit.Contents = strings.Trim(unit.Contents, "\"")
		values := r.FindStringSubmatch(unit.Contents)
		if len(values) < UnitContentsRegexMatchLen {
			return hugepagesByNuma{}, fmt.Errorf("unable to get hugepages values from mc (contents=%s)", unit.Contents)
		}

		numaNode, _ := strconv.Atoi(values[3])
		hpSize, _ := strconv.Atoi(values[2])
		hpCount, _ := strconv.Atoi(values[1])

		if sizeCounts, exists := hugepages[numaNode]; exists {
			sizeCounts[hpSize] = hpCount
		} else {
			hugepages[numaNode] = countBySize{hpSize: hpCount}
		}
	}

	if len(hugepages) > 0 {
		log.Info("Machineconfig's systemd.units hugepages: %v", hugepages)
	} else {
		log.Info("No hugepages found in machineconfig system.units")
	}

	return hugepages, nil
}

// logMcKernelArgumentsHugepages Logs the hugepage configuration extracted from machine‑config kernel arguments
//
// This function builds a human‑readable string that includes the default
// hugepage size and each configured size with its count. It then sends this
// message to the package logger at info level, providing visibility into how
// many hugepages of each size were requested by the node’s machine
// configuration.
func logMcKernelArgumentsHugepages(hugepagesPerSize map[int]int, defhugepagesz int) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("MC KernelArguments hugepages config: default_hugepagesz=%d-kB", defhugepagesz))
	for size, count := range hugepagesPerSize {
		sb.WriteString(fmt.Sprintf(", size=%dkB - count=%d", size, count))
	}
	log.Info("%s", sb.String())
}

// getMcHugepagesFromMcKernelArguments extracts hugepage configuration from kernel arguments
//
// The function parses the kernelArguments field of a MachineConfig to build a
// map that associates each hugepage size with its count, using RHEL defaults
// when necessary. It also determines the default hugepages size specified in
// the arguments or falls back to a system default. The resulting map and
// default size are returned for use by tests validating node hugepage settings.
func getMcHugepagesFromMcKernelArguments(mc *provider.MachineConfig) (hugepagesPerSize map[int]int, defhugepagesz int) {
	defhugepagesz = RhelDefaultHugepagesz
	hugepagesPerSize = map[int]int{}

	hugepagesz := 0
	for _, arg := range mc.Spec.KernelArguments {
		keyValueSlice := strings.Split(arg, "=")
		if len(keyValueSlice) != KernArgsKeyValueSplitLen {
			// Some kernel arguments do not come in name=value
			continue
		}

		key, value := keyValueSlice[0], keyValueSlice[1]
		if key == HugepagesParam && value != "" {
			hugepages, _ := strconv.Atoi(value)
			if _, sizeFound := hugepagesPerSize[hugepagesz]; sizeFound {
				// hugepagesz was parsed before.
				hugepagesPerSize[hugepagesz] = hugepages
			} else {
				// use RHEL's default size for this count.
				hugepagesPerSize[RhelDefaultHugepagesz] = hugepages
			}
		}

		if key == HugepageszParam && value != "" {
			hugepagesz = hugepageSizeToInt(value)
			// Create new map entry for this size
			hugepagesPerSize[hugepagesz] = 0
		}

		if key == DefaultHugepagesz && value != "" {
			defhugepagesz = hugepageSizeToInt(value)
			// In case only default_hugepagesz and hugepages values are provided. The actual value should be
			// parsed next and this default value overwritten.
			hugepagesPerSize[defhugepagesz] = RhelDefaultHugepages
			hugepagesz = defhugepagesz
		}
	}

	if len(hugepagesPerSize) == 0 {
		hugepagesPerSize[RhelDefaultHugepagesz] = RhelDefaultHugepages
		log.Warn("No hugepages size found in node's machineconfig. Defaulting to size=%dkB (count=%d)", RhelDefaultHugepagesz, RhelDefaultHugepages)
	}

	logMcKernelArgumentsHugepages(hugepagesPerSize, defhugepagesz)
	return hugepagesPerSize, defhugepagesz
}
