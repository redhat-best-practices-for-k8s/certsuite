isStatefulSetsMatchingAtLeastOneLabel`

| Item | Details |
|------|---------|
| **Package** | `autodiscover` – part of the CertSuite autodiscovery subsystem. |
| **Signature** | `func isStatefulSetsMatchingAtLeastOneLabel(labels []labelObject, key string, ss *appsv1.StatefulSet) bool` |
| **Exported** | No (unexported helper used by other functions in this package). |

### Purpose
The function checks whether a given Kubernetes StatefulSet contains **at least one label that matches the supplied key/value pair**.  
It is invoked when autodiscovering workloads to determine if a StatefulSet should be considered “targeted” for certificate management based on user‑supplied labels.

### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `labels` | `[]labelObject` | Slice of label objects extracted from the target resource (e.g., a Pod or Deployment). Each `labelObject` holds a key/value pair. |
| `key` | `string` | The label key to look for (typically one of the labels defined in `data`, e.g. `"tnfLabelPrefix"`). |
| `ss` | `*appsv1.StatefulSet` | Pointer to the StatefulSet under inspection; used only for logging. |

### Return Value

- **bool** – `true` if any label in `labels` has a key that equals `key`; otherwise `false`.

### Core Logic

```go
for _, l := range labels {
    // If we find a matching key, log and return true.
    if l.Key == key {
        Debug("Found matching label", "label", l)
        return true
    }
}
Info("No matching label found for StatefulSet", "name", ss.Name)
return false
```

1. Iterate over each `labelObject` in the supplied slice.  
2. Compare the label’s key with the target key.  
3. On a match, emit a debug log and return `true`.  
4. If no match is found after exhausting the list, log an informational message indicating that the StatefulSet does not contain the desired label and return `false`.

### Dependencies & Side‑Effects

| Dependency | Effect |
|------------|--------|
| `Debug` / `Info` (from package logger) | Produces log entries; no other side effects. |
| `labelObject` type | Not defined in snippet but assumed to have fields `Key`, `Value`. |

The function itself has **no state‑changing side effects** aside from logging. It is purely a predicate.

### Integration in the Package

- Called by higher‑level discovery routines that evaluate whether a StatefulSet should be considered for certificate provisioning.
- Works alongside similar helpers (`isDeploymentMatchingAtLeastOneLabel`, etc.) to provide consistent label‑matching logic across workload types.
- The result influences subsequent steps such as creating or updating ServiceAccounts, Secrets, and Custom Resource Definitions.

### Example Usage

```go
// Inside a discovery loop:
labels := extractLabelsFromPod(pod)
if isStatefulSetsMatchingAtLeastOneLabel(labels, "app.kubernetes.io/instance", ss) {
    // Proceed with cert provisioning for this StatefulSet
}
```

This helper keeps the label‑matching logic isolated, making it easier to unit‑test and modify without touching higher‑level discovery code.
