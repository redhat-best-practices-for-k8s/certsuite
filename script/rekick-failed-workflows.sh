#!/bin/bash

# Default number of days to look back for failed workflows
DAYS_BACK=1

# Parse command line arguments
while [[ $# -gt 0 ]]; do
	case $1 in
	-d | --days)
		DAYS_BACK="$2"
		shift 2
		;;
	-h | --help)
		echo "Usage: $0 [-d|--days <number>]"
		echo "  -d, --days <number>  Number of days back to check for failed workflows (default: 1)"
		echo "  -h, --help           Show this help message"
		exit 0
		;;
	*)
		echo "Unknown option: $1"
		echo "Use -h or --help for usage information"
		exit 1
		;;
	esac
done

# Validate days input
if ! [[ "$DAYS_BACK" =~ ^[0-9]+$ ]] || [ "$DAYS_BACK" -le 0 ]; then
	echo "Error: Days must be a positive integer"
	exit 1
fi

# Calculate the date cutoff
if [[ "$OSTYPE" == "darwin"* ]]; then
	# macOS date command
	CUTOFF_DATE=$(date -v-"${DAYS_BACK}"d -u +"%Y-%m-%dT%H:%M:%SZ")
else
	# Linux date command
	CUTOFF_DATE=$(date -d "${DAYS_BACK} days ago" -u +"%Y-%m-%dT%H:%M:%SZ")
fi

echo "Looking for failed workflows created after: $CUTOFF_DATE (last $DAYS_BACK day(s))"

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
	"QE OCP 4.16 Testing"
	"QE OCP 4.17 Testing"
	"QE OCP 4.18 Testing"
	"QE OCP 4.19 Testing"
	"QE OCP 4.20 Testing"
	# TODO: Enable when quick-ocp supports it
	# "QE OCP 4.21 Testing"
	"QE OCP 4.14 Intrusive Testing"
	"QE OCP 4.16 Intrusive Testing"
	"QE OCP 4.17 Intrusive Testing"
	"QE OCP 4.18 Intrusive Testing"
	"QE OCP 4.19 Intrusive Testing"
	"QE OCP 4.20 Intrusive Testing"
	# TODO: Enable when quick-ocp supports it
	# "QE OCP 4.21 Intrusive Testing"
	"OCP ARM64 4.16 QE Testing"
	"qe-ocp-hosted.yml"
)

# Loop through the workflows and rekick any failed runs.
for workflow in "${WORKFLOWS_TO_CHECK[@]}"; do
	echo "Checking workflow: $workflow"
	# Get workflow runs with date filtering
	failed_runs=$(
		gh run list --limit 200 --workflow "$workflow" --json conclusion,databaseId,createdAt,updatedAt |
			jq --arg cutoff "$CUTOFF_DATE" -r '
			.[] |
			select(.conclusion == "failure" or .conclusion == "timed_out" or .conclusion == "cancelled") |
			select((.updatedAt != null and .updatedAt >= $cutoff) or (.createdAt != null and .createdAt >= $cutoff)) |
			.databaseId'
	)

	if [ -z "$failed_runs" ]; then
		echo "  No failed runs found in the last $DAYS_BACK day(s) for workflow: $workflow"
		continue
	fi

	for run_id in $failed_runs; do
		echo "  Re-running failed workflow run: $run_id"
		gh run rerun "$run_id" --failed
	done
done
