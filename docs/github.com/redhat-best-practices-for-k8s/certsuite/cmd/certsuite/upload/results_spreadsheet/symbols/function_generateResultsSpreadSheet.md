generateResultsSpreadSheet` – Internal helper for the *results‑spreadsheet* command

| Aspect | Details |
|--------|---------|
| **Signature** | `func()()` – no parameters, no return value |
| **Visibility** | Unexported (private to package) |
| **Purpose** | Orchestrates the creation of a Google‑Sheets workbook that contains two sheets – *Raw Results* and *Conclusions*.  The spreadsheet is created in the user’s Google Drive under a folder derived from `rootFolderURL`.  Once finished, it is moved into that folder and basic filters/sorts are applied to the sheet. |

### High‑level flow

```mermaid
flowchart TD
    A[CreateDriveServices] --> B{Extract folder ID}
    B -->|error| C[log.Fatalf]
    B --> D[Create or reuse Drive folder]
    D --> E[Create spreadsheet (title + date)]
    E --> F[Add Raw Results sheet]
    F --> G[Add Conclusions sheet]
    G --> H[Move spreadsheet to folder]
    H --> I[Apply basic filter]
    I --> J[Apply descending sort on results]
```

1. **Drive & Sheets services** – `CreateSheetsAndDriveServices()` returns the two API clients needed for all subsequent calls.
2. **Folder resolution** – The command line flag `rootFolderURL` is parsed by `extractFolderIDFromURL`.  If it fails, the program aborts (`log.Fatalf`).  
   *If the folder does not exist*, `createDriveFolder()` creates a new one named after the current timestamp and the `ocpVersion`.
3. **Spreadsheet creation** – A fresh spreadsheet titled `"Results – <timestamp>"` is created with the `Create()` method of the Sheets API.
4. **Sheet population**  
   * `createRawResultsSheet`: adds the “Raw Results” sheet, writes header row (`rawResultsHeaders`) and appends data from the local JSON file referenced by `resultsFilePath`.  
   * `createConclusionsSheet`: adds the “Conclusions” sheet with its own headers (`conclusionSheetHeaders`).
5. **Move to folder** – The spreadsheet is moved into the target Drive folder via `MoveSpreadSheetToFolder`.
6. **Post‑processing** – Two helper functions add a simple filter and sort the data in descending order on the “Raw Results” sheet.

### Dependencies & side effects

| Dependency | Role |
|------------|------|
| `credentials` | Path to Google OAuth 2.0 credentials JSON; used by `CreateSheetsAndDriveServices`. |
| `rootFolderURL`, `ocpVersion` | Determine where in Drive the spreadsheet will live and what name it receives. |
| `resultsFilePath` | File containing raw test results that are imported into the spreadsheet. |

All API interactions use **panic‑style error handling** (`log.Fatalf`).  The function never returns; on any failure it terminates the program with an error message.

### How it fits the package

The *resultsspreadsheet* command is a CLI tool for uploading test results to Google Sheets.  
`generateResultsSpreadSheet` contains the heavy lifting:

* It hides the details of Drive/Sheets API calls from the rest of the codebase.
* The main command (`uploadResultSpreadSheetCmd`) simply parses flags, sets globals, and then calls this function.
* Because it has no return value, callers do not need to handle errors – they are already logged and fatal.

This design keeps the public API minimal while ensuring that all spreadsheet‑related logic is concentrated in a single, well‑documented place.
