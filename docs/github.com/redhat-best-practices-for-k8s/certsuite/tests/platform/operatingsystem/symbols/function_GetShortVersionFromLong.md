GetShortVersionFromLong`

| | |
|---|---|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/tests/platform/operatingsystem` |
| **Exported** | Yes |
| **Signature** | `func GetShortVersionFromLong(long string) (string, error)` |

### Purpose
Converts a *long* operating‑system version string into its corresponding *short* form that is used by the test suite.  
The long format typically contains a full RHCOS build identifier such as `"rhcos-4.15.0-0.nightly-2023-04-01-123456"`.  
`GetShortVersionFromLong` maps this to a more concise representation like `"4.15"`.

### Inputs
* `long string` – the full, untrimmed OS version string supplied by the environment or test configuration.

### Outputs
* `string` – the short, canonical OS version (e.g., `"4.15"`) if mapping succeeds.
* `error` – non‑nil if:
  * The input does not match any known RHCOS build pattern.
  * An internal lookup fails (e.g., malformed map data).

When no mapping is found the function returns an empty string and a descriptive error.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `GetRHCOSMappedVersions` | Parses the embedded `rhcos_version_map` file (`rhcosVersionMap`) to build a lookup table from long to short versions. This table is used for the actual mapping. |
| `rhcosVersionMap` (embedded string) | Holds the raw content of `files/rhcos_version_map`. It is parsed by `GetRHCOSMappedVersions`. |

### Side‑Effects
* No state mutation – the function only reads from the embedded data and returns a result.
* Relies on the presence and correctness of the `rhcos_version_map` file; malformed content will surface as an error.

### Package Context
The `operatingsystem` package supplies utilities for interpreting operating‑system metadata used in CertSuite tests.  
Other components may call `GetShortVersionFromLong` to normalize version strings before comparing them against expected values or selecting test suites that target specific OS releases.

---

#### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[User Input: long string] --> B{Parse long string}
    B -->|Success| C[Lookup in rhcos_version_map]
    C --> D[Return short version]
    B -->|Failure| E[Return error "NotFoundStr"]
```

This visualizes the two‑step process: parsing the input and then looking it up in the embedded map.
