checkValidOperatorInstallation`

```go
func checkValidOperatorInstallation(installPath string) (bool, []string, error)
```

**Purpose**

`checkValidOperatorInstallation` verifies that an Operator bundle has been installed correctly in the cluster under test.  
It examines all Cluster‑Service‑Version (CSV) resources produced by the bundle and checks:

* whether each CSV belongs to a supported operator type (single‑namespaced, multi‑namespaced or cluster‑wide);
* that the expected CSV(s) are present;
* that no pods belong to operators that were not part of the current bundle.

If any of these conditions fail the function reports the problem and returns `false`.  
On success it returns `true` with an empty error list.

---

### Parameters

| Name | Type   | Description |
|------|--------|-------------|
| `installPath` | `string` | Filesystem path to the Operator bundle that was just installed.  The function will look for CSVs under this directory (typically `<bundle>/manifests/`). |

---

### Return values

| Index | Type      | Meaning |
|-------|-----------|---------|
| 0     | `bool`    | `true` if all checks passed; otherwise `false`. |
| 1     | `[]string` | Slice of error messages collected during the validation.  Empty when the return value is `true`. |
| 2     | `error`   | A high‑level error that aborts the check (e.g., failure to read CSV files).  If this is non‑nil, the boolean result should be ignored. |

---

### Key dependencies

| Dependency | Role in function |
|------------|------------------|
| `getCsvsBy(path string) ([]string, error)` | Lists all CSV YAML files under `installPath`. |
| `Split(s string, sep rune) []string` (standard library) | Splits a comma‑separated list of expected CSV names. |
| `checkIfCsvUnderTest(csv string, expected []string) bool` | Checks if the current CSV is one of those that should be present for this bundle. |
| `isSingleNamespacedOperator(csv string) bool` | Detects whether the CSV represents a single‑namespaced operator. |
| `isMultiNamespacedOperator(csv string) bool` | Detects whether the CSV represents a multi‑namespaced operator. |
| `isCsvInNamespaceClusterWide(csv string, ns string) bool` | Verifies that a cluster‑wide CSV is indeed installed in the expected namespace (usually `"openshift-operators"`). |
| `findPodsNotBelongingToOperators() ([]string, error)` | Returns a list of pod names that belong to operators *not* part of the current bundle. |

---

### Workflow (high‑level)

1. **Discover CSVs**  
   Call `getCsvsBy` on `installPath`.  If it fails, return with the wrapped error.

2. **Determine expected CSVs**  
   Split the environment variable `expected-operator-csvs` into a slice of names (`expectedCsvs`).  
   (The env var is defined in the test suite configuration.)

3. **Validate each CSV**  
   For every discovered CSV:
   * Skip if it isn’t one of the expected CSVs.
   * Check that its operator type matches the supported patterns:
     - Single‑namespaced (`isSingleNamespacedOperator`)
     - Multi‑namespaced (`isMultiNamespacedOperator`)
     - Cluster‑wide (via `isCsvInNamespaceClusterWide`).
   * Accumulate any mismatches into an error slice.

4. **Check for stray operator pods**  
   Call `findPodsNotBelongingToOperators`.  If any pods are found, add a message to the error list.

5. **Return result**  
   If no errors were collected, return `(true, nil, nil)`.  
   Otherwise return `(false, errs, nil)` where `errs` contains all accumulated messages.

---

### Side effects & assumptions

* The function performs read‑only queries against the cluster; it does **not** modify any resources.
* It relies on environment configuration (`env`) that is populated by the test suite’s `beforeEachFn`.  
  If that global is unset, expected CSV names cannot be resolved and the function will return an error.
* The operator bundle must expose its CSVs under a known directory (typically `<bundle>/manifests/`); otherwise `getCsvsBy` will fail.

---

### Placement in the package

`checkValidOperatorInstallation` lives in **tests/operator/helper.go**.  
It is used by the test suite’s `BeforeEach` hook to assert that an Operator bundle was installed correctly before any functional tests run.  The helper encapsulates all CSV‑related validation logic, keeping the test code focused on higher‑level assertions.
