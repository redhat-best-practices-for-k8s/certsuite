loadNonScalableDeployments`

| Item | Detail |
|------|--------|
| **Signature** | `func([]string)()` |
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/config` |
| **Visibility** | Unexported – used only inside the config generator CLI. |

### Purpose
Collect a list of *non‑scalable* Kubernetes Deployment names from a slice of raw strings and store them in the global configuration structure (`certsuiteConfig`).  
The function is invoked when the user selects the “Non‑Scalable Deployments” menu option during interactive configuration generation.

### Parameters
- `deployments []string` – each element contains a single deployment name. The caller may pass an empty slice if the user skipped this step.

### Return value
None (void).  
The function updates global state and optionally prints a confirmation message.

### Core logic

```go
func loadNonScalableDeployments(deployments []string) () {
    // If no input was provided, nothing to do.
    if len(deployments) == 0 { return }

    for _, d := range deployments {
        // Allow comma‑separated names in a single entry.
        parts := strings.Split(d, ",")
        for _, p := range parts {
            trimmed := strings.TrimSpace(p)
            if trimmed != "" {
                certsuiteConfig.NonScalableDeployments = append(
                    certsuiteConfig.NonScalableDeployments,
                    trimmed,
                )
            }
        }
    }

    fmt.Println("Non‑scalable deployments loaded.")
}
```

1. **Early exit** – If the slice is empty, the function returns immediately to avoid unnecessary work.
2. **Comma handling** – Each string may contain multiple names separated by commas; `strings.Split` splits them.
3. **Trimming & validation** – Whitespace is removed with `strings.TrimSpace`; empty entries are ignored.
4. **State mutation** – Valid names are appended to the global slice `certsuiteConfig.NonScalableDeployments`.
5. **User feedback** – A simple `Println` confirms that the values were loaded.

### Dependencies
- Standard library:
  - `strings.Split`
  - `len`
  - `fmt.Println`
  - `append`
- Global variable:  
  - `certsuiteConfig` (type unknown from snippet, but contains a field named `NonScalableDeployments []string`).

### Side effects
- Modifies the global configuration state (`certsuiteConfig`).
- Emits a message to stdout.

### Role in the package
Within the CLI workflow, each menu item has an associated loader function.  
`loadNonScalableDeployments` is registered as the handler for the “non‑scalable deployments” option. It ensures that any user input for this setting is captured and persisted before the final configuration file is written. The resulting `certsuiteConfig.NonScalableDeployments` slice later influences which Deployments are excluded from scaling operations in the generated certsuite manifests.
