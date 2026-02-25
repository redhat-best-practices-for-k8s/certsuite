# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## Repository Overview

This is the Red Hat Best Practices Test Suite for Kubernetes (certsuite) - a comprehensive certification test suite for verifying CNF/Cloud Native Functions deployment best practices on OpenShift/Kubernetes clusters. The test suite validates workloads against Red Hat's best practices across multiple categories including networking, security, lifecycle, operators, and more.

## Key Commands

### Build
```bash
make build              # Build the certsuite binary
make build-darwin-arm64 # Build for macOS ARM64
```

### Testing
```bash
make test               # Run unit tests with coverage
make coverage-html      # Generate and view HTML coverage report
```

### Linting
```bash
make lint               # Run all linters (golangci-lint, hadolint, shfmt, markdownlint, yamllint, shellcheck, typos)
make markdownlint       # Run markdown linter only
make yamllint           # Run YAML linter only
```

### Code Generation
```bash
make generate           # Run go generate for code generation
```

### Catalog and Documentation
```bash
make build-catalog-md   # Generate test catalog in Markdown (CATALOG.md)
make coverage-qe        # Generate QE coverage report
```

### Container Images
```bash
make build-image-local  # Build local container image
make get-db             # Download offline certification database
make delete-db          # Remove offline database
```

### Running Tests
```bash
./certsuite run --config-file <path> --output-dir <path>
./certsuite check results              # Validate test results
./certsuite claim show failures        # Show failed test cases
./certsuite claim compare <file1> <file2>  # Compare two claim files
```

## Code Architecture

### Command Structure (cmd/certsuite/)
The CLI is built using Cobra and organized into subcommands:
- `run`: Execute test suites against a cluster
- `claim`: Manage and analyze claim files (show, compare)
- `generate`: Generate catalogs, configs, feedback reports
- `check`: Validate results and image certification status
- `info`: Display information about test cases
- `version`: Show version information
- `upload`: Upload results to external systems

### Core Packages (pkg/)

**autodiscover**: Automatically discovers pods, operators, CRDs, and other Kubernetes resources in target namespaces. The `DiscoveredTestData` structure contains all discovered resources that tests validate against:
- Pods (target, probe, operand)
- Operators (CSVs, install plans, catalog sources)
- Network resources (policies, attachments, SR-IOV)
- Storage (PVs, PVCs)
- RBAC (roles, bindings)
- Custom resources and CRDs

**configuration**: Test configuration parsing and validation. Reads `tnf_config.yaml` (or custom config file) to determine:
- Target namespaces to test
- Pod label selectors
- Operator label selectors
- Network attachment definitions to check
- Test-specific parameters

**provider**: Test execution providers and resource access. The `TestEnvironment` interface provides access to discovered resources for test implementations.

**checksdb**: Test case database and results tracking. Each test is registered as a `Check` with metadata, skip conditions, and execution functions.

**compatibility**: Version compatibility checks between cluster components and expected versions.

**collector**: Result collection and submission to external data collectors.

### Internal Packages (internal/)

**clientsholder**: Kubernetes client management and caching. Maintains singleton instances of various k8s clients (core, apps, networking, etc.).

**log**: Logging infrastructure used throughout the codebase.

**cli**: CLI framework and utilities for command-line interactions.

**results**: Results processing and HTML generation for test reports.

### Test Organization (tests/)

Test suites are organized by category:
- `accesscontrol`: Security context, namespace, and privilege tests
- `networking`: Network policies, ICMP, services, and connectivity tests
- `platform`: OS validation, sysctls, node taints, boot parameters
- `lifecycle`: Pod recreation, scaling, owner references
- `observability`: Logging, monitoring, pod disruption budgets
- `operator`: Operator lifecycle, installation, best practices
- `certification`: Container, operator, and helm chart certification checks
- `performance`: Performance-related validations
- `manageability`: Management and configuration tests
- `preflight`: Red Hat preflight certification integration

Each test suite contains:
- Individual test implementations (e.g., `tests/networking/icmp/icmp.go`)
- Test utilities and helpers (e.g., `tests/networking/netcommons/netcommons.go`)
- Ginkgo test suite setup (`suite_test.go`)
- A `suite.go` file that registers checks using the checksdb API

### Testing Framework

Tests use the Ginkgo/Gomega BDD framework. Test execution follows this pattern:

1. **Autodiscovery Phase**: The `autodiscover` package scans the cluster and builds `DiscoveredTestData`
2. **Check Registration**: Each test suite's `LoadChecks()` function registers test cases with checksdb
3. **Check Execution**: The test runner iterates through registered checks, evaluating skip conditions and executing check functions
4. **Results Collection**: Test results are collected into a "claim file" (JSON format) containing pass/fail/skip status, logs, and configuration snapshots

### Configuration File

Tests require a configuration file (default: `config/certsuite_config.yml`) specifying:
```yaml
targetNameSpaces:
  - name: tnf
podsUnderTestLabels:
  - "redhat-best-practices-for-k8s.com/generic: target"
operatorsUnderTestLabels:
  - "redhat-best-practices-for-k8s.com/operator:target"
targetCrdFilters:
  - nameSuffix: "group1.test.com"
    scalable: false
```

## Development Guidelines

### Go Version
This repository uses Go 1.26.0. Ensure your local environment matches this version.

### Testing and Mocks
The codebase uses native Go structs for mocking interfaces in tests. Mock implementations are hand-written and located alongside the interfaces they mock (e.g., `internal/clientsholder/command_mock.go`). This approach avoids external dependencies and makes the code easier to understand and maintain.

### Linting
All code must pass the configured linters before submission. Use `make lint` to run all linters. The project uses:
- `golangci-lint` (Go code quality)
- `hadolint` (Dockerfile linting)
- `shfmt` (Shell script formatting)
- `markdownlint` (Markdown formatting)
- `yamllint` (YAML validation)
- `shellcheck` (Shell script analysis)
- `typos` (Spelling checker)

### Code Organization
- **cmd/**: Main applications and CLI tools
- **pkg/**: Public/exported packages that can be imported by other projects
- **internal/**: Private packages not meant for external use
- **tests/**: Test suite implementations organized by category
- **script/**: Build and automation scripts

### Version Management
Version information is managed via scripts:
- `script/create-version-files.sh`: Creates version metadata
- `script/get-git-release.sh`: Gets current Git release tag
- `version.json`: Contains version information for dependencies

### Results and Claims
Test execution produces a "claim file" (JSON format) containing:
- Test results (pass/fail/skip)
- Configuration snapshots
- Resource inventories
- Logs and failure reasons

The claim format is defined in a separate package (`certsuite-claim`) and shared across tools.

## Common Workflows

### Running a Single Test Suite
```bash
./certsuite run --config-file=tnf_config.yaml --label-filter=networking
```

### Comparing Test Results
```bash
./certsuite claim compare claim1.json claim2.json
```

### Building and Testing Locally
```bash
make build
make test
make lint
```

### Checking Certification Status
```bash
./certsuite check image-cert-status --image <image-reference>
```
