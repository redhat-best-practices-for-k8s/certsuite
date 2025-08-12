UnmarshalConfigurations`

**Package:** `claimhelper`  
**Location:** `pkg/claimhelper/claimhelper.go:294`

---

## Purpose
`UnmarshalConfigurations` converts a raw byte slice (typically the contents of a JSON or YAML configuration file) into a Go map that can be consumed by other parts of the CertSuite claim‑generation logic. It is used wherever a configuration blob needs to be parsed once and then reused as a generic key/value store.

## Signature
```go
func UnmarshalConfigurations(data []byte, target map[string]interface{}) func()
```

* **Parameters**
  * `data` – raw bytes representing the configuration (JSON/YAML).  
  * `target` – an empty or partially‑filled map that will receive the unmarshaled key/value pairs.

* **Return value** – a zero‑argument function.  
  The returned closure is intended to be deferred; it calls `Fatal()` on any error, causing the program to terminate immediately with a log message. This pattern allows callers to write:

  ```go
  defer UnmarshalConfigurations(configBytes, configMap)()
  ```

  and keep the rest of the function body clean.

## Key Dependencies

| Dependency | Role |
|------------|------|
| `Unmarshal` (from `encoding/json` or a custom unmarshaler) | Parses `data` into `target`. |
| `Fatal` (likely from `log` or `fmt`) | Logs the error and exits the program. |

These functions are called *inside* the returned closure, ensuring that any failure to parse the configuration results in an immediate fatal log.

## Side‑Effects

* **Fatal exit** – On parsing failure, the process terminates with a logged error; no value is returned.
* **Mutation of `target`** – The map passed in is populated directly; callers must provide a mutable map (e.g., `make(map[string]interface{})`) before invoking.

No other global state or file system access occurs here; it is purely in‑memory transformation.

## How It Fits the Package

The `claimhelper` package centralizes utilities for generating and validating claims used by CertSuite.  
- Configuration data drives claim generation rules, validation parameters, and output formatting.  
- `UnmarshalConfigurations` provides a reusable, safe way to load these settings from external files.
- The function’s return‑closure pattern fits the package’s error‑handling style (immediate fatal on misconfiguration), ensuring that any downstream logic never runs with invalid or missing configuration.

---

### Suggested Mermaid Diagram

```mermaid
flowchart TD
  A[Caller] -->|defer UnmarshalConfigurations(data, cfg)()| B[UnmarshalClosure]
  B --> C{unmarshal}
  C --> |success| D[cfg populated]
  C --> |error| E[Fatal & exit]
```

This visualises the deferred call and its fatal‑on‑error behaviour.
