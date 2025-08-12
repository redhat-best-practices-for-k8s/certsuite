getFileTarHeader`

```go
func getFileTarHeader(filePath string) (*tar.Header, error)
```

| Aspect | Details |
|--------|---------|
| **Purpose** | Creates a `*tar.Header` that represents the file located at `filePath`. The header is used when packing files into a tarball (`*.tar.gz`). |
| **Inputs** | `filePath`: Path to an existing file on disk. |
| **Outputs** | * `*tar.Header` – a fully populated header describing the file’s metadata (mode, size, name, etc.).<br>* `error` – non‑nil if the file cannot be accessed or a header cannot be constructed. |
| **Key Steps** | 1. Call `os.Stat(filePath)` to obtain an `fs.FileInfo`. <br>2. If `Stat` fails → return `fmt.Errorf("unable to stat %s: %w", filePath, err)`. <br>3. Convert the `FileInfo` into a tar header via `tar.FileInfoHeader(fi, "")`. <br>4. Set the header’s name field to the base of `filePath` using `filepath.Base(filePath)`. <br>5. Return the header or an error from `FileInfoHeader`. |
| **Dependencies** | *Standard library*:<br>- `os.Stat` (stat file)<br>- `fmt.Errorf` for error formatting<br>- `archive/tar.FileInfoHeader` to convert `fs.FileInfo` into a tar header<br>- `path/filepath.Base` (implicitly via the `Name()` call in the code) |
| **Side‑effects** | None. The function only reads file metadata; it does not modify any state or write files. |
| **How it fits the package** | `results/archiver.go` builds a compressed archive of test results. For each result file, `getFileTarHeader` supplies the header that is written to the tar writer before the actual file contents are streamed. This keeps the archiving logic isolated from file‑system details and allows callers to handle errors uniformly. |

### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[Call getFileTarHeader(filePath)] --> B{os.Stat(filePath)}
    B -->|OK| C[tar.FileInfoHeader(fi, "")]
    B -->|Err| D[Return error]
    C --> E[Set header.Name = filepath.Base(filePath)]
    E --> F[Return (*tar.Header, nil)]
```

This function is a small but essential helper that abstracts the boilerplate required to generate tar headers for arbitrary files within the `results` package.
