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

// PrivilegedPodJSON test templates for privileged pods
var PrivilegedPodJSON = string(`{
  "testcase": [
    {
      "name": "HOST_NETWORK_CHECK",
      "skiptest": true,
      "command": "oc get pod  %s  -n %s -o json  | jq -r '.spec.hostNetwork'",
      "action": "allow",
	   "loop": 0,
      "expectedstatus": [
        "NULL_FALSE"
      ]
    },
    {
      "name": "HOST_PORT_CHECK",
      "skiptest": true,
      "loop": 1,
      "command": "oc get pod %s -n %s -o go-template='{{range (index .spec.containers %d).ports }}{{.hostPort}}{{end}}'",
      "action": "allow",
      "expectedstatus": [
        "^(<no value>)*$"
      ]
    },
    {
      "name": "HOST_PATH_CHECK",
      "skiptest": true,
       "loop": 0,
      "command": "oc get pods %s -n %s -o go-template='{{range .spec.volumes}}{{.hostPath.path}}{{end}}'",
      "action": "allow",
      "expectedstatus": [
        "^(<no value>)*$"
      ]
    },
    {
      "name": "HOST_IPC_CHECK",
      "skiptest": true,
      "loop": 0,
      "command": "oc get pod  %s  -n %s -o json  | jq -r '.spec.hostipc'",
      "action": "allow",
      "expectedstatus": [
        "NULL_FALSE"
      ]
    },
    {
      "name": "HOST_PID_CHECK",
      "skiptest": true,
       "loop": 0,
      "command": "oc get pod  %s  -n %s -o json  | jq -r '.spec.hostpid'",
      "action": "allow",
      "expectedstatus": [
        "NULL_FALSE"
      ]
    },
    {
      "name": "CAPABILITY_CHECK",
      "skiptest": true,
      "loop": 1,
      "command": "oc get pod %s -n %s -o json  | jq -r '.spec.containers[%d].securityContext.capabilities.add'",
      "resultType": "array",
      "action": "deny",
      "expectedstatus": [
        "NET_ADMIN",
        "SYS_ADMIN",
        "NET_RAW",
        "IPC_LOCK"
      ]
    },
    {
      "name": "ROOT_CHECK",
      "skiptest": true,
       "loop": 1,
      "command": "oc get pod %s -n %s -o json  | jq -r '.spec.containers[%d].securityContext.runAsUser'",
      "resulttype": "string",
      "action": "allow",
      "expectedstatus": [
        "NON_ZERO_NUMBER"
      ]
    },
    {
      "name": "PRIVILEGE_ESCALATION",
      "skiptest": true,
      "loop": 1,
      "command": "oc get pod  %s -n %s -o json  | jq -r '.spec.containers[%d].securityContext.allowPrivilegeEscalation'",
      "action": "allow",
      "expectedstatus": [
        "NULL_FALSE"
      ]
    }
  ]
}`)
