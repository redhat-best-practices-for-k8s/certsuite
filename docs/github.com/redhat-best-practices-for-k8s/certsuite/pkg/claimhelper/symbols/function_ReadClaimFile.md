ReadClaimFile`

```go
func ReadClaimFile(filePath string) ([]byte, error)
```

## Purpose

`ReadClaimFile` is a helper that reads the contents of a *claim file* (a JSON or other payload used by CertSuite to store test results).  
The function:

1. Reads the entire file at `filePath`.  
2. Returns the raw byte slice on success, otherwise an error.

It is part of the **claimhelper** package, which centralises all logic for interacting with claim files (writing, reading, parsing and validating).

---

## Parameters

| Name | Type   | Description |
|------|--------|-------------|
| `filePath` | `string` | The absolute or relative path to the claim file that should be read. |

---

## Return Values

| Index | Type   | Description |
|-------|--------|-------------|
| 0     | `[]byte` | Raw contents of the claim file. On error this slice is nil. |
| 1     | `error`  | `nil` on success; otherwise an error returned by `os.ReadFile`. |

---

## Dependencies & Side Effects

| Dependency | Role | Notes |
|------------|------|-------|
| `ReadFile` (from `os`) | Reads the file into memory. | Direct I/O operation; can fail with *file not found*, *permission denied*, etc. |
| `Error` (logging) | Logs an error message when reading fails. | Side‑effect: writes to the global logger. |
| `Info` (logging) | Logs a success message after successful read. | Side‑effect: writes to the global logger. |

No other package state is modified; the function is *pure* with respect to its input and output, except for logging.

---

## Usage Context

Within CertSuite, claim files are produced by various test runners and later consumed by the report generator or validation tools.  
`ReadClaimFile` is typically called when:

- A test result needs to be re‑loaded (e.g., for re‑validation).  
- The contents of a claim file must be inspected before being merged into a larger report.

Because it logs both success and failure, it also serves as a lightweight audit trail for file operations.

---

## Example

```go
data, err := claimhelper.ReadClaimFile("/tmp/claim.json")
if err != nil {
    // handle error – the function already logged it
}
fmt.Println("Claim size:", len(data))
```

---

## Diagram (optional)

```mermaid
flowchart TD
  A[Call ReadClaimFile(filePath)] --> B{ReadFile OK?}
  B -- yes --> C[Return data]
  B -- no --> D[Log Error & Return err]
  C --> E[Log Info]
```

This flow captures the key decision points: successful read → log info; failure → log error.
