newRootCmd` – Package‑level command builder

**File:** `cmd/certsuite/main.go`  
**Location:** line 18  
**Package:** `main`

## Purpose
Creates and returns the *root* `cobra.Command` that drives the CertSuite CLI.  
The function wires together all sub‑commands (e.g., `run`, `verify`, etc.) so that a single `*cobra.Command`
can be passed to `cmd.Execute()` or used by tests.

```go
func newRootCmd() *cobra.Command {
    root := NewCommand(/* … */)

    // Register child commands
    root.AddCommand(NewCommand(...))
    root.AddCommand(NewCommand(...))
    …
    return root
}
```

> **Note:** The function is unexported because it is only needed internally by the `main` package to bootstrap the CLI.

## Inputs / Outputs

| Direction | Type      | Description |
|-----------|-----------|-------------|
| **Input** | *none*    | None – all configuration comes from global variables or environment. |
| **Output** | `*cobra.Command` | A fully‑initialized root command with all sub‑commands attached. |

## Key Dependencies

| Dependency | Role |
|------------|------|
| `github.com/spf13/cobra` | Provides the `Command`, `AddCommand`, and `NewCommand` functions used to construct CLI commands. |
| Sub‑command constructors (`NewCommand`) | Each call creates a specific child command (e.g., `run`, `verify`). The exact signatures are hidden but they return `*cobra.Command`. |

## Side Effects

- **No external state mutation:**  
  The function only allocates new objects; it does not alter global variables or write to files.
- **Command hierarchy creation** – by calling `AddCommand` the root command gains a tree structure that Cobra will traverse when executing user input.

## Package Integration

1. **Bootstrap**: In `main()`, the package calls `newRootCmd()` and passes the result to `cmd.Execute()`.
2. **Testing**: Tests can call `newRootCmd()` directly to obtain an isolated command instance, enabling unit tests of sub‑commands without starting a full CLI.
3. **Extensibility**: Adding new features only requires creating a new `NewCommand` implementation and adding it via `AddCommand`.

## Suggested Mermaid Diagram

```mermaid
graph TD;
  A[main] --> B[newRootCmd];
  B --> C[root *cobra.Command];
  subgraph Sub‑commands
    D1[cmd1] --> B;
    D2[cmd2] --> B;
    D3[cmd3] --> B;
  end
```

*This diagram visualizes the root command as a central node with multiple child commands attached, all orchestrated by `newRootCmd`.*

---
