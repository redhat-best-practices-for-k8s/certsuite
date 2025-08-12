NewCommand` – Package‑Level Command Factory

## Purpose
`NewCommand` creates the top‑level **Cobra** command that powers the `certsuite run` subcommand.  
It bundles together all flag definitions and registers them on a shared global variable, `runCmd`, so that other parts of the application can access the parsed flag values.

> **Why a factory?**  
> The function encapsulates the boilerplate required to set up the command tree (flags, usage strings, etc.) and returns the ready‑to‑use `*cobra.Command`. This keeps the `main` package thin and allows unit tests to instantiate the command without starting the CLI.

---

## Signature
```go
func NewCommand() *cobra.Command
```
- **Input**: none.  
- **Output**: a pointer to the configured `cobra.Command`.

---

## Key Dependencies

| Dependency | Role |
|------------|------|
| `github.com/spf13/cobra` | Provides the command structure and flag helpers (`StringP`, `Bool`, etc.). |
| `runCmd` (global variable) | Holds the command instance so that other packages can reference it. |

---

## Flag Configuration
The function registers a mix of **persistent** and **local** flags:

- **Persistent Flags** – available to all subcommands of `certsuite run`.
  - Example: `StringP("config", "c", "", "Path to configuration file")`
- **Local Flags** – specific to this command.
  - Examples include boolean toggles (`Bool`) and string values.

The flags cover a wide range of runtime options (timeouts, logging levels, test selection, etc.).  
All flag names are passed through `StringP`, `String`, or `Bool` helpers, which internally bind the flag to the underlying Cobra command.

> **Note**: The exact list of flags is not shown here due to repetition in the call trace, but they follow the pattern above.

---

## Side Effects

1. **Global state mutation** – assigns the constructed `*cobra.Command` to the package‑wide variable `runCmd`.
2. **Flag registration** – modifies the command’s flag set; this is a one‑time side effect during program initialization.

No other external resources are touched (no file I/O, network calls, etc.).

---

## Integration with the Package

```
cmd/certsuite/run/
├── run.go          ← contains NewCommand
└── run_test.go     ← unit tests may call NewCommand()
```

`NewCommand` is the public entry point for the `run` package.  
Other packages (e.g., `main`) import `"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/run"` and invoke:

```go
rootCmd.AddCommand(run.NewCommand())
```

Thus, it stitches the `certsuite run` command into the overall CLI hierarchy.

---

## Suggested Mermaid Diagram

```mermaid
flowchart TD
    subgraph root "Root Command"
        A[root]
    end

    subgraph runpkg "run package"
        B[NewCommand()]
        C[runCmd (global)]
    end

    A -->|AddCommand| B
    B --> C
```

The diagram illustrates how `NewCommand` produces the command and stores it in `runCmd`, which is then integrated into the root command tree.

---
