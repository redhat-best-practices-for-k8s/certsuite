getOperandPodsFromTestCsvs`

### Purpose
`getOperandPodsFromTestCsvs` extracts the **operand pods** that belong to any of a set of test CSVs (ClusterServiceVersions).  
An *operand pod* is defined as a pod whose *top‑level* owner reference points to a custom resource instance that is managed by one of the supplied CSVs.  

This helper is used during operator auto‑discovery: once the list of operator pods is known, the function narrows it down to those that actually belong to operators under test.

---

### Signature
```go
func getOperandPodsFromTestCsvs(csvs []*olmv1Alpha.ClusterServiceVersion,
                                allPods []corev1.Pod) ([]*corev1.Pod, error)
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `csvs` | `[]*olmv1Alpha.ClusterServiceVersion` | List of test CSV objects. |
| `allPods` | `[]corev1.Pod` | All pods discovered in the cluster (or namespace). |

| Return | Type | Description |
|--------|------|-------------|
| `[]*corev1.Pod` | Slice of pointers to pod objects that are owned by CRs managed by one of the test CSVs. |
| `error` | Error if any problem occurs while determining ownership. |

---

### Key Steps & Dependencies

| Step | Description | Dependencies |
|------|-------------|--------------|
| **Iterate over all pods** | For each pod, determine its top‑level owner using `GetPodTopOwner`. | `GetPodTopOwner` (internal helper) |
| **Determine owning CR's CSV** | Look up the `ClusterServiceVersion` that manages the owning CR by checking the CR’s *managedBy* label. This uses the CSVs list to map `metadata.labels["operators.coreos.com/<csv-name>"]`. | CSV list, pod metadata |
| **Collect matching pods** | If the pod’s owner is managed by a test CSV, append it to the result slice. | `append` |
| **Return or error** | On any failure (e.g., missing labels), return an informative error via `fmt.Errorf`. | `Errorf`, `Join`, `Cut` |

---

### Side Effects

* No modification of input data structures – only reads from `csvs` and `allPods`.
* Uses logging helpers (`Info`) to emit diagnostic messages, but does not alter the global state.
* Returns an error if a pod’s owner cannot be resolved or lacks required labels.

---

### Package Context

The `autodiscover` package automates discovery of components in a Kubernetes cluster for certificate testing.  
Operators are identified by their CSVs; this function bridges from the generic list of pods to the operator‑specific subset needed for further analysis (e.g., TLS inspection). It is called after `getOperatorPods`, which first filters by known operator deployments.

---

### Mermaid Flow Diagram (optional)

```mermaid
flowchart TD
  A[Start] --> B{For each pod in allPods}
  B --> C[GetPodTopOwner(pod)]
  C --> D{Is owner CR?}
  D -- No --> E[Skip]
  D -- Yes --> F[Check if owner is managed by test CSVs]
  F --> G{Match?}
  G -- Yes --> H[Append pod to result]
  G -- No --> E
  B --> I[End]
```

This diagram illustrates the core decision path of `getOperandPodsFromTestCsvs`.
