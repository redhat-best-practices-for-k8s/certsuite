## Package main (github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite)

# certsuite – Command‑line entry point

`cmd/certsuite/main.go` is the sole executable package of the **certsuite** tool.  
It wires together a set of sub‑commands (check, claim, generate, info, run,
upload, version) and hands control over to Cobra’s command dispatcher.

---

## Core functions

| Function | Purpose | Key interactions |
|----------|---------|------------------|
| `newRootCmd() *cobra.Command` | Builds the top‑level command tree. | Calls `NewCommand()` from each sub‑package (`check`, `claim`, …) and registers them via `AddCommand`. |
| `main()` | Entry point for the binary. | Instantiates the root command, executes it with `Execute()`, logs any errors (`log.Error`) and exits with `os.Exit(1)` on failure. |

---

## How the command tree is built

```mermaid
graph TD;
  main --> newRootCmd
  newRootCmd -->|AddCommand| check.NewCommand()
  newRootCmd --> claim.NewCommand()
  newRootCmd --> generate.NewCommand()
  newRootCmd --> info.NewCommand()
  newRootCmd --> run.NewCommand()
  newRootCmd --> upload.NewCommand()
  newRootCmd --> version.NewCommand()
```

Each `NewCommand()` returns a fully configured `*cobra.Command` that knows how
to parse its own flags and perform the requested action.  
The root command simply acts as a dispatcher; all heavy lifting happens in
those sub‑packages.

---

## Error handling

```go
if err := cmd.Execute(); err != nil {
    log.Error(err)
    os.Exit(1)
}
```

* `log` is an internal wrapper around the standard logger, providing a unified
  error output format.  
* On any failure during command parsing or execution the program terminates with
  status code **1**.

---

## Summary

- The package contains only two functions: `main()` and `newRootCmd()`.  
- `newRootCmd` constructs a Cobra root command and attaches sub‑commands from
  dedicated packages.  
- `main` runs the command tree, logs errors, and exits appropriately.

This design keeps the executable thin; all domain logic lives in the
sub‑packages, making the CLI straightforward to maintain and extend.

### Functions


### Call graph (exported symbols, partial)

```mermaid
graph LR
```

### Symbol docs

