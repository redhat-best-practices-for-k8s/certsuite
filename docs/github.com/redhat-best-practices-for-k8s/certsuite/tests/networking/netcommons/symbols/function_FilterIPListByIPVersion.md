FilterIPListByIPVersion`

| | |
|---|---|
| **Package** | `netcommons` (github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/netcommons) |
| **Exported?** | ✅ |
| **Signature** | `func FilterIPListByIPVersion(ipList []string, version IPVersion) []string` |

### Purpose
Given a slice of IP address strings (`ipList`) that may contain both IPv4 and IPv6 addresses, this function returns a new slice containing only those addresses whose IP family matches the supplied `version`.  
It is used by tests that need to isolate traffic for a specific protocol version (e.g., verifying IPv4‑only routing).

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `ipList` | `[]string` | List of IP address strings. The function does **not** modify this slice. |
| `version` | `IPVersion` | Target IP family (`IPv4`, `IPv6`, `IPv4v6`, or others defined in the package). |

### Return Value
| Type | Description |
|------|-------------|
| `[]string` | New slice containing only those addresses from `ipList` that match `version`. The order of matching elements is preserved. If no addresses match, an empty slice is returned. |

### Key Steps & Dependencies
1. **Iterate over each IP string** in the input slice.
2. For each address, call **`GetIPVersion(ip)`** (another exported function in this package) to determine its actual `IPVersion`.
3. If the detected version equals the requested `version`, append the original string to a result slice using Go’s built‑in `append`.  
   ```go
   if GetIPVersion(ip) == version {
       filtered = append(filtered, ip)
   }
   ```
4. Return the accumulated slice.

### Side Effects & Guarantees
- **No mutation** of the input `ipList`.
- The function is pure: its output depends solely on the inputs and no global state.
- Relies only on `GetIPVersion` for IP parsing; any errors inside that helper are ignored here (they simply result in non‑matching entries).

### Usage Context
This helper is part of the test utilities for networking within CertSuite. Tests that generate mixed IPv4/IPv6 traffic often need to verify behavior per protocol family, and `FilterIPListByIPVersion` provides a lightweight filter without re‑implementing IP parsing logic.

---

#### Mermaid Diagram (optional)

```mermaid
flowchart TD
    A[Input ipList] --> B{Iterate}
    B --> C[GetIPVersion(ip)]
    C --> D{Matches?}
    D -- Yes --> E[append to result]
    D -- No --> F[skip]
    E --> B
    F --> B
    B --> G[Return filtered slice]
```

This diagram illustrates the linear flow of filtering each IP address.
