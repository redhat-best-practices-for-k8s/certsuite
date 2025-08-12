evaluateAPICompliance`

| Item | Detail |
|------|--------|
| **Package** | `observability` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/observability`) |
| **Visibility** | unexported (private to the package) |
| **Signature** | ```go
func evaluateAPICompliance(
    apiVersions map[string]map[string]string,
    nextK8sVersion string,
    excludedAPIs map[string]struct{},
) []*testhelper.ReportObject
``` |

### Purpose

`evaluateAPICompliance` compares a set of workload API usage against the upcoming Kubernetes minor release (`nextK8sVersion`).  
It produces a slice of `*testhelper.ReportObject`s that describe:

1. **Deprecated APIs** – those scheduled for removal in the next version.
2. **Removed APIs** – already gone in the current or earlier releases.
3. **Missing/Optional APIs** – not yet present but expected in future releases.

These reports are later consumed by the test framework to assert compliance of workloads with the new Kubernetes API surface.

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `apiVersions` | `map[string]map[string]string` | Top‑level key: workload name.  Nested map: API kind → API version string (e.g., `"Deployment" : "apps/v1"`). |
| `nextK8sVersion` | `string` | Target Kubernetes version for evaluation, e.g. `"v1.28.0"`. |
| `excludedAPIs` | `map[string]struct{}` | Set of API strings (`"<group>/<resource>.<version>"`) that should be ignored (e.g., known false positives). |

### Return Value

A slice of pointers to `testhelper.ReportObject`, each representing a compliance issue.  
Each report contains fields:

- `Name` – short identifier for the type of issue.
- `Severity` – e.g., `"Error"` or `"Warning"`.
- `Message` – human‑readable description.
- Optional extra fields (`APIs`, `Missing`) depending on the issue.

### Key Steps & Dependencies

1. **Parse the target version**  
   ```go
   curVer, _ := semver.NewVersion(nextK8sVersion)
   nextMinor := curVer.IncMinor()
   ```
   Uses `semver` package (`NewVersion`, `IncMinor`) to increment the minor component.

2. **Iterate over workloads and their APIs**  
   For each API version string:
   - Split into group‑resource and version.
   - Skip if in `excludedAPIs`.

3. **Determine compliance status**  
   Calls helper functions (not shown) that consult Kubernetes deprecation data:
   - If the API’s version is scheduled for removal before or at `nextMinor`, it’s marked *Deprecated*.
   - If already removed, marked *Removed*.
   - Otherwise, may be added to a *Missing* list if expected in future.

4. **Build report objects**  
   Uses `testhelper.NewReportObject` and `AddField` to construct structured reports:
   ```go
   ro := testhelper.NewReportObject("Deprecated APIs", "Error")
   ro.AddField("APIs", strings.Join(deprecated, ", "))
   ```
   Reports are appended to the result slice.

5. **Return** – The accumulated slice is returned for further processing by the test harness.

### Side‑Effects

- No global state is modified.
- Only read‑only operations on inputs and local variables.
- Produces output via `fmt.Printf` statements (for debugging/logging) but does not alter program flow.

### How It Fits in the Package

The *observability* package orchestrates test suites that verify workload compatibility with future Kubernetes releases.  
`evaluateAPICompliance` is a core helper used by higher‑level tests to:

- Gather API usage from workloads (`beforeEachFn` populates `apiVersions`).
- Invoke this function with the desired target version.
- Assert that the returned reports meet expectations (no deprecated or removed APIs remain).

By isolating the compliance logic in this function, other tests can reuse it without duplicating parsing and reporting code.
