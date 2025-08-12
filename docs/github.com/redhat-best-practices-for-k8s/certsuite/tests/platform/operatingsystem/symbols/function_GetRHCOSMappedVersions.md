GetRHCOSMappedVersions`

| Aspect | Detail |
|--------|--------|
| **Package** | `operatingsystem` (github.com/redhat-best-practices-for-k8s/certsuite/tests/platform/operatingsystem) |
| **Exported?** | ✅ Yes – available to callers outside the package. |
| **Signature** | `func GetRHCOSMappedVersions(raw string) (map[string]string, error)` |

---

## Purpose
`GetRHCOSMappedVersions` converts a raw string representation of RHEL‑CoreOS (RHCOS) version mappings into an in‑memory Go map.

The function is used by tests that need to resolve a *raw* mapping file (for example, the contents of `files/rhcos_version_map`) into a usable data structure without having to read or parse the file each time.

---

## Inputs
| Parameter | Type   | Description |
|-----------|--------|-------------|
| `raw`     | `string` | The raw text containing multiple lines. Each line is expected to be of the form:

```
<key> <value>
```

where `<key>` and `<value>` are strings separated by whitespace (tabs or spaces). Lines may have leading/trailing whitespace, which must be ignored.

---

## Outputs
| Return value | Type        | Description |
|--------------|-------------|-------------|
| `map[string]string` | A map where each key/value pair corresponds to a line in the input. The key is the trimmed first token, and the value is the trimmed second token. |
| `error` | If the function cannot produce a valid mapping (e.g., an empty input or malformed lines), it returns a non‑nil error. Otherwise `nil`. |

---

## Key Dependencies & Side Effects
- **String helpers**: Uses standard library functions `strings.Split`, `strings.TrimSpace` to parse and clean up each line.
- **No external state**: The function is pure with respect to the Go runtime; it only operates on its input argument.  
  - It does *not* read from or write to files, nor modify global variables.
- **Global variable**: `rhcosVersionMap` (type `string`) holds the embedded file content but is not referenced directly inside this function. The caller typically supplies `rhcosVersionMap` as the argument.

---

## How It Works (Step‑by‑step)

```mermaid
flowchart TD
    A[Start] --> B{Split raw by "\n"}
    B --> C{Trim each line}
    C --> D{Skip empty lines?}
    D -- No --> E[Parse key/value]
    E --> F[Add to map]
    D -- Yes --> G[Continue]
    G --> H{End of lines?}
    H -- No --> C
    H -- Yes --> I[Return map, nil error]
```

1. **Split** the raw string on newline characters (`\n`).
2. For each resulting line:
   * Trim leading/trailing whitespace.
   * Skip if the trimmed line is empty.
   * Split the line into two parts using any amount of whitespace as a delimiter (`strings.Split(line, " ")`).
   * Trim spaces from both parts to clean up accidental padding.
3. Store the pair in the output map.
4. After processing all lines, return the populated map and `nil`.

---

## Usage Example

```go
// Assume rhcosVersionMap is embedded via //go:embed
raw := operatingsystem.rhcosVersionMap

m, err := operatingsystem.GetRHCOSMappedVersions(raw)
if err != nil {
    log.Fatalf("failed to parse RHCOS map: %v", err)
}
fmt.Printf("RHCOS mapping: %+v\n", m)
```

---

## Integration in the Package
`GetRHCOSMappedVersions` is a small utility that supports test cases which need to validate or consume RHCOS version mappings.  
It keeps parsing logic isolated, making it easy to unit‑test and reuse across multiple tests without duplicating code.

---
