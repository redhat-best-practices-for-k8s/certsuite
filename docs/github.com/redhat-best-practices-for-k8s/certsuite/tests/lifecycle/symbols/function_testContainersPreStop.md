testContainersPreStop` – Lifecycle Test Helper

| Item | Detail |
|------|--------|
| **File** | `tests/lifecycle/suite.go:239` |
| **Package** | `lifecycle` (github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle) |
| **Visibility** | Unexported – used only inside the test suite. |

## Purpose

`testContainersPreStop` is a helper that verifies the behavior of *pre‑stop* lifecycle hooks for containers in the target cluster.

During a lifecycle test, a `checksdb.Check` object holds the metadata for a specific check (e.g., “containers pre‑stop should run”). The function receives this check and the current test environment (`provider.TestEnvironment`) and:

1. Generates two container report objects – one for the *pre‑start* hook and one for the *post‑stop* hook.
2. Appends those reports to the `Check`’s `Reports` slice.
3. Marks the check as successful or failed by calling `SetResult`.

The function itself does not execute any hooks; it simply records what should be verified. The actual verification logic is performed elsewhere (e.g., in a test runner that inspects the generated reports).

## Signature

```go
func testContainersPreStop(*checksdb.Check, *provider.TestEnvironment)()
```

* **Parameters**
  - `c *checksdb.Check` – mutable reference to the check object being executed.
  - `env *provider.TestEnvironment` – read‑only environment that holds runtime information (cluster state, logger, etc.).

* **Return value** – none. The function mutates the passed check in place.

## Key Operations

| Step | Code snippet | Effect |
|------|--------------|--------|
| Log test start | `LogInfo("Running pre-stop containers check")` | Emits a debug message for traceability. |
| Create report objects | `NewContainerReportObject(env, c)` twice | Builds two `ContainerReportObject`s: one for the *pre‑start* phase and another for *post‑stop*. The helper uses `env` to access the current pod set and other context. |
| Append reports | `c.Reports = append(c.Reports, report1, report2)` | Adds the generated reports to the check’s collection. |
| Set result | `c.SetResult(nil)` or `c.SetResult(err)` | Marks the check as passed (nil error) or failed (error). The actual success/failure is determined by later test logic that inspects the reports. |

## Dependencies

- **`checksdb.Check`** – holds metadata and results for a single compliance check.
- **`provider.TestEnvironment`** – supplies runtime context (logger, cluster info, etc.).
- **`NewContainerReportObject(env, c)`** – helper that constructs a container‑specific report object; it is defined elsewhere in the same package.
- Logging helpers `LogInfo` and `LogError`.

## Side Effects

* Mutates the passed `Check` by adding reports and setting its result.  
* Emits log entries via the test environment’s logger.

No global state is modified, and no external resources (files, network) are touched. The function is deterministic given the same inputs.

## How It Fits Into the Package

The `lifecycle` package orchestrates a series of compliance checks for Kubernetes workloads. Each check follows this pattern:

1. **Preparation** – set up the environment, create required objects.
2. **Execution** – run the test logic (often via helper functions like `testContainersPreStop`).
3. **Reporting** – record results in the `Check` object.

`testContainersPreStop` is invoked by a higher‑level test runner during the “containers pre‑stop” phase. It does not perform the actual verification but prepares the necessary data structures for subsequent validation steps. This separation keeps the test logic modular and allows each helper to focus on a single responsibility.
