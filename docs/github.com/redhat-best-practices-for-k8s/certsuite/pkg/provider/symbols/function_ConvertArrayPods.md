ConvertArrayPods`

### Purpose
`ConvertArrayPods` transforms a slice of Kubernetes *Pod* objects (`[]*corev1.Pod`) into the provider’s internal representation (`[]*Pod`).  
The conversion is performed by delegating to the constructor **`NewPod`** for each element, then collecting the results in a new slice.

### Signature
```go
func ConvertArrayPods(pods []*corev1.Pod) []*Pod
```

- **Input:** `pods` – an array of pointers to `corev1.Pod`, the type used by the Kubernetes API.
- **Output:** A slice of pointers to `Pod`, the provider’s custom struct that contains only the fields needed for certificate‑suite checks.

### Dependencies & Side Effects

| Dependency | Role |
|------------|------|
| `NewPod`   | Creates a new `*Pod` instance from a single `corev1.Pod`. It is responsible for mapping all relevant fields (metadata, status, containers, etc.) into the internal struct. |
| `append`   | Standard Go slice operation used to build the result list incrementally. |

The function has **no side effects**: it neither mutates the input slice nor any global state.

### How It Fits in the Package

- **Provider Layer:** The `provider` package abstracts Kubernetes objects for the certificate‑suite logic. Converting raw API objects into internal types is a common step before running tests.
- **Pod Handling:** Other functions in this file (`NewPod`, various pod‑specific checks) expect the custom `*Pod`. `ConvertArrayPods` is the bridge that prepares data for those functions after retrieving pods via client-go or other mechanisms.

### Typical Usage Flow

```go
rawPods, _ := k8sClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
pods := provider.ConvertArrayPods(rawPods.Items)
for _, p := range pods {
    // run certificate‑suite checks on each internal Pod struct
}
```

This pattern keeps the rest of the codebase agnostic to Kubernetes API changes and centralizes pod conversion logic.

---
