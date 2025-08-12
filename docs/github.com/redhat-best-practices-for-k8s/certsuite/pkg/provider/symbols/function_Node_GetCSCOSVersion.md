Node.GetCSCOSVersion`

| Aspect | Details |
|--------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider` |
| **Exported?** | Yes – available to callers of the provider package. |
| **Receiver type** | `Node` – represents a Kubernetes node that has already been enriched with status and metadata (e.g., labels, kubelet version). |

### Purpose
`GetCSCOSVersion` extracts the *CentOS Stream CoreOS* (CS‑COS) release number from a node’s kubelet version string.  
The function is used only for nodes identified as CS‑COS by `Node.IsCSCOS()`. It returns the numeric part of the image tag that follows the `-cscos-` suffix, e.g.:

```
k8s-csi/openshift-kube-node:v1.24.0-2022.10.12-0.cscos-5 => "5"
```

This value is later used by tests that validate that CS‑COS nodes are running the expected image version.

### Inputs / Outputs
| Parameter | Type | Notes |
|-----------|------|-------|
| Receiver (`n Node`) | – | Contains `KubeletVersion` string. |

| Return | Type | Description |
|--------|------|-------------|
| `string` | The numeric CS‑COS release (e.g., `"5"`). |
| `error` | Non‑nil if the node is not a CS‑COS node or the expected format cannot be parsed. |

### Algorithm
1. **Check CS‑COS** – call `n.IsCSCOS()`; return an error if false.  
2. **Split on “-cscos-”** – isolate the part after this suffix.  
3. **Trim whitespace** – clean any accidental spaces.  
4. **Extract first token** – split again on `"-"` to separate the release number from any following qualifiers (e.g., `"-1.cscos"`). Return that token.

### Key Dependencies
| Called function | Role |
|-----------------|------|
| `Node.IsCSCOS()` | Determines whether the node is a CS‑COS node. |
| `fmt.Errorf` | Formats error messages for callers. |
| `strings.Split`, `strings.TrimSpace` | String manipulation to parse the version tag. |

### Side Effects
* No modification of the receiver or global state.
* Only reads `n.KubeletVersion`.

### How it Fits the Package

- **Provider context** – The `provider` package implements logic that inspects Kubernetes objects (nodes, pods, etc.) for compliance tests.  
- **CS‑COS specific logic** – Tests that validate the image tag of CS‑COS nodes rely on this helper to isolate the numeric release component.  
- **Error handling** – If the node is not a CS‑COS or the format deviates, callers receive an error and can skip or flag the test accordingly.

### Example Usage

```go
node := provider.Node{KubeletVersion: "k8s-csi/openshift-kube-node:v1.24.0-2022.10.12-0.cscos-5"}
if node.IsCSCOS() {
    ver, err := node.GetCSCOSVersion()
    if err != nil { /* handle error */ }
    fmt.Println("CS‑COS release:", ver) // -> "5"
}
```

### Summary
`GetCSCOSVersion` is a small, pure helper that isolates the numeric CS‑COS image tag from a kubelet version string. It validates the node type first, then parses the string safely, returning an error when expectations are not met. This function supports higher‑level compliance checks within the `provider` package.
