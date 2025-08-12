TestScaleHPACrd`

**Location**

`github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/scaling/crd_scaling.go:113`

---

### Purpose
`TestScaleHPACrd` is a helper used by the **Scaling** test suite to verify that a `HorizontalPodAutoscaler` (HPA) created as a Custom Resource Definition (CRD) can be scaled correctly.  
It performs the following high‑level steps:

1. **Retrieve** an HPA client for the target namespace.
2. **Read** the current replica count of the referenced deployment.
3. **Scale** the HPA to a new desired replica value.
4. **Poll** until the underlying deployment reaches the expected replica count.
5. **Return** whether the scaling succeeded within the allotted timeout.

The function is deliberately side‑effect free apart from interacting with the Kubernetes API and emitting log output; it does not modify any global state.

---

### Signature

```go
func TestScaleHPACrd(
    cr *provider.CrScale,
    hpa *scalingv1.HorizontalPodAutoscaler,
    resource schema.GroupResource,
    timeout time.Duration,
    logger *log.Logger,
) bool
```

| Parameter | Type | Meaning |
|-----------|------|---------|
| `cr` | `*provider.CrScale` | Test context containing the Kubernetes client configuration. |
| `hpa` | `*scalingv1.HorizontalPodAutoscaler` | The HPA CRD instance to test. |
| `resource` | `schema.GroupResource` | API group/resource of the workload that the HPA targets (e.g., `apps/deployment`). |
| `timeout` | `time.Duration` | Maximum duration allowed for the scaling operation to complete. |
| `logger` | `*log.Logger` | Logger used for debugging output. |

**Return value**

- `true` – Scaling succeeded within `timeout`.
- `false` – Timeout or error prevented successful scaling.

---

### Key Dependencies

| Dependency | Role |
|------------|------|
| `provider.GetClientsHolder(cr)` | Returns a client holder that contains typed clients (e.g., autoscalingv1). |
| `GetNamespace(cr)` | Provides the namespace in which the HPA and its target workload live. |
| `AutoscalingV1()` | Factory for the `autoscaling/v1` API group. |
| `HorizontalPodAutoscalers(resource, ns)` | Returns a typed client for HPAs of the given resource type. |
| `scaleHpaCRDHelper(...)` | Internal helper that performs one scaling step and logs progress. |

---

### Flow Overview (pseudo‑code)

```text
clients = GetClientsHolder(cr)
ns      = GetNamespace(cr)

// Obtain current replica count
currentReplicas = clients.AutoscalingV1().
                    HorizontalPodAutoscalers(resource, ns).
                    Get(hpa.Name).Spec.MinReplicas

// Desired replicas: increment by 1 (or any test logic)
desired := *currentReplicas + 1

startTime := now()
for now() - startTime < timeout {
    // Attempt to scale
    if err := scaleHpaCRDHelper(clients, ns, hpa.Name, desired, logger); err != nil {
        logger.Error(err)
        continue
    }

    // Verify scaling
    newReplicas := clients.AutoscalingV1().
                    HorizontalPodAutoscalers(resource, ns).
                    Get(hpa.Name).Status.CurrentReplicas

    if newReplicas == desired { return true }
    time.Sleep(pollInterval)
}
return false
```

The `scaleHpaCRDHelper` function is invoked repeatedly; it updates the HPA’s `spec.minReplicas` field and logs each attempt. The outer loop keeps polling until either the target replica count is observed or the timeout expires.

---

### Side Effects

* Makes **read** and **write** calls to the Kubernetes API (HPA `GET/PUT`, deployment status reads).
* Emits debug and error messages via the supplied logger.
* No modification of global variables or package state.

---

### Integration into the Package

`TestScaleHPACrd` lives in the `scaling` test package.  
It is called by higher‑level tests that:

1. Create an HPA CRD instance (`provider.CrScale`) for a workload.
2. Invoke this function to confirm that scaling behaves as expected.

By abstracting the scaling logic into this helper, individual tests remain concise and focused on setup/teardown, while the heavy lifting of polling and client handling is centralized here.

---
