Pod.IsRunAsUserID`

### Purpose
`IsRunAsUserID` determines whether a container in a Pod is configured to run as the user ID passed to the function.

- **When true**: The container’s security context explicitly sets `runAsUser` to the supplied UID.
- **When false**: Either no explicit `runAsUser` is set, or it differs from the provided UID.

This check is used by CertSuite when validating that workloads run with the correct privilege level (e.g., non‑root) and for generating compliance reports.

### Signature
```go
func (p Pod) IsRunAsUserID(uid int64) bool
```

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `uid`     | `int64`| The UID to check against the container’s security context. |

| Return | Type  | Description |
|--------|-------|-------------|
| `bool` | true/false | Indicates whether the container runs as the specified UID. |

### How It Works
1. **Iterate Containers**  
   For each container in the Pod (both init and normal containers) the function:
   - Skips containers whose names are listed in `ignoredContainerNames`.
   - Checks if a security context exists.

2. **Check `runAsUser`**  
   If the security context contains a `RunAsUser` field, it is compared to the supplied `uid`.  
   - If equal → return `true`.
   - Otherwise continue scanning.

3. **Return Result**  
   If none of the containers match the UID, the function returns `false`.

### Dependencies & Side‑Effects
- Relies on the Pod’s container list (`p.Spec.InitContainers`, `p.Spec.Containers`).
- Uses the global slice `ignoredContainerNames` to filter out system containers.
- No state is mutated; purely read‑only analysis.

### Placement in Package

| File | Role |
|------|------|
| `pods.go` | Core Pod inspection utilities |

`IsRunAsUserID` sits among other helper methods that introspect Pod security settings. It is invoked by higher‑level checks such as:

- **Non‑root verification** – ensuring workloads do not run as UID 0.
- **Compliance reporting** – tagging pods with the effective user ID.

Because it only reads pod data, it can be safely called from any part of CertSuite that has a `Pod` instance and needs to validate its privilege configuration.
