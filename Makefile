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
TNF_IMAGE_NAME?=testnetworkfunction/cnf-certification-test
IMAGE_TAG?=localtest
TNF_VERSION?=0.0.1
RELEASE_VERSION?=4.12
.PHONY: all clean test
.PHONY: \
	build \
	build-cnf-tests \
	build-cnf-tests-debug \
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
GOLANGCI_VERSION=v1.55.1
LINKER_TNF_RELEASE_FLAGS=-X github.com/test-network-function/cnf-certification-test/pkg/versions.GitCommit=${GIT_COMMIT}
LINKER_TNF_RELEASE_FLAGS+= -X github.com/test-network-function/cnf-certification-test/pkg/versions.GitRelease=${GIT_RELEASE}
LINKER_TNF_RELEASE_FLAGS+= -X github.com/test-network-function/cnf-certification-test/pkg/versions.GitPreviousRelease=${GIT_PREVIOUS_RELEASE}
LINKER_TNF_RELEASE_FLAGS+= -X github.com/test-network-function/cnf-certification-test/pkg/versions.ClaimFormatVersion=${CLAIM_FORMAT_VERSION}
PARSER_RELEASE=$(shell jq .parserTag version.json)
BASH_SCRIPTS=$(shell find . -name "*.sh" -not -path "./.git/*")

all: build

# Runs the unit tests and build all binaries
build:
	make \
		build-cnf-tests \
		test

build-tnf-tool:
	go build -o tnf -v cmd/tnf/main.go

# Cleans up auto-generated and report files
clean:
	go clean && rm -f all-releases.txt cover.out claim.json cnf-certification-test/claim.json \
		cnf-certification-test/claimjson.js cnf-certification-test/cnf-certification-test.test \
		cnf-certification-test/cnf-certification-tests_junit.xml \
		cnf-certification-test/results.html jsontest-cli latest-release-tag.txt \
		release-tag.txt test-out.json tnf

# Runs configured linters
lint:
	checkmake --config=.checkmake Makefile
	golangci-lint run --timeout 10m0s
	hadolint Dockerfile
	shfmt -d *.sh script
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

coverage-qe: build-tnf-tool
	./tnf generate qe-coverage-report

# Generates the test catalog in Markdown
build-catalog-md: build-tnf-tool
	./tnf generate catalog markdown >CATALOG.md

# build the CNF test binary
build-cnf-tests: results-html
	PATH=${PATH}:${GOBIN} go build -ldflags "${LINKER_TNF_RELEASE_FLAGS}" -o ./cnf-certification-test

# build the CNF test binary for local development
dev:
	PATH=${PATH}:${GOBIN} go build -ldflags "${LINKER_TNF_RELEASE_FLAGS}" -o ./cnf-certification-test

# Builds the CNF test binary with debug flags
build-cnf-tests-debug: results-html
	PATH=${PATH}:${GOBIN} go build -gcflags "all=-N -l" -ldflags "${LINKER_TNF_RELEASE_FLAGS} -extldflags '-z relro -z now'" ./cnf-certification-test

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

OCT_IMAGE=quay.io/testnetworkfunction/oct:latest
REPO_DIR=$(shell pwd)

get-db:
	mkdir -p ${REPO_DIR}/offline-db
	docker pull ${OCT_IMAGE}
	docker run -v ${REPO_DIR}/offline-db:/tmp/dump:Z --user $(shell id -u):$(shell id -g) --env OCT_DUMP_ONLY=true ${OCT_IMAGE}
delete-db:
	rm -rf ${REPO_DIR}/offline-db

build-image-local:
	docker build --pull --no-cache \
		-t ${REGISTRY_LOCAL}/${TNF_IMAGE_NAME}:${IMAGE_TAG} \
		-t ${REGISTRY}/${TNF_IMAGE_NAME}:${IMAGE_TAG} \
		-f Dockerfile .

build-image-tnf:
	docker build --pull --no-cache \
		-t ${REGISTRY_LOCAL}/${TNF_IMAGE_NAME}:${IMAGE_TAG} \
		-t ${REGISTRY}/${TNF_IMAGE_NAME}:${IMAGE_TAG} \
		-t ${REGISTRY}/${TNF_IMAGE_NAME}:${TNF_VERSION} \
		-f Dockerfile .

results-html:
	script/get-results-html.sh ${PARSER_RELEASE}

check-results:
	./tnf check results
