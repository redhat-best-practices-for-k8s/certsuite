Pod.CheckResourceHugePagesSize`

| | |
|---|---|
| **Package** | `provider` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider`) |
| **Exported** | ✅ |
| **Receiver** | `p Pod` – a struct that represents a Kubernetes pod (defined elsewhere in the package). |
| **Signature** | `func (p Pod) CheckResourceHugePagesSize(size string) bool` |

#### Purpose
`CheckResourceHugePagesSize` is a helper that validates whether a given *huge‑pages* size string matches one of the two supported huge‑page sizes used by CertSuite:

* `2Mi` – 2 MiB pages (`HugePages2Mi`)
* `1Gi` – 1 GiB pages (`HugePages1Gi`)

The function is invoked when a pod declares a `resources.limits.hugepages-<size>` entry. It guarantees that the pod’s huge‑page request uses a size that CertSuite knows how to handle and report.

#### Parameters
* **`size string`** – The raw value from the pod spec, e.g. `"2Mi"` or `"1Gi"`.  
  It may contain leading/trailing whitespace; the function trims it before comparison.

#### Return Value
* **`bool`** – `true` if `size` equals either `HugePages2Mi` or `HugePages1Gi`; otherwise `false`.

#### Implementation details
```go
func (p Pod) CheckResourceHugePagesSize(size string) bool {
    size = strings.TrimSpace(size)

    // 1. Handle the special case of "2Mi" first
    if len(size) == 3 && strings.Contains(size, "Mi") && strings.Contains(size, "2") {
        return true
    }

    // 2. Handle the special case of "1Gi"
    if len(size) == 3 && strings.Contains(size, "Gi") && strings.Contains(size, "1") {
        return true
    }

    // 3. Fallback: compare against the exported constants
    return size == HugePages2Mi || size == HugePages1Gi
}
```
* The function uses only standard library functions (`strings.TrimSpace`, `strings.Contains`, and `len`).  
* No external state or globals are read; it is fully deterministic.

#### Side‑effects
None – the function does not modify any package state, nor does it perform I/O.

#### How it fits the package
Within `provider/pods.go` the `Pod` type represents a Kubernetes pod being evaluated by CertSuite.  
During the *resource validation* step, each container’s resource limits are inspected. When a huge‑page limit is found, `CheckResourceHugePagesSize` ensures that the requested size is supported before reporting it as compliant or not.

This keeps the huge‑page logic isolated from the rest of the pod‑inspection code and makes unit testing trivial.
