NewCommand` – CLI constructor for the results‑spreadsheet sub‑command

| Item | Detail |
|------|--------|
| **Package** | `resultsspreadsheet` (path: `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/upload/results_spreadsheet`) |
| **Exported?** | Yes (`func NewCommand() *cobra.Command`) |
| **Purpose** | Builds and returns a `*cobra.Command` that implements the `certsuite upload results-spreadsheet` CLI sub‑command. The command parses flags required to upload a test results spreadsheet to an OCP cluster. |

### Function signature
```go
func NewCommand() *cobra.Command
```
No parameters are passed; all configuration comes from global variables and flag bindings.

### Key flag definitions

| Flag | Variable bound | Type | Description |
|------|----------------|------|-------------|
| `--results-file` (`-r`) | `resultsFilePath` | string | Path to the local spreadsheet file that will be uploaded. **Required**. |
| `--root-folder-url` (`-u`) | `rootFolderURL` | string | URL of the root folder on the cluster where the spreadsheet should be stored. **Required**. |
| `--ocp-version` (`-o`) | `ocpVersion` | string | Target OpenShift cluster version. **Optional** – used when populating conclusion columns. |
| `--credentials` (`-c`) | `credentials` | string | Path to a kubeconfig or other credentials file used for authentication. **Required**. |

The command also defines a helper header array (`conclusionSheetHeaders`) that is later used by the upload routine.

### Dependencies & side effects

1. **Cobra package** – The function uses `cobra.Command` and its flag API (`Flags().StringVarP`, `MarkFlagRequired`).  
2. **Logging / error handling** – Calls `log.Fatalf` (via `Fatalf`) when a required flag is missing, causing the program to exit with an error message.  
3. **Global state mutation** – Each flag binding writes into package‑level variables (`resultsFilePath`, `rootFolderURL`, `ocpVersion`, `credentials`). These globals are later read by the command’s `Run` function (not shown in the snippet).  
4. **No I/O or network activity** – Construction is pure; all runtime effects happen when the returned command is executed.

### How it fits into the package

The `resultsspreadsheet` package implements a sub‑command that uploads test results to an OpenShift cluster. The typical flow in `certsuite upload` is:

```
certsuite upload results-spreadsheet [flags]
```

- `NewCommand()` creates the command object.
- The main `upload` command registers this object via `AddCommand`.
- When invoked, Cobra parses the flags (binding to globals) and runs the associated action defined elsewhere in the package.

Thus `NewCommand` is the entry point for configuring and exposing the spreadsheet upload functionality within the CLI tool.
