// Copyright (C) 2020-2021 Red Hat, Inc.
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

package operator

// OperatorJSON test templates for collecting operator status
var OperatorJSON = string(`{
  "testcase": [
    {
      "name": "CSV_INSTALLED",
      "skiptest": true,
      "command": "oc get csv %s -n %s -o json | jq -r '.status.phase'",
      "action": "allow",
      "resulttype": "string",
      "expectedstatus": [
        "Succeeded"
      ]
    },
    {
      "name": "CSV_SCC",
      "skiptest": true,
      "command": "oc get csv %s -n %s -o json | jq -r 'if .spec.install.spec.clusterPermissions == null then null else . end ` +
	`| if . == null then \"EMPTY\" else .spec.install.spec.clusterPermissions[].rules[].resourceNames end | if length == 0 then \"EMPTY\" else . end'",
      "action": "allow",
      "resulttype": "string",
      "expectedstatus": [
        "EMPTY"
      ]
    }
  ]
}`)
