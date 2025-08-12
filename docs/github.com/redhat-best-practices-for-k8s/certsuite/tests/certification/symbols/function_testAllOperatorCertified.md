testAllOperatorCertified`

| Item | Details |
|------|---------|
| **Package** | `certification` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/certification`) |
| **File** | `suite.go`, line 123 |
| **Visibility** | Unexported (used only inside the test suite) |
| **Signature** | ```go
func (*checksdb.Check, *provider.TestEnvironment, certdb.CertificationStatusValidator)()
```
The function is a closure returned by `testAllOperatorCertified`. It receives a pointer to a check definition (`*checksdb.Check`), the test environment (`*provider.TestEnvironment`) and a certification status validator (`certdb.CertificationStatusValidator`). The return type is `func()`, so the caller simply invokes it when the test framework schedules this particular check.

---

## Purpose

`testAllOperatorCertified` verifies that **every operator** installed on a cluster meets the Red‑Hat certification criteria.  
It does this by:

1. Determining whether the current cluster is an OpenShift (OCP) cluster.
2. Splitting the list of certified operators into two groups:
   * Operators that are already certified (`CertifiedOperator`).
   * Operators that are not yet certified (`Online` – they exist but lack certification data).
3. For each group, creating a structured report object that contains the operator names and the result status.
4. Finally setting the overall test result (pass/fail) based on whether any uncertified operators were found.

The function is intended to be executed as part of the certification test suite; it logs progress and errors for visibility in CI pipelines.

---

## Inputs

| Parameter | Type | Meaning |
|-----------|------|---------|
| `*checksdb.Check` | check definition (unused directly, but required by the closure signature) | Holds metadata about the current test. |
| `*provider.TestEnvironment` | environment context | Provides information such as whether the cluster is OpenShift (`IsOCPCluster`). |
| `certdb.CertificationStatusValidator` | validator interface | Supplies the logic for determining if an operator is certified. |

---

## Key Operations & Dependencies

1. **Cluster Type Check**  
   ```go
   if !env.IsOCPCluster() { ... }
   ```
   Uses `IsOCPCluster` from the test environment to gate certification checks only on OpenShift clusters.

2. **Operator List Splitting**  
   ```go
   certOps, onlineOps := SplitN(operators, count)
   ```
   `SplitN` partitions operators into certified and non‑certified lists based on a counter.

3. **Logging**  
   * `LogInfo`: General informational messages (e.g., number of operators).
   * `LogError`: Errors encountered while processing operators.

4. **Certification Check**  
   ```go
   if validator.IsOperatorCertified(op) { ... }
   ```
   Calls the validator to decide certification status for each operator.

5. **Report Construction**  
   * `NewOperatorReportObject` creates a structured report entry.
   * `AddField` attaches fields such as `"operator"` and `"status"`.
   * `SetResult` marks the overall test result (`Pass`/`Fail`) based on findings.

---

## Output / Side Effects

* **Logs** – Informational and error messages are emitted to the test runner’s console.
* **Test Result** – The closure calls `check.SetResult(...)`, which records pass/fail status in the framework.
* **Report Objects** – Two separate report objects (`certOps` and `onlineOps`) are created and populated with operator names and their certification status. These are typically collected by the test harness for reporting.

---

## How It Fits the Package

The `certification` package orchestrates end‑to‑end tests that validate whether a cluster meets Red‑Hat best‑practice guidelines.  
- **Other test functions** (e.g., `testOperatorInstalled`, `testHelmChartReleases`) are similarly structured as closures returning `func()`.  
- `testAllOperatorCertified` is the final step in the certification pipeline, ensuring that *every* operator on an OpenShift cluster has been evaluated and reported.  
- It relies on global test variables (`env`, `validator`) defined elsewhere in `suite.go`.

---

## Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[Start] --> B{Is OCP Cluster?}
    B -- No --> C[Log & Skip]
    B -- Yes --> D[Retrieve Operator List]
    D --> E[Split into Certified / Online]
    E --> F[For each Certified op]
    F --> G[Add to Cert Report]
    E --> H[For each Online op]
    H --> I[Add to Online Report]
    G & I --> J[Set Test Result (Pass/Fail)]
    J --> K[End]
```

This diagram visualizes the decision flow and report generation performed by `testAllOperatorCertified`.
