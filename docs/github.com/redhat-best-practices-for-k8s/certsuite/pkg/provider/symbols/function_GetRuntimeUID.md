GetRuntimeUID`

| Item | Detail |
|------|--------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider` |
| **Signature** | `func GetRuntimeUID(cs *corev1.ContainerStatus) string` |
| **Exported?** | Yes |

---

#### Purpose
`GetRuntimeUID` extracts the runtime‑specific identifier (UID) from a container’s status.  
In Kubernetes, each running container has a `ContainerID` field in its status of the form

```
<runtime>://<uid>
```

where `<runtime>` is the container engine (e.g., `docker`, `cri-o`) and `<uid>` is a unique string assigned by that runtime. This helper isolates the UID part, which is useful for:

* Logging / debugging – identify the exact container instance.
* Validation tests – compare expected UIDs or ensure uniqueness.
* Cross‑runtime compatibility – callers don’t need to know the prefix format.

---

#### Input
- `cs *corev1.ContainerStatus`  
  The status object of a single container (from a Pod).  
  **Assumptions**:  
  * `cs.State.Running` is non‑nil (container has started).  
  * `cs.State.Running.ContainerID` contains the full runtime ID string.

---

#### Output
- `string` – The substring after the first occurrence of `"://"` in `ContainerID`.  
  If the format does not contain `"//:"`, an empty string is returned.

---

#### Algorithm (in code)

```go
parts := strings.Split(cs.State.Running.ContainerID, "://")
if len(parts) == 2 && len(parts[1]) > 0 {
    return parts[1]
}
return ""
```

* Splits the `ContainerID` on `"//:"`.
* If two parts are produced and the second part is non‑empty, it is returned.
* Otherwise an empty string indicates a malformed or unknown format.

---

#### Dependencies
| Dependency | Role |
|------------|------|
| `strings.Split` | Tokenises the container ID. |
| `len` (twice)   | Checks array bounds and non‑emptiness of the UID part. |

No external packages beyond the Go standard library are used.

---

#### Side Effects & Safety
* **Read‑only** – The function only reads from the supplied status; it does not modify any state.
* **Nil safety** – It assumes `cs` is non‑nil and that `State.Running` exists. In production code, callers should guard against nil pointers to avoid panics.

---

#### Package Context
The `provider` package implements various helpers for probing Kubernetes clusters (e.g., node roles, pod status checks).  
`GetRuntimeUID` fits into this ecosystem as a small utility that other probe functions can use when they need the underlying container engine’s identifier—for example, during connectivity or runtime‑specific validation tests.

---

#### Suggested Mermaid Diagram
```mermaid
flowchart TD
    A[ContainerStatus] --> B{State.Running.ContainerID}
    B --> C["<runtime>://<uid>"]
    C --> D[Split on "://"]
    D --> E{len(parts)==2?}
    E -- Yes --> F[Return parts[1]]
    E -- No --> G[Return ""]
```

This visual captures the simple split-and-return logic that `GetRuntimeUID` performs.
