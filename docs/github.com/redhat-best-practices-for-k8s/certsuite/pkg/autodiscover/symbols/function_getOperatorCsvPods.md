getOperatorCsvPods`

**Purpose**  
Collects the Pods that are managed by each Operator CSV (ClusterServiceVersion) in a given cluster and returns them as a map keyed by the CSV’s namespaced name.

**Signature**

```go
func getOperatorCsvPods(csvs []*olmv1Alpha.ClusterServiceVersion) (
    map[types.NamespacedName][]*corev1.Pod, error)
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `csvs` | `[]*olmv1Alpha.ClusterServiceVersion` | Slice of pointers to CSV objects that were discovered in the cluster. Each CSV represents an installed Operator. |

| Return | Type | Description |
|--------|------|-------------|
| `map[types.NamespacedName][]*corev1.Pod` | Map from a CSV’s namespaced name (`metadata.Namespace/metadata.Name`) to the list of Pods that belong to that CSV’s installation namespace and are owned by it. |
| `error` | If any step fails (client creation, pod listing, etc.) an error is returned. |

**Key dependencies**

1. **`GetClientsHolder()`** – Retrieves a `clients.Clients` holder that contains Kubernetes client interfaces needed to query Pods.
2. **`getPodsOwnedByCsv(csv)`** – Helper that lists all Pods in the CSV’s namespace and filters those owned by the CSV (via owner references).  
   *It uses the core client from the clients holder.*  
3. **Standard library helpers** – `strings.TrimSpace`, `fmt.Errorf`.

**Implementation outline**

```go
func getOperatorCsvPods(csvs []*olmv1Alpha.ClusterServiceVersion) (
    map[types.NamespacedName][]*corev1.Pod, error) {

    // 1. Prepare result map.
    csvToPods := make(map[types.NamespacedName][]*corev1.Pod)

    // 2. Get shared clients (k8s clientset).
    c, err := GetClientsHolder()
    if err != nil {
        return nil, fmt.Errorf("failed to get clients: %w", err)
    }

    // 3. Iterate over each CSV.
    for _, csv := range csvs {
        // Ensure namespace is trimmed.
        ns := strings.TrimSpace(csv.Namespace)

        // List pods owned by this CSV in its namespace.
        pods, err := getPodsOwnedByCsv(c, csv)
        if err != nil {
            return nil, fmt.Errorf("failed to list pods for CSV %s/%s: %w",
                ns, csv.Name, err)
        }

        // Key is the CSV's namespaced name.
        key := types.NamespacedName{Namespace: ns, Name: csv.Name}
        csvToPods[key] = pods
    }

    return csvToPods, nil
}
```

**Side‑effects**

- No mutation of global state.  
- The function only performs read operations against the Kubernetes API server.

**Package context**

In `autodiscover`, operators are discovered by inspecting CSV objects. Knowing which Pods belong to an operator is essential for later steps:

* **Certificate discovery** – Only Pods that run the operator’s controllers may need certificates injected or validated.
* **Health checks / metrics** – The presence of operator Pods informs readiness probes and monitoring.

`getOperatorCsvPods` therefore bridges the high‑level Operator catalog (`ClusterServiceVersion`) with concrete runtime resources (Pods), enabling the rest of the autodiscover package to act on per‑operator data.
