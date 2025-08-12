GetPidsFromPidNamespace`

**Package:** `crclient`  
**Location:** `internal/crclient/crclient.go:157`

---

### Purpose
Retrieves the list of process IDs (PIDs) that belong to a specific PID namespace in a container running inside a Kubernetes pod.  
It is typically used by tests that need to validate the isolation or visibility of processes across namespaces.

---

### Signature

```go
func GetPidsFromPidNamespace(namespace string, c *provider.Container) ([]*Process, error)
```

| Parameter | Type                | Description |
|-----------|---------------------|-------------|
| `namespace` | `string` | The target PID namespace identifier (e.g., `"pid-1234"`). |
| `c`          | `*provider.Container` | A reference to the container whose processes are being queried. |

| Return | Type           | Description |
|--------|----------------|-------------|
| `[]*Process` | Slice of pointers to `Process` structs (defined elsewhere in the package) | Each element contains PID and associated metadata. |
| `error` | Error if any step fails. |

---

### Key Steps & Dependencies

1. **Environment & Context Setup**
   - Calls `GetTestEnvironment()` to acquire shared test configuration.
   - Uses `GetNodeProbePodContext()` to get a context for executing commands on the node that hosts the probe pod.

2. **Command Execution**
   - Executes `docker inspect` inside the target container (`c`) to read the namespace file `/proc/<pid>/ns/pid`.
   - The command is run via `ExecCommandContainer()`, which wraps kubernetes client-go calls and returns raw output.

3. **Namespace Matching**
   - Uses a regular expression (`PsRegex`) compiled with `MustCompile` to extract PIDs from the namespace string.
   - Matches all occurrences of `(<pid>)` in the inspect output, collecting candidate PIDs.

4. **PID Validation & Conversion**
   - For each matched PID:
     - Converts the string to an integer using `strconv.Atoi`.
     - On conversion error, returns a wrapped error via `Errorf`.

5. **Result Construction**
   - Appends valid `Process` objects (containing the PID) to a slice.
   - Returns the slice along with any encountered error.

---

### Side Effects

- No global state is modified; all data is returned in the result slice.
- The function may log or return errors, but does not alter the container or pod configuration.

---

### How It Fits the Package

`crclient` provides a lightweight wrapper around Kubernetes client‑go for test environments.  
This function is one of several utilities that:

- **Inspect containers** (`ExecCommandContainer`, `GetTestEnvironment`).
- **Parse process information** (regular expressions, PID extraction).
- **Return structured data** (`Process` structs) for higher‑level test logic.

It enables tests to assert that processes are correctly isolated or visible across PID namespaces, which is a core requirement for certifying container runtime security.
