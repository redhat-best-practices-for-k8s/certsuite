testIsRedHatRelease`

| Item | Detail |
|------|--------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/tests/platform` |
| **File / Line** | `suite.go:406` |
| **Exported?** | No – it is a helper used only within the test suite. |

### Purpose
`testIsRedHatRelease` runs the *“is Red‑Hat release”* check against a container image that has been built from the current test environment’s base image.  
It:

1. Instantiates a `BaseImageTester` for the target image.
2. Creates a `checksdb.Check` context with all necessary clients.
3. Executes the concrete test function `TestContainerIsRedHatRelease`.
4. Records the result in the check report.

### Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `check` | `*checksdb.Check` | The check object that holds the report and metadata for this run. |
| `env`   | `*provider.TestEnvironment` | Environment data (e.g., image name, registry credentials) used to locate the container image. |

### Return Value
None – the function mutates the supplied `check.Report` in‑place.

### Key Dependencies & Flow

```mermaid
flowchart TD
  A[testIsRedHatRelease] --> B[LogInfo("Starting test")]
  B --> C[NewBaseImageTester(env.Image)]
  C --> D[GetClientsHolder()]
  D --> E[NewContext(check, clients)]
  E --> F[TestContainerIsRedHatRelease(ctx, tester)]
  F --> G{error?}
  G -- yes --> H[LogError("...")]
  G -- no --> I[Append report]
```

1. **Logging** – `LogInfo` marks the start and end of the test.
2. **Tester creation** – `NewBaseImageTester` constructs a tester object that knows how to interact with the image defined in `env`.
3. **Client acquisition** – `GetClientsHolder` pulls all Kubernetes/OpenShift clientsets required by the check.
4. **Context assembly** – `NewContext` packages the check, clients, and logger into a reusable context for test functions.
5. **Test execution** – `TestContainerIsRedHatRelease` performs the actual inspection (e.g., checking `/etc/redhat-release`).  
   It returns an error if the image is not a Red‑Hat release or if any operation fails.
6. **Result handling** – On success, two `NewContainerReportObject`s are appended to the report with the check’s ID and result; on failure, errors are logged but the function continues.

### Side Effects
* Mutates `check.Report` by appending new objects.
* Emits log messages via the package‑wide logger (`LogInfo`, `LogError`).
* No global state is altered beyond what is stored in `env`.

### Integration with the Package
The `platform` test suite contains a set of helper functions that drive individual checks.  
`testIsRedHatRelease` is one such helper, wired to the *“is Red‑Hat release”* check ID (`CheckIDContainerIsRedHatRelease`).  
It is invoked by the high‑level `RunAllChecks()` routine (or similar orchestrator) when that specific check is enabled.

The function follows the same pattern as other helpers in `suite.go`, ensuring consistent report structure and error handling across all platform checks.
