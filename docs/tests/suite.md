# Package suite

**Path**: `tests`

## Table of Contents

- [Overview](#overview)

## Overview

The `suite` package provides a lightweight framework for executing certificate‑related test cases against Kubernetes clusters. It is intended to be used by integration tests in the CertSuite project and offers convenient orchestration of individual test functions, result aggregation, and basic reporting.

### Key Features

- Centralized execution engine that runs a collection of test functions with setup/teardown hooks
- Automatic discovery and registration of test cases based on naming conventions or explicit registration
- Structured result collection including pass/fail status, error messages, and optional metrics

### Design Notes

- Test execution is driven by the Kubernetes client configuration available in the environment; if no kubeconfig is found tests are skipped
- The framework deliberately keeps side‑effects minimal – it performs only the actions required for each test case and cleans up immediately afterward
- Users should register tests via the exported `Register` function or by embedding them in a type that implements the `TestCase` interface to ensure deterministic ordering
