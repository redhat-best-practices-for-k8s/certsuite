FindMajorMinor`

```go
// FindMajorMinor returns the major‑minor component of a Kubernetes/OCP version string.
//
// Example:
//   FindMajorMinor("4.12.3") → "4.12"
//   FindMajorMinor("1.22")   → "1.22"
//   FindMajorMinor("v1.21")  → "v1.21"   // unchanged if no dot is present
func FindMajorMinor(version string) string
```

### Purpose  
The function normalises an OCP/Kubernetes version string to its *major.minor* form.  
Many parts of the `compatibility` package (e.g., lifecycle date lookups, status mapping) need only the major‑minor pair, not patch or build metadata.

### Inputs  

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `version` | `string` | A version identifier such as `"4.12.3"`, `"1.22"` or `"v1.21"`.*

\*The function is tolerant of leading “v” or other non‑numeric prefixes; it simply splits on the first dot.

### Output  

| Return value | Type   | Description |
|--------------|--------|-------------|
| `string`     | The major and minor parts joined by a period, e.g., `"4.12"`. If the input contains fewer than two components (no dot), the original string is returned unchanged. |

### Key Dependencies  

* **`strings.Split`** – Splits the input on “.” to isolate the version components.

No other global variables or functions are accessed.

### Side Effects  

The function has no side effects: it does not modify any globals, perform I/O, or alter its argument.

### How It Fits the Package  
Within `pkg/compatibility`, many look‑ups (e.g., `ocpLifeCycleDates` map keys) use a normalized major.minor string.  
`FindMajorMinor` provides that canonical form so callers can reliably index those maps and compare versions across different parts of the suite.

---

**Mermaid diagram suggestion**

```mermaid
graph TD;
  A[Input version] -->|Split on "."| B{Components};
  B -->|≥2| C[Return first two joined by "."];
  B -->|<2| D[Return original string];
```

This visualises the simple branching logic of `FindMajorMinor`.
