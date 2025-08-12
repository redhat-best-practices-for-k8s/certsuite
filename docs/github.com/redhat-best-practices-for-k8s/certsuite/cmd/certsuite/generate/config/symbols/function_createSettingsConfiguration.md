createSettingsConfiguration`

| Item | Detail |
|------|--------|
| **Package** | `config` – part of the CertSuite generator command tree (`cmd/certsuite/generate/config`). |
| **Signature** | `func()()` – returns a closure that, when called, performs interactive configuration. |

### Purpose
`createSettingsConfiguration` builds an **interactive menu** for configuring *global* settings used by CertSuite.  
The returned function is invoked as part of the command‑line interface to prompt the user for values such as:

- The namespace where the probe daemon set will be installed (`probeDaemonSetNamespace`).  
- Other generic settings (currently only the namespace; more may be added later).  

After collecting answers it updates the in‑memory `certsuiteConfig` structure, which is later written out to a YAML file by `saveConfigHelp`.

### Workflow
1. **Print heading** – informs the user that they are configuring global settings.  
2. **Prompt for probe daemon set namespace**  
   - Calls `loadProbeDaemonSetNamespace()` to provide a default value (usually the current Kubernetes context’s namespace).  
   - Uses `getAnswer()` to display the prompt and read the user's input, allowing the default if the user presses *Enter*.  
3. **Update configuration** – assigns the returned string to `certsuiteConfig.ProbeDaemonSetNamespace`.  
4. **Return nil** – the closure has no return value; side effects are the updated global config.

### Dependencies & Side‑Effects
| Dependency | Role |
|------------|------|
| `Run` (from the Cobra command framework) | Executes this function as a sub‑command. |
| `Printf` | Displays instructions and prompts. |
| `loadProbeDaemonSetNamespace` | Provides the default namespace value. |
| `getAnswer` | Handles user input with validation/echo suppression. |

Side effects are limited to mutating the global variable `certsuiteConfig`. No files are written until `saveConfigHelp` is invoked elsewhere.

### Relation to the Package
- The `config` package defines a Cobra command tree for generating CertSuite configurations.  
- `createSettingsConfiguration` is one of several *menu* functions (e.g., `createResourcesConfiguration`, `createCRDFiltersConfiguration`) that collectively build the final YAML.  
- It plugs into the overall flow by being attached to the `generateConfigCmd` command via `Run`.

### Example Usage
```go
// In generate.go
cmd := &cobra.Command{
    Use:   "settings",
    Short: "Configure global settings",
    Run:   createSettingsConfiguration(),
}
generateConfigCmd.AddCommand(cmd)
```

When a user runs:

```
certsuite generate config settings
```

they will be prompted for the probe daemon set namespace, which is then stored in `certsuiteConfig` and later persisted.
