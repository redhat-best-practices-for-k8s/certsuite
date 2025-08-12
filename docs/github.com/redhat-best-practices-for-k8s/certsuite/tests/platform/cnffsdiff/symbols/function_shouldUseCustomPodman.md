shouldUseCustomPodman`

| Item | Description |
|------|-------------|
| **Signature** | `func( *checksdb.Check, string ) bool` |
| **Package** | `cnffsdiff` (path: `github.com/redhat-best-practices-for-k8s/certsuite/tests/platform/cnffsdiff`) |
| **Exported?** | No – used internally by the test harness. |

### Purpose
Determines whether a *custom* Podman binary (bundled inside probe pods) should be invoked instead of the system‑installed one that ships with OpenShift Container Platform (OCP) nodes.

The custom binary is required only on older, RHEL 8.x‑based OCP releases (≤ 4.12).  
From OCP ≥ 4.13 the default Podman implementation works correctly and no workaround is needed.

### Parameters

| Name | Type | Role |
|------|------|------|
| `check` | `*checksdb.Check` | Contains metadata about the node under test, notably its **OCP version** (`Version`) in semantic‑version format. |
| `ocpVersionString` | `string` | The OCP version extracted from the node’s *Kubernetes API* (e.g., `"4.12.34"`). This is passed by the caller to avoid re‑parsing the same string repeatedly. |

### Return value

- **`true`** – use the bundled custom Podman binary.  
- **`false`** – rely on the node’s preinstalled Podman.

### Core logic flow

1. **Parse OCP version**  
   ```go
   v, err := semver.NewVersion(ocpVersionString)
   if err != nil { … }
   ```
   The `semver` package (imported via `github.com/Masterminds/semver/v3`) is used to interpret the semantic‑version string.

2. **Determine major/minor components**  
   ```go
   major := v.Major()
   minor := v.Minor()
   ```

3. **Decision rule**  
   * If `major` == 4 and `minor` ≤ 12 → return `true`.  
   * Otherwise (e.g., 4.13+, 5.x) → return `false`.

   Any parsing error is logged (`LogError`) and the function conservatively returns `false`, assuming the node’s native Podman can be used.

### Dependencies

| Dependency | Role |
|------------|------|
| `semver.NewVersion` | Parses OCP version string into a comparable object. |
| `LogError` | Records parsing failures (from `checksdb.Check`). |
| `Major`, `Minor` | Retrieve numeric components from the parsed semantic version. |

### Side‑effects

- Emits an error log if the supplied OCP version cannot be parsed.
- No state mutation; purely deterministic based on inputs.

### Integration in the package

The `cnffsdiff` package implements a file‑system diffing mechanism for certificates inside OpenShift pods.  
Before performing operations that rely on Podman (e.g., mounting volumes or executing container commands), tests must decide which binary to invoke:

```go
if shouldUseCustomPodman(check, ocpVersion) {
    // use bundled podman in the probe pod
} else {
    // call system podman on the node
}
```

This helper centralizes that decision logic, ensuring consistent behavior across all test cases and simplifying future version updates.
