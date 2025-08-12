NewCommand` – Package *qecoverage*

| Item | Description |
|------|-------------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/qe_coverage` |
| **Exported?** | Yes |
| **Signature** | `func NewCommand() *cobra.Command` |

### Purpose
Creates the top‑level Cobra command used by the `certsuite generate qe-coverage` sub‑command.  
The command is responsible for wiring up the flags that control the generation of QE coverage reports and exposing a `RunE` handler (defined elsewhere in the package).  The function returns the fully configured command so it can be added to the main CLI tree.

### Inputs / Outputs
| Parameter | Type | Description |
|-----------|------|-------------|
| *none* | – | No arguments are accepted. |

| Return value | Type | Description |
|--------------|------|-------------|
| `*cobra.Command` | Pointer to a Cobra command struct | The fully initialized command, ready for registration with the root command. |

### Key Dependencies
- **Cobra**: Uses `github.com/spf13/cobra`.  
  - Calls `PersistentFlags()` on the new command to attach flags that should be inherited by any sub‑commands.
  - Calls `String()` to create a string flag (`report-path`) used by the QE coverage generation logic.
- **Global Variable** – `qeCoverageReportCmd`: The variable holds the command instance after creation. It is referenced by other package files (e.g., the run handler) to access shared state or flags.

### Side Effects
1. Instantiates a new `*cobra.Command`.
2. Stores that instance in the unexported global `qeCoverageReportCmd`.
3. Registers two persistent string flags:
   - `--report-path` – path where the coverage report will be written.
4. Does **not** set any `Run` or `RunE` handler directly; those are attached elsewhere.

### Integration into the Package
The `qecoverage` package is part of the *certsuite generate* command hierarchy:

```
certsuite
└─ generate
   └─ qe-coverage  <-- this subcommand
```

`NewCommand` is called from the package’s init routine (or from the root command builder) to create and expose the QE coverage generator. The returned command becomes a child of the `generate` command, enabling users to run:

```bash
certsuite generate qe-coverage --report-path /tmp/report.json
```

### Suggested Mermaid Diagram

```mermaid
graph TD;
  A[Root CLI] --> B[Generate Command];
  B --> C[qecoverage.NewCommand()];
  C --> D[qeCoverageReportCmd (global)];
  D --> E["Flags: report-path"];
  C --> F[Run handler defined elsewhere];
```

This diagram shows the relationship between the root command, the `generate` sub‑command, and the QE coverage command created by `NewCommand`.
