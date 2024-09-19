# There are four main groups of operations provided by this Makefile: build,
# clean, run and tasks.
#
# Build operations will create artefacts from code. This includes things such
# as binaries, mock files, or catalogs of CNF tests.
#
# Clean operations remove the results of the build tasks, or other files not
# considered permanent.
#
# Run operations provide shortcuts to execute built binaries in common
# configurations or with default options. They are part convenience and part
# documentation.
#
# Tasks provide shortcuts to common operations that occur frequently during
# development. This includes running configured linters and executing unit
# tests.
GO_PACKAGES=$(shell go list ./... | grep -v vendor)

# Default values
REGISTRY_LOCAL?=localhost
REGISTRY?=quay.io
CERTSUITE_IMAGE_NAME?=redhat-best-practices-for-k8s/certsuite
CERTSUITE_IMAGE_NAME_LEGACY?=testnetworkfunction/cnf-certification-test
IMAGE_TAG?=localtest
.PHONY: all clean test build
.PHONY: \
	build-certsuite-tool \
	build-certsuite-tool-debug \
	coverage-html \
	generate \
	install-moq \
	lint \
	update-rhcos-versions \
	vet

# Gets default value of $GOBIN if not explicitly set
GO_PATH=$(shell go env GOPATH)
ifeq (,$(shell go env GOBIN))
	GOBIN=${GO_PATH}/bin
else
	GOBIN=$(shell go env GOBIN)
endif

COMMON_GO_ARGS=-race
GIT_COMMIT=$(shell script/create-version-files.sh)
GIT_RELEASE=$(shell script/get-git-release.sh)
GIT_PREVIOUS_RELEASE=$(shell script/get-git-previous-release.sh)
CLAIM_FORMAT_VERSION=$(shell script/get-claim-version.sh)
GOLANGCI_VERSION=v1.60.3
LINKER_CERTSUITE_RELEASE_FLAGS=-X github.com/redhat-best-practices-for-k8s/certsuite/pkg/versions.GitCommit=${GIT_COMMIT}
LINKER_CERTSUITE_RELEASE_FLAGS+= -X github.com/redhat-best-practices-for-k8s/certsuite/pkg/versions.GitRelease=${GIT_RELEASE}
LINKER_CERTSUITE_RELEASE_FLAGS+= -X github.com/redhat-best-practices-for-k8s/certsuite/pkg/versions.GitPreviousRelease=${GIT_PREVIOUS_RELEASE}
LINKER_CERTSUITE_RELEASE_FLAGS+= -X github.com/redhat-best-practices-for-k8s/certsuite/pkg/versions.ClaimFormatVersion=${CLAIM_FORMAT_VERSION}
BASH_SCRIPTS=$(shell find . -name "*.sh" -not -path "./.git/*")
PARSER_RELEASE=$(shell jq -r .parserTag version.json)
RESULTS_HTML_URL=https://raw.githubusercontent.com/redhat-best-practices-for-k8s/parser/${PARSER_RELEASE}/html/results.html

all: build

build: build-certsuite-tool

build-certsuite-tool: results-html
	PATH="${PATH}:${GOBIN}" go build -ldflags "${LINKER_CERTSUITE_RELEASE_FLAGS}" -o certsuite -v cmd/certsuite/main.go
	git restore internal/results/html/results.html

build-darwin-arm64: results-html
	PATH="${PATH}:${GOBIN}" GOOS=darwin GOARCH=arm64 go build -ldflags "${LINKER_CERTSUITE_RELEASE_FLAGS}" -o certsuite -v cmd/certsuite/main.go
	git restore internal/results/html/results.html

# Cleans up auto-generated and report files
clean:
	go clean && rm -f all-releases.txt cover.out claim.json cnf-certification-test/claim.json \
		cnf-certification-test/claimjson.js cnf-certification-test/cnf-certification-tests_junit.xml \
		cnf-certification-test/results.html jsontest-cli latest-release-tag.txt \
		release-tag.txt test-out.json certsuite

# Runs configured linters
lint:
	checkmake --config=.checkmake Makefile
	golangci-lint run --timeout 10m0s
	hadolint Dockerfile
	shfmt -d script
	typos
	markdownlint '**/*.md'
	yamllint --no-warnings .
	shellcheck --format=gcc ${BASH_SCRIPTS}

# Builds and runs unit tests
test: coverage-qe
	./script/create-missing-test-files.sh
	go build ${COMMON_GO_ARGS} ./...
	UNIT_TEST=true go test -coverprofile=cover.out.tmp ./...

coverage-html: test
	cat cover.out.tmp | grep -v _moq.go >cover.out
	go tool cover -html cover.out

coverage-qe: build-certsuite-tool
	./certsuite generate qe-coverage-report

# Generates the test catalog in Markdown
build-catalog-md: build-certsuite-tool
	./certsuite generate catalog markdown >CATALOG.md

# Builds the Certsuite binary with debug flags
build-certsuite-tool-debug: results-html
	PATH="${PATH}:${GOBIN}" go build -gcflags "all=-N -l" -ldflags "${LINKER_CERTSUITE_RELEASE_FLAGS} -extldflags '-z relro -z now'" -o certsuite -v cmd/certsuite/main.go
	git restore internal/results/html/results.html

install-mac-brew-tools:
	brew install \
		checkmake \
		golangci-lint \
		hadolint \
		shfmt

# Installs linters
install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_VERSION}

install-shfmt:
	go install mvdan.cc/sh/v3/cmd/shfmt@latest

vet:
	go vet ${GO_PACKAGES}

install-moq:
	go install github.com/matryer/moq@latest

generate: install-moq
	find . | grep _moq.go | xargs rm
	go generate ./...

update-rhcos-versions:
	./script/rhcos_versions.sh

OCT_IMAGE=quay.io/redhat-best-practices-for-k8s/oct:latest
REPO_DIR=$(shell pwd)

get-db:
	mkdir -p ${REPO_DIR}/offline-db
	docker pull ${OCT_IMAGE}
	docker run -v ${REPO_DIR}/offline-db:/tmp/dump:Z --user $(shell id -u):$(shell id -g) --env OCT_DUMP_ONLY=true ${OCT_IMAGE}
delete-db:
	rm -rf ${REPO_DIR}/offline-db

# Runs against whatever architecture the host is
build-image-local:
	docker build --pull --no-cache \
		-t ${REGISTRY_LOCAL}/${CERTSUITE_IMAGE_NAME}:${IMAGE_TAG} \
		-t ${REGISTRY}/${CERTSUITE_IMAGE_NAME}:${IMAGE_TAG} \
		-t ${REGISTRY_LOCAL}/${CERTSUITE_IMAGE_NAME_LEGACY}:${IMAGE_TAG} \
		-t ${REGISTRY}/${CERTSUITE_IMAGE_NAME_LEGACY}:${IMAGE_TAG} \
		-f Dockerfile .

results-html:
	curl -s -O --output-dir internal/results/html ${RESULTS_HTML_URL}

check-results:
	./certsuite check results
