testPodsOwnerReference`

| Aspect | Detail |
|--------|--------|
| **Package** | `lifecycle` (github.com/redhat‑best-practices-for-k8s/certsuite/tests/lifecycle) |
| **Visibility** | Unexported – used only within the test suite. |
| **Signature** | `func(test *checksdb.Check, env *provider.TestEnvironment)` |

### Purpose

`testPodsOwnerReference` validates that every pod created by a test case has the correct Kubernetes OwnerReference set. In CertSuite, tests may create various resources (e.g., Deployments, StatefulSets). The pods spawned from those controllers should point back to their owning controller so that garbage‑collection and lifecycle events behave correctly.

The function is invoked as part of `RunTest` for a particular test case. It records the result in the test’s `Check` object and logs any failures.

### Inputs

| Parameter | Type | Description |
|-----------|------|-------------|
| `test` | `*checksdb.Check` | Holds metadata about the current test, including its name, description, and a slice to store results. |
| `env`  | `*provider.TestEnvironment` | Provides context for interacting with Kubernetes: clientsets, logger, and helper utilities. |

### Key Steps

1. **Logging**  
   The function logs an informational message indicating that it is beginning the owner‑reference check.

2. **OwnerReference Construction**  
   It calls `NewOwnerReference`, passing the test’s name to build a reference that should be attached to each pod created by this test.

3. **Run Test Logic (`RunTest`)**  
   The core logic is delegated to `RunTest`. This helper iterates over all pods discovered in the environment and compares their OwnerReferences against the expected one. For each pod:
   - If the reference matches, a successful report object is created via `NewPodReportObject`.
   - If it does not match, an error report object is generated.
   The reports are appended to the test’s results slice.

4. **Result Handling**  
   After iteration, `SetResult` marks the overall outcome (success or failure) in the test’s status.

5. **Error Logging**  
   Any errors encountered during pod retrieval or comparison are logged using `LogError`.

### Dependencies

| Dependency | Role |
|------------|------|
| `GetLogger()` | Retrieves a logger scoped to the current test. |
| `GetResults()` | Accesses the slice of existing results for appending new reports. |
| `NewOwnerReference(name string)` | Builds an OwnerReference struct that is expected to appear on each pod. |
| `RunTest(pods []v1.Pod, env *provider.TestEnvironment, check *checksdb.Check)` | Performs the actual comparison and result recording. |
| `NewPodReportObject(...)` | Creates a structured report for each pod outcome. |

### Side Effects

- **State Mutation**: Appends new `Result` objects to `test.Results`.
- **Logging**: Emits info/error messages via the test‑specific logger.
- **No External Changes**: Does not modify Kubernetes resources; it only reads them.

### How It Fits the Package

Within the `lifecycle` test suite, several checks ensure that resource lifecycle behaviors (e.g., pod recreation, deletion) are correct. `testPodsOwnerReference` is one of these checks and is typically invoked after a controller has spawned pods. By verifying OwnerReferences, it guarantees that:

- Kubernetes garbage collection will clean up orphaned pods when the owning controller is deleted.
- Test harnesses can reliably identify which pods belong to which test case.

The function complements other lifecycle tests such as `testPodsRecreation`, `testPodSetReadyTimeout`, etc., providing a holistic validation of resource management.
