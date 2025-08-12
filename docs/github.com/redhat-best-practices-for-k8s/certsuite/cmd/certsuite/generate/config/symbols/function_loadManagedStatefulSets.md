loadManagedStatefulSets`

| Item | Details |
|------|---------|
| **Location** | `config/config.go:402` |
| **Visibility** | Unexported (private to the package) |
| **Signature** | `func([]string) func()` |

### Purpose
Collects a slice of stateful‑set names that are *managed* by CertSuite and returns a function that, when executed, appends these names to the global configuration (`certsuiteConfig`). This helper is used during the interactive configuration wizard to persist user selections for managed stateful sets.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `statefulSets []string` | `[]string` | A list of stateful‑set names that the user has chosen (or that were auto‑detected). |

> The slice is passed by value; the function does **not** modify it.

### Return Value
A zero‑argument function (`func()`) that, when called, appends the supplied `statefulSets` to `certsuiteConfig.ManagedStatefulSets`.  
The returned closure captures the original slice and the global configuration variable.

```go
// Example usage inside the wizard:
onConfirm := loadManagedStatefulSets(selected)
onConfirm() // persists selection
```

### Key Dependencies & Side Effects

| Dependency | Effect |
|------------|--------|
| `certsuiteConfig` (global) | The closure modifies its field `ManagedStatefulSets`. This is the only side effect. |
| Standard library `append` | Used to concatenate the new slice onto the existing configuration slice. |

No other global variables or external packages are touched.

### How It Fits in the Package

The `config` package implements an interactive CLI for generating a CertSuite configuration file.  
During the wizard, each menu section (e.g., *Managed Stateful Sets*, *Deployments*, etc.) collects user input and stores it temporarily. When the user confirms a choice, helper functions like `loadManagedStatefulSets` are invoked to move that data into the persistent `certsuiteConfig`.  

This pattern is repeated for other resource types (`deployments`, `pods`, etc.), keeping the wizard logic clean: menus collect values → closures persist them → final configuration is written out.

### Mermaid Diagram (Optional)

```mermaid
flowchart TD
    A[User selects Stateful Sets] --> B{Confirm}
    B -- yes --> C[loadManagedStatefulSets(selected) -> closure]
    C --> D[closure() executes]
    D --> E[certsuiteConfig.ManagedStatefulSets updated]
```

---
