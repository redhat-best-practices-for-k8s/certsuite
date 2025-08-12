BetaRHCOSVersionsFoundToMatch`

> **Location**  
> `github.com/redhat-best-practices-for-k8s/certsuite/pkg/compatibility/compatibility.go:242`

## Purpose
`BetaRHCOSVersionsFoundToMatch` determines whether a pair of RHEL‑based OpenShift Container Platform (RH‑COS) release strings can be considered equivalent for the purposes of compatibility checks.  
The function is used when evaluating *beta* versions of RH‑COS against known stable releases to decide if they should be treated as matching.

## Signature
```go
func BetaRHCOSVersionsFoundToMatch(v1, v2 string) bool
```
- **Parameters**
  - `v1`, `v2`: strings containing version identifiers (e.g., `"4.15"` or `"4.16-beta"`).
- **Returns**: `true` if the two versions are considered a match; otherwise `false`.

## How it works

| Step | Action | Notes |
|------|--------|-------|
| 1 | Extract major/minor numbers from each input string using `FindMajorMinor`. | `FindMajorMinor` returns `(int, int)` for major and minor components. |
| 2 | Compare the extracted major/minor pairs. | If they are identical, a match is found immediately. |
| 3 | Handle “beta” qualifiers: |
|   | *If* either input contains `"beta"` (checked via `StringInSlice`), treat the version as a beta candidate. | The function checks if the string contains `"beta"` by scanning for that substring. |
|   | Compare the major/minor of the non‑beta and beta versions. | If they match, return `true`. |
| 4 | Default to `false`. | No match found. |

### Dependencies
- **`FindMajorMinor`** – parses a version string into its numeric components.
- **`StringInSlice`** – simple substring containment check (used here to detect `"beta"`).

No global state is read or modified; the function is pure aside from calling these helpers.

## Side‑effects
None. The function is deterministic and has no observable effect on package state.

## Package Context
The `compatibility` package maintains lists of OpenShift Container Platform (OCP) versions and their life‑cycle dates (`ocpBetaVersions`, `ocpLifeCycleDates`).  
`BetaRHCOSVersionsFoundToMatch` assists in reconciling beta RH‑COS releases with stable OCP releases during compatibility verification, enabling the suite to treat a beta release as compatible when its major/minor numbers align.

---

### Suggested Mermaid Diagram

```mermaid
flowchart TD
  A[Input v1] -->|FindMajorMinor| B(major1, minor1)
  C[Input v2] -->|FindMajorMinor| D(major2, minor2)
  B --> E{major1==major2 & minor1==minor2}
  E -- yes --> F[Return true]
  E -- no --> G{v1 contains "beta" or v2 contains "beta"}
  G -- yes --> H{major/minor match across beta/non‑beta}
  H -- yes --> F
  H -- no --> I[Return false]
  G -- no --> I
```

This diagram illustrates the decision flow of the function.
