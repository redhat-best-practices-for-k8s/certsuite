Event.String() string`

| Aspect | Details |
|--------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider` |
| **Receiver** | `e Event` (value receiver) |
| **Signature** | `func (e Event) String() string` |

### Purpose
`Event.String` implements the `fmt.Stringer` interface for the internal `Event` type. It produces a human‑readable representation of an event that is logged or displayed to users.

The function formats three fields of the event:
1. **Name** – the event identifier (e.g., `"NodeReady"`).
2. **Message** – a free‑form description.
3. **Timestamp** – the time the event was recorded, formatted as RFC3339.

This representation is used throughout the provider when printing events to console, logs, or test reports.

### Inputs
- The method receives an `Event` value by copy (`e Event`).  
  All fields of `Event` are read; no mutation occurs.

### Output
- Returns a single string in the form:
  ```
  "%s: %s (%s)"
  ```
  where the placeholders correspond to the event's name, message, and timestamp (RFC3339).

### Key Dependencies & Calls
| Dependency | Role |
|------------|------|
| `fmt.Sprintf` | Formats the output string. |
| `e.Timestamp.Format(time.RFC3339)` | Converts the time value to a standard textual representation. |

No external globals or package variables are accessed; the method is fully self‑contained.

### Side Effects
- **None** – The function only reads from the receiver and performs formatting. It does not modify any state, write logs, or affect other packages.

### Integration in the `provider` Package
The `Event` type (defined elsewhere in the package) represents status changes or significant occurrences during a CertSuite run.  
- **Logging** – When events are recorded, they are often printed using `fmt.Println(e)`; thanks to this method, the output is consistent and human‑friendly.
- **Testing** – Test harnesses may capture event strings for assertions.

By providing a clean string representation, `Event.String` ensures that logs remain readable without exposing internal struct layouts or requiring callers to format fields manually.
