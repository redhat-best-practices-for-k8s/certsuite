#!/bin/bash

# Test script to validate all links in CATALOG.md
# This script extracts URLs and tests them with curl

set -euo pipefail

CATALOG_FILE="${1:-CATALOG.md}"
TEMP_DIR=$(mktemp -d)
RESULTS_FILE="$TEMP_DIR/link_test_results.txt"
FAILED_LINKS_FILE="$TEMP_DIR/failed_links.txt"

# Configuration options (can be overridden by environment variables)
MAX_RETRIES=${LINK_TEST_MAX_RETRIES:-3}
INITIAL_BACKOFF=${LINK_TEST_INITIAL_BACKOFF:-1}
REQUEST_DELAY=${LINK_TEST_REQUEST_DELAY:-0.2} # Reduced from 0.5s since we have retries
CONNECT_TIMEOUT=${LINK_TEST_CONNECT_TIMEOUT:-15}
MAX_TIME=${LINK_TEST_MAX_TIME:-30}
# VERBOSE=${LINK_TEST_VERBOSE:-false}  # Currently unused, reserved for future use

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TOTAL_LINKS=0
PASSED_LINKS=0
FAILED_LINKS=0

cleanup() {
	# shellcheck disable=SC2317  # Function is called indirectly via trap
	rm -rf "$TEMP_DIR"
}

trap cleanup EXIT

log_info() {
	echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
	echo -e "${GREEN}[PASS]${NC} $1"
}

log_warning() {
	echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
	echo -e "${RED}[FAIL]${NC} $1"
}

