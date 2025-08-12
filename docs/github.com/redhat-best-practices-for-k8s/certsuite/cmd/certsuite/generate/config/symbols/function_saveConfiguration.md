saveConfiguration`

```go
func saveConfiguration(cfg *configuration.TestConfiguration) func()
```

### Purpose
`saveConfiguration` is a helper that serialises the current in‑memory test configuration (`cfg`) to a YAML file, writes that file to disk, and prints a colourful confirmation message.  
The function returns an empty closure so it can be used as a Cobra command action or passed around without executing immediately.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `cfg` | `*configuration.TestConfiguration` | The configuration object to persist. It is the same struct that backs the interactive CLI menu; the caller passes the current state before invoking this helper. |

### Return value
A zero‑argument function that, when called, performs the write operation.  
This design allows callers to defer execution (e.g., as a Cobra `Run` handler) while keeping the logic encapsulated.

### Key Steps

| Step | Code |
|------|------|
| 1️⃣ Marshal to YAML | `data, err := yaml.Marshal(cfg)` |
| 2️⃣ Log intent | `fmt.Printf("Generating config file %s... ", cfg.ConfigFileName)` |
| 3️⃣ Run templating (if needed) | `templates.ExecuteTemplate(os.Stdout, "config", cfg)` – this prints a preview of the generated config to stdout. |
| 4️⃣ Write file | `err = os.WriteFile(cfg.ConfigFileName, data, defaultConfigFilePermissions)` |
| 5️⃣ Success message | `fmt.Println(GreenString("OK"))` |

### Dependencies & Globals

| Dependency | Role |
|------------|------|
| `yaml.Marshal` | Serialises the struct to YAML. |
| `fmt.Printf/Println` | Provides console output. |
| `os.WriteFile` | Persists data to disk. |
| `templates.ExecuteTemplate` | Renders a Go template for previewing the config (uses the package‑level `templates` variable). |
| `defaultConfigFilePermissions` | File permission bits used when creating the file. |

### Side Effects

* Writes a new or overwrites an existing file named in `cfg.ConfigFileName`.  
* Prints progress and success messages to stdout, including coloured text via `GreenString`.

### How It Fits the Package

The `config` package implements an interactive configuration wizard for CertSuite.  
During the wizard, after the user finalises options, the command invokes this helper to persist the choices.  
By returning a closure, the function can be plugged into Cobra’s `RunE` field or other contexts that expect a `func()` signature.

---

#### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[User completes wizard] --> B{Save config?}
    B -- Yes --> C[Call saveConfiguration(cfg)]
    C --> D1[Marshal to YAML]
    C --> D2[Execute template preview]
    C --> D3[Write file to disk]
    C --> D4[Print success message]
```

This diagram visualises the flow from wizard completion to persistence.
