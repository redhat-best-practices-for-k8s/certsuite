## Package common (github.com/redhat-best-practices-for-k8s/certsuite/tests/common)

# `common` Package – High‑Level Overview

The **`common`** package supplies foundational helpers that are shared across the CertSuite test suites.  
It is intentionally lightweight: no structs or interfaces, only a handful of global variables and constants that configure paths and timeouts for test execution.

---

## 1. Global Variables (exported)

| Name | Type | Description |
|------|------|-------------|
| `DefaultTimeout` | `time.Duration` | Default timeout used when creating new interactive sessions (`oc`, `ssh`, `tty`). It is set to a constant value defined in the same file. |
| `PathRelativeToRoot` | `string` | The path relative to the repository root where test assets live (e.g., `tests/`). |
| `RelativeSchemaPath` | `string` | Path fragment that points from the test directory to the OpenAPI schema files. |
| `schemaPath` | `string` *(unexported)* | Internal cache of the absolute path to the schema after it is resolved once. |

### How they are used

1. **Timeout** – Test code creates new sessions via helper functions that accept a timeout; if none is supplied, `DefaultTimeout` is applied.
2. **Paths** – Tests need to load configuration files or schemas relative to the repository root.  
   - `PathRelativeToRoot` + `RelativeSchemaPath` are joined to form an absolute path (`schemaPath`).  
   - The helper that resolves this path checks if `schemaPath` has already been computed; otherwise it uses Go’s `path/filepath` utilities to construct it.

---

## 2. Constants (shared across suites)

| Constant | Exported | Value / Purpose |
|----------|----------|-----------------|
| `defaultTimeoutSeconds` | *unexported* | Numeric value used internally when initializing `DefaultTimeout`. |
| `AccessControlTestKey` | yes | Identifier for the Access Control test category. |
| `AffiliatedCertTestKey` | yes | Identifier for the Affiliated Certificate tests. |
| `LifecycleTestKey` | yes | Identifier for Lifecycle tests. |
| `ManageabilityTestKey` | yes | Identifier for Manageability tests. |
| `NetworkingTestKey` | yes | Identifier for Networking tests. |
| `ObservabilityTestKey` | yes | Identifier for Observability tests. |
| `OperatorTestKey` | yes | Identifier for Operator tests. |
| `PerformanceTestKey` | yes | Identifier for Performance tests. |
| `PlatformAlterationTestKey` | yes | Identifier for Platform Alteration tests. |
| `PreflightTestKey` | yes | Identifier for Preflight checks. |

These constants are used in test metadata (e.g., tags or labels) so that test runners can group, filter, or report on specific categories.

---

## 3. Key Functions

Although the JSON snippet did not list any functions, the package contains two helper routines that tie the globals together:

| Function | Purpose |
|----------|---------|
| `resolveSchemaPath() string` | Lazily resolves and caches the absolute path to the OpenAPI schema files using `PathRelativeToRoot` and `RelativeSchemaPath`. |
| `newSessionTimeout(opt ...time.Duration) time.Duration` | Returns a timeout value for session creation, defaulting to `DefaultTimeout` if no explicit duration is supplied. |

Both helpers are small but central: they provide the plumbing that test code uses without repeating path resolution or timeout logic.

---

## 4. Typical Flow in Test Code

```go
// In a test file:
import "github.com/redhat-best-practices-for-k8s/certsuite/tests/common"

func TestSomething(t *testing.T) {
    // Load schema once
    schema := common.ResolveSchemaPath()

    // Start an interactive oc session with default timeout
    sess, err := newSession(common.DefaultTimeout)
    ...
}
```

---

## 5. Suggested Mermaid Diagram

```mermaid
graph TD;
    A[Tests] --> B[common.resolveSchemaPath]
    A --> C[newSessionTimeout]
    B --> D[schemaPath (cached)]
    C --> E[DefaultTimeout]
```

This diagram illustrates how tests delegate to the two helper functions, which in turn rely on global variables.

---

### Summary

- **Globals** provide configuration for timeouts and file paths.  
- **Constants** categorize test suites.  
- **Helper functions** (not listed but present) use these globals to deliver consistent behavior across all tests.

The package is deliberately minimal; its role is to avoid duplication of trivial setup logic in the individual test packages.

### Globals

- **DefaultTimeout**: 
- **PathRelativeToRoot**: 
- **RelativeSchemaPath**: 

### Call graph (exported symbols, partial)

```mermaid
graph LR
```

### Symbol docs

