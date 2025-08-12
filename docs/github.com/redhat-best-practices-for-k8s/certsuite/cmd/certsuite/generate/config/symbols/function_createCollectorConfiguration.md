createCollectorConfiguration`

```go
func createCollectorConfiguration() func()
```

| Item | Detail |
|------|--------|
| **Purpose** | Builds a *collector configuration* command that is added to the CLI’s “generate” sub‑command. The returned closure is executed when the user selects the “collector” option in the interactive menu. |
| **Inputs** | None – the function captures all required data from package globals (`certsuiteConfig`, `templates`) and the surrounding command context (`generateConfigCmd`). |
| **Output** | A zero‑argument function (the closure). When invoked it prints a configuration file for a Collector to stdout or writes it to disk, depending on how the CLI is used. |
| **Key operations inside the closure** | 1. **Template rendering** – uses `templates.Lookup("collector").Execute()` to fill a Go template with values from `certsuiteConfig`. <br>2. **String sanitisation** – runs several string replacements (`strings.ReplaceAll`, `strings.ToLower`) on the rendered text (e.g., normalising names). <br>3. **Validation** – checks that required fields are present using `strings.Contains`. <br>4. **Execution** – writes the final configuration to a file with permissions set by `defaultConfigFilePermissions` and prints progress via `fmt.Printf`. |
| **Dependencies** | *Standard library*: `fmt`, `os`, `strings`. <br>*Project internals*: `certsuiteConfig` (configuration struct), `templates` (parsed Go templates map). |
| **Side effects** | • Creates/overwrites a file named by the collector configuration. <br>• Prints status messages to stdout. <br>• No return value or error propagation – any failure is signalled only via console output. |
| **How it fits the package** | The `config` package implements a wizard‑style CLI for generating CertSuite configuration files. Each menu option (e.g., *collector*, *operator*, *service*) has an associated closure that performs the actual file generation. `createCollectorConfiguration` is the factory function that wires up the Collector section of this wizard, registering it with `generateConfigCmd`. |

---

### Suggested Mermaid diagram

```mermaid
flowchart TD
  A[User selects “collector”] --> B{createCollectorConfiguration()}
  B --> C[Render template from templates["collector"]]
  C --> D[Sanitise string (ReplaceAll/ToLower)]
  D --> E[Validate required fields (Contains)]
  E --> F[Write file with default permissions]
  F --> G[Print success message]
```

This diagram visualises the linear flow of the closure returned by `createCollectorConfiguration`.
