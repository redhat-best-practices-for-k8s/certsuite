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

package cnf

// RolesJSON test templates for testing permission
//nolint:lll
var RolesJSON = string(`{
  "testcase": [
    {
      "name": "CLUSTER_ROLE_BINDING_BY_SA",
      "skiptest": true,
      "command": "oc get clusterrolebinding -n %s -o json | jq --arg name 'ServiceAccount' --arg null ',null,' --arg subjects 'subjects' --arg ns '%s' --arg sa '%s' -jr 'if (.items|length)>0 then .items[] | if (has($subjects)) then .subjects[] | select((.namespace==$ns) and (.kind==$name) and (.name==$sa)).name else $null end else $null end'",
      "action": "deny",
	  "loop": 0,
      "resulttype": "array",
      "expectedtype": "function",
      "expectedstatus": [
        "FN_SERVICE_ACCOUNT_NAME"
      ]
    },
    {
      "name": "ROLE_BINDING_BY_SA",
      "skiptest": true,
      "loop": 0,
      "command": "oc get rolebinding -n %s -o json | jq --arg name 'ServiceAccount' --arg null ',null,' --arg ns '%s' --arg subjects 'subjects' --arg sa '%s' -jr 'if (.items|length)>0 then .items[] | if (has($subjects)) then .subjects[] | select((.namespace==$ns) and (.kind==$name) and (.name==$sa)).name else $null end else $null end'",
      "action": "allow",
      "resulttype": "array",
      "expectedtype": "function",
      "expectedstatus": [
        "FN_SERVICE_ACCOUNT_NAME",
        "null"
      ]
    }
  ]
}`)
