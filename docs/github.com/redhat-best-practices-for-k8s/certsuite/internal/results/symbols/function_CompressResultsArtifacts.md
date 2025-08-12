CompressResultsArtifacts`

> **Package**: `internal/results`  
> **File**: `archiver.go` (line 41)  
> **Signature**

```go
func CompressResultsArtifacts(outputDir string, filePaths []string) (string, error)
```

## Purpose

`CompressResultsArtifacts` packages a set of result files into a single compressed archive.  
It is used by the test‑execution pipeline to bundle generated HTML reports and other artefacts for later download or inspection.

The function

1. **Creates** a new `.zip` file inside `outputDir`.  
2. **Adds** every path in `filePaths` as an entry in that ZIP archive, preserving relative paths.  
3. **Returns** the absolute path of the created ZIP file or an error if any step fails.

## Parameters

| Name      | Type     | Description |
|-----------|----------|-------------|
| `outputDir` | `string` | Directory where the resulting ZIP should be stored. The directory must exist; otherwise, the function will return an error. |
| `filePaths` | `[]string` | List of file system paths to include in the archive. Paths may be relative or absolute. |

## Return Values

| Value | Type    | Description |
|-------|---------|-------------|
| `zipPath` | `string` | Absolute path of the created ZIP file on success. |
| `err`      | `error`  | Non‑nil if any I/O or archive operation fails. |

## Key Steps & Dependencies

```text
1. generateZipFileName(outputDir)
   └─ returns a unique name like "results-<timestamp>.zip".

2. os.Create(zipPath) → zipFile
3. zip.NewWriter(zipFile) → zw (ZIP writer)

4. For each file in filePaths:
     - Open the source file.
     - Create a ZIP header via getFileTarHeader(file).
     - Write the header and copy file contents into `zw`.

5. Close writers & file handles.
6. Return absolute path: filepath.Abs(zipPath)
```

### Function Calls

| Call | Role |
|------|------|
| `generateZipFileName` | Builds a unique ZIP filename. |
| `filepath.Join`, `filepath.Abs` | Path manipulation. |
| `os.Create`, `os.Open` | File I/O. |
| `zip.NewWriter`, `zw.WriteHeader`, `io.Copy`, `zw.Close` | ZIP archive construction. |
| `log.Info`, `log.Debug`, `log.Errorf` | Structured logging (assumed from context). |

### Global Variables & Constants

- **`tarGzFileNamePrefixLayout` / `tarGzFileNameSuffix`** – *unused in this function* but part of the same package.
- **`htmlResultsFileContent`** – unrelated; contains embedded HTML used elsewhere.

## Side Effects

- Creates a new ZIP file on disk.  
- Does **not** delete or modify source files.  
- Emits log messages via the package’s logger (if configured).

## How It Fits in the Package

The `results` package orchestrates generation, storage, and packaging of test results:

1. Individual tests write their outputs to temporary directories.
2. After all tests finish, the pipeline calls `CompressResultsArtifacts` to bundle these files.
3. The resulting ZIP is then served (e.g., via an HTTP endpoint) or stored for archival.

Thus, `CompressResultsArtifacts` is a central utility that turns raw artefacts into a consumable artifact for downstream consumers.
