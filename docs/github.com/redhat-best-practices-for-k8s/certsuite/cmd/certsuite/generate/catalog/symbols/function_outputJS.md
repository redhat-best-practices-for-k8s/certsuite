outputJS` – Internal helper for JSON catalog export

```go
func outputJS()()
```

| Feature | Detail |
|---------|--------|
| **Purpose** | Serialises the command‑level configuration (`generateCmd`) into pretty‑printed JSON and writes it to standard output. It is invoked when the `--output js` flag is supplied for the *generate catalog* sub‑command. |
| **Inputs** | None – the function operates on the package‑wide variable `generateCmd`. |
| **Outputs** | The function does not return a value; its side effect is writing JSON to `stdout`. |
| **Key dependencies** | • `json.MarshalIndent` – formats the data structure.<br>• `fmt.Printf` – emits the formatted string.<br>• `log.Error` (via the package’s logger) – reports marshalling failures. |
| **Side effects** | 1. Calls `json.MarshalIndent(generateCmd, "", "    ")`.<br>2. If marshalling fails, logs an error and exits the program (`log.Fatal`).<br>3. On success prints the JSON string followed by a newline to `stdout`. |
| **Package context** | The function lives in `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/catalog`. It is used exclusively by the *generate catalog* command when the user requests output in JavaScript/JSON format. The surrounding CLI infrastructure (`cobra` commands such as `generateCmd`, `markdownGenerateClassification`, etc.) registers this function as a handler for the `--output js` flag. |

### How it fits

```
generate
 ├─ catalog.go          ← defines generateCmd and outputJS
 │   └─ outputJS()      ← called when --output js is set
 └─ other sub‑commands
```

When the user runs:

```bash
certsuite generate catalog --output js
```

the CLI parses flags, sets `generateCmd` appropriately, then invokes `outputJS`. The function emits a JSON representation of the entire command configuration (including any parsed arguments or defaults), allowing downstream tooling to consume it programmatically.

> **Note**: No other part of the package calls `outputJS`; it is intentionally unexported because it is an implementation detail tied to the CLI flag handling.
