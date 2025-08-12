CreateSheetsAndDriveServices`

**Package:** `resultsspreadsheet` – `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/upload/results_spreadsheet`  
**File:** `results_spreadsheet.go:75`  
**Signature**

```go
func CreateSheetsAndDriveServices(credsPath string) (*sheets.Service, *drive.Service, error)
```

---

## Purpose

Creates and returns authenticated Google **Sheets** and **Drive** service clients that are used to populate a results spreadsheet.  
The function expects the path to a JSON credentials file (service‑account key) and builds two separate services:

1. `sheets.Service` – for writing data into a Google Sheet.
2. `drive.Service`  – for handling Drive operations such as creating folders or moving files.

These clients are later passed to other functions that generate the spreadsheet structure, upload raw results, and store the finished file in a specified Drive folder.

---

## Parameters

| Name | Type   | Description |
|------|--------|-------------|
| `credsPath` | `string` | Filesystem path to the JSON credentials file. The function reads this file each time it is called; it does **not** cache the result. |

> *The caller typically obtains this value from a flag or environment variable.*

---

## Return Values

| Index | Type            | Meaning |
|-------|-----------------|---------|
| `0`   | `*sheets.Service` | Authenticated Sheets API client. |
| `1`   | `*drive.Service`  | Authenticated Drive API client. |
| `2`   | `error`           | Non‑nil if authentication fails or the credentials file cannot be read. |

---

## Key Steps

1. **Build a Sheets client**  
   ```go
   sheetsSvc, err := sheets.NewService(ctx,
       option.WithCredentialsFile(credsPath))
   ```
   *Uses* `google.golang.org/api/sheets/v4` and `google.golang.org/api/option`.

2. **Handle error** – if the Sheets service cannot be created, wrap the underlying error with a helpful message.

3. **Build a Drive client** in the same way using `drive.NewService`.

4. **Return** both clients or an error.

---

## Dependencies

| Dependency | Purpose |
|------------|---------|
| `google.golang.org/api/sheets/v4` | Sheets API client library. |
| `google.golang.org/api/drive/v3`  | Drive API client library. |
| `google.golang.org/api/option`    | Helper to set credentials via file path. |

The function does **not** depend on any of the package‑level globals (`credentials`, `ocpVersion`, etc.).  
It is a pure helper for authentication.

---

## Side Effects

- No global state is modified.
- The only external effect is I/O: reading the credentials JSON file from disk and establishing network connections to Google APIs.

---

## How It Fits the Package

`CreateSheetsAndDriveServices` is the first step in the upload workflow:

1. **Command Line** → `uploadResultSpreadSheetCmd` parses flags and calls this function with the supplied credential path.
2. The returned services are passed to higher‑level functions that:
   - Create the result spreadsheet (`createResultsSpreadsheet`).
   - Populate it with headers and raw data.
   - Upload the final file to a specified Drive folder.

Because authentication is isolated in its own helper, the rest of the package can focus on business logic (generating sheet names, writing rows, etc.) without duplicating credential handling code.
