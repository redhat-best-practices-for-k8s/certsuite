NewCommand` – Package *info*

| Item | Details |
|------|---------|
| **File** | `cmd/certsuite/info/info.go` (line 61) |
| **Signature** | `func NewCommand() *cobra.Command` |
| **Exported?** | ✅ Yes – public constructor for the *info* sub‑command |

## Purpose

`NewCommand` builds and returns a fully configured `*cobra.Command` that implements the `certsuite info` CLI command.  
The command displays static information about the CertSuite tool (e.g., version, author, license). It also accepts a `--verbose` flag to control the amount of output.

## Inputs & Outputs

| Parameter | Type | Description |
|-----------|------|-------------|
| *none* | – | The function receives no arguments. |
| **Return** | `*cobra.Command` | A pointer to a Cobra command ready for registration in the root command tree. |

## Core Logic

1. **Command Construction**  
   ```go
   cmd := &cobra.Command{
       Use:   "info",
       Short: "Show CertSuite info",
       RunE:  runInfo,
   }
   ```
   * `Use` and `Short` provide the help text.  
   * The actual action is delegated to a helper called `runInfo` (not shown in the snippet).

2. **Persistent Flags** – Two flags are added to every sub‑command that inherits this command:
   * `--output, -o` (`StringP`) – chooses output format (`json`, `yaml`, …).  
   * `--verbose, -v` (`BoolP`) – toggles detailed logging.

3. **Required Flag** – The `output` flag is marked as required via `MarkPersistentFlagRequired`.

4. **Help Formatting** – A custom help template is printed using `fmt.Fprintf`.  
   This prints the command’s description with a maximum line width of `lineMaxWidth` (constant defined elsewhere) and left‑justified padding `linePadding`.

5. **Return** – The fully built `cmd` is returned to be attached to the root command.

## Dependencies

| Dependency | Role |
|------------|------|
| `github.com/spf13/cobra` | Provides the `Command` struct, flag helpers (`StringP`, `BoolP`) and persistent flag API. |
| `fmt.Fprintf` | Writes formatted help text to the command’s output stream. |

## Side Effects

* Adds two persistent flags to the command; these flags are inherited by all child commands.
* Prints a custom help message when invoked with `--help`.
* No global state is mutated; it only constructs and returns a new object.

## How It Fits the Package

The *info* package contains all logic for the `certsuite info` sub‑command.  
- **`NewCommand`** builds the command instance.  
- The returned command is registered by the root command (likely in `cmd/certsuite/root.go`).  
- Other files in the package provide helper functions (`runInfo`, flag handlers, etc.) that implement the actual information display logic.

### Mermaid Flow

```mermaid
flowchart TD
    A[NewCommand] --> B{Build cobra.Command}
    B --> C[Set Use & Short]
    B --> D[Define RunE = runInfo]
    D --> E[Add persistent flags]
    E --> F[Mark output flag required]
    F --> G[Print custom help (Fprintf)]
    G --> H[Return cmd]
```

## Summary

`NewCommand` is a straightforward constructor that creates the `certsuite info` command, attaches its flags, ensures proper usage enforcement, and prepares it for integration into the main CLI tree. It leverages Cobra’s flag mechanisms and outputs a tailored help message while remaining free of side effects beyond object construction.
