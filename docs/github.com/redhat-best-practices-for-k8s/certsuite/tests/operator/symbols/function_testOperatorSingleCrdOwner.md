testOperatorSingleCrdOwner`

**Location**

`tests/operator/suite.go:354`

**Purpose**

Verifies that a *single* Custom Resource Definition (CRD) is correctly owned by the operator under test.  
The function creates two CRD reports – one for an expected “owner” and another for an unexpected “non‑owner”.  
It then checks that only the owner CRD passes the ownership test while the non‑owner fails.

**Signature**

```go
func testOperatorSingleCrdOwner(check *checksdb.Check, env *provider.TestEnvironment)
```

| Parameter | Type                      | Description |
|-----------|--------------------------|-------------|
| `check`   | `*checksdb.Check`        | The check record that will receive the results. |
| `env`     | `*provider.TestEnvironment` | Test harness providing utilities such as logging and result handling. |

**Key Steps**

1. **Logging Setup**  
   * Logs a start message (`LogInfo`) for traceability.

2. **CRD Report Construction**  
   * Two reports are built using `NewCrdReportObject`.  
     * *Owner report* – contains fields that should satisfy the ownership check.  
     * *Non‑owner report* – contains fields that intentionally violate ownership (e.g., wrong owner reference).  

3. **Result Assignment**  
   * For each report, the function calls `SetResult` on the corresponding `check`.  
   * The owner report is expected to succeed (`Result: true`), while the non‑owner should fail (`Result: false`).  
   * Each outcome is logged via `LogDebug`.

4. **Final Logging**  
   * Summarises the test outcome with `LogInfo` or `LogError` if an unexpected condition arises.

**Dependencies**

- **Logging helpers** – `LogInfo`, `LogDebug`, `LogError`.
- **Result handling** – `SetResult` on a `*checksdb.Check`.
- **CRD report construction** – `NewCrdReportObject`.
- **Utility functions** – standard Go functions like `append`, `len`, and `strings.Join`.

**Side Effects**

The function mutates the passed‑in `check` object by setting its result for each CRD report.  
No global state is modified; it only interacts with the provided `env` logger.

**How It Fits in the Package**

Within the `operator` test suite, this helper implements a unit test for ownership logic of CRDs.  
It is invoked by higher‑level test runners that iterate over all checks defined for the operator component.  

```mermaid
flowchart TD
    A[testOperatorSingleCrdOwner] --> B[Create owner report]
    A --> C[Create non‑owner report]
    B --> D[SetResult (expected true)]
    C --> E[SetResult (expected false)]
    D & E --> F[Log outcome]
```

*Note*: The function relies on the surrounding test environment (`env`) for logging but otherwise operates independently.
