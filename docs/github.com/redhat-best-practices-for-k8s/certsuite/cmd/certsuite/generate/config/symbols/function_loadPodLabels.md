loadPodLabels`

| Item | Detail |
|------|--------|
| **Signature** | `func([]string)()`. The function accepts a slice of strings and returns an unnamed closure with no parameters and no return value. |
| **Purpose** | To register a *configuration callback* that will populate the global `certsuiteConfig` structure with pod‑label information. In the CLI workflow this is invoked when the user chooses the *Pod Labels* menu item, allowing the program to capture the list of labels that should be applied to generated Pods. |
| **Parameters** | - `labels []string` – a slice containing key/value pairs (e.g. `"app=web"`). The function does not modify this slice directly; it simply passes the values into the callback closure. |
| **Return value** | A closure (`func()`) that, when executed, assigns the provided labels to `certsuiteConfig.PodLabels`. This pattern is used throughout the command package so that each menu option can defer configuration until the user has finished interacting with the UI. |

### How it works (inferred)

1. The function creates a local copy of the incoming `labels` slice.
2. It returns an anonymous function that, when called, sets  
   ```go
   certsuiteConfig.PodLabels = labelsCopy
   ```
3. This closure is stored in the `generateConfigCmd` command tree and executed during the configuration‑generation phase.

### Dependencies & Side‑effects

| Dependency | Effect |
|------------|--------|
| `certsuiteConfig` (global) | The closure mutates its `PodLabels` field, affecting subsequent steps that generate manifests. |
| `generateConfigCmd` | The closure is registered as a callback on this command; the package’s CLI infrastructure will invoke it at the appropriate time. |

### Package context

The `config` package implements an interactive configuration wizard for CertSuite. Each menu option (e.g., *Namespaces*, *Operators*, *Pod Labels*) returns a closure that mutates the shared `certsuiteConfig`. The returned closure is executed later when generating the final YAML files, ensuring all user selections are captured.

In summary, `loadPodLabels` is a small helper that turns a slice of pod‑label strings into a deferred configuration step for the CertSuite generator.
