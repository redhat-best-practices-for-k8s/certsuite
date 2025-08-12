GetTestParameters`

### Purpose
`GetTestParameters` is the primary accessor for the **test‑configuration** data used by the Certsuite test framework.  
It returns a pointer to the singleton `TestParameters` instance that holds all runtime configuration values (e.g., paths, timeouts, Kubernetes cluster information). The function guarantees that the returned structure has been loaded exactly once and is safe for concurrent reads.

### Signature
```go
func GetTestParameters() *TestParameters
```

| Item | Description |
|------|-------------|
| **Return** | `*TestParameters` – a pointer to the configuration object. |

> **Note**: The function has no parameters; all required data is derived from package‑level globals.

### Key Dependencies

| Global / Type | Role in the function |
|---------------|---------------------|
| `confLoaded`  | Boolean flag that indicates whether the configuration file has already been parsed. |
| `configuration` | Holds raw configuration values (e.g., a map or struct) populated during an earlier load step. |
| `parameters`   | The cached, fully‑validated `TestParameters` instance returned by this function. |

The function relies on other package routines (not shown in the snippet) that perform the actual file read and parsing; those routines set `confLoaded`, populate `configuration`, and construct `parameters`.

### Side Effects & Invariants

* **No state mutation after first call** – Subsequent invocations simply return the already‑initialized `parameters` pointer.
* **Thread safety** – The implementation uses a package‑level read/write guard (e.g., `sync.Once`) to guarantee that initialization happens only once even under concurrent access.  
  *If such guarding is missing, callers should invoke this function after any explicit configuration load step.*

### How It Fits the Package

The `configuration` package encapsulates all logic for loading and exposing test settings. `GetTestParameters` is the public entry point used by:

* Test executors (`pkg/runner`, `pkg/tests/...`) to read runtime options.
* CLI commands that need to introspect current configuration before launching tests.

By centralizing access through this function, the package ensures a single source of truth for test parameters and avoids duplicated parsing logic across the codebase.
