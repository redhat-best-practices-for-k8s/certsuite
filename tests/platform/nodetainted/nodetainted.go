// Copyright (C) 2021-2024 Red Hat, Inc.
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
	"regexp"
	"strconv"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

// NodeTainted provides access to kernel taint information for a node
//
// It holds the context and node name used to query system files and run shell
// commands that expose kernel taints. The struct offers methods to retrieve the
// numeric taint mask, list modules that set taints, and parse those module
// taints from /sys/module. These functions enable inspection of tainted states
// on a target node.
type NodeTainted struct {
	ctx  *clientsholder.Context
	node string
}

var runCommand = func(ctx *clientsholder.Context, cmd string) (string, error) {
	ch := clientsholder.GetClientsHolder()
	output, outerr, err := ch.ExecCommandContainer(*ctx, cmd)
	if err != nil {
		log.Error("can not execute command on container, err=%v", err)
		return "", err
	}
	if outerr != "" {
		log.Error("Error when running nodetainted command err=%v", outerr)
		return "", errors.New(outerr)
	}
	return output, nil
}

// NewNodeTaintedTester Creates a tester for checking kernel taints on a node
//
// This function constructs and returns a new instance of the NodeTainted type.
// It stores the provided client context and node name so that subsequent
// methods can interact with the node’s kernel taint state via the Kubernetes
// API. The returned object is used by test logic to retrieve and analyze taints
// for compliance checks.
func NewNodeTaintedTester(context *clientsholder.Context, node string) *NodeTainted {
	return &NodeTainted{
		ctx:  context,
		node: node,
	}
}

// NodeTainted.GetKernelTaintsMask Retrieves the kernel taints bitmask from a node
//
// This method runs a command to read /proc/sys/kernel/tainted, cleans up any
// whitespace characters, then parses the resulting string as an unsigned
// integer in base ten. If parsing fails it returns an error indicating the
// malformed value. On success it returns the taints mask as a uint64 and a nil
// error.
func (nt *NodeTainted) GetKernelTaintsMask() (uint64, error) {
	output, err := runCommand(nt.ctx, `cat /proc/sys/kernel/tainted`)
	if err != nil {
		return 0, err
	}
	output = strings.ReplaceAll(output, "\n", "")
	output = strings.ReplaceAll(output, "\r", "")
	output = strings.ReplaceAll(output, "\t", "")

	// Convert to number.
	taintsMask, err := strconv.ParseUint(output, 10, 64) // base 10 and uint64
	if err != nil {
		return 0, fmt.Errorf("failed to decode taints mask %q: %w", output, err)
	}

	return taintsMask, nil
}

// KernelTaint Represents an individual kernel taint
//
// This structure holds the human-readable description of a taint as well as its
// identifying letters used by the kernel to mark nodes. The Description field
// explains why the taint exists, while Letters contains the short string that
// is applied to node metadata. Instances are typically collected and examined
// when evaluating node health or scheduling constraints.
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
	12: {"externally-built (\"out-of-tree\") module was loaded", "O"},
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

// GetTaintMsg Retrieves a descriptive message for a kernel taint bit
//
// This function looks up the given integer bit in a predefined map of known
// kernel taints. If found, it returns the taint's description along with the
// bit number; otherwise it indicates the bit is reserved. The output string is
// used to label taint information throughout the test suite.
func GetTaintMsg(bit int) string {
	if taintMsg, exists := kernelTaints[bit]; exists {
		return fmt.Sprintf("%s (tainted bit %d)", taintMsg.Description, bit)
	}

	return fmt.Sprintf("reserved (tainted bit %d)", bit)
}

// DecodeKernelTaintsFromBitMask Converts a bitmask into human‑readable kernel taint messages
//
// The function iterates over all 64 bits of the supplied unsigned integer,
// checking each bit for a set value. For every bit that is on, it calls a
// helper to retrieve a descriptive message and appends that string to a slice.
// The resulting list of strings represents the active kernel taints
// corresponding to the original mask.
func DecodeKernelTaintsFromBitMask(bitmask uint64) []string {
	taints := []string{}
	for i := 0; i < 64; i++ {
		bit := (bitmask >> i) & 1
		if bit == 1 {
			taints = append(taints, GetTaintMsg(i))
		}
	}
	return taints
}

// RemoveAllExceptNumbers strips all non-digit characters from a string
//
// This function takes an input string, compiles a regular expression that
// matches any non‑digit sequence, and replaces those sequences with nothing.
// The result is a new string containing only the numeric characters that were
// present in the original input.
func RemoveAllExceptNumbers(incomingStr string) string {
	// example string ", bit:10)"
	// return 10

	// remove all characters except numbers
	re := regexp.MustCompile(`\D+`)
	return re.ReplaceAllString(incomingStr, "")
}

