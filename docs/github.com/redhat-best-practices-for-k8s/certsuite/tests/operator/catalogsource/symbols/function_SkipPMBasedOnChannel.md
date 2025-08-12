SkipPMBasedOnChannel` – CatalogSource Test Helper

## Purpose
In the **catalogsource** test suite this helper determines whether a particular Package‑Manager (PM) should be skipped based on the set of *PackageChannels* that are currently enabled for an operator’s catalog source.  
The logic is intentionally simple: if any channel name matches the supplied PM name, the function signals to skip that PM.

## Signature
```go
func SkipPMBasedOnChannel(channels []olmpkgv1.PackageChannel, pm string) bool
```

| Parameter | Type                           | Description |
|-----------|--------------------------------|-------------|
| `channels`| `[]olmpkgv1.PackageChannel`   | Slice of channel objects describing the catalog source’s available channels. |
| `pm`       | `string`                       | The name of the package‑manager being considered for skipping. |

**Return value**

- `true` – The PM should be skipped (a matching channel was found).  
- `false` – No match; proceed with testing that PM.

## Core Logic
```go
func SkipPMBasedOnChannel(channels []olmpkgv1.PackageChannel, pm string) bool {
    if len(channels) == 0 {           // No channels defined → nothing to skip.
        return false
    }
    for _, ch := range channels {     // Inspect each channel name.
        Debug("Checking channel %s against PM %s", ch.Name, pm)
        if strings.EqualFold(ch.Name, pm) {
            Debug("Channel matches PM – skipping")
            return true
        }
    }
    return false
}
```

### Key Points
1. **Early exit**: An empty channel list immediately returns `false`.  
2. **Case‑insensitive match** (`EqualFold`): Channel names and the PM name are compared without regard to case, ensuring robustness against naming variations.  
3. **Debug logging**: The function calls a package‑level `Debug` helper (likely a wrapper around `log.Printf`) four times:
   * When no channels exist.
   * For each channel inspected.
   * Upon finding a match.
   * When the loop finishes without a match.

These logs aid test diagnostics but have no functional side effects beyond console output.

## Dependencies
- **`olmpkgv1.PackageChannel`** – The type from the OLM (Operator Lifecycle Manager) API, providing at least a `Name` field.  
- **Standard library**:
  - `strings.EqualFold`
  - `len`

No global variables or external state are accessed; the function is pure aside from its debug prints.

## Integration in the Package
The `catalogsource` package contains tests that validate operator catalogs. When iterating over a list of known package‑manager names, the test harness calls `SkipPMBasedOnChannel` to filter out PMs whose corresponding channel already exists in the catalog source. This prevents duplicate or conflicting tests for the same PM.

```
for _, pm := range allPackageManagers {
    if SkipPMBasedOnChannel(channelsFromCatalogSource, pm) {
        t.Skipf("Skipping %s – channel present", pm)
        continue
    }
    // Run PM‑specific tests…
}
```

By centralising this check in a dedicated helper, the test suite stays concise and avoids duplicating matching logic across multiple test files.

## Summary

- **What it does**: Decides whether to skip testing a PM based on catalog channel names.  
- **Inputs/outputs**: Receives channels & PM name → returns `bool`.  
- **Dependencies**: OLM API, standard string comparison, debug logger.  
- **Side effects**: Emits debug logs only; otherwise pure.  
- **Package role**: Keeps catalog‑source tests DRY and deterministic by avoiding redundant PM checks.
