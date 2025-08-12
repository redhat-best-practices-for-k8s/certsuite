LoadChecks` – Load all manageability test checks

**Package**: `github.com/redhat-best-practices-for-k8s/certsuite/tests/manageability`

---

## Purpose
`LoadChecks` is the entry point that registers every *manageability* check in CertSuite.  
It builds a hierarchy of check groups and individual checks, wiring each check with its
execution function, skip logic, and test metadata (ID & labels).  The returned closure
is intended to be executed by the CertSuite framework during the test‑suite initialisation.

---

## Signature

```go
func LoadChecks() func()
```

* **Return value** – a zero‑argument function that, when called, will perform all
  registrations.  This pattern keeps `LoadChecks` itself free of side effects while
  still allowing the framework to defer registration until runtime.

---

## Key Steps

1. **Debug log**
   ```go
   Debug("Loading checks")
   ```
   Emits a debug‑level message indicating that check loading has started.

2. **Set up before‑each hook**
   ```go
   WithBeforeEachFn(beforeEachFn)
   ```
   Associates the global `beforeEachFn` (defined elsewhere in this package) with every
   subsequent test, ensuring common setup logic is run prior to each check.

3. **Create a root group**  
   ```go
   NewChecksGroup("manageability")
   ```
   All checks are added under the `"manageability"` top‑level group.

4. **Add individual checks** – The function repeatedly calls `NewCheck` and adds it to
   the current group via `Add`.  For each check:
   * **Metadata**: `GetTestIDAndLabels()` supplies a unique test ID and labels.
   * **Execution**: `WithCheckFn(<checkFunc>)` registers the function that performs the
     actual test logic.
   * **Skip condition**: `WithSkipCheckFn(<skipFn>)` allows conditional skipping (e.g.,
     when no containers are present).
   * Two checks are added here:
     1. **Test image tag** – verifies that container images have a proper tag format.
        Uses the helper `testContainersImageTag`.
     2. **Container port name format** – ensures ports follow the
        `<protocol>[-<suffix>]` naming convention, referencing `allowedProtocolNames`
        via `testContainerPortNameFormat`.

5. **Return closure** – The function returned by `LoadChecks` encapsulates all the above
   registration logic; calling it performs the actual setup.

---

## Dependencies

| Dependency | Role |
|------------|------|
| `Debug` | Logging helper (unknown implementation, assumed to log at debug level). |
| `WithBeforeEachFn`, `Add`, `WithCheckFn`, `WithSkipCheckFn` | Fluent API for building test suites. |
| `NewChecksGroup`, `NewCheck` | Constructors for groups and checks. |
| `GetTestIDAndLabels` | Generates unique IDs and labels for each check. |
| `testContainersImageTag`, `testContainerPortNameFormat` | The actual test functions executed by the checks. |

These helpers are part of CertSuite’s internal testing DSL; their exact signatures
are not shown here but they follow a builder pattern.

---

## Side Effects

* **Registration side effect** – Adds checks to the global test registry via `Add`.
* **Logging side effect** – Emits a debug message.
* No modification of global variables other than registering hooks.

---

## Package Context

The `manageability` package contains tests that verify best‑practice
behaviour of container images and Kubernetes manifests (e.g., image tagging,
port naming).  `LoadChecks` is called during the CertSuite initialisation phase
to make these checks discoverable by the test runner.  
It is intentionally isolated from execution logic; the actual check bodies are
implemented in separate helper functions (`testContainersImageTag`,
`testContainerPortNameFormat`, etc.).

---

## Suggested Mermaid Diagram

```mermaid
graph TD;
  A[LoadChecks] --> B{Register Debug};
  A --> C{Set beforeEachFn};
  A --> D[Create "manageability" group];
  D --> E[Add Check: Image Tag];
  D --> F[Add Check: Port Name Format];
```

This diagram illustrates the high‑level flow of check registration.
