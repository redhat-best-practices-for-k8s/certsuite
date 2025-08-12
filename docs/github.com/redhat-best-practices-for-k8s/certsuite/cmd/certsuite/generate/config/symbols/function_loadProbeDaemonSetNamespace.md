loadProbeDaemonSetNamespace`

```go
func loadProbeDaemonSetNamespace(args []string) func()
```

### Purpose

`loadProbeDaemonSetNamespace` is a **menu‑action factory** used by the
interactive configuration tool in the *generate* command of CertSuite.
It builds and returns a closure that, when executed, reads the namespace
for the Probe DaemonSet from the user (or from defaults) and stores it
in the global `certsuiteConfig` structure.

### Inputs

| Parameter | Type     | Description |
|-----------|----------|-------------|
| `args`    | `[]string` | A slice of command‑line arguments supplied by the CLI menu system.  
  In practice the slice is usually empty because the action only needs
  to prompt the user; it is kept for API consistency with other actions.

### Output

A **zero‑argument function** (`func()`) that performs the actual work:

1. Prompts the user for a Kubernetes namespace (via `templates` or
   similar helper).
2. Validates the input (non‑empty, no illegal characters, etc.).
3. Stores the resulting value in `certsuiteConfig.ProbeDaemonSet.Namespace`.

The returned function is executed by the menu dispatcher when the
user selects *“Probe DaemonSet Namespace”*.

### Dependencies & Side Effects

| Dependency | Role |
|------------|------|
| `certsuiteConfig` (global) | Holds all configuration values; this function writes to its `ProbeDaemonSet.Namespace` field. |
| `templates` (global) | Provides the interactive prompt/validation utilities used by the closure. |
| `generateConfigCmd` (global) | Not directly used in the snippet, but part of the surrounding command context that registers the action. |

The function has no external side effects beyond updating the global
configuration and possibly printing to stdout/stderr via the prompting
helpers.

### Relationship to Package

- Located in **`github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/config`**.
- The package implements an interactive CLI for generating CertSuite
  configuration files.  
- `loadProbeDaemonSetNamespace` is one of many similar “load” functions
  (e.g., `loadNamespace`, `loadOperators`) that populate different
  sections of the configuration based on user input.

### Example Flow

```go
// Inside menu registration:
generateConfigCmd.Flags().StringVar(&certsuiteConfig.ProbeDaemonSet.Namespace,
    "probe-daemonset-namespace", "", "Namespace for Probe DaemonSet")

action := loadProbeDaemonSetNamespace(nil)
action() // User is prompted, result stored in certsuiteConfig
```

The returned closure encapsulates the logic so that the menu system can
invoke it without caring about the internal prompt/validation code.

--- 

**Key takeaway:**  
`loadProbeDaemonSetNamespace` creates a reusable action for editing the
namespace of the Probe DaemonSet within the interactive configuration
workflow, wiring user input into the global `certsuiteConfig`.
