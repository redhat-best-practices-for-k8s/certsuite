NewCommand` – CSV Dump Sub‑command

| Item | Value |
|------|-------|
| **Package** | `csv` (`github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/show/csv`) |
| **Exported?** | ✅ |
| **Signature** | `func NewCommand() *cobra.Command` |

---

## Purpose

`NewCommand` builds the **`csv` sub‑command** for the CertSuite CLI.  
The command produces a CSV representation of claim data (or CNF list) and writes it to standard output or a file.

```bash
certsuite claim show csv --claim-file=... [--cnf-name=...] [--cnf-list-file=...] [--add-header]
```

---

## How It Works

1. **Create the command**  
   ```go
   cmd := &cobra.Command{Use: "csv", Short: "..."}
   ```

2. **Declare flags**  
   - `claimFilePathFlag` (`string`) – path to a claim file (required).  
   - `CNFNameFlag` (`string`) – optional CNF name filter.  
   - `CNFListFilePathFlag` (`string`) – optional file containing CNFs list.  
   - `addHeaderFlag` (`bool`) – whether to prefix the CSV with a header row.

3. **Mark required flags**  
   Flags that are mandatory for the command to run are marked with `MarkFlagRequired`.  
   If marking fails, the program aborts via `log.Fatalf`.

4. **Return the configured command**  
   The returned `*cobra.Command` is later added to the root CLI tree.

---

## Inputs & Outputs

| Parameter | Type | Description |
|-----------|------|-------------|
| *none* | – | The function takes no runtime arguments. |

| Return | Type | Description |
|--------|------|-------------|
| `*cobra.Command` | pointer to a Cobra command object | Represents the fully configured CSV sub‑command ready for registration. |

---

## Key Dependencies

| Dependency | Role |
|------------|------|
| `github.com/spf13/cobra` | CLI framework; provides `Command`, flag helpers, and error handling. |
| Standard library (`log`) | Used to abort on flag‑registration errors with `Fatalf`. |

The function itself does not perform any CSV generation; it only sets up the command interface. The actual logic resides in the command’s `Run` or `RunE` closure (not shown here).

---

## Side Effects

- **Global state mutation**: Sets the values of package‑level flag variables (`claimFilePathFlag`, `CNFNameFlag`, `CNFListFilePathFlag`, `addHeaderFlag`) via Cobra’s `StringVarP` / `BoolVarP`.  
- **Process termination**: If any flag cannot be marked as required, the function calls `log.Fatalf`, terminating the program.

---

## Package Context

The `csv` package lives under:

```
cmd/certsuite/claim/show/csv
```

It is one of several “show” sub‑commands (e.g., `json`, `yaml`).  
Each command follows the same pattern: a constructor that registers flags and returns a Cobra command, while actual data processing occurs elsewhere in the package.

---

## Suggested Mermaid Diagram

```mermaid
graph TD
  A[NewCommand] --> B[Create *cobra.Command]
  B --> C[Define flags (claimFilePathFlag, CNFNameFlag, CNFListFilePathFlag, addHeaderFlag)]
  C --> D[Mark required flags]
  D --> E[Return command]
```

This diagram illustrates the sequential setup steps performed by `NewCommand`.
