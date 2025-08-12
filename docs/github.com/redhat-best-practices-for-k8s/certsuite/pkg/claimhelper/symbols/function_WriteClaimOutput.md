WriteClaimOutput`

```go
func WriteClaimOutput(filePath string, payload []byte) ()
```

### Purpose  
`WriteClaimOutput` persists the JSON (or binary) payload that represents a *claim* to disk.  
The function is used after a claim has been generated – for example, in tests or
in the CertSuite command‑line tool – and the resulting data must be stored so
that downstream tools can read it.

### Parameters  

| Name | Type   | Description |
|------|--------|-------------|
| `filePath` | `string` | Destination file on disk. The function will create or overwrite this file. |
| `payload`  | `[]byte` | Raw bytes to write (typically JSON). |

### Return Value  
None – the function is declared with a zero‑value return type and simply
terminates the process if something goes wrong.

### Core Logic

```text
1. Log: "Writing claim output to <filePath>"
2. Attempt to write `payload` to `<filePath>` using os.WriteFile.
   - Permissions are set by the unexported const `claimFilePermissions`.
3. If WriteFile fails:
     a. Convert error to string via fmt.Sprintf("%v", err)
     b. Call Fatal(errString) – this aborts the program (calls log.Fatalf internally).
```

### Dependencies & Side‑Effects  

| Dependency | Role |
|------------|------|
| `Info` (from package logger) | Records an informational message before writing. |
| `WriteFile` (`os.WriteFile`) | Performs the actual file write with permissions `claimFilePermissions`. |
| `Fatal` (logger) | Aborts execution if the write fails; propagates error to stdout/stderr. |
| `string` conversion | Formats the error for Fatal output. |

*Side‑effects*:  
- Creates/overwrites a file at `filePath`.  
- Logs an info message and may terminate the program on failure.

### How It Fits in `claimhelper`

The package orchestrates the creation, validation, and persistence of *claims*
used by CertSuite to prove that certain compliance checks passed.  
`WriteClaimOutput` is the final step that writes the claim data out so other
components (e.g., uploaders or test harnesses) can consume it.

```
mermaid
flowchart TD
  A[Generate Claim Payload] --> B{Is Valid?}
  B -- yes --> C[WriteClaimOutput(filePath, payload)]
  C --> D[Persisted to disk]
  B -- no --> E[Log / Skip]
```

### Notes  

* The function is intentionally *void* because a failure in writing a claim
should stop the process immediately – it would be unsafe to continue with an
incomplete or missing claim file.  
* `claimFilePermissions` is unexported; its value is defined elsewhere in the package
and dictates the UNIX permission bits for the created file.  

---
