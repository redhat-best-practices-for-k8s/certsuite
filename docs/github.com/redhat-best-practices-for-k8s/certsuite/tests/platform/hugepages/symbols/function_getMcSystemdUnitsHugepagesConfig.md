getMcSystemdUnitsHugepagesConfig` |
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/tests/platform/hugepages` |
| **Signature** | `func(*provider.MachineConfig) (hugepagesByNuma, error)` |
| **Exported** | No – used internally by the huge‑page test harness. |

### Purpose

`getMcSystemdUnitsHugepagesConfig` extracts huge‑page configuration information from a `MachineConfig` instance’s systemd unit files.  
The function scans each systemd unit for `vm.nr_hugepages` and `vm.hugetlb_shm_group` entries, parses their values, and returns a mapping of NUMA node indices to the number of configured huge pages.

### Inputs

| Parameter | Type | Description |
|-----------|------|-------------|
| `mc` | `*provider.MachineConfig` | A pointer to a MachineConfig object that contains a collection of systemd unit definitions. The function inspects the `Units` slice for relevant configuration directives. |

### Outputs

| Return Value | Type | Meaning |
|--------------|------|---------|
| `hugepagesByNuma` | custom map type (`map[int]int`) | Maps NUMA node index → number of huge pages configured on that node. A missing entry indicates the default (zero) value for that node. |
| `error` | `error` | Non‑nil if parsing fails or required data cannot be extracted. The error message includes a short description and any relevant line numbers. |

### Key Dependencies

| Dependency | Role |
|------------|------|
| `provider.MachineConfig` | Supplies the systemd unit definitions to parse. |
| `HugepagesParam`, `HugepageszParam` constants | Define the exact parameter names that the function looks for in unit files (`vm.nr_hugepages`, `vm.hugetlb_shm_group`). |
| Regular expressions (`outputRegex`) | Used to capture key‑value pairs from unit file lines. |
| Standard library functions | `strings.Trim`, `strings.Contains`, `strconv.Atoi`, `regexp.MustCompile`, etc., for string manipulation and numeric conversion. |

### Core Logic

1. **Prepare regex** – Compile a regular expression that matches a parameter line (`key=value`) in systemd unit files.
2. **Iterate units** – For each unit in `mc.Units`:
   * Skip non‑relevant units (e.g., those not containing huge‑page parameters).
3. **Parse lines** – Split the unit’s content into lines, trim whitespace, and check for parameter names (`HugepagesParam`, `HugepageszParam`).
4. **Extract values** – For each matching line:
   * Use regex to capture the value after the equals sign.
   * Convert the captured string to an integer with `strconv.Atoi`.
5. **Populate map** – Store the parsed huge‑page count in a local map keyed by NUMA node index (derived from the unit name or comment).
6. **Return result** – If all parsing succeeds, return the populated map; otherwise, return an error indicating the problematic unit and line.

### Side Effects

* The function only reads data from `mc`; it does not modify any state.
* Logging via `Info` statements is performed for debugging but has no impact on the returned values.

### How It Fits the Package

The `hugepages` package implements tests that validate huge‑page configuration on a node.  
During test setup, a `MachineConfig` object representing the node’s current configuration is supplied to various helper functions.  

`getMcSystemdUnitsHugepagesConfig` is one of those helpers; it provides a low‑level view of how many huge pages are configured per NUMA node by inspecting systemd unit files directly.  
Higher‑level test logic then compares this map against expected defaults or user‑specified values, ensuring that the system’s huge‑page settings match policy requirements.

### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[MachineConfig] --> B{Iterate Units}
    B -->|Unit contains param| C[Parse line]
    C --> D[Extract value]
    D --> E[Convert to int]
    E --> F[Store in map[numaIndex]]
```

This diagram visualizes the step‑by‑step extraction of huge‑page settings from systemd unit files.
