// Copyright (C) 2020-2026 Red Hat, Inc.
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

package bootparams

import (
	"testing"

	mcv1 "github.com/openshift/api/machineconfiguration/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func TestGetMcKernelArguments_SingleKeyValue(t *testing.T) {
	env := &provider.TestEnvironment{
		Nodes: map[string]provider.Node{
			"node1": {
				Mc: provider.MachineConfig{
					MachineConfig: &mcv1.MachineConfig{
						Spec: mcv1.MachineConfigSpec{
							KernelArguments: []string{"nosmt=true"},
						},
					},
				},
			},
		},
	}

	result := GetMcKernelArguments(env, "node1")
	assert.Equal(t, "true", result["nosmt"])
}

func TestGetMcKernelArguments_MultipleArgs(t *testing.T) {
	env := &provider.TestEnvironment{
		Nodes: map[string]provider.Node{
			"node1": {
				Mc: provider.MachineConfig{
					MachineConfig: &mcv1.MachineConfig{
						Spec: mcv1.MachineConfigSpec{
							KernelArguments: []string{"nosmt=true", "hugepagesz=1G", "hugepages=16"},
						},
					},
				},
			},
		},
	}

	result := GetMcKernelArguments(env, "node1")
	assert.Len(t, result, 3)
	assert.Equal(t, "true", result["nosmt"])
	assert.Equal(t, "1G", result["hugepagesz"])
	assert.Equal(t, "16", result["hugepages"])
}

func TestGetMcKernelArguments_FlagWithoutValue(t *testing.T) {
	env := &provider.TestEnvironment{
		Nodes: map[string]provider.Node{
			"node1": {
				Mc: provider.MachineConfig{
					MachineConfig: &mcv1.MachineConfig{
						Spec: mcv1.MachineConfigSpec{
							KernelArguments: []string{"nosmt"},
						},
					},
				},
			},
		},
	}

	result := GetMcKernelArguments(env, "node1")
	assert.Len(t, result, 1)
	assert.Equal(t, "", result["nosmt"])
}

func TestGetMcKernelArguments_EmptyArgs(t *testing.T) {
	env := &provider.TestEnvironment{
		Nodes: map[string]provider.Node{
			"node1": {
				Mc: provider.MachineConfig{
					MachineConfig: &mcv1.MachineConfig{
						Spec: mcv1.MachineConfigSpec{
							KernelArguments: []string{},
						},
					},
				},
			},
		},
	}

	result := GetMcKernelArguments(env, "node1")
	assert.Empty(t, result)
}
