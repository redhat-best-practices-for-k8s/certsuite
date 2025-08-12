claimCompare` – CLI handler for comparing two claim files

| Item | Details |
|------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare` |
| **Signature** | `func (*cobra.Command, []string) error` |
| **Exported?** | No – internal command implementation |

### Purpose
`claimCompare` is the entry point for the `certsuite claim compare` sub‑command.  
It parses the two file paths supplied via the global flags (`Claim1FilePathFlag`, `Claim2FilePathFlag`) and delegates the actual comparison logic to the helper function `claimCompareFilesfunc`. The function returns any error that occurs during parsing or comparison.

### Parameters
| Parameter | Type | Role |
|-----------|------|------|
| `cmd` | `*cobra.Command` | The Cobra command instance (unused inside the body, but required by the signature). |
| `args` | `[]string` | Positional arguments – not used; the command expects file paths via flags instead. |

### Return Value
```go
error
```
- Returns an error if the comparison fails or if a fatal condition is encountered (e.g., missing flag values).  
- On success it returns `nil`.

### Key Dependencies & Side‑Effects

| Dependency | How It’s Used |
|------------|---------------|
| `claimCompareFilesfunc` | Called with the two file path flags; performs the actual claim comparison logic. |
| `Fatal` (from the same package) | Invoked when a fatal error occurs to terminate the program and print an error message. |

### Global Variables Involved

```go
var (
    Claim1FilePathFlag string // Path to first claim file (set via flag)
    Claim2FilePathFlag string // Path to second claim file (set via flag)

    // Helper that performs the comparison; implementation elsewhere in this package.
    claimCompareFiles func(string, string) error
)
```

- `Claim1FilePathFlag` and `Claim2FilePathFlag` are populated by Cobra’s flag parsing mechanism before `claimCompare` runs.  
- `claimCompareFiles` is a function variable that can be overridden (e.g., for testing). The actual comparison logic resides in the implementation of this variable.

### Flow Summary

```mermaid
flowchart TD
    A[cmd starts] --> B{args used?}
    B -- no --> C[Call claimCompareFilesfunc(Claim1FilePathFlag, Claim2FilePathFlag)]
    C --> D{error?}
    D -- yes --> E[Fatal(error)]
    D -- no --> F[return nil]
```

### How It Fits the Package
`claimCompare` ties together Cobra command parsing with the claim‑comparison logic.  
- The package exposes a `compare.go` file that defines the sub‑command, its flags, and this handler.  
- Other parts of the CLI register the command; when a user runs `certsuite claim compare`, this function is executed, ultimately calling `claimCompareFilesfunc` to produce the comparison output.

---