// DecodeKernelTaintsFromLetters Converts a string of taint letters into descriptive taint strings
//
// This routine iterates over each character in the input, matching it against a
// predefined list of kernel taints. For matched letters it builds a
// human‑readable description that includes the taint’s name, the letter
// used, and its bit index. If a letter is unknown it records an "unknown taint"
// entry. The resulting slice contains one entry per letter.
func DecodeKernelTaintsFromLetters(letters string) []string {
	taints := []string{}

	for _, l := range letters {
		taintLetter := string(l)
		found := false

		for i := range kernelTaints {
			kernelTaint := kernelTaints[i]
			if strings.Contains(kernelTaint.Letters, taintLetter) {
				taints = append(taints, fmt.Sprintf("%s (taint letter:%s, bit:%d)",
					kernelTaint.Description, taintLetter, i))
				found = true
				break
			}
		}

		// The letter does not belong to any known (yet) taint...
		if !found {
			taints = append(taints, fmt.Sprintf("unknown taint (letter %s)", taintLetter))
		}
	}

	return taints
}

// getBitPosFromLetter Finds the bit index of a kernel taint letter
//
// The function accepts a single-character string representing a module taint
// and searches through a predefined list of known kernel taints to determine
// its corresponding bit position. It returns that integer index if found,
// otherwise it produces an error indicating the letter is invalid or unknown.
// Input validation ensures only one character is processed.
func getBitPosFromLetter(letter string) (int, error) {
	if letter == "" || len(letter) > 1 {
		return 0, fmt.Errorf("input string must contain one letter")
	}

	for bit, taint := range kernelTaints {
		if strings.Contains(taint.Letters, letter) {
			return bit, nil
		}
	}

	return 0, fmt.Errorf("letter %s does not belong to any known kernel taint", letter)
}

// GetTaintedBitsByModules Collects kernel taint bits from module letters
//
// This function receives a map of modules to their taint letter strings. It
// iterates over each letter, converts it to the corresponding bit position
// using a helper, and records that bit as true in a result map. Errors are
// returned if any letter cannot be mapped to a known taint.
func GetTaintedBitsByModules(tainters map[string]string) (map[int]bool, error) {
	taintedBits := map[int]bool{}

	for tainter, letters := range tainters {
		// Save taint bits from this module.
		for i := range letters {
			letter := string(letters[i])
			bit, err := getBitPosFromLetter(letter)
			if err != nil {
				return nil, fmt.Errorf("module %s has invalid taint letter %s: %w", tainter, letter, err)
			}

			taintedBits[bit] = true
		}
	}

	return taintedBits, nil
}

// GetOtherTaintedBits Identifies kernel taint bits not associated with any module
//
// The function examines a 64‑bit mask of currently set kernel taints and
// compares each bit to a map that records which bits have been set by known
// modules. It iterates over all possible bit positions, collecting those that
// are active in the mask but absent from the module record. The result is a
// slice of integers representing the indices of these orphaned taint bits.
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

// NodeTainted.getAllTainterModules Retrieves all kernel modules that are tainting the node
//
// The function runs a shell command to list every module in /sys/module, reads
// each module's taint file if present, and collects non‑empty taints into a
// map keyed by module name. It returns this mapping or an error if the command
// fails or parsing encounters duplicate entries or malformed lines.
func (nt *NodeTainted) getAllTainterModules() (map[string]string, error) {
	const (
		command = "modules=`ls /sys/module`; for module_name in $modules; do taint_file=/sys/module/$module_name/taint; " +
			"if [ -f $taint_file ]; then taints=`cat $taint_file`; " +
			"if [[ ${#taints} -gt 0 ]]; then echo \"$module_name `cat $taint_file`\"; fi; fi; done"

		numFields       = 2
		posModuleName   = 0
		posModuleTaints = 1
	)

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

		// Save to the all tainters list.
		if taints, exist := tainters[moduleName]; exist {
			return nil, fmt.Errorf("module %s (taints %s) has already been parsed (taints %s)",
				moduleName, moduleTaints, taints)
		}

		tainters[moduleName] = moduleTaints
	}

	return tainters, nil
}

// NodeTainted.GetTainterModules Retrieves non-allowlisted modules that set kernel taint bits
//
// The method runs a command on the node to list all modules with taint letters,
// then filters out those present in an allowlist. It returns a map of module
// names to their taint letter strings and another map indicating which taint
// bits are set across all modules. Errors from command execution or parsing are
// wrapped and returned.
func (nt *NodeTainted) GetTainterModules(allowList map[string]bool) (tainters map[string]string, taintBits map[int]bool, err error) {
	// First, get all the modules that are tainting the kernel in this node.
	allTainters, err := nt.getAllTainterModules()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get tainter modules: %w", err)
	}

	filteredTainters := map[string]string{}
	for moduleName, moduleTaintsLetters := range allTainters {
		moduleTaints := DecodeKernelTaintsFromLetters(moduleTaintsLetters)
		log.Debug("%s: Module %s has taints (%s): %s", nt.node, moduleName, moduleTaintsLetters, moduleTaints)

		// Apply allowlist.
		if allowList[moduleName] {
			log.Debug("%s module %s is tainting the kernel but it has been allowlisted (taints: %v)",
				nt.node, moduleName, moduleTaints)
		} else {
			filteredTainters[moduleName] = moduleTaintsLetters
		}
	}

	// Finally, get all the bits that all the modules have set.
	taintBits, err = GetTaintedBitsByModules(allTainters)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get taint bits by modules: %w", err)
	}

	return filteredTainters, taintBits, nil
}
