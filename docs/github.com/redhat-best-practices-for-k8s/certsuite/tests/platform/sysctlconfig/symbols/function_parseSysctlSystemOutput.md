parseSysctlSystemOutput`

| Item | Detail |
|------|--------|
| **Package** | `sysctlconfig` – a helper for the CertSuite test harness that validates system‑level sysctl settings. |
| **Signature** | `func parseSysctlSystemOutput(output string) map[string]string` |
| **Exported?** | No – it is an internal utility used only within this package. |

### Purpose
`parseSysctlSystemOutput` takes the raw text returned by executing `sysctl --system` and turns it into a convenient lookup table of *final* key‑value pairs.

When `sysctl --system` runs, it iterates over all sysctl configuration files in order of priority (e.g. `/etc/sysctl.conf`, `/usr/lib/sysctl.d/*.conf`, etc.). Each file may override keys defined earlier. The command prints the **effective** value for each key once the entire cascade has been processed.  
The function parses that output, discarding comments and blank lines, and returns a `map[string]string` where:

* **key** – the sysctl name (e.g. `"kernel.pid_max"`)
* **value** – the effective string value as printed by the command.

This map is later compared against expected values defined in test data to assert that the system’s sysctl configuration matches policy.

### Inputs
| Parameter | Type | Notes |
|-----------|------|-------|
| `output` | `string` | Multiline text exactly as produced by `sysctl --system`. |

### Outputs
| Return | Type | Description |
|--------|------|-------------|
| map[string]string | The mapping of sysctl keys to their resolved values. If a line cannot be parsed, it is ignored silently. |

### Core Logic

```go
func parseSysctlSystemOutput(output string) map[string]string {
    // Initialise an empty map.
    result := make(map[string]string)

    // Split the entire output into individual lines.
    for _, l := range strings.Split(output, "\n") {
        // Skip blank lines and comment lines (those starting with '#').
        if l == "" || strings.HasPrefix(l, "#") {
            continue
        }

        // Regex to capture "key = value" pairs.
        //   ^([a-zA-Z0-9_.]+)\s*=\s*(.*)$
        re := regexp.MustCompile(`^([a-zA-Z0-9_.]+)\s*=\s*(.*)$`)
        if !re.MatchString(l) {
            continue
        }

        // Extract key and value.
        submatches := re.FindStringSubmatch(l)
        result[submatches[1]] = submatches[2]
    }
    return result
}
```

#### Step‑by‑step

1. **`make(map[string]string)`** – Creates the empty map that will be returned.  
2. **`strings.Split(output, "\n")`** – Breaks the multi‑line output into individual lines for processing.  
3. **`strings.HasPrefix(l, "#")`** – Detects comment lines; they are ignored.  
4. **`regexp.MustCompile(...)`** – Compiles a pattern that matches the canonical `key = value` format used by sysctl.  
5. **`re.MatchString(l)`** – Quick sanity check to skip malformed lines.  
6. **`re.FindStringSubmatch(l)`** – Captures the key and value groups; they are stored in the map.

### Dependencies

| Function | Package | Role |
|----------|---------|------|
| `make` | built‑in | Allocate map. |
| `strings.Split` | `strings` | Split output into lines. |
| `strings.HasPrefix` | `strings` | Detect comment lines. |
| `regexp.MustCompile` | `regexp` | Compile regex once per call. |
| `re.MatchString` | `regexp` | Test if a line matches the expected pattern. |
| `re.FindStringSubmatch` | `regexp` | Extract key/value groups. |

### Side Effects
None – the function is pure: it only reads its argument and returns a new map.

### How It Fits the Package

The `sysctlconfig` package contains tests that verify system sysctl values against a policy definition.  
* The test runner executes `sysctl --system`.  
* The raw output is passed to `parseSysctlSystemOutput`, producing a map of effective settings.  
* Test logic then compares this map with the expected key/value pairs loaded from test fixtures.

Thus, `parseSysctlSystemOutput` acts as the bridge between the shell command and the Go‑based assertion logic.
