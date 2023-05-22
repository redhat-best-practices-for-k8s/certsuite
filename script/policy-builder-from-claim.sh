#!/bin/bash

# Only collect the common tests
jq '.[] | .results | .[] | select(.[].testID.tags=="common") | .[] | .testID' cnf-certification-test/claim.json >temp.txt

# Display the JQ to the user
#shellcheck disable=SC2005,SC2002
RESULT=$(echo "$(cat temp.txt | jq -n '.grades.requiredPassingTests |= [inputs] | .grades.gradeName = "good"')")
rm temp.txt
rm generated_policy.json
echo "$RESULT" >generated_policy.json
