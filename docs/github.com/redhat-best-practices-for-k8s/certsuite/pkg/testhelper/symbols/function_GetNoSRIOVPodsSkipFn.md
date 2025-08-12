GetNoSRIOVPodsSkipFn`

```go
func (*provider.TestEnvironment) GetNoSRIOVPodsSkipFn() func() (bool, string)
```

### Purpose
`GetNoSRIOVPodsSkipFn` returns a closure that determines whether a test should be **skipped** because the cluster contains pods using SR‑I/O‑V networking.  
The returned function is used by the test harness to decide if a test that expects *no* SR‑I/O‑V pods can proceed.

### Inputs
- The receiver `*provider.TestEnvironment` – holds the current test environment, including the kube client and cluster state.

### Output
A zero‑argument function returning:
1. **bool** – `true` if any SR‑I/O‑V pod exists (i.e., the test should be skipped).
2. **string** – a human‑readable message explaining why the skip happened, e.g. `"Found X pods using SRIOV"`.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `GetPodsUsingSRIOV` | Queries the cluster for pods that declare an SR‑I/O‑V device and returns their count. |
| `fmt.Sprintf` | Formats the skip message with the pod count. |
| `len` | Obtains the number of SR‑I/O‑V pods returned by `GetPodsUsingSRIOV`. |

### Side Effects
None – the function is read‑only; it merely queries cluster state and returns a closure.

### How It Fits in the Package
- The **testhelper** package supplies utilities for constructing test environments and making skip decisions.  
- Tests that must run only when SR‑I/O‑V is absent can call `GetNoSRIOVPodsSkipFn()` to get the skip logic, then use it with the harness’ skip mechanism.  

### Example Usage
```go
skip := env.GetNoSRIOVPodsSkipFn()
if skipNeeded, msg := skip(); skipNeeded {
    t.Skip(msg)
}
```

This keeps test code concise and centralises SR‑I/O‑V detection logic within `testhelper`.
