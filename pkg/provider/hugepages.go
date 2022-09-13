// Copyright (C) 2022 Red Hat, Inc.
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

package provider

import (
	"strings"
)

const (
	hugePages2Mi = "hugepages-2Mi"
	hugePages1Gi = "hugepages-1Gi"
	hugePages    = "hugepages"
)

func (p *Pod) HasHugepages() bool {
	// Pods may contain more than one container.  All containers must conform to the CPU isolation requirements.
	for _, cut := range p.Containers {
		for name := range cut.Resources.Requests {
			if strings.Contains(name.String(), hugePages) {
				return true
			}
		}
		for _, name := range cut.Resources.Limits {
			if strings.Contains(name.String(), hugePages) {
				return true
			}
		}
	}
	return false
}

func (p *Pod) CheckResourceOnly2MiHugePages() (bool, error) {
	// Pods may contain more than one container.  All containers must conform to the CPU isolation requirements.
	for _, cut := range p.Containers {
		// Resources must be specified
		if len(cut.Resources.Requests) == 0 || len(cut.Resources.Limits) == 0 {
			continue
		}
		for name := range cut.Resources.Requests {
			if strings.Contains(name.String(), hugePages) && name != hugePages2Mi {
				return false, nil
			}
		}
		for name := range cut.Resources.Limits {
			if strings.Contains(name.String(), hugePages) && name != hugePages2Mi {
				return false, nil
			}
		}
	}
	return true, nil
}
