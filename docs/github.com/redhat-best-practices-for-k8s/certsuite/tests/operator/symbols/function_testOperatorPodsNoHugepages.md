testOperatorPodsNoHugepages`

**Location**

`github.com/redhat-best-practices-for-k8s/certsuite/tests/operator/suite.go:396`

```go
func testOperatorPodsNoHugepages(c *checksdb.Check, env *provider.TestEnvironment) {
    // …
}
```

---

## Overview

`testOperatorPodsNoHugepages` is a **private helper** used by the operator test suite.  
Its purpose is to verify that none of the Operator pods in the target cluster are
configured with HugePages.  It reports any offending pods as failures and marks the
overall check accordingly.

The function consumes:

| Parameter | Type                        | Role |
|-----------|-----------------------------|------|
| `c`       | `*checksdb.Check`           | The test case metadata object that will be updated with results. |
| `env`     | `*provider.TestEnvironment`| Runtime context that provides access to the Kubernetes client and the cluster namespace. |

It does **not** return a value; instead it mutates `c` via its `SetResult` method.

---

## Key Steps

1. **Retrieve operator pod list**  
   The function obtains all pods belonging to the Operator (`operatorPods`) from
   `env`.  (The retrieval itself is hidden in helper code not shown here.)

2. **Iterate over each pod**  
   For every pod it:
   * Logs a brief informational message.
   * Calls `HasHugepages(pod)` – a predicate that returns true if the pod has any
     HugePages requests or limits set.

3. **Collect failures**  
   If `HasHugepages` is true, the pod is appended to a slice of failing pods.
   For each failure a `NewPodReportObject(pod)` is created and added to a
   report list (`podReports`).

4. **Finalize the check**  
   * If there are any failures, it logs an error, sets the result to
     `"FAILED"`, and attaches the detailed pod reports.
   * Otherwise it marks the result as `"PASSED"`.

---

## Dependencies

| Dependency | What it does |
|------------|--------------|
| `SplitCsv` | Parses comma‑separated strings (used when building the final report). |
| `LogInfo`  | Emits informational logs to the test output. |
| `HasHugepages` | Determines if a pod uses HugePages. |
| `LogError` | Emits error logs for failures. |
| `NewPodReportObject` | Builds a structured object summarizing a pod’s status for reporting. |
| `SetResult` | Records the final verdict (`PASSED`/`FAILED`) and any attached data on the check. |

All these are part of the same package or imported test utilities.

---

## Side‑Effects

* **Logs** – The function writes informational and error logs via the test logger.
* **State mutation** – It mutates the `c *checksdb.Check` object, setting its result
  and attaching any pod reports.
* No external resources are modified; it only reads pod information from the cluster.

---

## Place in the Package

The `operator` package contains end‑to‑end tests for Kubernetes Operators.  
`testOperatorPodsNoHugepages` is one of many *check functions* that run as part of a
suite defined by `checksdb.Check`.  The function is invoked automatically by a test harness
that iterates over all checks and supplies the current environment.

In short, it enforces a security/quality‑of‑service rule: **Operator pods should not request HugePages**.  
When this rule fails, the suite records a clear failure and provides the offending pod list for debugging.
