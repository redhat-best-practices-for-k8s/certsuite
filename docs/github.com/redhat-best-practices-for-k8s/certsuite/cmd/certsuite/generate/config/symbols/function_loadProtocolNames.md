loadProtocolNames`

| | |
|-|-|
| **Package** | `config` (github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/config) |
| **Visibility** | unexported (`func`) |
| **Signature** | `func([]string)()`. It takes a slice of strings and returns a zero‑argument function. |

### Purpose

`loadProtocolNames` is used during the interactive configuration wizard to initialise the *protocol names* field in the internal `certsuiteConfig` structure.  
The wizard collects protocol names from the user as a comma‑separated string, splits it into a slice of strings and then passes that slice to this helper.

### How it works

1. **Input** – a slice of protocol name strings (`[]string`).  
   The caller already parsed the user input (e.g. `"http,https"` → `[]string{"http","https"}`).

2. **Return value** – a closure with signature `func()`.  
   This closure is intended to be executed later by the wizard infrastructure when it is time to commit the configuration values.

3. **Side effect** – inside the returned function the slice is assigned to the `ProtocolNames` field of the global `certsuiteConfig` variable.  
   No other state changes occur; no external resources are touched.

4. **Error handling** – The helper itself does not return an error; it assumes that the supplied slice is already validated by earlier wizard steps. If the slice were empty, the assignment would simply set an empty list in the configuration.

### Dependencies

| Dependency | Role |
|------------|------|
| `certsuiteConfig` (global) | Holds all configuration values being built up during the wizard run. The closure mutates its `ProtocolNames` field. |
| `templates` / `generateConfigCmd` | Not directly used by this function; they are part of the same package and may call `loadProtocolNames`. |

### Placement in the package

The `config` package implements an interactive CLI for generating a CertSuite configuration file.  
Each menu step returns a closure that applies the user’s choice to `certsuiteConfig`.  
`loadProtocolNames` is one such step, responsible specifically for populating the protocol names list before the final configuration is written out.

### Example flow

```go
// In wizard code:
protocols := strings.Split(userInput, ",") // e.g. []string{"http","https"}
applyProtocols := loadProtocolNames(protocols)

// Later, when the wizard reaches the “commit” point:
applyProtocols()   // sets certsuiteConfig.ProtocolNames = protocols
```

This pattern keeps the wizard logic declarative: each step supplies a simple mutation closure that is executed at commit time.
