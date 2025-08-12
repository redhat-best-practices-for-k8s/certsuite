getSummaryAllOperators`

**Location:** `pkg/provider/operators.go:301`  
**Package:** `provider`

### Purpose
`getSummaryAllOperators` aggregates a human‑readable summary of the state of every operator in an OpenShift/Kubernetes cluster.  
The function receives a slice of pointers to `Operator` structs (defined elsewhere in the package) and returns a slice of strings, each string summarizing one operator’s status.

Typical usage is when the test harness needs to output concise diagnostics for all operators – e.g., after a connectivity or health check run – so that the user can quickly see which operators are ready, pending, or failed.

### Signature
```go
func getSummaryAllOperators(operators []*Operator) []string
```

| Parameter | Type                | Description |
|-----------|---------------------|-------------|
| `operators` | `[]*Operator` | Slice of operator pointers to summarize. |

| Return value | Type        | Description |
|--------------|------------|-------------|
| `[]string`   | One string per operator, containing a formatted summary line. |

### Core Logic
1. **Iterate over all operators**  
   For each `op` in the input slice:
   - Extract `Namespace`, `Name`, and the current `Status.State`.
2. **Format status line**  
   Use `fmt.Sprintf` to build a string of the form:

   ```text
   <namespace>/<name> : <state>
   ```

   where `<state>` is one of the values from `op.Status.State` (e.g., `"Available"`, `"Progressing"`, `"Failed"`).
3. **Collect results**  
   Append each formatted string to a result slice.
4. **Return** the fully populated slice.

The function does not modify any input data and has no side effects beyond constructing strings.

### Dependencies
| Dependency | Role |
|------------|------|
| `fmt.Sprintf` | String formatting of the summary line. |
| `append` | Adds each formatted string to the result slice. |
| `strings` (from `strings.Join`) | **Not used directly** in this function; listed as a call because the package imports it for other utilities, but it is not invoked here. |

### Interaction with the rest of the package
- The returned summary slice is typically fed into higher‑level reporting functions that print or log operator statuses.
- `Operator` objects are produced by discovery code elsewhere in `provider`, such as the `getAllOperators` function (not shown).  
- This helper keeps the reporting logic isolated, enabling unit tests to verify formatting without touching operator discovery.

### Example
```go
ops := getAllOperators()          // returns []*Operator
summary := getSummaryAllOperators(ops)
for _, line := range summary {
    fmt.Println(line)             // e.g., "openshift-apiserver-operator/apiserver: Available"
}
```

---

**Key Takeaway:**  
`getSummaryAllOperators` is a pure, formatting helper that turns operator objects into concise status lines for diagnostics or logs. It depends only on standard library functions and the `Operator` struct defined elsewhere in the package.
