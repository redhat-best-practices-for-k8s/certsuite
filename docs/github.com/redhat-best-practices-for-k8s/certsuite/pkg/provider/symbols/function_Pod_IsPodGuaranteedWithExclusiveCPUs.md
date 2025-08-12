Pod.IsPodGuaranteedWithExclusiveCPUs`

| Item | Details |
|------|---------|
| **Package** | `provider` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider`) |
| **Exported?** | Yes |
| **Receiver** | `pod Pod` |
| **Signature** | `func (p Pod) IsPodGuaranteedWithExclusiveCPUs() bool` |

### Purpose

Determines whether a pod is *guaranteed* and has **exclusive CPU resources** allocated.  
In OpenShift/Kubernetes this means:

1. The pod’s containers request the same amount of CPU and memory as they are limited to (no oversubscription).  
2. All requested CPUs are whole‑number units – no fractional CPUs (`0.5`, `1.25`, etc.) are used.

If both conditions hold, the pod is eligible for *exclusive* CPU usage on a node.

### Inputs & Outputs

| Parameter | Type | Notes |
|-----------|------|-------|
| `p` (receiver) | `Pod` | The pod instance to evaluate. |

**Return value**

- `bool`:  
  - `true` – the pod is guaranteed **and** requests whole‑number CPU units.  
  - `false` – otherwise.

### Dependencies

The function relies on two helper functions defined elsewhere in the same package:

| Function | Role |
|----------|------|
| `AreCPUResourcesWholeUnits(pod Pod) bool` | Checks that every container’s CPU request is a whole number. |
| `AreResourcesIdentical(pod Pod) bool` | Verifies that each container’s CPU/memory requests equal its limits (i.e., the pod is guaranteed). |

No global variables are read or written.

### How It Works

```go
func (p Pod) IsPodGuaranteedWithExclusiveCPUs() bool {
    return AreCPUResourcesWholeUnits(p) && AreResourcesIdentical(p)
}
```

1. **`AreCPUResourcesWholeUnits(p)`**  
   Iterates over the pod’s containers, examining their CPU requests. If any request contains a fractional part (e.g., `500m`, `1.5`), it returns `false`.

2. **`AreResourcesIdentical(p)`**  
   Compares each container’s CPU and memory *request* values to its *limit* values. The pod is guaranteed only if all requests equal limits.

The function short‑circuits: if the first check fails, the second isn’t invoked.

### Side Effects

None – pure evaluation.

### Placement in Package

`IsPodGuaranteedWithExclusiveCPUs` lives in `pods.go`.  
It is part of a small set of predicates used by the provider to filter pods when:

- Determining which workloads can run on nodes with specific CPU policies.  
- Running *pod‑level* conformance checks that rely on guaranteed resource guarantees (e.g., verifying exclusive CPU allocation).

Because it aggregates two fundamental pod‑resource properties, other tests and metrics in the `provider` package call this function to quickly decide if a pod meets strict scheduling constraints.

### Example Usage

```go
for _, p := range allPods {
    if p.IsPodGuaranteedWithExclusiveCPUs() {
        fmt.Println("Pod", p.Name, "has exclusive CPUs")
    }
}
```

In summary, `IsPodGuaranteedWithExclusiveCPUs` is a lightweight helper that combines two core pod‑resource predicates to identify workloads suitable for exclusive CPU allocation on a node.
