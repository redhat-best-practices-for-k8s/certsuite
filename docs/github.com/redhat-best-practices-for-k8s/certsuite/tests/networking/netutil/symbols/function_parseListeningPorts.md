parseListeningPorts`

| Aspect | Detail |
|--------|--------|
| **Package** | `netutil` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/netutil`) |
| **Signature** | `func parseListeningPorts(data string) (map[PortInfo]bool, error)` |
| **Exported?** | No – internal helper used by the test suite to interpret raw net‑stat output. |

### Purpose
`parseListeningPorts` turns a raw multi‑line string produced by the `netstat -tulpn` command into a map keyed on `PortInfo`. Each entry in the map records whether the port is actively listening (`true`) or not (`false`). The function performs basic validation of each line and reports malformed input via an error.

### Inputs
| Parameter | Type | Description |
|-----------|------|-------------|
| `data` | `string` | Raw output from `netstat -tulpn`, one connection per line, with a header that should be ignored. |

### Outputs
| Return value | Type | Meaning |
|--------------|------|---------|
| `map[PortInfo]bool` | A map where the key is a `PortInfo` (protocol, port number, and state) and the value is `true` if the line indicates an active listening socket. |
| `error` | `nil` on success; otherwise descriptive error indicating why parsing failed. |

### Key Dependencies & Constants
The function uses several package‑level constants that mirror the column indices in a typical `netstat -tulpn` output:

- `getListeningPortsCmd` – command string (unused directly here but part of the package).  
- `indexPort`, `indexProtocol`, `indexState` – integer positions for the “Local Address”, “Proto” and “State” columns.  
- `portStateListen` – literal `"LISTEN"` used to detect listening sockets.

Other helpers:
* `TrimSuffix`, `Split`, `Fields` – string manipulation from `strings`.  
* `ParseInt` – converts port number strings to integers (`int32`).  
* `Errorf`, `ToUpper` – for error formatting and normalizing protocol names.

### How It Works (Step‑by‑step)

1. **Trim the trailing newline**: removes an optional final `\n` that may be present.
2. **Split into lines**: each line is processed independently; the first line (header) is skipped if it contains `"Proto"`.
3. For every remaining line:
   - Split by whitespace (`Fields`) to get tokens.
   - Verify that at least 6 columns exist; otherwise return an error.
   - Extract protocol, local address and state using the predefined indices.
   - Parse the port number from the `Local Address` field (format `"IP:port"` or `"[IPv6]:port"`).
   - Build a `PortInfo{Protocol, Port, State}` struct.
   - Mark the entry as listening (`true`) only if the state equals `"LISTEN"`.
4. Return the assembled map and a nil error.

### Side Effects
None beyond normal Go runtime behaviour – it allocates a new map and may return an error; no global state is modified.

### How It Fits the Package

`netutil` contains helper utilities for networking tests.  
Other parts of the package (e.g., `GetListeningPorts`) likely invoke this function after running `netstat`. The returned map feeds into test assertions that verify whether expected ports are listening or closed, forming a core part of the network‑state validation logic.

### Suggested Mermaid Diagram

```mermaid
flowchart TD
  A[Raw netstat string] --> B{Split lines}
  B --> C{Skip header?}
  C -->|yes| D[Process each line]
  D --> E{Parse tokens}
  E --> F[Extract protocol, port, state]
  F --> G{State == LISTEN?}
  G -->|yes| H[Set map[key] = true]
  G -->|no | I[Set map[key] = false]
  H & I --> J[Return map + nil]
```

This diagram visualises the linear parsing logic and decision points.
