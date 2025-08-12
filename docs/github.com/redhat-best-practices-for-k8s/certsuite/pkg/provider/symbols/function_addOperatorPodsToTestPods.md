addOperatorPodsToTestPods`

```go
func addOperatorPodsToTestPods(pods []*Pod, testEnv *TestEnvironment) func()
```

### Purpose
Creates a **closure** that, when invoked, appends all operator‑related pods from the current test environment into the slice of pods that will be used for connectivity testing.  
The closure allows the caller to defer the addition until the list is finalized (e.g., after other pod filters have run).

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `pods` | `[]*Pod` | Slice holding pods that are already selected for tests. The function will append operator pods **in‑place** to this slice. |
| `testEnv` | `*TestEnvironment` | Context object that holds the list of all pods discovered in the cluster (`testEnv.AllPods`) and other environment state. |

### Return Value

A zero‑argument function (`func()`).  
The returned function has no return value; it mutates the supplied `pods` slice by appending operator pods.

### Key Dependencies & Calls

| Call | Role |
|------|------|
| `searchPodInSlice(pods, pod)` | Checks if a given pod is already present in the `pods` slice to avoid duplicates. |
| `Info(...)` (twice) | Logs debug information about adding operator pods. |
| `append(pods, pod)` | Adds the pod to the test list. |

### Algorithm Overview

1. **Capture the current state** – The closure closes over the original `pods` slice and the `testEnv`.
2. **Iterate over all pods in the environment** (`testEnv.AllPods`).
3. For each pod:
   - If its name contains any of the operator‑specific identifiers (e.g., `"operator"`), proceed.
   - Use `searchPodInSlice` to ensure it isn’t already part of the test slice.
   - Log that the pod will be added and append it using Go’s built‑in `append`.
4. Return the closure, leaving actual mutation to the caller.

### Side Effects

- **Mutates** the `pods` slice passed in; operator pods are appended directly.
- Emits informational logs via `Info`, which may affect log output but not program state.

### How It Fits the Package

Within `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider`, pod selection for connectivity tests is a multi‑step process:

1. **Initial filtering** – based on labels, namespaces, etc.
2. **Operator pods inclusion** – handled by this function to ensure that critical operator workloads are always tested.
3. **Final test execution** – the accumulated `pods` slice is used to run connectivity checks.

By returning a closure rather than performing the addition immediately, the package allows callers to decide when to incorporate operator pods—typically after all other filtering steps—to avoid inadvertently excluding them from tests.
