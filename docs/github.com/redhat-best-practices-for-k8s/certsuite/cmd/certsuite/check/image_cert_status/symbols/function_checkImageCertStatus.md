checkImageCertStatus`

| Item | Details |
|------|---------|
| **Package** | `imagecert` (`github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/check/image_cert_status`) |
| **Exported** | No – it is an internal command handler. |
| **Signature** | `func(*cobra.Command, []string) error` |

### Purpose
This function implements the *image‑certificate status* sub‑command of the Certsuite CLI.  
When a user runs:

```bash
certsuite check image-cert-status <flags>
```

the Cobra framework calls this handler to:
1. Parse command flags (`image`, `namespace`, `container`, `all-containers`).
2. Validate that required parameters are supplied.
3. Determine whether the specified container(s) inside a pod are certified by consulting the *certification service* via `IsContainerCertified`.
4. Print a concise status report to stdout, color‑coding success (green) and failure (red).

### Inputs
| Parameter | Type | Description |
|-----------|------|-------------|
| `cmd` | `*cobra.Command` | The Cobra command instance; used to retrieve flag values via `Flags().GetString(...)`. |
| `_args` | `[]string` | Positional arguments – not used by this function (all data comes from flags). |

### Outputs
- Returns an `error`:
  - `nil` when the command completes successfully.
  - A descriptive error (via `Errorf`) if mandatory flags are missing or the certification check fails.

The function also writes directly to stdout via `Println`/`Printf`, so callers do not capture output programmatically.

### Key Dependencies
| Dependency | Role |
|------------|------|
| **Cobra** (`github.com/spf13/cobra`) | Provides flag handling and command registration. |
| **certsuite internal helpers** | `GetString`, `GetValidator`, `IsContainerCertified` – utilities for retrieving flag values, validating inputs, and querying certification status. |
| **color helpers** (`GreenString`, `RedString`) | Wrap output strings in ANSI color codes for terminal display. |

### Side‑Effects
- Prints to standard output.
- Uses global command variable `checkImageCertStatusCmd` only for registration; the handler itself does not modify globals.

### How it Fits the Package
The `imagecert` package groups all image‑certificate related CLI commands.  
`checkImageCertStatus` is one of those commands, invoked via Cobra when a user requests status information. It delegates to shared helpers (`IsContainerCertified`) that encapsulate the business logic for certification verification. The function is intentionally read‑only regarding the codebase; it only reads flags and writes output.

### Suggested Mermaid Diagram
```mermaid
flowchart TD
    A[User runs certsuite check image-cert-status] --> B[Cobra command dispatch]
    B --> C{checkImageCertStatus}
    C --> D1[Retrieve flags: image, namespace, container, all-containers]
    D1 --> E{Validate required fields}
    E -- valid --> F[Call IsContainerCertified for each target]
    F --> G[Print status (green/red)]
    G --> H[Return nil]
    E -- invalid --> I[Errorf -> return error]
```

This diagram illustrates the high‑level control flow of the function.
