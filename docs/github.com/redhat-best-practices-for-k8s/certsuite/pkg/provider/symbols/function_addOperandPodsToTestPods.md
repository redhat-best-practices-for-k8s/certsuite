addOperandPodsToTestPods`

| Item | Detail |
|------|--------|
| **Package** | `provider` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider`) |
| **Signature** | `func addOperandPodsToTestPods(pods []*Pod, env *TestEnvironment)` |
| **Visibility** | unexported (internal helper) |

### Purpose
During a test run the framework gathers all pods that need to be checked for connectivity.  
`addOperandPodsToTestPods` expands this list by adding any *operator*‑managed Pods that belong to the same namespace as the supplied `env`.  

In practice this means:
1. For each operator pod in the environment, locate its corresponding entry in the global test‑pod slice (`pods`).  
2. If a match is found, append the operator’s full `Pod` object (with all containers and annotations) to the slice that will be used for subsequent tests.

This allows operator workloads (e.g., `operator-lifecycle-manager`, custom operators) to participate in connectivity checks without requiring explicit user configuration.

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `pods` | `[]*Pod` | Slice of already‑collected test pods. The function appends matched operator pods to this slice. |
| `env` | `*TestEnvironment` | Context holding the current test environment, including its namespace and list of operator Pods (`OperatorPods`). |

### Return value
None – the function mutates the supplied slice in place.

### Key Operations & Dependencies

1. **Searching**  
   *Calls:* `searchPodInSlice(pods, op)`  
   Searches for an existing pod with the same name and namespace as the operator pod `op`. The helper returns a pointer to the found pod or `nil`.

2. **Logging**  
   Two calls to `Info` log whether an operator pod was found or not:
   - If found: logs “Operator pod already in test list”.
   - If not found: logs “Adding operator pod”.

3. **Appending**  
   *Calls:* builtin `append(pods, op)` – adds the operator pod to the slice.

4. **No external globals used** – the function operates purely on its arguments and local variables.

### Interaction with Other Parts of the Package

| Related Piece | Relationship |
|---------------|--------------|
| `TestEnvironment.OperatorPods` | Source list of operator pods that may need to be tested. |
| `Pod` struct (in `pods.go`) | The objects appended; contain full container information for later connectivity tests. |
| Logging utilities (`Info`, etc.) | Provide visibility into the selection process during test execution. |

### Example Flow

```go
// Existing list of pods that will be tested
testPods := collectAllPods()

// Environment containing operator pods
env := &TestEnvironment{
    Namespace: "openshift-operators",
    OperatorPods: []*Pod{op1, op2},
}

// Add any operator pods to the test set
addOperandPodsToTestPods(testPods, env)

// testPods now contains original pods + op1/op2 if not already present
```

### Summary

`addOperandPodsToTestPods` is a lightweight helper that ensures all operator‑managed Pods in the current test environment are included in the connectivity test set. It performs a simple membership check and appends missing pods, logging its actions for debugging purposes. The function is purely functional on its inputs and has no side effects beyond mutating the provided slice.
