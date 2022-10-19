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
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
)

// NodeTainted holds information about tainted nodes.
type NodeTainted struct {
	ClientHolder *clientsholder.ClientsHolder
}

type TaintedFuncs interface {
	runCommand(ctx clientsholder.Context, cmd string) (string, error)
	GetKernelTaintInfo(ctx clientsholder.Context) (string, error)
	GetModulesFromNode(ctx clientsholder.Context) []string
	ModuleInTree(moduleName string, ctx clientsholder.Context) bool
	GetOutOfTreeModules(modules []string, ctx clientsholder.Context) []string
}

// NewNodeTainted creates a new NodeTainted tester
func NewNodeTaintedTester(client *clientsholder.ClientsHolder) *NodeTainted {
	return &NodeTainted{
		ClientHolder: client,
	}
}

func (nt *NodeTainted) runCommand(ctx clientsholder.Context, cmd string) (string, error) {
	output, outerr, err := nt.ClientHolder.ExecCommandContainer(ctx, cmd)
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

func (nt *NodeTainted) GetKernelTaintInfo(ctx clientsholder.Context) (string, error) {
	output, err := nt.runCommand(ctx, `cat /proc/sys/kernel/tainted`)
	if err != nil {
		return "", err
	}
	output = strings.ReplaceAll(output, "\n", "")
	output = strings.ReplaceAll(output, "\r", "")
	output = strings.ReplaceAll(output, "\t", "")
	return output, nil
}

func (nt *NodeTainted) GetModulesFromNode(ctx clientsholder.Context) []string {
	// Get the 1st column list of the modules running on the node.
	// Split on the return/newline and get the list of the modules back.
	command := `chroot /host lsmod | awk '{ print $1 }' | grep -v Module`
	output, _ := nt.runCommand(ctx, command)
	output = strings.ReplaceAll(output, "\t", "")
	moduleList := strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n")
	return stringhelper.RemoveEmptyStrings(moduleList)
}

func (nt *NodeTainted) ModuleInTree(moduleName string, ctx clientsholder.Context) bool {
	command := `chroot /host cat /sys/module/` + moduleName + `/taint`
	cmdOutput, _ := nt.runCommand(ctx, command)
	return !strings.Contains(cmdOutput, "O")
}

var kernelTaints = map[int]string{
	// Linux standard kernel taints
	0:  "proprietary module was loaded",
	1:  "module was force loaded",
	2:  "kernel running on an out of specification system",
	3:  "module was force unloaded",
	4:  "processor reported a Machine Check Exception (MCE)",
	5:  "bad page referenced or some unexpected page flags",
	6:  "taint requested by userspace application",
	7:  "kernel died recently, i.e. there was an OOPS or BUG",
	8:  "ACPI table overridden by user",
	9:  "kernel issued warning",
	10: "staging driver was loaded",
	11: "workaround for bug in platform firmware applied",
	12: "externally-built (“out-of-tree”) module was loaded",
	13: "unsigned module was loaded",
	14: "soft lockup occurred",
	15: "kernel has been live patched",
	16: "auxiliary taint, defined for and used by distros",
	17: "kernel was built with the struct randomization plugin",
	18: "an in-kernel test has been run",

	// RedHat custom taints for RHEL/CoreOS
	// https://access.redhat.com/solutions/40594
	27: "Red Hat extension: Hardware for which support has been removed. / OMGZOMBIES easter egg.",
	28: "Red Hat extension: Unsupported hardware. Refer to \"UNSUPPORTED HARDWARE DEVICE:\" kernel log entry for details.",
	29: "Red Hat extension: Technology Preview code was loaded; cf. Technology Preview features support scope description. Refer to \"TECH PREVIEW:\" kernel log entry for details.",
	30: "Red Hat extension: reserved taint bit 30",
	31: "Red Hat extension: reserved taint bit 31",
}

func getTaintMsg(bit int) string {
	if taintMsg, exists := kernelTaints[bit]; exists {
		return taintMsg
	}

	return fmt.Sprintf("reserved kernel taint bit %d", bit)
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

func (nt *NodeTainted) GetOutOfTreeModules(modules []string, ctx clientsholder.Context) []string {
	taintedModules := []string{}
	for _, module := range modules {
		logrus.Debug(fmt.Sprintf("Looking for module in tree: %s", module))
		if !nt.ModuleInTree(module, ctx) {
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
