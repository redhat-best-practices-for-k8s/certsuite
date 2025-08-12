isDeploymentsPodsMatchingAtLeastOneLabel`

| Aspect | Detail |
|--------|--------|
| **File** | `pkg/autodiscover/autodiscover_podset.go:58` |
| **Signature** | `func (labels []labelObject, podName string, deployment *appsv1.Deployment) bool` |
| **Visibility** | unexported (internal helper) |

### Purpose
Checks whether a given pod (identified by its name) that belongs to a Kubernetes Deployment contains at least one label from a supplied list of expected labels.  
The function is used by the autodiscover logic when filtering pods for further inspection (e.g., certificate discovery or health checks). It returns `true` only if **any** label in `labels` matches a key/value pair present on the pod.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `labels` | `[]labelObject` | A slice of label objects that represent expected key/value pairs. Each `labelObject` contains two string fields: `Key` and `Value`. |
| `podName` | `string` | The name of the pod to inspect. It is used only for logging purposes. |
| `deployment` | `*appsv1.Deployment` | Pointer to the Deployment that owns the pod. The function looks at `deployment.Spec.Template.Labels` – the labels that will be applied to all pods created by this deployment. |

### Return
| Value | Meaning |
|-------|---------|
| `true` | At least one of the supplied label objects matches a key/value pair in the pod template labels. |
| `false` | No match was found. |

### Key Dependencies & Side‑Effects
* **Logging** – The function calls two package helpers, `Debug` and `Info`, to emit diagnostic information about why a pod did or did not match. These functions are part of the same `autodiscover` package and write to the standard logger; they have no side‑effects beyond logging.
* No mutation is performed on any input object; the function is read‑only.

### How it fits the package
In the autodiscover workflow, after a Deployment is identified as relevant (e.g., by its own labels or CSV metadata), the code enumerates all pods belonging to that deployment.  
`isDeploymentsPodsMatchingAtLeastOneLabel` is invoked for each pod name; if it returns `true`, the pod proceeds to the next stage of processing (certificate extraction, probe generation, etc.). If it returns `false`, the pod is skipped.

The function therefore acts as a gatekeeper that filters pods based on label matching, ensuring that only those pods that carry at least one expected label are considered for further analysis.
