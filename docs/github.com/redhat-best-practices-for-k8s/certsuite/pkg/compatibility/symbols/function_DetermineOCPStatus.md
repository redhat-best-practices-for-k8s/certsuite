DetermineOCPStatus` – Compatibility Package

**Location**

```
pkg/compatibility/compatibility.go:321
```

## Purpose
`DetermineOCPStatus` translates an OpenShift Container Platform (OCP) version string and a reference date into one of five lifecycle status labels:

| Status constant | Meaning |
|-----------------|---------|
| `OCPStatusPreGA` | Version is still in “pre‑General Availability” (beta or release candidate). |
| `OCPStatusGA`     | Version has reached General Availability. |
| `OCPStatusMS`     | Version is in a maintenance‑support period. |
| `OCPStatusEOL`    | Version has reached End‑of‑Life – no more support. |
| `OCPStatusUnknown`| The version string cannot be mapped to the lifecycle table. |

The function is used by other parts of CertSuite when evaluating whether a cluster’s OCP release falls within supported ranges (e.g., in policy checks or audit reports).

## Signature
```go
func DetermineOCPStatus(ocpVersion string, refDate time.Time) string
```

* **`ocpVersion`** – Human‑readable OpenShift version such as `"4.12"` or `"4.12.1"`.  
  The function accepts any non‑empty string; it will split on `.` and use the major/minor part to look up dates.

* **`refDate`** – A reference point (usually current time) against which lifecycle thresholds are compared.

* **Return value** – One of the five status constants defined in this package.  
  The function never panics; it returns `OCPStatusUnknown` if parsing fails or data is missing.

## How It Works

1. **Input validation**  
   * If `ocpVersion` is empty, return `OCPStatusUnknown`.  
   * Use `strings.Split(ocpVersion, ".")` to isolate the major/minor components (e.g., `"4"` and `"12"`).

2. **Retrieve lifecycle dates**  
   * Call `GetLifeCycleDates` (defined elsewhere in the same file) with the split parts.  
   * The helper returns a struct containing `PreGA`, `GA`, `MS`, and `EOL` timestamps, or zero values if the version is unknown.

3. **Determine status relative to `refDate`**  
   * If all returned dates are zero → `OCPStatusUnknown`.  
   * Compare `refDate` against each threshold in order:
     1. `PreGA`: `refDate.Before(PreGA)` or `refDate.Equal(PreGA)` → `OCPStatusPreGA`.
     2. `GA`: `refDate.Before(GA)` or `refDate.Equal(GA)` → `OCPStatusGA`.
     3. `MS`: `refDate.Before(MS)` or `refDate.Equal(MS)` → `OCPStatusMS`.
     4. `EOL`: `refDate.After(EOL)` (or equal) → `OCPStatusEOL`.
   * If none of the above matches, fall back to `OCPStatusUnknown`.

All date comparisons use Go’s standard library (`time.Time.Before`, `.Equal`, `.After`).

## Dependencies & Side‑Effects

| Dependency | Role |
|------------|------|
| `strings.Split` | Extract major/minor version components. |
| `GetLifeCycleDates` | Provides the lifecycle dates for a given OCP release. |
| `time.Time` methods (`Before`, `Equal`, `After`) | Perform chronological comparisons. |

No global state is mutated; the function is pure with respect to its inputs and the read‑only lifecycle tables (`ocpBetaVersions`, `ocpLifeCycleDates`). It only reads from these maps via `GetLifeCycleDates`.

## Integration

- **Policy evaluation** – When a policy references supported OCP versions, it calls `DetermineOCPStatus` to verify that the cluster’s release is at least GA and not EOL.
- **Reporting** – Audit reports may list each node’s status; this function feeds those strings.

The mapping logic mirrors Red‑Hat’s official OCP lifecycle schedule, enabling CertSuite to enforce compliance automatically.
