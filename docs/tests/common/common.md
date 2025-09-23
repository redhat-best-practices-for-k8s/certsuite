# Package common

**Path**: `tests/common`

## Table of Contents

- [Overview](#overview)

## Overview

The `common` package supplies shared configuration and identifiers for the CertSuite test suites, such as test keys, default timeouts, and path utilities used across multiple test packages.

### Key Features

- Defines a set of exported string constants that act as keys to group and identify individual test categories (e.g., AccessControlTestKey, PerformanceTestKey).
- Provides globally accessible variables like `DefaultTimeout` for session creation and `PathRelativeToRoot`/`RelativeSchemaPath` for locating resources relative to the repository root.
- Includes internal helpers (e.g., a private default timeout constant) that support consistent behavior across test packages without exposing implementation details.

### Design Notes

- Constants are exported so that all test suites can reference the same identifiers, ensuring consistency in test grouping and reporting.
- The package intentionally exposes only read‑only globals; mutable state is avoided to keep tests deterministic.
- Best practice: import this package for any shared configuration needed by a test suite, but avoid adding test logic here to maintain separation of concerns.
