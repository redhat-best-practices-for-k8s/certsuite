MarshalClaimOutput`

| Feature | Detail |
|---------|--------|
| **Purpose** | Serialises a *claim* tree (`*claim.Root`) into a JSON byte slice for downstream consumption (e.g., CLI output, API responses). It guarantees that the claim is encoded in an indented format and aborts the program if marshaling fails. |
| **Signature** | `func MarshalClaimOutput(root *claim.Root) []byte` |
| **Inputs** | - `root`: a pointer to a `claim.Root` value representing the entire claim hierarchy. The function assumes this is non‑nil; passing `nil` will cause a panic when `MarshalIndent` attempts to access fields. |
| **Outputs** | Returns a byte slice containing the pretty‑printed JSON representation of the claim. On failure, the function never returns – it logs the error and terminates the process via `Fatal`. |
| **Key dependencies** | - `encoding/json.MarshalIndent`: performs the actual conversion from Go structs to JSON with indentation.<br>- `log.Fatal` (via a local wrapper named `Fatal`): used for fatal error handling. |
| **Side effects** | 1. **Logging & termination** – if marshaling fails, the function writes an error message and exits the program immediately (`os.Exit(1)` via `Fatal`).<br>2. No mutation of the input claim tree occurs; it is read‑only. |
| **How it fits the package** | The `claimhelper` package provides utilities for manipulating and inspecting claims. `MarshalClaimOutput` is a thin wrapper that standardises how claims are presented to users or other components, ensuring consistent JSON formatting and robust error handling. It is typically invoked by command‑line tools or HTTP handlers when they need to output the claim data in a human‑readable form. |

### Usage example

```go
root := claimhelper.LoadClaim("path/to/claim.yaml") // returns *claim.Root
jsonBytes := claimhelper.MarshalClaimOutput(root)
fmt.Println(string(jsonBytes))
```

If `MarshalIndent` fails (e.g., due to an unsupported type in the claim), the program will log an error and exit, preventing downstream components from receiving malformed data.

### Diagram

```mermaid
flowchart TD
    A[User or API] --> B{Want claim output}
    B --> C[claimhelper.MarshalClaimOutput(root)]
    C --> D[encoding/json.MarshalIndent]
    D --> E[JSON bytes]
    D -- error --> F[log.Fatal & os.Exit(1)]
```

This function is intentionally simple but crucial for ensuring that all outputs from the `claimhelper` package are consistently formatted and reliably produced.
