LoadChecks` – Certification Test Loader

`LoadChecks` is the entry point that builds the entire certification test suite for the **certification** package.  
It creates a hierarchy of checks, attaches helper functions (e.g., `beforeEachFn`, skip‑conditions) and registers each check with the underlying testing framework.

---

## Purpose
- **Build** the full set of tests that verify whether an operator is certified.
- **Return** a closure that, when invoked, performs the actual test registration.  
  The outer function (`LoadChecks`) is used by higher‑level test runners (e.g., `ginkgo` or custom orchestrators) to lazily instantiate the suite.

---

## Signature
```go
func LoadChecks() func()
```
- **Input:** None.
- **Output:** A zero‑argument function that, when called, registers all checks with the testing framework.  
  The returned closure is intended for deferred execution (e.g., `defer` or test runner setup).

---

## Key Steps & Dependencies

| Step | Description | Called Functions |
|------|-------------|------------------|
| **1** | Create a group of certification checks (`NewChecksGroup`) named `"certification"` and attach the global *before‑each* hook (`WithBeforeEachFn(beforeEachFn)`). | `NewChecksGroup`, `WithBeforeEachFn` |
| **2** | Add three top‑level checks: <br>• *Helm version check*<br>• *All operators certified*<br>• *Helm Certified* | `Add` |
| **3** | For each check:<br>- Build a `Check` object (`NewCheck`) with ID/labels from `GetTestIDAndLabels`. <br>- Attach the actual test logic via `WithCheckFn` (e.g., `testHelmVersion`).<br>- Attach skip logic using `WithSkipCheckFn` (e.g., `skipIfNoHelmChartReleasesFn`, `skipIfNoOperatorsFn`). | `NewCheck`, `GetTestIDAndLabels`, `WithCheckFn`, `WithSkipCheckFn` |
| **4** | Add a sub‑check that verifies container certification status by digest. It uses the helper `GetNoContainersUnderTestSkipFn` to skip when no containers are present. | `Add`, `NewCheck`, `GetTestIDAndLabels`, `GetNoContainersUnderTestSkipFn`, `WithCheckFn`, `WithSkipCheckFn` |

---

## Global Dependencies

| Variable | Type | Role |
|----------|------|------|
| `env` | `provider.TestEnvironment` | Supplies environment details (e.g., Helm releases, operator list) to the checks. |
| `validator` | `certdb.CertificationStatusValidator` | Validates certification status in container‑level tests. |
| `beforeEachFn` | `func()` | Executes setup logic before each check (e.g., resetting state). |
| `skipIfNoHelmChartReleasesFn`, `skipIfNoOperatorsFn` | `func() error` | Skip functions that abort a test when prerequisites are missing. |

---

## How It Fits the Package

- **Certification Tests** – The package implements operator certification logic; `LoadChecks` wires all concrete tests together.
- **Test Runner Integration** – By returning a closure, the function can be plugged into any test runner that expects a lazily executed suite (e.g., Ginkgo’s `DescribeTable`, custom orchestrators).
- **Extensibility** – Adding new checks only requires inserting an `Add` block; the rest of the orchestration remains unchanged.

---

## Suggested Mermaid Diagram

```mermaid
flowchart TD
  A[LoadChecks] --> B{Create Checks Group}
  B --> C["certification group"]
  C --> D[beforeEachFn]
  C --> E{Add Top‑Level Checks}
  E --> F["Helm Version Check"]
  E --> G["All Operators Certified"]
  E --> H["Helm Certified"]
  E --> I["Container Certification Status by Digest"]

  subgraph "Check Construction"
    F -.-> J[NewCheck]
    G -.-> K[NewCheck]
    H -.-> L[NewCheck]
    I -.-> M[NewCheck]
  end

  subgraph "Skip Logic"
    J --> N[WithSkipCheckFn(skipIfNoHelmChartReleasesFn)]
    K --> O[WithSkipCheckFn(skipIfNoOperatorsFn)]
    L --> P[WithSkipCheckFn(skipIfNoHelmChartReleasesFn)]
    M --> Q[WithSkipCheckFn(GetNoContainersUnderTestSkipFn)]
  end

  subgraph "Execution"
    J --> R[WithCheckFn(testHelmVersion)]
    K --> S[WithCheckFn(testAllOperatorCertified)]
    L --> T[WithCheckFn(testHelmCertified)]
    M --> U[WithCheckFn(testContainerCertificationStatusByDigest)]
  end
```

This diagram visualizes the flow from `LoadChecks` to individual checks, their skip conditions, and execution functions.
