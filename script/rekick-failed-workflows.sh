#!/bin/bash

# Check if the user is logged in
if ! gh auth status; then
	echo "You are not logged in. Please log in with 'gh auth login'."
	exit 1
fi

# Set the default repository to the current repository.
gh repo set-default redhat-best-practices-for-k8s/certsuite

# This script will rekick failed workflows in this project with the 'gh' command line tool.
WORKFLOWS_TO_CHECK=(
	"QE OCP 4.14 Testing"
	"QE OCP 4.15 Testing"
	"QE OCP 4.16 Testing"
	"QE OCP 4.17 Testing"
	"QE OCP 4.18 Testing"
	"QE OCP 4.14 Intrusive Testing"
	"QE OCP 4.15 Intrusive Testing"
	"QE OCP 4.16 Intrusive Testing"
	"QE OCP 4.17 Intrusive Testing"
	"QE OCP 4.18 Intrusive Testing"
	"OCP ARM64 4.16 QE Testing"
	"qe-ocp-hosted.yml"
)

# Loop through the workflows and rekick any failed runs.
for workflow in "${WORKFLOWS_TO_CHECK[@]}"; do
	echo "Checking workflow: $workflow"
	for run_id in $(gh run list --limit 20 --workflow "$workflow" --json conclusion,databaseId | jq -r '.[] | select(.conclusion == "failure" or .conclusion == "timed_out") | .databaseId'); do
		gh run rerun "$run_id" --failed
	done
done