extract_urls() {
	local file="$1"

	log_info "Extracting URLs from $file..."

	# Extract URLs with improved regex and cleanup
	# First extract all potential HTTPS URLs
	grep -o "https://[^|[:space:]]*" "$file" >"$TEMP_DIR/raw_urls.txt" || true

	# Further clean URLs and split multiple URLs on same line
	true >"$TEMP_DIR/all_urls.txt" # Clear the file

	if [[ -f "$TEMP_DIR/raw_urls.txt" ]]; then
		while IFS= read -r url_line; do
			if [[ -n "$url_line" ]]; then
				# Split on "and", "or", comma, semicolon to handle multiple URLs per line
				echo "$url_line" | sed 's/ and /\n/g; s/ or /\n/g; s/,/\n/g; s/;/\n/g' |
					while IFS= read -r url; do
						# Clean up each URL - remove leading/trailing whitespace
						url=$(echo "$url" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')
						# Remove leading punctuation that's not part of URLs
						url=${url#*([|(\[{])}
						# Remove trailing punctuation step by step (more reliable than complex regex)
						url=${url%%.}  # Remove trailing dot
						url=${url%%)}  # Remove trailing )
						url=${url%%]}  # Remove trailing ]
						url=${url%%\}} # Remove trailing }
						url=${url%%,}  # Remove trailing comma
						url=${url%%;}  # Remove trailing semicolon
						url=${url%%:}  # Remove trailing colon
						url=${url%%!}  # Remove trailing !
						url=${url%%\?} # Remove trailing ?
						url=${url%%|}  # Remove trailing |
						# Repeat for multiple trailing punctuation (like ").")
						url=${url%%.} # Remove trailing dot again
						url=${url%%)} # Remove trailing ) again
						# Only keep valid URLs
						if [[ "$url" =~ ^https://[^[:space:]]+$ ]] && [[ ${#url} -gt 10 ]]; then
							echo "$url" >>"$TEMP_DIR/all_urls.txt"
						fi
					done
			fi
		done <"$TEMP_DIR/raw_urls.txt"
	fi

	# Remove duplicates and count
	sort -u "$TEMP_DIR/all_urls.txt" >"$TEMP_DIR/unique_urls.txt"
	mv "$TEMP_DIR/unique_urls.txt" "$TEMP_DIR/all_urls.txt"

	TOTAL_LINKS=$(wc -l <"$TEMP_DIR/all_urls.txt")
	log_info "Found $TOTAL_LINKS unique URLs to test"

	# Debug: show first few URLs if in verbose mode
	if [[ "${LINK_TEST_VERBOSE:-}" == "true" ]]; then
		log_info "Sample URLs found:"
		head -5 "$TEMP_DIR/all_urls.txt" | while read -r url; do
			log_info "  - $url"
		done
	fi
}

test_url() {
	local url="$1"
	local base_url
	local anchor
	local http_status
	local content_file="$TEMP_DIR/page_content.html"
	local max_retries=$MAX_RETRIES
	local retry_count=0
	local curl_exit_code
	local backoff_delay=$INITIAL_BACKOFF

	# Split URL into base URL and anchor
	if [[ "$url" == *"#"* ]]; then
		base_url="${url%#*}"
		anchor="${url#*#}"
	else
		base_url="$url"
		anchor=""
	fi

	# Test HTTP status with retry mechanism
	log_info "Testing: $url"

	while [[ $retry_count -lt $max_retries ]]; do
		# Use curl to test the base URL with improved options
		set +e # Temporarily disable exit on error for curl
		http_status=$(curl -s -o "$content_file" -w "%{http_code}" \
			--location \
			--max-time "$MAX_TIME" \
			--connect-timeout "$CONNECT_TIMEOUT" \
			--retry 1 \
			--retry-delay 1 \
			--retry-max-time 60 \
			--retry-connrefused \
			--user-agent "Mozilla/5.0 (compatible; LinkChecker/1.0)" \
			--header "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8" \
			"$base_url" 2>/dev/null)
		curl_exit_code=$?
		set -e # Re-enable exit on error

		if [[ $curl_exit_code -eq 0 ]]; then
			# Success! Break out of retry loop
			break
		else
			((retry_count++))
			if [[ $retry_count -lt $max_retries ]]; then
				log_warning "Attempt $retry_count failed for $url, retrying in ${backoff_delay}s..."
				sleep "$backoff_delay"
				# Exponential backoff: 1s, 2s, 4s
				backoff_delay=$((backoff_delay * 2))
			else
				log_error "All $max_retries attempts failed for $url"
			fi
		fi
	done

	if [[ $curl_exit_code -eq 0 ]]; then
		if [[ "$http_status" =~ ^[2-3][0-9][0-9]$ ]]; then
			# HTTP status is success (2xx or 3xx)

			# If there's an anchor, check if it exists in the content
			if [[ -n "$anchor" ]]; then
				# Check for anchor in HTML content - handle both quoted and unquoted attributes
				# Look for id="anchor", id='anchor', id=anchor, name="anchor", name='anchor', name=anchor, or href="#anchor"
				if grep -qE "id=[\"']?${anchor}[\"']?" "$content_file" ||
					grep -qE "name=[\"']?${anchor}[\"']?" "$content_file" ||
					grep -q "href=[\"']#${anchor}[\"']" "$content_file"; then
					log_success "✓ $url (HTTP $http_status, anchor found)"
					echo "PASS: $url" >>"$RESULTS_FILE"
					((PASSED_LINKS++))
					return 0
				else
					log_error "✗ $url (HTTP $http_status, anchor '#$anchor' not found)"
					echo "FAIL: $url - Anchor not found" >>"$FAILED_LINKS_FILE"
					((FAILED_LINKS++))
					return 1
				fi
			else
				log_success "✓ $url (HTTP $http_status)"
				echo "PASS: $url" >>"$RESULTS_FILE"
				((PASSED_LINKS++))
				return 0
			fi
		else
			log_error "✗ $url (HTTP $http_status)"
			echo "FAIL: $url - HTTP $http_status" >>"$FAILED_LINKS_FILE"
			((FAILED_LINKS++))
			return 1
		fi
	else
		log_error "✗ $url (Connection failed)"
		echo "FAIL: $url - Connection failed" >>"$FAILED_LINKS_FILE"
		((FAILED_LINKS++))
		return 1
	fi
}

test_all_urls() {
	log_info "Starting URL validation..."
	echo "=== Link Test Results ===" >"$RESULTS_FILE"
	echo "=== Failed Links ===" >"$FAILED_LINKS_FILE"

	set +e # Don't exit on failed test_url calls
	while IFS= read -r url; do
		if [[ -n "$url" ]]; then
			test_url "$url" || true # Continue even if test_url fails
		fi
		# Small delay to be respectful to servers (reduced since we have retry logic)
		sleep "$REQUEST_DELAY"
	done <"$TEMP_DIR/all_urls.txt"
	set -e # Re-enable exit on error
}

print_summary() {
	echo ""
	echo "=========================================="
	echo "           LINK TEST SUMMARY"
	echo "=========================================="
	echo "Total links tested: $TOTAL_LINKS"
	echo -e "Passed: ${GREEN}$PASSED_LINKS${NC}"
	echo -e "Failed: ${RED}$FAILED_LINKS${NC}"
	echo "Success rate: $((TOTAL_LINKS > 0 ? (PASSED_LINKS * 100) / TOTAL_LINKS : 0))%"
	echo ""

	if [[ $FAILED_LINKS -gt 0 ]]; then
		echo "Failed links:"
		echo "============="
		cat "$FAILED_LINKS_FILE"
		echo ""
	fi

	# Copy results for CI systems
	if [[ -n "${GITHUB_WORKSPACE:-}" ]]; then
		cp "$RESULTS_FILE" "$GITHUB_WORKSPACE/link_test_results.txt" 2>/dev/null || true
		cp "$FAILED_LINKS_FILE" "$GITHUB_WORKSPACE/failed_links.txt" 2>/dev/null || true
	fi
}

main() {
	if [[ ! -f "$CATALOG_FILE" ]]; then
		log_error "Catalog file '$CATALOG_FILE' not found!"
		exit 1
	fi

	log_info "Starting link validation for $CATALOG_FILE"

	extract_urls "$CATALOG_FILE"
	test_all_urls
	print_summary

	# Report results but always exit successfully (informational only)
	if [[ $FAILED_LINKS -gt 0 ]]; then
		log_warning "Link validation completed: $FAILED_LINKS broken links found (informational only)"
	else
		log_success "All links are valid!"
	fi

	# Always exit with success for informational reporting
	exit 0
}

# Show usage if help requested
if [[ "${1:-}" == "-h" ]] || [[ "${1:-}" == "--help" ]]; then
	echo "Usage: $0 [CATALOG_FILE]"
	echo ""
	echo "Test all URLs in CATALOG.md for validity and anchor existence."
	echo ""
	echo "Arguments:"
	echo "  CATALOG_FILE    Path to catalog markdown file (default: CATALOG.md)"
	echo ""
	echo "Environment variables:"
	echo "  GITHUB_WORKSPACE               If set, results will be copied there for CI"
	echo "  LINK_TEST_MAX_RETRIES         Maximum retry attempts per URL (default: 3)"
	echo "  LINK_TEST_INITIAL_BACKOFF     Initial retry delay in seconds (default: 1)"
	echo "  LINK_TEST_REQUEST_DELAY       Delay between different URLs in seconds (default: 0.2)"
	echo "  LINK_TEST_CONNECT_TIMEOUT     Connection timeout in seconds (default: 15)"
	echo "  LINK_TEST_MAX_TIME            Maximum time per request in seconds (default: 30)"
	echo "  LINK_TEST_VERBOSE             Show debug output during URL extraction (default: false)"
	echo ""
	exit 0
fi

main "$@"
