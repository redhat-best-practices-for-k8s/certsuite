FindPodsByLabels`

**Package:** `autodiscover`  
**File:** `autodiscover_pods.go:44`  
**Signature**

```go
func FindPodsByLabels(
    client corev1client.CoreV1Interface,
    labels []labelObject,
    namespaces []string,
) []corev1.Pod
```

---

### Purpose

`FindPodsByLabels` retrieves all Kubernetes Pods that match **at least one** of a set of label selectors across one or more namespaces.  
It is used by the autodiscover logic to locate Pods that belong to particular operators (e.g., Istio, cert-manager) or have specific probe‑helper labels.

---

### Parameters

| Name        | Type                     | Description |
|-------------|--------------------------|-------------|
| `client`    | `corev1client.CoreV1Interface` | Kubernetes client used to list Pods. |
| `labels`    | `[]labelObject`          | Slice of label selectors (`labelObject`) that the function will try to match. Each selector is a key/value pair. |
| `namespaces`| `[]string`               | List of namespace names in which to search for matching Pods. If empty, all namespaces are searched. |

> **Note**: The type `labelObject` is defined elsewhere in the package as a simple struct with `Key` and `Value` fields.

---

### Returns

- `[]corev1.Pod`: A slice containing every Pod that satisfies at least one of the supplied label selectors across any of the specified namespaces.  
  Pods are appended in the order they are discovered; duplicates are **not** removed.

---

### Key Steps & Dependencies

| Step | Action | Called Function / Variable |
|------|--------|----------------------------|
| 1 | Determine which namespaces to query | `len(namespaces)` – if zero, use all namespaces (`client.Pods("").List`) |
| 2 | For each namespace, list Pods | `client.Pods(namespace).List(...)` |
| 3 | Filter the listed Pods against the label selectors | Calls helper `findPodsMatchingAtLeastOneLabel` (not shown) which iterates over pods and returns those that contain any of the supplied labels. |
| 4 | Accumulate matching Pods | Uses Go's built‑in `append` to build a result slice. |
| 5 | Log debug information | Calls `Debug` (a logging helper). |
| 6 | Handle errors | If listing fails, logs with `Error`; otherwise continues. |

The function relies on the global `data` map (from `autodiscover.go`) only indirectly: the label selectors passed in are typically derived from that map elsewhere in the package.

---

### Side‑Effects & Observations

* **No mutation** of input slices or client state – purely read‑only.
* **Logging**: The function emits debug and error logs. These calls rely on a global logger (`Debug`, `Error`) defined in the same package.
* **Performance**: For many namespaces and large cluster sizes, this may perform a large number of API calls; however, it is intentionally straightforward to keep discovery logic simple.

---

### How It Fits the Package

`FindPodsByLabels` is a core utility for the autodiscover subsystem:

1. **Operator Discovery** – Other helpers (e.g., `findIstioPods`, `findCertManagerPods`) build label selectors and invoke this function to locate operator Pods.
2. **Probe‑Helper Detection** – The probe helper pods are identified by specific labels; this function aggregates them for health checks.
3. **Extensibility** – By exposing a generic “label‑based Pod lookup”, new discovery modules can reuse it without duplicating Kubernetes client logic.

In short, `FindPodsByLabels` abstracts the plumbing of querying the API server and filtering Pods by label, allowing higher‑level autodiscover code to focus on *what* labels matter rather than *how* to fetch the matching resources.
