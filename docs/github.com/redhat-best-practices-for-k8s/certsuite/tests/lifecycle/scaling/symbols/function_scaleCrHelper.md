scaleCrHelper` – Internal CRD‑Scaling Helper

| Item | Description |
|------|-------------|
| **File** | `tests/lifecycle/scaling/crd_scaling.go:79` |
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/scaling` |
| **Exported?** | No – used only inside the test suite |

## Purpose

`scaleCrHelper` implements a single‑step scaling operation for a Custom Resource (CR).  
It is invoked by higher‑level tests that want to change the replica count of an
arbitrary CR, wait for the scale to take effect, and report whether the operation
succeeded.

The function:

1. Reads the current `scale` subresource of the CR.
2. Updates its `Spec.Replicas` field to a new target value.
3. Persists the change via the Kubernetes API (`Update`).
4. Waits for the scale subresource to reach the desired state (`WaitForScalingToComplete`).

It returns `true` when scaling completes successfully, otherwise `false`.

## Signature

```go
func scaleCrHelper(
    scalesGetter scale.ScalesGetter,
    gvr schema.GroupResource,
    crScale *provider.CrScale,
    target int32,
    shouldLog bool,
    timeout time.Duration,
    logger *log.Logger) (bool)
```

| Parameter | Type | Role |
|-----------|------|------|
| `scalesGetter` | `scale.ScalesGetter` | Client capable of reading/updating the scale subresource. |
| `gvr` | `schema.GroupResource` | Group/Resource tuple identifying the CR type. |
| `crScale` | `*provider.CrScale` | Holds namespace and name of the target CR. |
| `target` | `int32` | Desired replica count. |
| `shouldLog` | `bool` | Whether to emit debug logs. |
| `timeout` | `time.Duration` | Max wait time for scaling to finish. |
| `logger` | `*log.Logger` | Logger used for optional debug output. |

## Key Dependencies

- **Kubernetes Scale API** – via `scale.ScalesGetter`, `Get`, `Update`.
- **Retry logic** – `RetryOnConflict` handles concurrent update conflicts.
- **Waiting helper** – `WaitForScalingToComplete` polls until the scale subresource reflects the target count.
- **Logging utilities** – `Debug`, `Error`, and `TODO` are used to emit diagnostic messages.

## Flow (simplified)

```mermaid
flowchart TD
  A[Get current Scale] --> B{Conflict?}
  B -- yes --> C[RetryOnConflict]
  C --> A
  B -- no --> D[Set Spec.Replicas = target]
  D --> E[Update Scale via client]
  E --> F[WaitForScalingToComplete(timeout)]
  F --> G{Success?}
  G -- yes --> H[Return true]
  G -- no --> I[Log error, Return false]
```

1. **Read current scale**  
   `scalesGetter.Scales(crScale.Namespace).Get(gvr, crScale.Name)`.

2. **Retry on conflict** – if an update fails due to a conflict, the operation is retried until it succeeds or another error occurs.

3. **Update replica count** – modify `scale.Spec.Replicas` and send the change with `scalesGetter.Scales(...).Update(...)`.

4. **Wait for completion** – call `wait.WaitForScalingToComplete(gvr, crScale.Namespace, crScale.Name, target, timeout)` to block until the actual replicas match the desired count.

5. **Return status** – `true` on success, `false` otherwise (after logging errors).

## Side Effects

- The function mutates the Kubernetes cluster by writing a new scale value for the specified CR.
- It may log messages if `shouldLog` is true.
- No other global state is altered.

## How it fits the package

The `scaling` test package validates that various resources correctly expose and honour scaling via the Kubernetes Scale API.  
`scaleCrHelper` is a low‑level utility used by higher‑level tests (e.g., `TestCRDScaling`) to perform an actual scale operation on a CR, encapsulating the common pattern of read‑modify‑update‑wait.

By keeping this logic in a single helper:

- Tests remain concise and focused on assertions.
- Error handling and retry logic are centralized.
- Future changes to scaling semantics (e.g., new fields) can be made here without touching each test.
