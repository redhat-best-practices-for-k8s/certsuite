parsePingResult` – Internal Result Parser

### Purpose
`parsePingResult` converts the raw output of a Linux `ping` command into a structured `PingResults` value and reports any problems that prevented a valid ping run.

The function is **unexported**; it is used only by tests in this package to turn the textual stdout/stderr produced by a helper (`TestPing`) into machine‑readable data.

### Signature
```go
func parsePingResult(stdout string, stderr string) (PingResults, error)
```

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `stdout`  | `string` | The standard output captured from the ping process. |
| `stderr`  | `string` | The standard error stream – used only for detecting malformed command invocation. |

### Return Values
- **PingResults**  
  A struct (defined elsewhere in this package) that holds:
  - `TotalPackets`: number of packets sent,
  - `ReceivedPackets`: number received,
  - `LossPct`: packet loss percentage,
  - `Min`, `Avg`, `Max`, `Stddev`: round‑trip time statistics.

- **error**  
  An error is returned if either the command failed to run correctly or the output could not be parsed.  
  Typical errors include:
  * “invalid ping arguments” – when stderr contains a message matched by `ConnectInvalidArgumentRegex`.
  * “unable to parse ping result” – when stdout does not match the expected pattern.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `regexp.MustCompile` | Compiles two regular expressions at runtime: one for detecting an invalid‑argument error, another for extracting statistics from a successful run. |
| `FindStringSubmatch` | Extracts captured groups from the regex matches. |
| `strconv.Atoi` | Converts string numeric fields into integers (packet counts) or floats (percentages). |
| `fmt.Errorf` | Wraps errors with context. |

The function uses two package‑level constants defined in the same file:

```go
const (
    ConnectInvalidArgumentRegex = `ping: .*`
    SuccessfulOutputRegex       = `(\d+) packets transmitted, (\d+) received, ([\d\.]+)% packet loss.*min/avg/max/mdev = ([\d\.]+)/([\d\.]+)/([\d\.]+)/([\d\.]+) ms`
)
```

These regexes are compiled on each call (via `MustCompile`) because the function is short‑lived and invoked only during tests.

### Side Effects
- None.  
  The function performs no I/O, does not modify global state, and returns a pure result or an error.

### How It Fits the Package

The `icmp` package contains end‑to‑end network tests for verifying ICMP connectivity between containers.  
`TestPing` runs an external `ping` command and captures its output.  
Afterward, it calls `parsePingResult` to turn that raw output into a `PingResults` struct so the test can assert on:

- Packet loss,
- RTT statistics, etc.

Because this function is internal, changes to its behaviour or signature are confined to the package; callers rely only on the documented return values and error handling.
