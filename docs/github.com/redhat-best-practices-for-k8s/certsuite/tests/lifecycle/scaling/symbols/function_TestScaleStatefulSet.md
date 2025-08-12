TestScaleStatefulSet`

| Aspect | Detail |
|--------|--------|
| **Package** | `scaling` (github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/scaling) |
| **Exported?** | Yes – intended for use by other test packages or the test harness. |
| **Signature** | `func TestScaleStatefulSet(set *appsv1.StatefulSet, timeout time.Duration, logger *log.Logger) bool` |

### Purpose
`TestScaleStatefulSet` validates that a Kubernetes StatefulSet can be scaled from its current replica count to a target count and back again within the supplied timeout. It is used by the CertSuite lifecycle tests to ensure scaling operations do not leave resources in an inconsistent state.

### Inputs

| Parameter | Type | Role |
|-----------|------|------|
| `set` | `*appsv1.StatefulSet` | The StatefulSet object whose replicas will be altered. Only its namespace, name, and current `Spec.Replicas` are used. |
| `timeout` | `time.Duration` | Maximum duration allowed for each scaling operation (scale‑up, scale‑down). |
| `logger` | `*log.Logger` | Used to emit debug information during the test; no state is altered outside of log output. |

### Output
Returns a single `bool`:
- `true` – all scaling steps succeeded within their timeouts.
- `false` – at least one step failed or timed out.

The function never panics; failures are reported via the logger and the boolean return value.

### Key Dependencies & Flow

```mermaid
flowchart TD
    A[TestScaleStatefulSet] --> B[GetClientsHolder]
    B --> C[AppsV1 client]
    C --> D[AppsV1.StatefulSets(namespace).UpdateScale]
    subgraph Scale‑Up
        D --> E[scaleStatefulsetHelper(target)]
        E --> F[Debug/ Error handling]
    end
    subgraph Scale‑Down
        G[Set Spec.Replicas back] --> H[AppsV1.StatefulSets(namespace).UpdateScale]
        H --> I[scaleStatefulsetHelper(original)]
        I --> J[Debug/ Error handling]
    end
```

1. **Client acquisition** – `GetClientsHolder` returns a holder that contains an AppsV1 client for the test namespace.
2. **Record original replicas** – the function stores the current replica count (`originalReplicas`) to restore later.
3. **Scale‑up**  
   * Compute target: `originalReplicas + 1`.  
   * Call `scaleStatefulsetHelper` with this target, passing the AppsV1 client, timeout, and logger.  
   * Log success or error via `Debug` / `Error`.
4. **Scale‑down** – revert to `originalReplicas` using the same helper function.
5. **Return** – the function returns `true` only if both scaling steps succeeded; otherwise `false`.

### Side Effects

* The StatefulSet’s replica count is temporarily changed during the test and then restored.
* No persistent state outside of Kubernetes resources is modified.
* Logs are emitted for debugging but do not affect program flow.

### Package Context
The `scaling` package contains lifecycle tests that exercise scaling behaviour of Deployments, DaemonSets, and StatefulSets.  
`TestScaleStatefulSet` is the counterpart to functions like `TestScaleDeployment` or `TestScaleDaemonSet`, providing a uniform interface for test harnesses to verify scale operations across resource kinds.
