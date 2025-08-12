GetTestSuites`

**Location:** `github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb/checksdb.go:207`  
**Signature:** ```go
func GetTestSuites() []string
```

### Purpose

`GetTestSuites` returns a list of all distinct *test suite* identifiers that are represented in the checks database.  
In CertSuite each check belongs to one or more test suites; this helper aggregates those names so callers can enumerate the available suites (e.g., for CLI help, filtering, or reporting).

### Inputs / Outputs

| Parameter | Type | Notes |
|-----------|------|-------|
| *none* | — | The function does not accept any arguments. |

| Return value | Type | Description |
|--------------|------|-------------|
| `[]string` | slice of strings | Each element is the name of a test suite that has at least one check associated with it. The slice may contain duplicates if multiple checks belong to the same suite; those are removed by the function. |

### Key Dependencies

* **`dbByGroup`** – an internal map (`map[string]*ChecksGroup`) that holds all loaded `ChecksGroup`s indexed by their group name. Each `ChecksGroup` contains a slice of `Check` objects, and each `Check` has a `TestSuites []string` field.
* **`StringInSlice`** – a small helper used to test whether a suite name is already in the output slice.

The function iterates over all groups in `dbByGroup`, then over every check in those groups. For each check it examines its `TestSuites` slice and appends any new suite names to the result list.

### Side Effects

* **No mutation** – The function reads from global state but never writes to it.
* **Thread‑safety** – It does not acquire `dbLock`. In the current code base this is safe because `GetTestSuites` is intended for read‑only, post‑initialisation use. If concurrent modifications are possible in the future, callers should guard the call with a lock.

### How it fits the package

The `checksdb` package manages an in‑memory database of checks (`ChecksGroup`, `Check`, etc.).  
Other parts of CertSuite need to know which test suites exist:

* The CLI can expose a “list‑suites” command.
* Filters that restrict execution to specific suites query this function first.
* Reporting modules may iterate over all suites to aggregate results.

`GetTestSuites` is the single source of truth for that information, derived directly from the current contents of `dbByGroup`. It therefore acts as a lightweight view into the underlying checks database.
