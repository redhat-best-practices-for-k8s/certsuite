TestEnvironment.IsPreflightInsecureAllowed`

```go
func (te *TestEnvironment) IsPreflightInsecureAllowed() bool
```

### Purpose

`IsPreflightInsecureAllowed` reports whether the test harness is permitted to skip or relax security checks performed during the OpenShift/Red‑Hat‑Enterprise Kubernetes *pre‑flight* validation stage.  
The function does **not** perform any checks itself; it simply reads a configuration value that was loaded into the `TestEnvironment` instance when the provider was initialized.

### Inputs

| Parameter | Type | Description |
|-----------|------|-------------|
| `te`      | `*TestEnvironment` | Receiver containing the test‑environment state, including all parsed flags and environment variables. |

> No other arguments are required; the method relies entirely on fields of `TestEnvironment`.

### Outputs

| Return value | Type   | Description |
|--------------|--------|-------------|
| `bool` | Indicates whether insecure pre‑flight mode is enabled (`true`) or not (`false`). |

### Key Dependencies & Side‑Effects

- **Dependency on configuration** – The decision is derived from a field inside the `TestEnvironment`. That field is set during provider initialization (see `provider.go:200` where `loaded` is assigned).  
  It typically reflects an environment variable such as `CERTSUITE_ALLOW_INSECURE_PREFLIGHT`, or a CLI flag passed to the test binary.
- **No state mutation** – The method reads only; it does not modify any fields of `TestEnvironment` nor trigger external side effects.  

### How It Fits Into the Package

The *provider* package orchestrates the execution of certsuite tests against an OpenShift cluster.  
During the boot‑strap phase, a series of **pre‑flight** checks validate that the target cluster meets basic security and configuration requirements (e.g., network policy enforcement, proper RBAC).  

When running in environments where those strict checks cannot be satisfied (for instance, in CI pipelines or on custom clusters), users can opt‑in to an “insecure” mode.  
`IsPreflightInsecureAllowed` is consulted by the pre‑flight logic to decide whether to skip certain hard failures.

Typical usage pattern:

```go
if env.IsPreflightInsecureAllowed() {
    // Skip or downgrade some checks
}
```

Thus, this helper centralises the policy decision and keeps the rest of the codebase free from scattered `if` statements checking environment variables directly.
