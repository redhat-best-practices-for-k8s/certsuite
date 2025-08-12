NewCommand` – CLI “generate” sub‑command

**File:** `cmd/certsuite/generate/generate.go`  
**Package:** `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate`

---

## Purpose
`NewCommand` builds and returns the top‑level *cobra* command that implements the `generate` sub‑command of the CertSuite CLI.  
The function wires together a hierarchical set of child commands (e.g., `cert`, `config`, etc.) by repeatedly calling `AddCommand`.  The resulting command is what the main application registers with the root command.

---

## Signature
```go
func NewCommand() *cobra.Command
```
* **Returns** – a pointer to a fully‑configured `*cobra.Command` that can be added to the CLI tree.  
* **No inputs** – all configuration comes from the package’s internal setup (the `generate` variable defined in the same file).

---

## Key dependencies

| Dependency | Role |
|------------|------|
| `github.com/spf13/cobra` | Provides the command struct and helper methods (`AddCommand`, flags, run handlers). |
| `generate` (package‑level variable) | Holds sub‑command definitions that are added to the root `generate` command. The file declares it on line 12; its concrete type is inferred from usage in `NewCommand`. |

---

## Implementation outline

```go
func NewCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "generate",
        Short: "Generate certificates, configs, etc.",
    }

    // Add each sub‑command defined in the generate package.
    for _, subCmd := range generate {          // ← `generate` is a slice of *cobra.Command
        cmd.AddCommand(subCmd)                 // Attach child command
    }
    return cmd
}
```

* The function creates a new `cobra.Command` with basic usage text.  
* It iterates over the package‑level `generate` collection, adding each element as a sub‑command via `AddCommand`.  
* No other state is mutated; the function is purely functional.

---

## Side effects

* **No external side effects** – only constructs in‑memory objects.  
* The returned command may contain flags and run functions defined elsewhere (in the child commands).  

---

## Integration with the package

`NewCommand` is the public entry point for the `generate` package.  In the root of the CertSuite CLI (`cmd/certsuite/main.go` or equivalent), this function will be called to obtain the command tree:

```go
rootCmd.AddCommand(generate.NewCommand())
```

Thus, `NewCommand` acts as a bridge between the generic Cobra framework and the domain‑specific “generate” functionality implemented in this package.
