GetVersionOcp`

**Signature**

```go
func GetVersionOcp() string
```

### Purpose

`GetVersionOcp` retrieves the OpenShift Container Platform (OCP) version running in the current test environment, if any.  
It is a helper used by diagnostics to adjust behavior based on the OCP release.

### How It Works

1. **Detect Test Environment** – Calls `GetTestEnvironment()` to obtain a struct describing the cluster type and configuration.  
2. **Check for OCP** – Uses `IsOCPCluster(env)` (where `env` is the result from step 1) to determine whether the current cluster is an OCP instance.
3. **Return Version** – If the environment is OCP, it returns the value of `env.Version`.  
   When the cluster is not OCP or no version information is available, the function falls back to returning an empty string.

### Inputs / Outputs

| Parameter | Type | Description |
|-----------|------|-------------|
| *none*    | –    | The function relies on global test‑environment detection; no direct arguments. |

| Return value | Type   | Meaning |
|--------------|--------|---------|
| `string`     | OCP version (e.g., `"4.12"`) or empty string if not applicable. |

### Dependencies

- **`GetTestEnvironment()`** – Provides the current cluster environment details.
- **`IsOCPCluster(env)`** – Checks whether a given environment is an OpenShift cluster.

Both functions are part of the same diagnostics package and operate on the internal `TestEnv` struct (not shown in the snippet).

### Side Effects

None. The function only reads state; it does not modify any global variables or perform external I/O.

### Role within `diagnostics`

This helper is used throughout the diagnostics module to:

- Conditional‑compile tests that are specific to OCP.
- Emit version‑specific diagnostic messages.
- Decide which feature checks are relevant for the running cluster.

Because diagnostics may run against multiple Kubernetes distributions, exposing a simple `GetVersionOcp()` keeps the rest of the codebase agnostic of how the environment is discovered.
