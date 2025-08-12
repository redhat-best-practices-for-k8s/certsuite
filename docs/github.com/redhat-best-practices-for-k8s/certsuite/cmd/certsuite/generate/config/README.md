## Package config (github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/config)

# Certsuite – Configuration Generator CLI

The *generate/config* package implements the `certsuite generate config` command, which walks a user through an interactive wizard to build a YAML configuration for the Certsuite test harness.

Below is a concise, structured view of the key pieces – data structures, global state, and functions – and how they fit together.  
All code is read‑only; no modification is performed.

---

## 1. Core Data Structure

| Type | Purpose |
|------|---------|
| **`configOption`** (unexported) | Holds a single menu option for the wizard. Two fields: `Help` – a short description shown in the prompt, and `Option` – the value that will be stored in the configuration when chosen. |

> The struct is used only inside the interactive prompts; it is not exported to other packages.

---

## 2. Global State

| Variable | Type (inferred) | Role |
|----------|-----------------|------|
| **`certsuiteConfig`** | `*configuration.TestConfiguration` | Holds the configuration being built in memory while the wizard runs. It is populated by successive prompt functions and finally written to disk. |
| **`generateConfigCmd`** | `*cobra.Command` | The Cobra command object that registers the wizard under `certsuite generate config`. Its `RunE` field calls `generateConfig()`. |
| **`templates`** | `map[string]string` | A map of prompts (key) → example syntax (value). It is used to provide context‑aware help strings during the wizard. |

> All globals are private; they exist only in this package and are initialized in `init()`.

---

## 3. Command Creation

```go
func NewCommand() *cobra.Command {
    return generateConfigCmd
}
```

`NewCommand` is exported so other parts of Certsuite can embed the config generator into the CLI tree.  
The command itself is built at init time:

```go
generateConfigCmd = &cobra.Command{
    Use:   "config",
    Short: "Generate a configuration file for certsuite",
    RunE:  func(_ *cobra.Command, _ []string) error { generateConfig(); return nil },
}
```

---

## 4. The Wizard Flow

`generateConfig()` orchestrates the wizard:

1. **Run** – starts the interactive prompts (`createConfiguration()`).  
2. **Show** – prints the current configuration to stdout (`showConfiguration()`).  
3. **Save** – writes the config to disk (`saveConfiguration()`).

### 4.1. Building the Configuration

`createConfiguration()` calls three sub‑routines in order:

| Function | Responsibility |
|----------|----------------|
| `createCertSuiteResourcesConfiguration()` | Prompts for namespaces, labels, CRD filters, Helm charts, etc. |
| `createExceptionsConfiguration()` | Gathers exception lists: kernel taints, protocols, services, non‑scalable objects. |
| `createSettingsConfiguration()` | Collects runtime settings such as the probe DaemonSet namespace. |

Each sub‑routine follows a similar pattern:

1. **Prompt** – use `promptui.Select` or `promptui.Prompt`.  
2. **Process answer** – helper functions (`getAnswer`, `load*`) parse comma‑separated strings into slices.  
3. **Store** – append to the corresponding field in `certsuiteConfig`.

#### Example: Loading Namespaces

```go
func loadNamespaces(names []string) {
    for _, ns := range names {
        certsuiteConfig.Namespaces = append(certsuiteConfig.Namespaces, strings.TrimSpace(ns))
    }
}
```

The helper `getAnswer(prompt, example, syntax)` displays the prompt with colored help and returns a slice of trimmed answers.

---

## 5. Utility Functions

| Function | Key Operations |
|----------|----------------|
| **`saveConfiguration(cfg *configuration.TestConfiguration)`** | Marshals to YAML, writes to `<defaultConfigFileName>` with permissions `0644`. Uses `log.Printf` for feedback. |
| **`showConfiguration(cfg *configuration.TestConfiguration)`** | Pretty‑prints the marshalled YAML to stdout. |
| **`getAnswer()`** | Handles user input; supports multi‑entry comma separation, trimming, and echoing back the parsed list. |

Other helpers (`loadCRDfilters`, `loadHelmCharts`, etc.) perform minimal parsing (e.g., splitting on commas, converting `"true"`/`"false"` strings to booleans).

---

## 6. Constants & Prompt Metadata

The package contains a large `const.go` with:

* **Menu names** – e.g., `create`, `exceptions`, `settings`.  
* **Help text** – user‑friendly explanations shown by the wizard.  
* **Example syntax** – used in `templates` to show how to format answers.

These constants are referenced throughout the prompts to keep UI consistent.

---

## 7. Interaction Flow Diagram (Mermaid)

```mermaid
flowchart TD
    A[generateConfig()] --> B[createConfiguration()]
    B --> C[createCertSuiteResourcesConfiguration()]
    C --> D{Prompt user}
    D --> E[getAnswer() -> parse]
    E --> F[append to certsuiteConfig]
    B --> G[createExceptionsConfiguration()]
    G --> H{Prompt user}
    H --> I[getAnswer() -> parse]
    I --> J[append to certsuiteConfig]
    B --> K[createSettingsConfiguration()]
    K --> L{Prompt user}
    L --> M[getAnswer() -> parse]
    M --> N[append to certsuiteConfig]
    A --> O[showConfiguration()]
    A --> P[saveConfiguration()]
```

---

## 8. Summary

* The package implements an **interactive CLI wizard** for building a Certsuite configuration file.  
* All state lives in the private `certsuiteConfig` variable, populated by a series of prompt helpers.  
* After completion, the config is displayed and written to disk.  
* The code is tightly coupled: each helper function directly mutates `certsuiteConfig`; there are no abstractions beyond the wizard flow.

This overview should help new contributors understand how the configuration generation process is wired together without diving into every line of boilerplate prompt handling.

### Structs

- **configOption**  — 2 fields, 0 methods

### Functions

- **NewCommand** — func()(*cobra.Command)

### Globals


### Call graph (exported symbols, partial)

```mermaid
graph LR
```

### Symbol docs

- [function NewCommand](symbols/function_NewCommand.md)
