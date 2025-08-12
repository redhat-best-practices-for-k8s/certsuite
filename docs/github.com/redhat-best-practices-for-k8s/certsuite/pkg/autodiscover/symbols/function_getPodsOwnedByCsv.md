getPodsOwnedByCsv`

| Item | Detail |
|------|--------|
| **Package** | `autodiscover` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/autodiscover`) |
| **Visibility** | Unexported – internal helper used by other autodiscovery routines. |
| **Signature** | `func getPodsOwnedByCsv(csvName, ns string, clients *clientsholder.ClientsHolder) ([]*corev1.Pod, error)` |

### Purpose
`getPodsOwnedByCsv` retrieves the operator/controller pods that are owned by a specific ClusterServiceVersion (CSV) within a given namespace.  
In an Operator‑based deployment, each CSV represents one version of an operator and owns the Pods that run its controllers.  This helper is used to gather those pods so that other discovery logic can inspect their labels, annotations, or runtime state.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `csvName` | `string` | The name of the CSV whose owned Pods are desired. |
| `ns` | `string` | Namespace where the CSV and its pods reside (usually the operator’s installation namespace). |
| `clients` | `*clientsholder.ClientsHolder` | Holds Kubernetes clients; this function uses it to access the CoreV1 API. |

### Return Values
| Value | Type | Meaning |
|-------|------|---------|
| `[]*corev1.Pod` | slice of pointers to `Pod` objects | The set of Pods that are owned by the specified CSV. |
| `error` | error | Non‑nil if any Kubernetes API call fails or the owner lookup logic encounters an issue. |

### Key Dependencies
- **Kubernetes CoreV1 Client** (`clients.CoreV1().Pods(ns).List`) – fetches all pods in the namespace.
- **Owner Reference Helper** (`GetPodTopOwner`) – determines the top‑level owner of a pod; used to filter pods whose owner matches the target CSV.
- **Logging / Error helpers** (`logrus.Errorf`, `fmt.Errorf`) – report problems.

### Algorithm (high‑level)
1. List all Pods in namespace `ns`.  
2. Iterate over each Pod:
   * Use `GetPodTopOwner` to find its highest‑level owner reference.
   * If that owner is a CSV and its name equals `csvName`, add the pod to the result slice.
3. Return the collected pods or an error if the list operation failed.

### Side Effects
- No state mutation: only reads from Kubernetes API, no writes.
- Logs errors via `logrus.Errorf` but does not terminate execution unless a critical failure occurs (e.g., failing to list pods).

### Usage Context
This function is called by higher‑level autodiscovery logic that needs to:
- Verify that operator pods are running.
- Inspect pod labels/annotations for service discovery metadata.
- Ensure the correct CSV version is controlling the expected set of pods.

By isolating the owner‑lookup logic, the package keeps its responsibilities modular: the discovery engine can rely on `getPodsOwnedByCsv` to return a deterministic list of operator pods without duplicating Kubernetes API interactions.
