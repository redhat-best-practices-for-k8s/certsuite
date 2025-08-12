NewCommand` – package‑level command constructor

| Item | Description |
|------|-------------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim` |
| **Signature** | `func NewCommand() *cobra.Command` |
| **Exported?** | Yes |

### Purpose
`NewCommand` builds the root Cobra command that represents the *certsuite claim* sub‑command.  
It is used by the main application (`cmd/certsuite`) to register the “claim” functionality in the CLI tree.

### Inputs / Outputs
- **Inputs**: none (no parameters).
- **Output**: a fully configured `*cobra.Command` instance that can be attached to a parent command with `AddCommand`.

### Key Behavior
1. **Lazy initialization of `claimCommand`.**  
   The package declares an unexported variable:

   ```go
   var claimCommand *cobra.Command
   ```

   `NewCommand` checks whether this variable is already initialized; if not, it creates a new `cobra.Command`, configures its flags and sub‑commands (by calling other constructors in the same package), and assigns the result to `claimCommand`. Subsequent calls simply return the cached instance.

2. **Sub‑command wiring**  
   Inside the constructor the function typically performs:

   ```go
   claimCommand = &cobra.Command{
       Use:   "claim",
       Short: "...",
       RunE:  runClaim,
   }
   claimCommand.AddCommand(NewValidateCommand())
   claimCommand.AddCommand(NewListCommand())
   ```

   Each call to `NewXCommand()` is a *function* that returns another `*cobra.Command` for a nested sub‑command (e.g., `validate`, `list`). These are added via `AddCommand`.

3. **Side effects**  
   - Mutates the package‑level `claimCommand` variable.  
   - No external state is modified beyond this cache; all other side effects happen when the returned command executes its `RunE` handler.

### Dependencies
- **Cobra library** (`github.com/spf13/cobra`) – for command construction and flag handling.
- Other internal constructors in the same package (`NewValidateCommand`, `NewListCommand`, etc.) – used to assemble sub‑commands.
- No global variables outside this file are touched; only `claimCommand` is written.

### How It Fits the Package
The `claim` package implements a group of CLI commands that operate on *claims* (likely Kubernetes admission or compliance claims).  
`NewCommand` is the entry point for exposing this functionality to the top‑level application. Once called, it returns a command tree that can be attached to the main root (`certsuite`) command via:

```go
rootCmd.AddCommand(claim.NewCommand())
```

Subsequent invocations of `NewCommand()` are idempotent because the result is cached in `claimCommand`. This pattern keeps construction logic centralized and avoids repeated allocation or flag re‑definition.
