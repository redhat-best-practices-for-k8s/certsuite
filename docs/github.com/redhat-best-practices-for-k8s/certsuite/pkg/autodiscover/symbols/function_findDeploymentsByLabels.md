findDeploymentsByLabels`

| Attribute | Value |
|-----------|-------|
| **File** | `autodiscover_podset.go` (line 70) |
| **Exported** | No – internal helper used by the autodiscovery logic |
| **Signature** | `func(findDeploymentsByLabels(appv1client.AppsV1Interface, labels []labelObject, namespaces []string) ([]appsv1.Deployment))` |

### Purpose
`findDeploymentsByLabels` scans a set of Kubernetes namespaces for Deployments whose pods match *at least one* label from the supplied list.  
It is used by the autodiscover package to locate operator‑managed resources (e.g., Istio, CNF operators) that expose probe helper pods via specific labels.

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `client` | `appv1client.AppsV1Interface` | A client for the Apps v1 API. Used to list Deployments (`Deployments(namespace).List`) in each namespace. |
| `labels` | `[]labelObject` | Slice of label descriptors (key/value pairs) that represent probe helper pod labels. The function checks whether a Deployment’s pods contain *any* of these labels. |
| `namespaces` | `[]string` | Namespaces to search. Each namespace is queried independently. |

### Return Value

- `[]appsv1.Deployment` – A slice containing all Deployments that satisfy the label match criterion across the supplied namespaces.

If an error occurs while listing a namespace, it is logged (via `log.Error`) and the function continues with the next namespace; the offending Deployment is simply omitted from the result set.  
The function never propagates errors to its caller.

### Key Dependencies

| Dependency | Usage |
|------------|-------|
| `client.Deployments(ns).List(...)` | Retrieves all Deployments in a namespace. |
| `isDeploymentsPodsMatchingAtLeastOneLabel(deployment, labels)` | Determines if any pod within the Deployment carries one of the specified labels. |
| Logging helpers (`Warn`, `Info`, `Debug`) | Emit diagnostic information at various stages (empty result sets, errors, etc.). |

### Control Flow Overview

```mermaid
flowchart TD
    A[Start] --> B{for each namespace}
    B --> C[List Deployments]
    C --> D{error?}
    D -- yes --> E[log Error; continue next ns]
    D -- no --> F{no deployments?}
    F -- yes --> G[Warn “No deployments in ns”]
    F -- no --> H[for each deployment]
    H --> I{isMatch(deployment)}
    I -- true --> J[append to result]
    I -- false --> K[next deployment]
    K --> H
    G --> B
    J --> H
    B --> Z[Return result slice]
```

### Side‑Effects & Logging

* **Logging** – The function writes warnings when a namespace yields no Deployments or when the overall result set is empty. Errors during API calls are logged as errors.
* **No mutation of inputs** – The supplied `labels` and `namespaces` slices are read‑only; the returned slice contains copies of Deployment objects from the client.

### Integration into the Package

The autodiscover package orchestrates detection of various operator deployments (e.g., Istio, CNF).  
`findDeploymentsByLabels` is a building block for those routines:

1. **Probe helper discovery** – Operators expose probe pods with well‑known labels (`probeHelperPodsLabelName/Value`).  
2. The higher‑level functions call `findDeploymentsByLabels` with the appropriate label set and namespaces, then inspect the returned Deployments to extract the operator’s CSV or other metadata.

Because it abstracts the common pattern of “list all deployments in a namespace, filter by pod labels,” this helper keeps the package DRY and isolates Kubernetes API interaction from higher‑level logic.
