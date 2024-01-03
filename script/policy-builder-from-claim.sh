#!/bin/bash

# Only collect the tests that are performed in Kind by github workflows
jq '.[] | .results | .[] | select(.[].testID.tags!="affiliated-certification-container-is-certified-digest" and .[].testID.tags!="access-control-security-context") | .[] | .testID' cnf-certification-test/claim.json >temp.txt

# Display the JQ to the user
#shellcheck disable=SC2005,SC2002
RESULT=$(echo "$(cat temp.txt | jq -n '.grades.requiredPassingTests |= [inputs] | .grades.gradeName = "good"')")
rm temp.txt
rm generated_policy.json
echo "$RESULT" >generated_policy.json
