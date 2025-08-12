CheckResult.String`

```go
func (r CheckResult) String() string
```

### Purpose  
`CheckResult` is an enumerated type representing the outcome of a test check.  
The `String()` method converts that enum value into a human‑readable, colored
string that can be printed in logs or UI output.

> **Why it matters** – All other parts of the package (e.g., rendering functions,
> status tables, and API responses) rely on this representation to communicate
> results back to users. A consistent string format also allows downstream tools
> (like CI dashboards) to parse test results automatically.

### Inputs / Receiver  
- **Receiver**: `r CheckResult` – a value of the enum type defined in
  `check.go`.  
- No other parameters are taken; the method is pure with respect to its
  receiver.

### Output  
A single `string` that contains:

1. The plain textual representation (`PASSED`, `FAILED`, etc.).
2. A colored prefix using ANSI escape codes (e.g., green for success,
   red for failure, yellow for skip).

The color code is chosen based on the constants defined in `checksdb.go`
(`PASSED`, `FAILED`, `SKIPPED`). If a value does not match any known enum,
the function simply returns `"unknown"`.

### Key Dependencies  
| Dependency | Role |
|------------|------|
| `string` (built‑in) | Converts the internal representation into a Go string. |
| Constants `PASSED`, `FAILED`, `SKIPPED` | Determine which color and text to use. |
| ANSI escape codes (hard‑coded inside the method) | Provide terminal colouring. |

No global state is read or modified; the function has **no side effects**.

### How it fits the package  
- **Checks execution** – After a check finishes, its result is stored as
  `CheckResult`.  
- **Reporting** – Functions that produce textual reports (e.g., CLI output,
  logs) call `.String()` to present the outcome.  
- **API exposure** – When results are marshalled into JSON or other formats,
  this method ensures consistent, readable strings for consumers.

### Usage Example

```go
res := CheckResultPassed
fmt.Println("Check status:", res.String())
// → "Check status: PASSED" (with green colour in a terminal)
```

---

**Mermaid diagram suggestion**

```mermaid
flowchart TD
    A[Check Execution] --> B{Store Result}
    B --> C[CheckResult enum]
    C --> D[CheckResult.String()]
    D --> E[Human‑readable, coloured string]
```

This method is a small but essential part of the `checksdb` package,
bridging internal test state to user‑facing output.
