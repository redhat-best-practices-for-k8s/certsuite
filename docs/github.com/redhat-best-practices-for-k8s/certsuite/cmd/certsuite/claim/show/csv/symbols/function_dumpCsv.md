dumpCsv` – CSV Export Sub‑Command

### Purpose
The `dumpCsv` function implements the *export* action of the `certsuite claim show csv` command.  
It collects claims for one or more CNFs, builds a catalog in memory and then writes that catalog to standard output as a CSV file.

> **Use case** – When a user wants to inspect all claims associated with a set of CNFs (identified by name, ID list or a file containing IDs) in plain‑text format suitable for downstream tooling.

### Signature
```go
func dumpCsv(cmd *cobra.Command, args []string) error
```
* `cmd` – The Cobra command instance that triggered the function.  
  Used only to set its output stream (`SetOutput`) and for flag access via `cmd.Flags()`.
* `args` – Positional arguments supplied on the CLI (unused in this implementation).

### Workflow & Key Dependencies

| Step | Action | Called Function(s) | Notes |
|------|--------|--------------------|-------|
| 1 | Set command output to `os.Stdout` | `cmd.SetOutput(os.Stdout)` | Guarantees CSV is printed to console. |
| 2 | Parse command‑line flags | `Parse(cmd.Flags())` | Loads flag values into the package variables (`CNFNameFlag`, `CNFListFilePathFlag`, etc.). |
| 3 | Validate claim file exists & version | `CheckVersion(claimFilePathFlag)` | Ensures the claim data file is readable and compatible. |
| 4 | Load CNF ID → type mapping | `loadCNFTypeMap()` | Populates internal map used when building the catalog. |
| 5 | Build catalog from claim IDs | `buildCatalogByID(cnfIDs, addHeaderFlag)` | Returns a slice of `Claim` structs that will be turned into CSV rows. |
| 6 | Convert catalog to CSV data | `buildCSV(catalog)` | Serialises claims into a byte buffer (`[]byte`). |
| 7 | Write CSV to stdout | `NewWriter(os.Stdout).WriteAll(csvBytes)` | Handles any write errors. |
| 8 | Flush writer & report status | `Flush()`, `Error()` | Ensures all data is sent and reports I/O errors. |

### Inputs / Outputs

* **Inputs** – Flags:
  * `--claim-file` (`claimFilePathFlag`) – Path to the claims JSON file.
  * `--cnf-name` (`CNFNameFlag`) – Name of a single CNF to export.
  * `--cnf-list-file` (`CNFListFilePathFlag`) – File containing a list of CNF IDs.
  * `--add-header` (`addHeaderFlag`) – Whether the CSV should include a header row.

* **Outputs** – The function writes a CSV representation of all matching claims to standard output.  
  On success it returns `nil`; on failure it returns an error that will be printed by Cobra and cause a non‑zero exit status.

### Side Effects & Error Handling

| Source | Effect | Mitigation |
|--------|--------|------------|
| `Fatalf` calls | Terminate the process immediately with an error message. | Used for unrecoverable failures (e.g., failed file read). |
| `Errorf` / returned errors | Propagate to Cobra which prints them and exits with status 1. | Allows callers to handle or log errors uniformly. |
| Writer operations (`WriteAll`, `Flush`) | May return I/O errors. | Checked and wrapped into a fatal error if unrecoverable. |

### Integration in the Package

* The command is defined by `CSVDumpCommand` (a `cobra.Command`).  
  `dumpCsv` is set as its `RunE` handler.
* It relies on other helpers (`loadCNFTypeMap`, `buildCatalogByID`, `buildCSV`) that are also part of the same package, ensuring a clear separation between *data loading*, *catalog construction*, and *serialization*.

### Suggested Mermaid Diagram

```mermaid
graph TD;
    A[User runs certsuite claim show csv] --> B[dumpCsv];
    B --> C[Parse flags];
    C --> D{CNF selection?};
    D -->|Name| E[Single CNF];
    D -->|List file| F[List of IDs];
    D -->|None| G[Error];
    E & F --> H[CheckVersion];
    H --> I[loadCNFTypeMap];
    I --> J[buildCatalogByID];
    J --> K[buildCSV];
    K --> L[NewWriter(os.Stdout)];
    L --> M[WriteAll];
    M --> N[Flush];
    N --> O[Success/Print CSV];
```

This diagram illustrates the linear flow of data from command line input to CSV output, highlighting the role of each helper function.
