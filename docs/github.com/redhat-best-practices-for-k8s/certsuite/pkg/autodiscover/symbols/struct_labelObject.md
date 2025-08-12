labelObject` – a lightweight key/value pair used by autodiscover

| Feature | Detail |
|---------|--------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/autodiscover` |
| **Visibility** | Unexported (only usable within the package) |
| **Purpose** | Represents a single Kubernetes label in the form `<key>=<value>`.  It is the fundamental unit that drives the discovery logic: all other helpers (`CreateLabels`, `findPodsMatchingAtLeastOneLabel`, etc.) build on top of it. |

### Fields

| Field | Type   | Role |
|-------|--------|------|
| `LabelKey`   | `string` | The label key (e.g., `"app.kubernetes.io/name"`). |
| `LabelValue` | `string` | The corresponding value (e.g., `"cert-manager"`). |

### Typical lifecycle

1. **Parsing** – `CreateLabels([]string)` takes a slice of strings like `"key=value"`, uses regex to split each string into key/value, and builds a slice of `labelObject`.  
2. **Filtering** – Helper functions such as `findPodsMatchingAtLeastOneLabel` iterate over a set of `labelObject`s and query the Kubernetes API (`List`) with label selectors constructed from these objects.  
3. **Aggregation** – The results (pods, deployments, statefulsets, operators) are collected into higher‑level structures for further processing.

### Key dependencies

| Dependency | Why it matters |
|------------|----------------|
| `CreateLabels` | Generates the slice of `labelObject`s used by all discovery functions. |
| Kubernetes client APIs (`corev1client.CoreV1Interface`, `appv1client.AppsV1Interface`, etc.) | The label objects are turned into selector strings for API calls. |
| Logging helpers (`Debug`, `Info`, `Warn`, `Error`) | Functions that consume `labelObject` log diagnostic information when parsing or filtering fails. |

### Side‑effects

- **No state mutation** – `labelObject` itself is immutable; it only holds data.
- **Logging** – The discovery functions that use this type emit logs on parse errors, missing labels, or API call failures.

### How it fits the package

```
autodiscover
├─ labelObject            ← raw key/value pair
├─ CreateLabels()         ← parses strings into []labelObject
├─ findPodsMatchingAtLeastOneLabel()
├─ findDeploymentsByLabels()
├─ findStatefulSetsByLabels()
└─ findOperatorsByLabels()
```

All discovery helpers share the same abstraction: a slice of `labelObject`s that describes what to look for in the cluster.  This centralization keeps the code DRY and makes it straightforward to add new resource types later—just write a finder that accepts `[]labelObject`.

---

**Mermaid suggestion (optional)**

```mermaid
flowchart LR
    strings[Input: []string] --> parse[CreateLabels]
    parse --> labelObjs[labelObject array]
    subgraph discovery[Discovery functions]
        podFinder(findPodsMatchingAtLeastOneLabel)
        depFinder(findDeploymentsByLabels)
        stsFinder(findStatefulSetsByLabels)
        opFinder(findOperatorsByLabels)
    end
    labelObjs --> podFinder
    labelObjs --> depFinder
    labelObjs --> stsFinder
    labelObjs --> opFinder
```

This diagram illustrates how a list of string labels is turned into `labelObject`s and then fed to each discovery routine.
