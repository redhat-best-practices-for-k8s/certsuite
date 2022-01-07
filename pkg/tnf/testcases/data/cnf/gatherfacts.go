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

// GatherPodFactsJSON test templates for collecting facts
var GatherPodFactsJSON = string(`{
  "testcase": [
    {
      "name": "NAME",
      "skiptest": false,
      "command": "oc get pod %s -n %s -o json | jq -r '.metadata.name'",
      "action": "allow",
      "resulttype": "string",
      "expectedtype": "regex",
      "expectedstatus": [
        "ALLOW_ALL"
      ]
    },
    {
      "name": "CONTAINER_COUNT",
      "skiptest": false,
      "command": "oc get pod %s -n %s -o json | jq -r '.spec.containers | length'",
      "action": "allow",
      "resulttype": "integer",
      "expectedType": "regex",
      "expectedstatus": [
        "DIGIT"
      ]
    },
    {
      "name": "SERVICE_ACCOUNT_NAME",
      "skiptest": false,
      "command": "oc get pod %s -n %s -o json | jq -r '.spec.serviceAccountName'",
      "action": "allow",
      "resulttype": "string",
      "expectedType": "regex",
      "expectedstatus": [
        "ALLOW_ALL"
      ]
    }
  ]
}`)
