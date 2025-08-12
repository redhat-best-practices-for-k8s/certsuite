showConfiguration`

| Item | Detail |
|------|--------|
| **Signature** | `func(showConfiguration(*configuration.TestConfiguration))()` |
| **Visibility** | Unexported (internal helper) |

### Purpose
Displays the current test configuration in a human‑readable format.  
It is invoked by the *generate* command when the user selects the “show” menu item.

### Parameters
- `cfg *configuration.TestConfiguration` – the configuration instance to print.  
  The caller passes the global `certsuiteConfig`, which holds all values gathered from prompts, defaults, and environment variables.

### Behaviour
1. **Marshal** – Uses `json.MarshalIndent` (implicit via the imported `"encoding/json"`) to convert `cfg` into a pretty‑printed JSON string.  
   *If marshaling fails, the error is ignored; the function continues with an empty string.*
2. **Output** – Prints three lines:
   - A header (`"+---------- show configuration ----------+"`)
   - The indented JSON representation
   - A footer (`"+---------------------------------------+"`)

All output goes to standard out via `fmt.Printf`/`Println`.

### Dependencies
- `encoding/json.MarshalIndent`
- `fmt.Printf`, `fmt.Println`

No other global state is modified.  
The function only reads the passed configuration and writes to stdout.

### Side‑effects
- Writes to the console; no file or network I/O.
- Does **not** modify the configuration object.

### Context within package
`config` provides a CLI for generating test suites.  
`showConfiguration` is one of several menu callbacks that operate on `certsuiteConfig`.  
It offers users an immediate visual confirmation of what will be written to disk by the *save* command.

---

#### Mermaid flow (optional)

```mermaid
flowchart TD
    A[User selects "Show"] --> B[showConfiguration(cfg)]
    B --> C{Marshal JSON}
    C -->|success| D[Print header]
    C -->|success| E[Print JSON]
    C -->|success| F[Print footer]
```

This diagram illustrates the linear path from menu selection to console output.
