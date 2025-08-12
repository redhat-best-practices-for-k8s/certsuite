findStatefulSetsByLabels`

| Symbol | Value |
|--------|-------|
| **File** | `autodiscover_podset.go:118` |
| **Exported** | No (private helper) |
| **Signature** | `func findStatefulSetsByLabels(client appv1client.AppsV1Interface, labels []labelObject, namespaces []string) ([]appsv1.StatefulSet)` |

### Purpose
Collect all Kubernetes StatefulSets that match *any* of a set of label selectors across a list of namespaces.  
The function is used by the autodiscover logic when it needs to determine which workloads should be instrumented or monitored based on their labels.

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `client` | `appv1client.AppsV1Interface` | Kubernetes Apps‑V1 client used to query StatefulSets. |
| `labels` | `[]labelObject` | A slice of label selectors (`labelObject`) that describe the key/value pairs a StatefulSet must contain to be considered relevant. |
| `namespaces` | `[]string` | List of namespaces in which to search for StatefulSets. If empty, all namespaces are queried. |

### Return Value
- `[]appsv1.StatefulSet`: Slice containing every StatefulSet that matches at least one label selector from the input list.  
  The slice is ordered by discovery order (namespace iteration then list API order).

### Key Dependencies & Flow

```mermaid
flowchart TD
    A[Client] -->|StatefulSets()| B[List API]
    B --> C{Iterate Namespaces}
    C --> D[For each ns: List StatefulSets]
    D --> E{Check labels}
    E -->|Match| F[Append to result]
```

1. **Namespace loop** – For each namespace in `namespaces`, the function calls  
   `client.StatefulSets(ns).List(ctx, opts)` (implicit context is omitted for brevity) to fetch all StatefulSets.
2. **Empty list handling** – If a namespace yields zero results, a warning is logged (`Warn`) and the loop continues.
3. **Label matching** – Each returned StatefulSet is passed to `isStatefulSetsMatchingAtLeastOneLabel`, which checks whether its labels satisfy any selector in `labels`.  
   - If *no* match: skip.
   - If *match*: log debug info (`Debug`) and append the set to the result slice.
4. **Result** – After iterating all namespaces, the accumulated list is returned.

### Side Effects & Logging

- Uses the package‑level logger (e.g., `log.Debug`, `log.Info`, `log.Warn`).  
- No state mutation outside of the returned slice; the function is pure aside from logging.
- If a namespace has no StatefulSets or an error occurs, a warning message is emitted. Errors are not propagated – they are logged and ignored.

### Package Context

`findStatefulSetsByLabels` lives in `autodiscover`, which is responsible for automatically detecting workloads that should receive certificates or be otherwise managed by CertSuite.  
Other helpers (e.g., `isStatefulSetsMatchingAtLeastOneLabel`) rely on this function to filter relevant StatefulSets before further processing, such as generating service endpoints or applying policies.

---
