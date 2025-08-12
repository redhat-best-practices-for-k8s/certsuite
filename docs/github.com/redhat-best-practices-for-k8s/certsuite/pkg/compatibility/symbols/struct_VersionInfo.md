VersionInfo`

| Feature | Detail |
|---------|--------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/compatibility` |
| **Location** | `compatibility.go:43-53` |
| **Exported** | ✅ |

## Purpose
`VersionInfo` is a data container that describes the compatibility requirements for a particular version of Red Hat‑Enterprise‑Core‑OS (RHCOS) and its associated RHEL releases.  
It is used by the `GetLifeCycleDates` function to expose life‑cycle dates (FSE, GA, MSE) and minimal OS versions required for certification.

## Fields

| Field | Type | Meaning |
|-------|------|---------|
| `FSEDate` | `time.Time` | **First Supported Edition** date – the earliest release that can be certified. |
| `GADate` | `time.Time` | **General Availability** date – when the edition becomes generally available to users. |
| `MSEDate` | `time.Time` | **Minimum Supported Edition** date – the latest release after which older versions are no longer supported. |
| `MinRHCOSVersion` | `string` | Minimum RHCOS version string that satisfies this life‑cycle window. |
| `RHELVersionsAccepted` | `[]string` | List of RHEL major/minor releases that are considered compatible with the corresponding RHCOS version. |

> **Note**: All dates are stored as UTC `time.Time`; callers should format them as needed.

## Dependencies
* Relies on Go’s standard library (`time` package).
* Populated by hard‑coded data in `GetLifeCycleDates`, which returns a map keyed by edition names (e.g., `"v4.13"`) mapping to `VersionInfo`.

## Side Effects & Usage

| Function | Effect |
|----------|--------|
| `GetLifeCycleDates` | Returns a read‑only map of edition identifiers → `VersionInfo`. No mutation occurs; the returned map is safe for concurrent reads. |

Typical usage:

```go
dates := compatibility.GetLifeCycleDates()
info, ok := dates["v4.13"]
if !ok {
    // handle unknown edition
}
fmt.Println("GA date:", info.GADate.Format(time.RFC3339))
```

## Integration in the Package

`compatibility.go` contains both the `VersionInfo` struct and the helper function `GetLifeCycleDates`.  
The struct encapsulates all information needed for compatibility checks performed elsewhere in *certsuite*, such as validating a cluster’s RHCOS version against certification requirements.

```mermaid
graph TD
    GetLifeCycleDates -->|returns map[string]VersionInfo| CompatibilityMap
    CompatibilityMap --lookup--> VersionInfo
```

> **Key Takeaway**: `VersionInfo` is the canonical representation of life‑cycle metadata for a certified RHCOS edition, enabling other components to query support windows and OS version constraints.
