CreateLabels`

### Purpose
`CreateLabels` parses a slice of label strings (in the form `"key=value"`) and converts each into a `labelObject`.  
The function is used by the autodiscover package to turn raw label definitions that may contain regular‑expression syntax into structured objects that can be attached to Kubernetes resources during discovery.

### Signature
```go
func CreateLabels(labels []string) []labelObject
```

| Parameter | Type      | Description |
|-----------|-----------|-------------|
| `labels`  | `[]string`| Raw label definitions supplied by the user or from configuration. |

| Return value | Type            | Description |
|--------------|-----------------|-------------|
| `[]labelObject` | Slice of parsed labels | Each element contains a key, a raw value, and an optional compiled regex (if the value matches the global `labelRegex`). |

### Key Steps

1. **Pre‑compile regex**  
   ```go
   re := MustCompile(labelRegex)
   ```
   The package level variable `labelRegex` is a regular expression that captures three groups:
   * key (`[^=]+`)
   * separator (`=` or `:`)
   * value (`.+`)

2. **Iterate over input strings**  
   For each string `s` in `labels`:
   * Use `re.FindStringSubmatch(s)` to extract the captured groups.
   * If a match is found, create a `labelObject` with:
     * `Key`: first capture group
     * `Value`: third capture group (the actual value)
     * `Regex`: if the matched string differs from the original input, compile that substring into a regex (`MustCompile`).  
       This allows later code to test whether the label matches dynamic values.

3. **Collect results**  
   Append each successfully parsed object to a slice and return it at the end of the function.

### Dependencies & Side Effects
* Relies on package globals:
  * `labelRegex` – pattern used for parsing.
  * `labelTemplate`, `tnfLabelPrefix`, etc., are not directly touched but may be used elsewhere in the same file to generate label keys.
* Uses standard library functions: `MustCompile`, `FindStringSubmatch`, `append`.
* No external state is mutated; all work is done locally and the returned slice is independent of any package‑level variables.

### How It Fits the Package
`autodiscover` builds dynamic Kubernetes resource manifests that need to be annotated with user‑supplied labels.  
`CreateLabels` is a helper that normalises those label strings into objects that downstream code (e.g., `buildPodSpec`, `applyManifest`) can consume without re‑parsing or regex compiling repeatedly.

```mermaid
flowchart LR
  A[Input []string] --> B{Parse each string}
  B -->|Valid| C[labelObject]
  B -->|Invalid| D[Skip / log error]
  C --> E[Collect into slice]
  E --> F[Return []labelObject]
```

> **Note:** If a label string does not match `labelRegex`, it is silently ignored – callers should validate input elsewhere if strict enforcement is required.
