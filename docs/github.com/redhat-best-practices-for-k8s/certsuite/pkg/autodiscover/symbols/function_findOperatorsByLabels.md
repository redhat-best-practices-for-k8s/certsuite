findOperatorsByLabels`

```go
func findOperatorsByLabels(
    ops v1alpha1.OperatorsV1alpha1Interface,
    labelObjs []labelObject,
    namespaces []configuration.Namespace,
) []*olmv1Alpha.ClusterServiceVersion
```

### Purpose
`findOperatorsByLabels` is a helper that retrieves all **Cluster Service Versions (CSVs)** from an operator catalog that match at least one of the supplied labels and exist in any of the given Kubernetes namespaces.

The function is used by the auto‑discovery logic to determine which operators are present on a cluster so that their CSVs can be inspected for certificate or TLS configuration.

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `ops` | `v1alpha1.OperatorsV1alpha1Interface` | Client interface for the Operator Lifecycle Manager (OLM) API. It is used to list CSVs in a namespace. |
| `labelObjs` | `[]labelObject` | A slice of label descriptors that specify which labels to look for on CSV objects. Each `labelObject` contains a key/value pair and an optional regular‑expression flag. |
| `namespaces` | `[]configuration.Namespace` | The namespaces to search in. Only CSVs whose `metadata.namespace` matches one of these are considered. |

### Return Value

- A slice of pointers to `olmv1Alpha.ClusterServiceVersion`.  
  Each element represents a CSV that matched at least one label and was found in one of the supplied namespaces.

If no CSVs match, an empty slice is returned (not `nil`).

### Key Dependencies & Flow

| Step | Called Function | What It Does |
|------|-----------------|--------------|
| 1. `len(labelObjs)` | builtin | Determines if any labels were provided. |
| 2. `findOperatorsMatchingAtLeastOneLabel(ops, labelObjs, namespaces)` | local helper | Performs the actual OLM API calls: for each namespace it lists CSVs and keeps those whose annotations or labels match at least one supplied label (via regex or exact match). |
| 3. Logging (`Debug`, `Info`) | from `logrus` | Emits debug info about how many operators were found. |
| 4. Error handling | builtin `Error` | If listing fails, logs the error and continues with other namespaces. |

### Side Effects

- **Logging**: The function writes to the shared logger (via `Debug`, `Info`, `Error`).  
- **No state mutation**: It does not modify any global variables or the passed‑in data structures; it only reads from them.

### How It Fits Into the Package

`findOperatorsByLabels` is a low‑level routine used by higher‑level discovery logic (`DiscoverOperatorCSV`, etc.). The package first builds a list of labels that indicate which operators are relevant for certificate validation (e.g., Istio, OpenShift Service Mesh). It then calls this function to pull the corresponding CSVs from the cluster. Those CSV objects are later examined for `spec.install.spec.clusterServiceVersionOverrides` and other fields that influence how certificates should be validated.

In short: **findOperatorsByLabels → returns matching CSVs → used by discovery logic to decide what operators need certificate checks.**
