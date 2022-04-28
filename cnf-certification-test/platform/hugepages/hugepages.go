package hugepages

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	v1 "k8s.io/api/core/v1"
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

type hugePagesConfig struct {
	hugepagesSize  int // size in kb
	hugepagesCount int
}

// numaHugePagesPerSize maps a numa id to an array of hugePagesConfig structs.
type numaHugePagesPerSize map[int][]hugePagesConfig

// String is the stringer implementation for the numaHugePagesPerSize type so debug/info
// lines look better.
func (numaHugepages numaHugePagesPerSize) String() string {
	// Order numa ids/indexes
	numaIndexes := make([]int, 0)
	for numaIdx := range numaHugepages {
		numaIndexes = append(numaIndexes, numaIdx)
	}
	sort.Ints(numaIndexes)

	str := ""
	for _, numaIdx := range numaIndexes {
		hugepagesPerSize := numaHugepages[numaIdx]
		str += fmt.Sprintf("Numa=%d ", numaIdx)
		for _, hugepages := range hugepagesPerSize {
			str += fmt.Sprintf("[Size=%dkB Count=%d] ", hugepages.hugepagesSize, hugepages.hugepagesCount)
		}
	}
	return str
}

type Tester struct {
	node    *provider.Node
	context clientsholder.Context

	nodeNumaHugePagesPerSize numaHugePagesPerSize
	mcSystemdHugepages       numaHugePagesPerSize
}

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

func NewTester(node *provider.Node, debugPod *v1.Pod) (*Tester, error) {
	tester := &Tester{
		node: node,
		context: clientsholder.Context{
			Namespace:     debugPod.Namespace,
			Podname:       debugPod.Name,
			Containername: debugPod.Spec.Containers[0].Name,
		},
	}

	logrus.Infof("Getting node %s numa's hugepages values.", node.Data.Name)
	var err error
	tester.nodeNumaHugePagesPerSize, err = tester.getNodeNumaHugePages()
	if err != nil {
		return nil, fmt.Errorf("unable to get node hugepages, err: %v", err)
	}

	logrus.Info("Parsing machineconfig's kernelArguments and systemd's hugepages units.")
	tester.mcSystemdHugepages, err = getMcSystemdUnitsHugepagesConfig(&tester.node.Mc)
	if err != nil {
		return nil, fmt.Errorf("failed to get MC systemd hugepages config, err: %v", err)
	}

	return tester, nil
}

func (tester *Tester) HasMcSystemdHugepagesUnits() bool {
	return len(tester.mcSystemdHugepages) > 0
}

func (tester *Tester) Run() error {
	if tester.HasMcSystemdHugepagesUnits() {
		logrus.Info("Comparing MachineConfig Systemd hugepages info against node values.")
		if pass, err := tester.TestNodeHugepagesWithMcSystemd(); !pass {
			return fmt.Errorf("failed to compare machineConfig systemd's unit hugepages config with node values, err: %v", err)
		}
	} else {
		logrus.Info("Comparing MC KernelArguments hugepages info against node values.")
		if pass, err := tester.TestNodeHugepagesWithKernelArgs(); !pass {
			return fmt.Errorf("failed to compare machineConfig KernelArguments with node ones, err: %v", err)
		}
	}
	return nil
}

// TestNodeHugepagesWithMcSystemd compares the node's hugepages values against the mc's systemd units ones.
func (tester *Tester) TestNodeHugepagesWithMcSystemd() (bool, error) {
	// Iterate through mc's numas and make sure they exist and have the same sizes and values in the node.
	for mcNumaIdx, mcNumaHugepageCfgs := range tester.mcSystemdHugepages {
		nodeNumaHugepageCfgs, exists := tester.nodeNumaHugePagesPerSize[mcNumaIdx]
		if !exists {
			return false, fmt.Errorf("hugepages config not found for machine config's numa %d", mcNumaIdx)
		}

		// For this numa, iterate through each of the mc's hugepages sizes and compare with node ones.
		for _, mcHugepagesCfg := range mcNumaHugepageCfgs {
			configMatching := false
			for _, nodeHugepagesCfg := range nodeNumaHugepageCfgs {
				if nodeHugepagesCfg.hugepagesSize == mcHugepagesCfg.hugepagesSize && nodeHugepagesCfg.hugepagesCount == mcHugepagesCfg.hugepagesCount {
					logrus.Infof("MC numa=%d, hugepages count:%d, size:%d match node ones: %s",
						mcNumaIdx, mcHugepagesCfg.hugepagesCount, mcHugepagesCfg.hugepagesSize, tester.nodeNumaHugePagesPerSize)
					configMatching = true
					break
				}
			}
			if !configMatching {
				return false, fmt.Errorf("MC numa=%d, hugepages (count:%d, size:%d) not matching node ones: %s",
					mcNumaIdx, mcHugepagesCfg.hugepagesCount, mcHugepagesCfg.hugepagesSize, tester.nodeNumaHugePagesPerSize)
			}
		}
	}

	return true, nil
}

// TestNodeHugepagesWithKernelArgs compares node hugepages against kernelArguments config.
// The total count of hugepages of the size defined in the kernelArguments must match the kernArgs' hugepages value.
// For other sizes, the sum should be 0.
func (tester *Tester) TestNodeHugepagesWithKernelArgs() (bool, error) {
	kernelArgsHugepagesPerSize, _ := getMcHugepagesFromMcKernelArguments(&tester.node.Mc)

	for size, count := range kernelArgsHugepagesPerSize {
		total := 0
		for numaIdx, numaHugepages := range tester.nodeNumaHugePagesPerSize {
			found := false
			for _, hugepages := range numaHugepages {
				if hugepages.hugepagesSize == size {
					total += hugepages.hugepagesCount
					found = true
					break
				}
			}
			if !found {
				return false, fmt.Errorf("numa %d has no hugepages of size %d", numaIdx, size)
			}
		}

		if total == count {
			logrus.Infof("kernelArguments' hugepages count:%d, size:%d match total node ones for that size.", count, size)
		} else {
			return false, fmt.Errorf("total hugepages of size %d won't match (node count=%d, expected=%d)", size, total, count)
		}
	}

	return true, nil
}

