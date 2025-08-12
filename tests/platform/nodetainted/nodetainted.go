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

// NodeTainted holds information about tainted nodes.
//
// It contains a context used to run commands on the node and the name of
// the node being inspected. The methods on this type query the kernel
// taint mask, list modules that set taints, and combine these results
// into useful diagnostics.
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

// NewNodeTaintedTester creates a new NodeTainted tester.
//
// It accepts a context holder and a node name, then returns a NodeTainted instance
// that can be used to verify whether the specified node has any kernel taints.
// The returned object provides methods for checking taint status against the
// known list of expected taints.
func NewNodeTaintedTester(context *clientsholder.Context, node string) *NodeTainted {
	return &NodeTainted{
		ctx:  context,
		node: node,
	}
}

// GetKernelTaintsMask returns a bitmask of kernel taints present on the node.
//
// It executes a system command to retrieve the current kernel taint state,
// parses the output, and converts each taint into its corresponding
// numeric mask value. The combined mask is returned as a uint64 along with
// an error if any step fails, such as command execution or parsing failures.
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

// KernelTaint represents a kernel taint with its descriptive text and corresponding letter codes.
//
// It holds two string fields: Description provides a human‑readable explanation of the taint,
// while Letters contains the single or multi‑character code used by the kernel to identify it.
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

// GetTaintMsg returns a formatted message describing the taint status for a given node index.
//
// It takes an integer representing the node index and looks up the corresponding
// kernel taint information from the internal kernelTaints slice.
// The function formats this data into a human‑readable string using Sprintf,
// returning the resulting message. If no taint is found for the index, it returns
// an empty string.
func GetTaintMsg(bit int) string {
	if taintMsg, exists := kernelTaints[bit]; exists {
		return fmt.Sprintf("%s (tainted bit %d)", taintMsg.Description, bit)
	}

	return fmt.Sprintf("reserved (tainted bit %d)", bit)
}

// DecodeKernelTaintsFromBitMask converts a bit mask into a slice of kernel taint strings.
//
// It interprets each set bit in the provided uint64 value as a specific kernel taint,
// retrieves the corresponding message using GetTaintMsg, and appends it to a
// string slice. The function returns this slice of taint messages.
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

// RemoveAllExceptNumbers strips all non-numeric characters from a string and returns the result.
//
// It accepts a single string argument and uses a regular expression to replace any character that is not a digit with an empty string.
// The resulting string contains only numeric characters, preserving their original order.
func RemoveAllExceptNumbers(incomingStr string) string {
	// example string ", bit:10)"
	// return 10

	// remove all characters except numbers
	re := regexp.MustCompile(`\D+`)
	return re.ReplaceAllString(incomingStr, "")
}

// DecodeKernelTaintsFromLetters converts a string of single‑letter taint codes into a slice of kernel taint strings.
//
// It interprets each character in the input as a code for a specific kernel taint,
// maps it to its full name, and returns a slice containing those names.
// If an unknown letter is encountered, it is ignored. The function never returns nil; if no valid letters are found it returns an empty slice.
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

// getBitPosFromLetter returns the kernel taint bit position for a given letter.
//
// It takes a single string argument containing the letter that represents a
// module's taint and returns its corresponding zero‑based bit index as an
// integer. If the input is empty or does not correspond to a known taint,
// an error is returned. The function looks up the letter in the internal
// kernelTaints slice to determine the position.
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

// GetTaintedBitsByModules retrieves taint bits for each module.
//
// It accepts a map of module names to their associated taint letters and
// returns a map where the key is the bit position and the value indicates
// whether that bit is set. If any letter cannot be converted to a bit
// position, an error is returned.
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

// GetOtherTaintedBits returns a slice of bit positions that are tainted but not associated with any module.
//
// It takes a 64-bit value representing the current taint state and a map from module indices to booleans indicating
// whether each module is present. The function iterates over all bits in the value, checks if the corresponding module
// exists, and appends the bit index to the result slice when the bit is set and its module is absent.
// The returned slice contains only those bit positions that are considered "other" tainted bits.
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

// getAllTainterModules retrieves a map of taint module names to their descriptions.
//
// It runs the underlying command that lists all available taint modules,
// parses the output into key/value pairs, and returns them in a map.
// If the command fails or the output is malformed, an error is returned.
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

// GetTainterModules retrieves modules that set kernel taint bits on a node.
//
// It executes a command on the node to list all taint‑setting modules, then parses
// the output into two maps: one mapping module names to the letters representing
// their individual taints (excluding any modules in the allowlist), and another
// mapping bit positions to true for every kernel taint bit set by any module,
// including those from the allowlist. The function returns these maps and an
// error if command execution or parsing fails.
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
