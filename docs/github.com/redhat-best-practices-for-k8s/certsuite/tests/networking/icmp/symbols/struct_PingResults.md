PingResults` – Result container for an ICMP ping

| Element | Type | Description |
|---------|------|-------------|
| `errors` | `int` | Number of packets that failed to reach the target (e.g., time‑outs, unreachable). |
| `outcome` | `int` | Status code returned by the underlying ping binary. 0 usually means success; non‑zero indicates a failure. |
| `transmitted` | `int` | Total number of ICMP echo requests sent. |
| `received` | `int` | Number of echo replies received back from the target. |

### Purpose
`PingResults` aggregates all quantitative information produced by an external ping command.  
It is used in two places:

1. **Parsing** – `parsePingResult()` extracts these fields from the raw output of a ping invocation (e.g., Linux `ping -c N <host>`).  
2. **Reporting** – `PingResults.String()` formats the data into a human‑readable summary for logs or test reports.

### Dependencies
* **`parsePingResult`** relies on regular expressions compiled via `regexp.MustCompile`. It uses:
  * `FindStringSubmatch` to pull numeric values from the ping output.
  * `strconv.Atoi` to convert those strings into integers.
  * `fmt.Errorf` for error handling when parsing fails.

* **`PingResults.String()`** depends on:
  * The standard library’s `fmt.Sprintf`.
  * A helper function `ResultToString` (not shown in the snippet) that probably converts the numeric fields into a concise string format.

### Side effects
The struct itself is immutable; it only holds data.  
All side effects happen outside:

* `parsePingResult()` returns an error if the ping output does not match expected patterns.
* `PingResults.String()` has no side effects beyond formatting.

### How it fits in the `icmp` package
```mermaid
graph TD;
    PingCommand[External ping command] -->|output string| parsePingResult();
    parsePingResult() --> PingResults;
    PingResults --> PingResults.String();
```

* The **ping command** is executed by higher‑level test logic (not shown here).  
* `parsePingResult()` turns the raw output into a typed struct.  
* Test assertions or logs can then use `PingResults` directly or call its `String()` method for human‑friendly output.

In summary, `PingResults` is a lightweight DTO that captures the essential statistics of an ICMP ping and provides convenient formatting for reporting within the networking tests.
