NewCommand` – CLI sub‑command constructor

| Aspect | Details |
|--------|---------|
| **Package** | `check` (path: `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/check`) |
| **Exported** | Yes (`func NewCommand() *cobra.Command`) |
| **Purpose** | Creates and returns the root Cobra command that implements the *certsuite check* sub‑command. It wires together all nested commands (e.g., `list`, `run`, etc.) so they can be invoked from the top‑level CLI. |

### Function signature

```go
func NewCommand() *cobra.Command
```

- **Input**: none.
- **Output**: a pointer to a fully populated `*cobra.Command` ready for registration with the parent command tree.

### Implementation walk‑through

1. **Declare a local variable `cmd`**  
   ```go
   cmd := &cobra.Command{
       Use:   "check",
       Short: "Run checks against certificates and configuration",
   }
   ```
   - This sets the base command name (`check`) and a short description.

2. **Add child commands** – The function calls `cmd.AddCommand(...)` repeatedly to register sub‑commands that live in the same package (e.g., `list`, `run`). Each call uses the constructor from another file:  
   ```go
   cmd.AddCommand(NewListCommand())
   cmd.AddCommand(NewRunCommand())
   ```

3. **Return** – The fully constructed command is returned to the caller.

### Key dependencies

| Dependency | Role |
|------------|------|
| `github.com/spf13/cobra` | Provides the `*cobra.Command` type and `AddCommand` method. |
| Sub‑command constructors (`NewListCommand`, `NewRunCommand`, …) | Provide the nested commands that are added to this root command. |

### Side effects

- **No global state mutation**: The function only creates local variables; it does not alter package globals or external state.
- **Registers child commands**: By calling `AddCommand`, it modifies the returned `cmd` object, which will later be integrated into the overall CLI tree by the caller.

### Package integration

In the larger `certsuite` binary, the root command (`certsuite`) imports this package and calls its `NewCommand()` to obtain the *check* sub‑command. The returned command is then added to the main command tree:

```go
rootCmd.AddCommand(check.NewCommand())
```

This design keeps each functional area (e.g., checking certificates, installing resources) in its own package while exposing a clean, modular CLI hierarchy.

### Suggested Mermaid diagram

```mermaid
graph TD;
    certsuite --> checkCmd[check.NewCommand()];
    subgraph Check Sub‑commands
        checkCmd --> listCmd[NewListCommand()]
        checkCmd --> runCmd[NewRunCommand()]
    end
```

This illustrates how the *certsuite* binary composes the `check` command and its children.
