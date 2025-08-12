loadHelmCharts`

### Purpose
`loadHelmCharts` builds a *post‑processing* function that appends the supplied Helm chart names into the current configuration object (`certsuiteConfig`).  
The returned function is intended to be executed after the main command has gathered all user input, ensuring that the list of charts is stored in the final configuration.

### Signature
```go
func loadHelmCharts(charts []string) func()
```
* **Input** – `charts`: a slice of strings representing Helm chart names that the user selected.
* **Output** – an anonymous function (`func()`) that, when called, mutates the global `certsuiteConfig` by appending each element from `charts`.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `append`   | Used inside the returned closure to add chart names to the config slice. |
| `certsuiteConfig` | Global configuration struct that holds all user‑selected options; the closure updates its `HelmCharts` field (or equivalent). |

### Side Effects
* Mutates the global `certsuiteConfig`, adding new entries.
* No return value other than the closure itself, so the side effect is the only observable change.

### How It Fits in the Package
The `config` package orchestrates interactive configuration generation for CertSuite.  
During the command flow (`generateConfigCmd`), various helper functions collect user choices (e.g., namespaces, services).  
Each helper returns a function that performs a specific mutation on `certsuiteConfig`.  
`loadHelmCharts` follows this pattern: it is called with the slice of chart names chosen by the user and supplies a closure to be executed later in the command lifecycle. This keeps the main logic clean while ensuring all collected data ends up in the final configuration file.

```mermaid
graph TD;
    subgraph User Interaction
        A[Select Helm Charts] --> B[loadHelmCharts(charts)]
    end
    subgraph Configuration Build
        B --> C{Closure}
        C --> D[certsuiteConfig.HelmCharts = append(...)]
    end
```

> **Note** – The actual field name in `certsuiteConfig` that receives the charts is not shown here; it is inferred to be a slice that can be appended to.

This function is intentionally unexported because its use is confined to the internal command workflow.
