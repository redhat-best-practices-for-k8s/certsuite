NewCommand` – Sub‑command factory for **certsuite**’s *check results* feature

| Item | Details |
|------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/check/results` |
| **Signature** | `func() *cobra.Command` |
| **Exported** | ✅ |

### Purpose
`NewCommand` constructs the *check‑results* sub‑command that is attached to the top‑level `certsuite check` command.  
It wires together all flags that control how test results are rendered and where they are stored, then returns a fully configured `*cobra.Command`. The returned command is later added to the CLI tree by the parent package.

### Inputs / Outputs
- **Input** – None (the function only reads package‑level state).
- **Output** – A pointer to a new `cobra.Command` instance ready for registration with the CLI.

The created command internally references the package‑scoped variable `checkResultsCmd`.  This global holds the last instance returned by `NewCommand`; it is used elsewhere in the package (e.g. for flag overrides or programmatic access).

### Flag configuration
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--template` (`-t`) | string | `TestResultsTemplateFileName` (default file name) | Path to a custom Go template used when rendering results. |
| `--output-dir` (`-o`) | string | current working directory | Directory where rendered output files are written. |
| `--json` (`-j`) | bool | `false` | When set, output is produced in JSON format instead of the default text/template rendering. |

The function also calls `MarkFlagsMutuallyExclusive` to ensure that `--template` and `--json` cannot be used together (the command must choose one representation).

### Key dependencies
- **Cobra** (`github.com/spf13/cobra`) – for command creation, flag registration, and mutual‑exclusion logic.
- The package’s constants (`TestResultsTemplateFileName`, `TestResultsTemplateFilePermissions`) are used as default values and file permission settings.

### Side effects
1. **Global state mutation** – assigns the newly created command to the package variable `checkResultsCmd`.
2. **No I/O** – all side‑effects happen at flag registration time; actual file or network operations occur when the command’s `Run` function is invoked elsewhere in the codebase.

### How it fits the package
The *results* sub‑command is one of several components under `certsuite check`.  
- The top‑level `check` command aggregates commands such as `run`, `list`, and this `results` command.  
- Once registered, users can run:  

  ```bash
  certsuite check results --template=my.tmpl --output-dir=/tmp/results
  ```

  or

  ```bash
  certsuite check results --json
  ```

  to produce formatted test result reports.

Below is a simplified Mermaid diagram showing the relationship:

```mermaid
graph LR
  A[certsuite] --> B[check]
  B --> C[results (NewCommand)]
  subgraph Flags
    D[--template] 
    E[--output-dir] 
    F[--json]
  end
  C --> D
  C --> E
  C --> F
```

---

**Summary**  
`NewCommand` is a factory that prepares the *check results* CLI command with its flags, enforces flag compatibility, stores the created command in a package‑level variable for later use, and returns it ready to be attached to the larger certsuite command hierarchy.
