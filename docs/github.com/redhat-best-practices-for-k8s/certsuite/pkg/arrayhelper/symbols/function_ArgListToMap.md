ArgListToMap`

**Package:** `arrayhelper`  
**File:** `pkg/arrayhelper/arrayhelper.go` – line 25

## Purpose
Converts a slice of strings that each contain a single key/value pair in the form `"key=value"` into a Go map (`map[string]string`).  
The function is a small helper used by command‑line parsers and configuration loaders to transform user supplied arguments (e.g. `--set foo=bar`) into an easily consumable data structure.

## Signature
```go
func ArgListToMap(args []string) map[string]string
```

| Parameter | Type         | Description |
|-----------|--------------|-------------|
| `args`    | `[]string`   | Slice of strings, each expected to contain exactly one `"="` separator. |

| Return | Type              | Description |
|--------|-------------------|-------------|
| map    | `map[string]string` | Key/value pairs extracted from the input slice; keys are trimmed of surrounding spaces, values are left as‑is after removing any surrounding double quotes (`"`). |

## Algorithm & Dependencies
1. **Allocation**  
   ```go
   result := make(map[string]string, len(args))
   ```
   A map is pre‑allocated with the same capacity as the input slice to avoid re‑allocations.

2. **Iteration** – For each element `arg` in `args`:  
   a. Split on the first `"="`:
      ```go
      parts := strings.Split(arg, "=")
      ```
      If the split yields fewer than 2 elements, the entry is ignored (invalid format).  

   b. Clean up the value:  
      ```go
      cleaned := strings.ReplaceAll(parts[1], "\"", "")
      ```
      All double‑quote characters are removed so that values like `"foo"` become `foo`.  

   c. Store in map:  
      ```go
      result[parts[0]] = cleaned
      ```

3. **Return** – The populated map.

The function relies on standard library functions:
- `make` (map allocation)
- `strings.Split`
- `strings.ReplaceAll`
- `len` for capacity calculation

No global state is accessed or modified; the function is pure from a functional‑programming perspective.

## Side Effects
- **None** – No I/O, no modification of external variables.
- It silently ignores malformed entries (those without an `"="`) instead of returning an error.

## Integration in the Package
`ArgListToMap` is exported for use by other packages that need to transform a list of key/value strings into a map.  
Typical callers include:
- CLI flag parsers that expose `--set` options.
- Configuration loaders that accept user‑supplied overrides.

It is intentionally simple and has no dependencies beyond the Go standard library, making it safe for reuse in multiple contexts within the `certsuite` codebase.

---

### Mermaid Diagram (Optional)
```mermaid
flowchart TD
    A[Input Slice] --> B{For each element}
    B --> C[Split on "="]
    C --> D{Has two parts?}
    D -- Yes --> E[Remove quotes from value]
    E --> F[Insert key/value into map]
    D -- No --> G[Skip]
    G --> H[Continue loop]
    F --> H
    H --> I[Return map]
```
This diagram illustrates the core logic flow of `ArgListToMap`.
