Add test case template  file here
1. Define a variable with json template (TODO create schema)
var TEST_CASE_NAME_HERE = string(`{
"testcase": [
    {
      "name": "name of the test that has to defined, a constant",
      "skiptest": true,
      "command": "oc get pod  %s  -n %s -o json  | jq -r '.spec.hostNetwork'",
      "resultType": "string|array", //return result type from the command
      "action": "allow|deny",
      "expectedtype" "string|function",
      "expectedstatus": [
        "NULL_FALSE"
      ]
    },
}')'
2.update data/config.go with struct