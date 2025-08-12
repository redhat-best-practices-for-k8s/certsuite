GetLifeCycleDates`

**Package:** `github.com/redhat-best-practices-for-k8s/certsuite/pkg/compatibility`  
**Exported:** Yes

## Purpose
`GetLifeCycleDates` aggregates OpenShift Container Platform (OCP) version information and returns a map keyed by the OCP release string. Each entry contains a `VersionInfo` struct that holds lifecycle metadata such as end‑of‑life dates, GA status, and any other relevant fields.

The function is used throughout the package whenever callers need quick access to the current set of known OCP releases and their lifecycle state. It essentially acts as a read‑only accessor for the internal data structures that track beta versions and full lifecycle dates.

## Signature
```go
func GetLifeCycleDates() map[string]VersionInfo
```

- **Return value** – A `map[string]VersionInfo` where:
  - The key is an OCP release string (e.g., `"4.12"`).
  - The value is a `VersionInfo` struct containing lifecycle metadata.

There are no input parameters; the function relies entirely on internal package globals.

## Key Dependencies
| Global | Type | Role |
|--------|------|------|
| `ocpBetaVersions` | `map[string]bool` (inferred) | Tracks which releases are still in beta. The function consults this map to set the appropriate status flag in the returned `VersionInfo`. |
| `ocpLifeCycleDates` | `map[string]VersionInfo` (inferred) | Holds pre‑computed lifecycle dates for each release. The function likely copies or merges these entries into the result. |

Both globals are defined at package level, initialized elsewhere in `compatibility.go`.

## Side Effects
- **Read‑only**: No mutation of global state occurs; the returned map is a copy (or shallow copy) of internal data.
- **Thread safety**: Since the function only reads from immutable globals and returns a new map, it can be safely called concurrently.

## Usage Context
The function is invoked by other components that need to:

1. Determine if a particular OCP release has reached GA or is still in beta.
2. Retrieve the end‑of‑life date for a release when validating certificates against supported platforms.
3. Populate UI elements or logs with lifecycle information.

Example (pseudo):
```go
lifecycle := GetLifeCycleDates()
info, ok := lifecycle["4.12"]
if ok && !info.IsEOL() {
    // proceed with compatibility checks
}
```

## Suggested Mermaid Diagram

```mermaid
flowchart TD
  A[GetLifeCycleDates()] --> B{Read Globals}
  B -->|ocpBetaVersions| C[Determine Beta Status]
  B -->|ocpLifeCycleDates| D[Fetch Lifecycle Dates]
  C & D --> E[Construct VersionInfo Map]
  E --> F[Return map[string]VersionInfo]
```

This diagram illustrates the read‑only data flow from package globals into the returned map.
