deployDaemonSet`

```go
func deployDaemonSet(name string) error
```

Deploys a DaemonSet identified by **`name`** into the cluster that is under test.

### Purpose

The function creates or updates a DaemonSet so that it can be used for
subsequent connectivity / policy tests.  
It is invoked by the test harness during the *pre‑test* phase to ensure
the required sidecar/agent is running on all target nodes.

### Inputs & Output

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `name`    | string | Name of the DaemonSet definition that should be applied. The value comes from a test parameter map (`GetTestParameters()`). |

| Return | Type  | Meaning |
|--------|-------|---------|
| `error`| error | Non‑nil if any step in creating or waiting for the DaemonSet fails. |

### Key Steps & Dependencies

1. **Client Setup**  
   * `SetDaemonSetClient()` – registers a client that will be used to
     create the DaemonSet and later check its status.

2. **Retrieve Test Parameters**  
   The function pulls several values from `GetTestParameters()`:
   - `Namespace` (where the DaemonSet lives)
   - `Image`, `Port`, `Args`, `Env` – runtime configuration for the pod.
   These parameters are defined in the test YAML or overridden by
   environment variables.

3. **Construct DaemonSet Object**  
   A new `appsv1.DaemonSet` is built via a call to Go’s `make`
   (the code actually uses an inline struct literal; the JSON shows only
   `make`).  
   The spec includes:
   - `Selector`, `Template`, container definition (`Image`, `Port`,
     `Args`, `Env`)
   - Node selector/affinity from test parameters.

4. **Create or Update**  
   * `CreateDaemonSet()` – attempts to create the DaemonSet in the
     target namespace. If it already exists, the call will error; this
     is handled by the caller (not shown here).

5. **Verify Readiness**  
   * `IsDaemonSetReady()` – quick check whether all desired pods are
     running.
   * If not ready, `WaitDaemonsetReady()` blocks until the DaemonSet
     reaches the desired state or times out.

6. **Error Handling**  
   Errors from any of the above calls are wrapped with a contextual
   message via `fmt.Errorf`.

### Side‑Effects

- **Cluster State:** A new DaemonSet is created (or an existing one
  updated) in the specified namespace.
- **Global Variables Modified:**
  - The client registered by `SetDaemonSetClient` becomes available to
    other test functions that need to inspect the DaemonSet status.
- **Logging / Metrics:** None directly; errors are returned for the
  caller to log.

### Package Context

Within the `provider` package, `deployDaemonSet` is part of the *setup*
phase that prepares the testing environment.  
Other functions in this file (e.g., `GetTestParameters`, `WaitDaemonsetReady`)
share the same client‑holder infrastructure (`GetClientsHolder`).  
The DaemonSet created here may host sidecar containers such as
Istio proxies or custom network plugins, which later tests will query
to verify connectivity, policy enforcement, or resource usage.

### Mermaid Diagram (Suggested)

```mermaid
flowchart TD
    A[deployDaemonSet(name)] --> B{Get Test Params}
    B --> C[Namespace]
    B --> D[Image, Port, Args, Env]
    B --> E[Affinity / NodeSelector]
    C & D & E --> F[Construct DaemonSet]
    F --> G[CreateDaemonSet()]
    G --> H{Success?}
    H -->|No| I[Return error]
    H -->|Yes| J[IsDaemonSetReady()]
    J --> K{Ready?}
    K -->|Yes| L[Done]
    K -->|No| M[WaitDaemonsetReady()]
    M --> N[Timeout or success]
```

This diagram illustrates the high‑level flow of creating and verifying
the DaemonSet within the test harness.
