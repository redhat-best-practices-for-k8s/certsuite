GetNoCPUPinningPodsSkipFn`

**Purpose**

`GetNoCPUPinningPodsSkipFn` builds a *skip function* that can be passed to
test‑framework helpers (e.g., `RunTestWithRetry`) to **exclude** test runs
that involve CPU‑pinning pods.  
The returned closure returns

```go
func() (bool, string)
```

- The first value (`bool`) indicates whether the current test should be skipped.
- The second value (`string`) is an explanatory message.

When called, the skip function will:

1. Retrieve the list of CPU‑pinning pods that are using DPDK from the supplied
   `*provider.TestEnvironment` via `GetCPUPinningPodsWithDpdk`.
2. If **any** such pods exist, it signals a *skip* with the message
   `"No CPU pinning pods skip: <count> CPU‑pinning pod(s) found"`.
3. Otherwise it returns `(false, "")`, meaning the test should proceed.

This helper is used in scenarios where a test’s logic requires the absence of
CPU‑pinning pods (e.g., when testing features that cannot coexist with DPDK
resource allocation).  

**Signature**

```go
func GetNoCPUPinningPodsSkipFn(env *provider.TestEnvironment) func() (bool, string)
```

- **Parameter**
  - `env`: the test environment context that holds cluster information and
    helper methods.
- **Return value**
  - A closure of type `func() (bool, string)` that performs the skip check at
    runtime.

**Dependencies**

| Dependency | Role |
|------------|------|
| `len` | Counts elements in a slice. |
| `GetCPUPinningPodsWithDpdk(env)` | Returns a slice of pod names that are CPU‑pinning with DPDK enabled. |

No global variables or other side effects are used.

**How it fits the package**

`testhelper` provides utilities for orchestrating and asserting test conditions
in CertSuite.  
Functions like `GetNoCPUPinningPodsSkipFn` allow tests to declaratively skip
execution when environmental preconditions are not met, keeping test logic
cleaner and more maintainable.

**Example usage**

```go
// In a test file
skipFn := testhelper.GetNoCPUPinningPodsSkipFn(env)
err := framework.RunTestWithRetry(t, "MyFeature", skipFn, func() error {
    // Test body that assumes no CPU‑pinning pods are present.
})
```

If the cluster contains any CPU‑pinning DPDK pods, `RunTestWithRetry` will
skip the test and log the provided message.
