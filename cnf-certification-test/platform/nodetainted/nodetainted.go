// Copyright (C) 2021 Red Hat, Inc.
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
)

// NodeTainted holds information about tainted nodes.
type NodeTainted struct {
	ClientHolder *clientsholder.ClientsHolder
	OCPContext   clientsholder.Context
}

//go:generate moq -out nodetainted_moq.go . TaintedFuncs
type TaintedFuncs interface {
	runCommand(cmd string, ctx clientsholder.Context) (string, error)
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

func (nt *NodeTainted) runCommand(cmd string, ctx clientsholder.Context) (string, error) {
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
	output, err := nt.runCommand(`cat /proc/sys/kernel/tainted`, ctx)
	if err != nil {
		return "", err
	}
	output = strings.ReplaceAll(output, "\n", "")
	output = strings.ReplaceAll(output, "\r", "")
	output = strings.ReplaceAll(output, "\t", "")
	return output, nil
}

func removeEmptyStrings(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func (nt *NodeTainted) GetModulesFromNode(ctx clientsholder.Context) []string {
	// Get the 1st column list of the modules running on the node.
	// Split on the return/newline and get the list of the modules back.
	//nolint:goconst // used only once
	command := `chroot /host lsmod | awk '{ print $1 }' | grep -v Module`
	output, _ := nt.runCommand(command, ctx)
	output = strings.ReplaceAll(output, "\t", "")
	moduleList := strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n")
	return removeEmptyStrings(moduleList)
}

func (nt *NodeTainted) ModuleInTree(moduleName string, ctx clientsholder.Context) bool {
	command := `chroot /host cat /sys/module/` + moduleName + `/taint`
	cmdOutput, _ := nt.runCommand(command, ctx)
	return !strings.Contains(cmdOutput, "O")
}

func GetTaintedBitValues() []string {
	return []string{"proprietary module was loaded",
		"module was force loaded",
		"kernel running on an out of specification system",
		"module was force unloaded",
		"processor reported a Machine Check Exception (MCE)",
		"bad page referenced or some unexpected page flags",
		"taint requested by userspace application",
		"kernel died recently, i.e. there was an OOPS or BUG",
		"ACPI table overridden by user",
		"kernel issued warning",
		"staging driver was loaded",
		"workaround for bug in platform firmware applied",
		"externally-built (“out-of-tree”) module was loaded",
		"unsigned module was loaded",
		"soft lockup occurred",
		"kernel has been live patched",
		"auxiliary taint, defined for and used by distros",
		"kernel was built with the struct randomization plugin",
	}
}

//nolint:gocritic
func DecodeKernelTaints(bitmap uint64) (string, []string) {
	values := GetTaintedBitValues()
	var out string
	individualTaints := []string{}
	for i := 0; i < 32; i++ {
		bit := (bitmap >> i) & 1
		if bit == 1 {
			out += fmt.Sprintf("%s, ", values[i])
			// Storing the individual taint messages for extra parsing.
			individualTaints = append(individualTaints, values[i])
		}
	}
	return out, individualTaints
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
