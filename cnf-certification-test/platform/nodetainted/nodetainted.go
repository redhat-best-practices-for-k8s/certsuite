// Copyright (C) 2021-2022 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package nodetainted

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
)

// NodeTainted holds information about tainted nodes.
type NodeTainted struct {
	ctx *clientsholder.Context
}

var runCommand = func(ctx *clientsholder.Context, cmd string) (string, error) {
	ch := clientsholder.GetClientsHolder()
	output, outerr, err := ch.ExecCommandContainer(*ctx, cmd)
	if err != nil {
		logrus.Errorln("can't execute command on container ", err)
		return "", err
	}
	if outerr != "" {
		logrus.Errorln("error when running nodetainted command ", outerr)
		return "", errors.New(outerr)
	}
	return output, nil
}

// NewNodeTainted creates a new NodeTainted tester
func NewNodeTaintedTester(context *clientsholder.Context) *NodeTainted {
	return &NodeTainted{
		ctx: context,
	}
}

func (nt *NodeTainted) GetKernelTaintsMask() (uint64, error) {
	output, err := runCommand(nt.ctx, `cat /proc/sys/kernel/tainted`)
	if err != nil {
		return 0, err
	}
	output = strings.ReplaceAll(output, "\n", "")
	output = strings.ReplaceAll(output, "\r", "")
	output = strings.ReplaceAll(output, "\t", "")

	// Convert o number.
	taintsMask, err := strconv.ParseUint(output, 10, 64) // base 10 and uint64
	if err != nil {
		return 0, fmt.Errorf("failed to decode taints mask %q: %w", output, err)
	}

	return taintsMask, nil
}

func (nt *NodeTainted) GetModulesFromNode() []string {
	// Get the 1st column list of the modules running on the node.
	// Split on the return/newline and get the list of the modules back.
	command := `chroot /host lsmod | awk '{ print $1 }' | grep -v Module`
	output, _ := runCommand(nt.ctx, command)
	output = strings.ReplaceAll(output, "\t", "")
	moduleList := strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n")
	return stringhelper.RemoveEmptyStrings(moduleList)
}

func (nt *NodeTainted) ModuleInTree(moduleName string) bool {
	command := `chroot /host cat /sys/module/` + moduleName + `/taint`
	cmdOutput, _ := runCommand(nt.ctx, command)
	return !strings.Contains(cmdOutput, "O")
}

type KernelTaint struct {
	Description string
	Letters     string
}

var kernelTaints = map[int]KernelTaint{
	// Linux standard kernel taints
	0:  {"proprietary module was loaded", "GP"},
	1:  {"module was force loaded", "F"},
	2:  {"kernel running on an out of specification system", "S"},
	3:  {"module was force unloaded", "R"},
	4:  {"processor reported a Machine Check Exception (MCE)", "M"},
	5:  {"bad page referenced or some unexpected page flags", "B"},
	6:  {"taint requested by userspace application", "U"},
	7:  {"kernel died recently, i.e. there was an OOPS or BUG", "D"},
	8:  {"ACPI table overridden by user", "A"},
	9:  {"kernel issued warning", "W"},
	10: {"staging driver was loaded", "C"},
	11: {"workaround for bug in platform firmware applied", "I"},
	12: {"externally-built (“out-of-tree”) module was loaded", "O"},
	13: {"unsigned module was loaded", "E"},
	14: {"soft lockup occurred", "L"},
	15: {"kernel has been live patched", "K"},
	16: {"auxiliary taint, defined for and used by distros", "X"},
	17: {"kernel was built with the struct randomization plugin", "T"},
	18: {"an in-kernel test has been run", "N"},

	// RedHat custom taints for RHEL/CoreOS
	// https://access.redhat.com/solutions/40594
	27: {"Red Hat extension: Hardware for which support has been removed. / OMGZOMBIES easter egg", "Zrh"},
	28: {"Red Hat extension: Unsupported hardware. Refer to \"UNSUPPORTED HARDWARE DEVICE:\" kernel log entry for details", "H"},
	29: {"Red Hat extension: Technology Preview code was loaded; cf. Technology Preview features support scope description. Refer to \"TECH PREVIEW:\" kernel log entry for details", "Tt"},
	30: {"BPF syscall has either been configured or enabled for unprivileged users/programs", "u"},
	31: {"BPF syscall has either been configured or enabled for unprivileged users/programs", "u"},
}

