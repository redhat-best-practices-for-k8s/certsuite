#!/bin/bash

# Note: This script is currently unused and is not being called from anywhere.
# It is being kept here for future reference.

# Loop through all of the test suites to be able to build the JSON correctly
SUITES=(access-control affiliated-certification lifecycle manageability networking observability operator performance platform-alteration)

for SUITE_VAL in "${SUITES[@]}"; do
	# Write the JSON blobs to a temp file
	#shellcheck disable=SC2013,SC2002
	for i in $(cat CATALOG.md | grep "Unique ID" | grep "$SUITE_VAL" | sed 's/Unique ID|//'); do
		jq --null-input --arg id "$i" --arg suite "$SUITE_VAL" '{"id": $id, "suite": $suite}' >>temp.txt
	done
done

# Display the JQ to the user
#shellcheck disable=SC2005,SC2002
RESULT=$(echo "$(cat temp.txt | jq -n '.grades.requiredPassingTests |= [inputs] | .grades.gradeName = "good"')")
rm temp.txt
rm generated_policy.json
echo "$RESULT" >generated_policy.json
