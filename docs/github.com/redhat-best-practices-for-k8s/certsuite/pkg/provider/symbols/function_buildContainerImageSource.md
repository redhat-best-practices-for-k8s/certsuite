buildContainerImageSource`

| Aspect | Detail |
|--------|--------|
| **Location** | `pkg/provider/provider.go:497` |
| **Visibility** | unexported (internal helper) |
| **Signature** | `func buildContainerImageSource(registry, image string) ContainerImageIdentifier` |

### Purpose
Transforms a registry URL and an image name into a fully‑qualified container image reference (`ContainerImageIdentifier`).  
The function is used by the provider to map a *raw* image description (e.g. `"quay.io/owner/repo:tag"`) into the canonical form that can be compared against the set of images actually pulled on a node or pod.

### Inputs
| Parameter | Type   | Description |
|-----------|--------|-------------|
| `registry` | `string` | The registry host part (e.g. `"quay.io"`). May contain port (`"localhost:5000"`). |
| `image`    | `string` | Image name that may already be fully qualified or just the repository/tag portion. |

### Output
- **Value** – a `ContainerImageIdentifier`, which is an alias for `string`.  
  It contains the full image reference in the form `<registry>/<repo>:<tag>`.

### Algorithm Overview
1. **Parse registry**  
   ```go
   regRe := regexp.MustCompile(`(?i)^([^/]+)/?(.*)$`)
   ```  
   *Matches a registry prefix (e.g., `"quay.io"`) optionally followed by a `/`. The remainder is captured for later use.*

2. **Parse image**  
   ```go
   imgRe := regexp.MustCompile(`(?i)^(?:([a-z0-9.-]+(?:/[a-z0-9._-]+)*):?([a-zA-Z0-9_.-]+)?$`)
   ```  
   *Captures the repository path and an optional tag.*

3. **Apply defaults**  
   - If `registry` is empty after parsing, it falls back to the registry part captured from `image`.  
   - If no tag is supplied, `"latest"` is assumed.

4. **Construct identifier**  
   ```go
   return ContainerImageIdentifier(fmt.Sprintf("%s/%s:%s", registry, repo, tag))
   ```

5. **Debug logging** – the final identifier is logged with `log.Debug`.

### Dependencies & Side‑Effects
- Uses Go's `regexp` package (`MustCompile`, `FindStringSubmatch`).  
- Calls `log.Debug` for diagnostics; no other global state is modified.
- No external side effects (no I/O, no modification of shared variables).

### Interaction with the Package
`buildContainerImageSource` is a helper used by higher‑level functions that gather container image data from nodes or pods. By normalizing image references it enables consistent comparison against expected images and simplifies reporting in test results.

---

#### Suggested Mermaid diagram (usage flow)

```mermaid
flowchart TD
  A[Call buildContainerImageSource(reg, img)]
  B{Parse registry}
  C{Parse image}
  D[Apply defaults]
  E[Return ContainerImageIdentifier]
  
  A --> B --> C --> D --> E
```

This diagram shows the linear transformation from raw inputs to a fully qualified identifier.
