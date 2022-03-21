// Copyright (C) 2020-2022 Red Hat, Inc.
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

package declaredandlistening

import (
	"strconv"
	"strings"
)

const (
	indexprotocolname = 0
	indexport         = 4
)

type Key struct {
	Port     int
	Protocol string
}

func ParseListening(res string, listeningPorts map[Key]bool) {
	var k Key
	lines := strings.Split(res, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if !strings.Contains(line, "LISTEN") {
			continue
		}
		if indexprotocolname > len(fields) || indexport > len(fields) {
			return
		}
		s := strings.Split(fields[indexport], ":")
		p, _ := strconv.Atoi(s[1])
		k.Port = p
		k.Protocol = strings.ToUpper(fields[indexprotocolname])
		k.Protocol = strings.ReplaceAll(k.Protocol, "\"", "")
		listeningPorts[k] = true
	}
}

func CheckIfListenIsDeclared(listeningPorts, declaredPorts map[Key]bool) map[Key]bool {
	res := make(map[Key]bool)
	if len(listeningPorts) == 0 {
		return res
	}
	for k := range listeningPorts {
		_, ok := declaredPorts[k]
		if !ok {
			res[k] = listeningPorts[k]
		}
	}
	return res
}