// getNodeNumaHugePages gets the actual node's hugepages config based on /sys/devices/system/node/nodeX files.
func (tester *Tester) getNodeNumaHugePages() (hugepages numaHugePagesPerSize, err error) {
	client := clientsholder.GetClientsHolder()

	// This command must run inside the node, so we'll need the node's context to run commands inside the debug daemonset pod.
	stdout, stderr, err := client.ExecCommandContainer(tester.context, cmd)
	if err != nil {
		return numaHugePagesPerSize{}, err
	}
	if stderr != "" {
		return numaHugePagesPerSize{}, errors.New(stderr)
	}

	hugepages = numaHugePagesPerSize{}
	r := regexp.MustCompile(outputRegex)
	for _, line := range strings.Split(stdout, "\n") {
		if line == "" {
			continue
		}

		values := r.FindStringSubmatch(line)
		if len(values) != numRegexFields {
			return numaHugePagesPerSize{}, fmt.Errorf("failed to parse node's numa hugepages output line:%s (stdout: %s)", line, stdout)
		}

		numaNode, _ := strconv.Atoi(values[1])
		hpSize, _ := strconv.Atoi(values[2])
		hpCount, _ := strconv.Atoi(values[3])

		hugepagesCfg := hugePagesConfig{
			hugepagesCount: hpCount,
			hugepagesSize:  hpSize,
		}

		if numaHugepagesCfg, exists := hugepages[numaNode]; exists {
			numaHugepagesCfg = append(numaHugepagesCfg, hugepagesCfg)
			hugepages[numaNode] = numaHugepagesCfg
		} else {
			hugepages[numaNode] = []hugePagesConfig{hugepagesCfg}
		}
	}

	logrus.Infof("Node %s hugepages: %s", tester.node.Data.Name, hugepages)
	return hugepages, nil
}

// getMcSystemdUnitsHugepagesConfig gets the hugepages information from machineconfig's systemd units.
func getMcSystemdUnitsHugepagesConfig(mc *provider.MachineConfig) (hugepages numaHugePagesPerSize, err error) {
	const UnitContentsRegexMatchLen = 4
	hugepages = numaHugePagesPerSize{}

	r := regexp.MustCompile(`(?ms)HUGEPAGES_COUNT=(\d+).*HUGEPAGES_SIZE=(\d+).*NUMA_NODE=(\d+)`)
	for _, unit := range mc.Config.Systemd.Units {
		unit.Name = strings.Trim(unit.Name, "\"")
		if !strings.Contains(unit.Name, "hugepages-allocation") {
			continue
		}
		unit.Contents = strings.Trim(unit.Contents, "\"")
		values := r.FindStringSubmatch(unit.Contents)
		if len(values) < UnitContentsRegexMatchLen {
			return numaHugePagesPerSize{}, fmt.Errorf("unable to get hugepages values from mc (contents=%s)", unit.Contents)
		}

		numaNode, _ := strconv.Atoi(values[3])
		hpSize, _ := strconv.Atoi(values[2])
		hpCount, _ := strconv.Atoi(values[1])

		hugepagesCfg := hugePagesConfig{
			hugepagesCount: hpCount,
			hugepagesSize:  hpSize,
		}

		if numaHugepagesCfg, exists := hugepages[numaNode]; exists {
			numaHugepagesCfg = append(numaHugepagesCfg, hugepagesCfg)
			hugepages[numaNode] = numaHugepagesCfg
		} else {
			hugepages[numaNode] = []hugePagesConfig{hugepagesCfg}
		}
	}

	if len(hugepages) > 0 {
		logrus.Infof("Machineconfig's systemd.units hugepages: %v", hugepages)
	} else {
		logrus.Infof("No hugepages found in machineconfig system.units")
	}

	return hugepages, nil
}

func logMcKernelArgumentsHugepages(hugepagesPerSize map[int]int, defhugepagesz int) {
	logStr := fmt.Sprintf("MC KernelArguments hugepages config: default_hugepagesz=%d-kB", defhugepagesz)
	for size, count := range hugepagesPerSize {
		logStr += fmt.Sprintf(", size=%dkB - count=%d", size, count)
	}
	logrus.Info(logStr)
}

// getMcHugepagesFromMcKernelArguments gets the hugepages params from machineconfig's kernelArguments
func getMcHugepagesFromMcKernelArguments(mc *provider.MachineConfig) (hugepagesPerSize map[int]int, defhugepagesz int) {
	defhugepagesz = RhelDefaultHugepagesz
	hugepagesPerSize = map[int]int{}

	hugepagesz := 0
	for _, arg := range mc.Spec.KernelArguments {
		keyValueSlice := strings.Split(arg, "=")
		if len(keyValueSlice) != KernArgsKeyValueSplitLen {
			// Some kernel arguments don't come in name=value
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
		logrus.Warnf("No hugepages size found in node's machineconfig. Defaulting to size=%dkB (count=%d)", RhelDefaultHugepagesz, RhelDefaultHugepages)
	}

	logMcKernelArgumentsHugepages(hugepagesPerSize, defhugepagesz)
	return hugepagesPerSize, defhugepagesz
}
