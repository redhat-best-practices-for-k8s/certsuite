Pod.GetTopOwner`

```go
func (p Pod) GetTopOwner() (map[string]podhelper.TopOwner, error)
```

| Item | Description |
|------|-------------|
| **Purpose** | Returns a map of *top‑level* owners for every pod represented by the receiver.  The key is the pod name; the value contains the top‑owner information (`podhelper.TopOwner`).  This helper is used by the test harness to identify which controller (Deployment, ReplicaSet, StatefulSet, etc.) ultimately manages a pod. |
| **Inputs** | None – operates on the `Pod` receiver which internally holds the list of pods fetched from the cluster. |
| **Outputs** | *Map*: `<podName> → podhelper.TopOwner`.  If any error occurs while walking the pod list or invoking the helper, an error is returned and the map may be incomplete. |
| **Key dependencies** | - `GetPodTopOwner` (local function in the same file).<br>- The `podhelper` package which defines the `TopOwner` struct.<br>- The receiver’s internal state (`p.pods`) – a slice of Kubernetes pod objects. |
| **Side effects** | None; purely read‑only. It does not modify cluster state or local data structures. |
| **Package context** | Part of the `provider` package that abstracts interactions with a Kubernetes cluster.  This method is used by higher‑level test flows to map pods back to their controllers, which is essential for many connectivity and configuration checks. |

---

### How it works (high‑level)

1. Iterate over all pods stored in the receiver (`p.pods`).  
2. For each pod, call `GetPodTopOwner(pod)` to compute its top owner.  
3. Store the result in a map keyed by the pod’s name.  
4. Return the map and any error that may have occurred during processing.

---

### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[Pod struct] --> B{for each pod}
    B --> C[GetPodTopOwner(pod)]
    C --> D[Return TopOwner]
    D --> E[Map[podName] = TopOwner]
```

This visual clarifies that `GetTopOwner` is essentially a thin wrapper around `GetPodTopOwner`, iterating over all pods and building the resulting map.
