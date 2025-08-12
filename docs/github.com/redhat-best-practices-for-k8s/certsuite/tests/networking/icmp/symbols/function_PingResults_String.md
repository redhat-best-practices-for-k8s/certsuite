PingResults.String` – Package `icmp`

## Overview
`PingResults.String` is a method that serialises a `PingResults` value into a human‑readable string.  
It is used by the test harness to log or display the outcome of an ICMP ping run, e.g., in the `TestPing` test case.

```go
func (r PingResults) String() string
```

- **Receiver**: `PingResults` – a struct defined elsewhere in this package that holds all metrics returned by a ping command.
- **Return type**: `string`.

The method is *exported* so callers can embed the output directly into test reports or logs.

## Functionality

1. **Format construction**  
   The method uses `fmt.Sprintf` to build a formatted string containing key fields of the `PingResults`.  
   It typically looks like:

   ```text
   PingResults{Success: true, PacketsSent: 10, PacketsRecv: 10, AvgRTT: "1.23 ms"}
   ```

2. **Delegated conversion**  
   Internally it calls a helper function `ResultToString`, which actually performs the field‑by‑field formatting.  
   This separation allows `ResultToString` to be reused by other parts of the package (e.g., JSON marshalling, custom reporters).

3. **No external side effects** – the method only reads from the receiver; it does not modify state or perform I/O.

## Dependencies

| Dependency | Role |
|------------|------|
| `fmt.Sprintf` | String formatting |
| `ResultToString` | Converts a `PingResults` struct into its string representation |

Both dependencies are standard Go library functions or package‑internal helpers, so the method has no external runtime requirements.

## How it fits in the package

- **Testing**: The test `TestPing` (located at line 230 of `icmp.go`) executes an ICMP ping between two containers and obtains a `PingResults`.  
  It then calls `results.String()` to print or log the outcome.
- **Reporting**: Any component that needs a quick textual snapshot of ping metrics can call this method, making test logs concise and readable.

## Suggested Mermaid diagram

```mermaid
flowchart TD
    A[Ping Test] --> B{Run ping}
    B --> C[PingResults]
    C --> D[PingResults.String()]
    D --> E[Formatted string]
```

This diagram shows the flow from running a ping to obtaining and serialising the results.
