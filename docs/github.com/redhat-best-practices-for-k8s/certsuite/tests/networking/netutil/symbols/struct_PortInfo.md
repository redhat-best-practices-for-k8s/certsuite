PortInfo` – Lightweight representation of a network listener

| Item | Detail |
|------|--------|
| **File** | `netutil.go` (line 36) |
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/netutil` |
| **Exported?** | Yes – used by test helpers to expose listening‑port information. |

## Purpose

`PortInfo` is a small value type that describes a single TCP/UDP listener inside a container.  
It is the key unit returned by `GetListeningPorts`, which queries a container’s runtime state (via `ExecCommandContainerNSEnter`) and parses the output of `ss -tuln`.  The struct allows callers to:

1. **Index** listeners efficiently – used as map keys (`map[PortInfo]bool`).
2. **Compare** port/protocol combinations across containers or test runs.
3. **Report** which ports are open when a test fails.

## Fields

| Field | Type   | Meaning |
|-------|--------|---------|
| `PortNumber` | `int32` | The numeric port number (e.g., 80, 443). Stored as an integer for quick comparisons and map keys. |
| `Protocol`   | `string` | Protocol name in upper‑case (`"TCP"` or `"UDP"`). Normalized by `parseListeningPorts`. |

> **Note:** Both fields are exported so that test code can build maps or slices of `PortInfo` without reflection.

## Typical Usage Flow

1. **Call** `GetListeningPorts(container *provider.Container) (map[PortInfo]bool, error)`
   - Executes `ss -tuln` inside the container.
2. **Parse** the command output with `parseListeningPorts`.
3. **Return** a map keyed by `PortInfo`.  
   Each key’s value is always `true`; the map acts as a set.

```go
ports, err := netutil.GetListeningPorts(container)
if err != nil { /* handle error */ }
for p := range ports {
    fmt.Printf("Container listens on %s/%d\n", p.Protocol, p.PortNumber)
}
```

## Key Dependencies

| Dependency | Role |
|------------|------|
| `provider.Container` | The container whose listening ports are queried. |
| `ExecCommandContainerNSEnter` | Runs the `ss -tuln` command inside the container. |
| `parseListeningPorts` | Converts raw command output into a map of `PortInfo`. |
| `Errorf`, `ParseInt`, etc. | Used within `parseListeningPorts` to validate and convert data. |

## Side‑Effects & Constraints

* **Immutable** – `PortInfo` is a plain struct; no methods mutate it.
* **Map Key** – Because both fields are comparable, instances can be used as keys in Go maps without additional hashing logic.
* **Protocol Normalization** – The parser guarantees that `Protocol` is upper‑case; callers should rely on this invariant.

## Integration with the Package

`netutil` provides utilities for networking diagnostics inside containers.  
`PortInfo` sits at the heart of those diagnostics, enabling:

- Test cases to assert that required services are listening.
- Comparisons between expected and actual port sets.
- Generation of readable logs when a container fails to expose a port.

The struct’s simplicity keeps the rest of the package lightweight while still offering precise control over network state inspection.
