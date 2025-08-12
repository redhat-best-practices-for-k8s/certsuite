NewCommand` – CLI helper for *show failures*

### Purpose
Creates and configures the **cobra** command that powers the
`certsuite claim show failures` sub‑command.  
The command reads three flags:

| Flag | Variable | Description |
|------|----------|-------------|
| `--claim-file` (`-f`) | `claimFilePathFlag` | Path to a claim file containing test results. |
| `--output-format` (`-o`) | `outputFormatFlag` | Desired output format (`text`, `json`). |
| `--test-suites` (`-t`) | `testSuitesFlag` | Comma‑separated list of test suite names to filter on. |

The command validates the flag values, ensures required flags are present,
and sets up a handler that will read the claim file and print the requested
failure data.

### Signature

```go
func NewCommand() *cobra.Command
```

- **Input** – none (relies solely on package‑level variables).
- **Output** – a fully configured `*cobra.Command` ready to be added to the
  root command tree.

### Key Dependencies & Side Effects

| Dependency | Role |
|------------|------|
| `github.com/spf13/cobra` | Provides the `Command`, flag helpers (`StringVarP`, `Flags`) and error handling (`MarkFlagRequired`). |
| `log.Fatalf` | Prints a fatal message if required flags are missing. |
| `fmt.Sprintf` | Used to format the list of available output formats in the help text. |

The function **does not** execute any business logic; it merely wires up
flags and basic validation. The actual execution occurs in the command’s
`RunE` closure, defined elsewhere in the package.

### Flow Overview

```mermaid
flowchart TD
    A[Start] --> B{Create cobra.Command}
    B --> C[StringVarP for claim-file]
    C --> D[StringVarP for output-format]
    D --> E[Set valid formats (text,json)]
    E --> F[StringVarP for test-suites]
    F --> G[Mark flags required]
    G --> H[Return command]
```

### Integration with the Package

- The global `showFailuresCommand` variable holds the returned value.
- Other parts of the CLI register this command under `claim show failures`.
- When invoked, it delegates to the execution logic that reads
  `claimFilePathFlag`, filters by `testSuitesFlag`, and prints in the chosen
  `outputFormatFlag`.

---

**Bottom line:**  
`NewCommand` is the entry point for the *show failures* feature, exposing
three user‑configurable flags, enforcing required arguments, and preparing a
cobra command that will later perform the failure reporting.