func getTaintMsg(bit int) string {
	if taintMsg, exists := kernelTaints[bit]; exists {
		return fmt.Sprintf("%s (tainted bit %d)", taintMsg.Description, bit)
	}

	return fmt.Sprintf("reserved (tainted bit %d)", bit)
}

func DecodeKernelTaints(bitmap uint64) []string {
	taints := []string{}
	for i := 0; i < 64; i++ {
		bit := (bitmap >> i) & 1
		if bit == 1 {
			taints = append(taints, getTaintMsg(i))
		}
	}
	return taints
}

func GetBitPosFromLetter(letter string) (int, error) {
	for bit, taint := range kernelTaints {
		if strings.Contains(taint.Letters, letter) {
			return bit, nil
		}
	}

	return 0, fmt.Errorf("letter %s does not belong to any known kernel taint", letter)
}

// GetTaintedBitsByModules helper function to gets, for each module, the taint bits from its taint letters.
func GetTaintedBitsByModules(tainters map[string]string) (map[int]bool, error) {
	taintedBits := map[int]bool{}

	for tainter, letters := range tainters {
		// Save taint bits from this module.
		for i := range letters {
			letter := string(letters[i])
			bit, err := GetBitPosFromLetter(letter)
			if err != nil {
				return nil, fmt.Errorf("module %s has invalid taint letter %s: %w", tainter, letter, err)
			}

			taintedBits[bit] = true
		}
	}

	return taintedBits, nil
}

// GetOtherTaintedBits helper function to get the tainted bits that are not related to
// any module.
func GetOtherTaintedBits(taintsMask uint64, taintedBitsByModules map[int]bool) []int {
	otherTaintedBits := []int{}
	// Lastly, check that all kernel taint bits come from modules.
	for i := 0; i < 64; i++ {
		// helper var that is true if bit "i" is set.
		bitIsSet := (taintsMask & (1 << i)) > 0

		if bitIsSet && !taintedBitsByModules[i] {
			otherTaintedBits = append(otherTaintedBits, i)
		}
	}

	return otherTaintedBits
}

// GetTainterModules runs a command in the node to get all the modules that
// have set a taint bit. The resturn value maps a module to a string of taint
// letters. Each letter maps to a single bit in the taint mask.
func (nt *NodeTainted) GetTainterModules() (map[string]string, error) {
	const (
		numFields       = 2
		posModuleName   = 0
		posModuleTaints = 1
	)

	command := "modules=`ls /sys/module`; for module_name in $modules; do taint_file=/sys/module/$module_name/taint; if [ -f $taint_file ]; then taints=`cat $taint_file`; if [[ ${#taints} -gt 0 ]]; then echo \"$module_name `cat $taint_file`\"; fi; fi; done"
	cmdOutput, err := runCommand(nt.ctx, command)
	if err != nil {
		return nil, fmt.Errorf("failed to run command: %w", err)
	}

	lines := strings.Split(cmdOutput, "\n")

	// Parse line by line: "module_name taints"
	tainters := map[string]string{}

	for _, line := range lines {
		if line == "" {
			continue
		}

		elems := strings.Split(line, " ")
		if len(elems) != numFields {
			return nil, fmt.Errorf("failed to parse line %q (output=%s)", line, cmdOutput)
		}

		moduleName := elems[posModuleName]
		moduleTaints := elems[posModuleTaints]

		if _, exist := tainters[moduleName]; exist {
			return nil, fmt.Errorf("module %s (taints %s) has already been parsed", moduleName, moduleTaints)
		}

		logrus.Infof("module %s has taints %s", moduleName, moduleTaints)
		tainters[moduleName] = moduleTaints
	}

	return tainters, nil
}

func (nt *NodeTainted) GetOutOfTreeModules(modules []string) []string {
	taintedModules := []string{}
	for _, module := range modules {
		logrus.Debug(fmt.Sprintf("Looking for module in tree: %s", module))
		if !nt.ModuleInTree(module) {
			taintedModules = append(taintedModules, module)
		}
	}
	return taintedModules
}

func TaintsAccepted(confTaints []configuration.AcceptedKernelTaintsInfo, taintedModules []string) bool {
	for _, taintedModule := range taintedModules {
		found := false
		logrus.Debug("Accepted Taints from Config: ", confTaints)
		for _, confTaint := range confTaints {
			logrus.Debug(fmt.Sprintf("Comparing confTaint: %s to taintedModule: %s", confTaint.Module, taintedModule))
			if confTaint.Module == taintedModule {
				found = true
				break
			}
		}

		if !found {
			// Tainted modules were not found to be in the allow-list.
			return false
		}
	}
	return true
}
