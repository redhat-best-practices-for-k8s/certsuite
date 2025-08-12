NewCommand` – Command Builder for Claim Comparison

| Item | Detail |
|------|--------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare` |
| **Signature** | `func NewCommand() *cobra.Command` |
| **Exported** | Yes (capitalized) |

### Purpose
Creates and configures the **CLI sub‑command** that performs a diff between two claim files.  
The command is later attached to the top‑level `certsuite` binary, enabling usage like:

```bash
certsuite claim compare --claim1 path/to/first.json --claim2 path/to/second.json
```

### How It Works

1. **Instantiate** a new `cobra.Command` with:
   * `Use: "compare"`
   * `Short`: brief description (omitted in snippet but normally present)
   * `Long`: full help text from the package‑level constant `longHelp`.

2. **Define Flags**  
   Two string flags are added to the command’s flag set:

   ```go
   cmd.Flags().StringVarP(&Claim1FilePathFlag, "claim1", "", "", "path to first claim file")
   cmd.Flags().StringVarP(&Claim2FilePathFlag, "claim2", "", "", "path to second claim file")
   ```

   * `Claim1FilePathFlag` and `Claim2FilePathFlag` are exported package variables that hold the values supplied by the user.
   * The empty string for the shorthand indicates **no short flag**.

3. **Enforce Requirements**  
   Both flags are marked as required:

   ```go
   cmd.MarkFlagRequired("claim1")
   cmd.MarkFlagRequired("claim2")
   ```

4. **Return** the fully configured command instance.

### Key Dependencies

| Dependency | Role |
|------------|------|
| `github.com/spf13/cobra` | Provides the `Command`, flag helpers, and error handling used throughout the CLI. |
| Package‑level constants/variables (`longHelp`, `Claim1FilePathFlag`, `Claim2FilePathFlag`) | Supply help text and store parsed flag values. |

### Side Effects

* **Global state mutation** – The two exported string variables are updated when the command parses flags.
* **Error handling** – If a required flag is missing, Cobra will automatically print an error and exit; the function itself does not return errors.

### Integration in the Package

The `compare` package implements a sub‑command for the `claim` group.  
`NewCommand` is called from the parent command (likely in `cmd/certsuite/claim/claim.go`) to add this functionality to the overall CLI tree.  

```go
// Pseudocode in parent command:
certsuiteCmd.AddCommand(compare.NewCommand())
```

Thus, `NewCommand` acts as a **factory** that wires up user input handling and help text for claim comparison, while delegating actual diff logic to other functions (e.g., `claimCompareFiles`, which is invoked elsewhere when the command runs).
