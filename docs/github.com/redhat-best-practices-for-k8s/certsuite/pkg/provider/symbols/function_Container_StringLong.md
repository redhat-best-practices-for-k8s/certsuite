Container.StringLong`

**Location**

`pkg/provider/containers.go:149`

```go
func (c Container) StringLong() string {
    return fmt.Sprintf("%s:%d/%s", c.HostIP, c.HostPort, c.Name)
}
```

### Purpose

Creates a human‑readable identifier for a container that is unique within the test run.  
The format is:

```
<host IP>:<host port>/<container name>
```

This string is used throughout the provider package to log diagnostics, build
resource names and to reference containers in error messages.

### Receiver

| Name | Type     |
|------|----------|
| `c`  | `Container` |

The receiver is a value copy; no mutation occurs.

### Return Value

| Type   | Description |
|--------|-------------|
| `string` | The formatted string described above. |

### Key Dependencies

* **`fmt.Sprintf`** – formatting routine from the standard library.
* **Fields of `Container`:**
  * `HostIP` (string) – IP address on which the container is reachable.
  * `HostPort` (int) – Port number exposed by the container.
  * `Name` (string) – Container’s name inside its pod.

No external packages or globals are accessed, making this function pure and side‑effect free.

### Side Effects

None. The method only reads from the receiver and returns a new string.

### Package Context

The `provider` package implements runtime checks for OpenShift/Kubernetes clusters.
`Container.StringLong()` is part of the **container inspection utilities** used by:

* **Diagnostic logs** – when a test fails, the full container address is logged.
* **Test result aggregation** – results are keyed by this string to avoid collisions.
* **Human‑readable output** – when printing lists of containers (e.g., in CLI tools).

Because it is pure and deterministic, it can be safely used in concurrent
contexts without locking.

---

### Example

```go
c := Container{
    HostIP:   "10.0.0.5",
    HostPort: 8080,
    Name:     "nginx-proxy",
}
fmt.Println(c.StringLong())
// → "10.0.0.5:8080/nginx-proxy"
```

This concise representation is preferred over the longer `Container` struct
serialization when readability matters.
