// Copyright (C) 2022-2024 Red Hat, Inc.
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

package postmortem

import (
	"fmt"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	corev1 "k8s.io/api/core/v1"
)

// Log retrieves the current test environment state and signals that it may need to be refreshed.
//
// The function first obtains the current test environment information, then marks the
// environment as needing a refresh. It again fetches the environment data, formats it into
// a readable string using Sprintf, and returns this string. The returned value is intended
// for use in postmortem reports or debugging output.
func Log() (out string) {
	// Get current environment
	env := provider.GetTestEnvironment()

	// Set refresh
	env.SetNeedsRefresh()

	// Get up-to-date environment
	env = provider.GetTestEnvironment()

	out += "\nNode Status:\n"
	for _, n := range env.Nodes {
		out += fmt.Sprintf("node name=%s taints=%+v", n.Data.Name, n.Data.Spec.Taints) + "\n"
	}
	out += "\nPending Pods:\n"
	for _, p := range env.AllPods {
		if p.Status.Phase != corev1.PodSucceeded && p.Status.Phase != corev1.PodRunning {
			out += p.String() + "\n"
		}
	}
	out += "\nAbnormal events:\n"
	for _, e := range env.AbnormalEvents {
		out += e.String() + "\n"
	}
	return out
}
