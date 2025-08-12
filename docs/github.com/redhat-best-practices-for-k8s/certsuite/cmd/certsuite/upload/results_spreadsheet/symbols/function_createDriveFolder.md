createDriveFolder`

| | |
|-|-|
|**Package**|`resultsspreadsheet` (internal helper in the *upload* sub‑command) |
|**Signature**|`func createDriveFolder(svc *drive.Service, parentID, folderName string) (*drive.File, error)` |
|**Exported?**|No – used only within this package |

---

#### Purpose
Creates a new Google Drive folder under the specified `parentID`.  
If a folder with the same name already exists in that location, the existing folder is returned instead of creating a duplicate.

This helper abstracts the two‑step logic required by the Drive API:

1. Search for an existing folder (`Files.List` with a query on `name`, `mimeType`, and `parents`).
2. If none found, create it (`Files.Create`).

---

#### Parameters

| Name | Type | Description |
|------|------|-------------|
| `svc` | `*drive.Service` | Authenticated Drive client used to issue API calls. |
| `parentID` | `string` | The ID of the parent folder where the new folder should be placed. |
| `folderName` | `string` | Desired name for the folder (displayed in Google Drive). |

---

#### Return Values

| Name | Type | Description |
|------|------|-------------|
| `*drive.File` | *drive.File | A pointer to the Drive metadata of the created or found folder. The returned struct contains at least an `Id`. |
| `error` | error | Non‑nil if any API call fails, or if a duplicate name exists but cannot be retrieved. |

---

#### Key Dependencies

- **Drive API** (`google.golang.org/api/drive/v3`) – the function uses:
  - `Files.List`, `Q`, `Fields`, `Do` to search.
  - `Files.Create`, `Do` to create a folder.
- Standard library:
  - `fmt.Sprintf` for query string formatting.
  - `errors.Errorf` (from `github.com/pkg/errors`) for error wrapping.

---

#### Side‑Effects

- **Network**: Makes HTTP requests to Google Drive; may be rate‑limited.
- **Drive State**: Creates a folder if not present, otherwise leaves the existing folder untouched.
- No local state is modified – purely read/write on the Drive service.

---

#### Flow (pseudocode)

```go
func createDriveFolder(svc, parentID, name string) (*drive.File, error) {
    // 1. Look for an existing folder with same name under parent
    query := fmt.Sprintf("name='%s' and mimeType='application/vnd.google-apps.folder' " +
                         "and '%s' in parents and trashed=false", name, parentID)
    listRes, err := svc.Files.List().Q(query).Fields(...).Do()
    if err != nil { return nil, errors.Wrap(err, "listing drive files") }

    // 2. If found, return it
    if len(listRes.Files) > 0 {
        return listRes.Files[0], nil
    }

    // 3. Create new folder
    f := &drive.File{
        Name:     name,
        MimeType: "application/vnd.google-apps.folder",
        Parents: []string{parentID},
    }
    created, err := svc.Files.Create(f).Do()
    if err != nil { return nil, errors.Wrap(err, "creating drive folder") }

    return created, nil
}
```

---

#### How It Fits the Package

The `resultsspreadsheet` command uploads CSV/JSON test results to a Google Drive hierarchy.  
Before uploading files, it must ensure that all necessary folders exist:

1. Root project folder (e.g., “CertSuite Results”).
2. Sub‑folders per OCP version, operator version, etc.

`createDriveFolder` is invoked repeatedly while constructing this tree, guaranteeing idempotent creation and avoiding duplicate folders on successive runs. It centralises the Drive API interaction, making the rest of the upload logic simpler and more testable.
