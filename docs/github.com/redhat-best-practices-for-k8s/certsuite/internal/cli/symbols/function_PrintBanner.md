PrintBanner`

```go
func PrintBanner() func()
```

### Purpose
`PrintBanner` displays the ASCII art banner that introduces CertSuite in the terminal.  
The function is intended to be called once during program startup (e.g., in `main()` or an init routine). It prints the banner immediately and returns a *cleanup* function that can be used to stop any background activity started by the banner display.

> **Note** – The actual implementation of the returned cleanup function is not shown in the snippet.  
> If the banner is animated (e.g., a spinner or ticker), the cleanup would likely signal the goroutine via `stopChan`. In the current file, the only side‑effect visible is a call to the standard library’s `Print` function.

### Inputs
None – the function has no parameters.

### Outputs
* Returns a value of type `func()`, i.e. a zero‑argument function that performs cleanup when called.  
  The caller should invoke this returned function (e.g., via `defer`) if any background goroutine is spawned by the banner logic.

### Key Dependencies & Globals
| Global | Type | Role |
|--------|------|------|
| `banner` | string | Holds the ASCII art that will be printed. |
| `lineLength`, `tickerPeriodSeconds` | int | Configuration values used when animating the banner (not directly referenced in this snippet). |
| `CliCheckLogSniffer` | *unknown* | Exported global; not used by `PrintBanner`. |
| `checkLoggerChan`, `stopChan` | channels | Not accessed directly by `PrintBanner`; likely used elsewhere to coordinate logging or termination. |

The function itself only calls the standard library’s `Print` (or possibly `fmt.Print`) to output the banner text.

### Side Effects
* Prints the ASCII art stored in `banner` to `stdout`.
* Returns a cleanup closure; if background goroutines are spawned, they can be terminated by invoking this returned function.

### How It Fits the Package
The `cli` package is responsible for all command‑line interactions.  
`PrintBanner` provides visual feedback when the tool starts, improving user experience.  
It is typically used early in the application's lifecycle and paired with other CLI utilities such as logging, progress indicators, or interactive prompts.

---

#### Suggested Mermaid Diagram

```mermaid
flowchart TD
    Start[Start Program] --> Banner[PrintBanner()]
    Banner --> Cleanup[Return cleanup func]
    Cleanup --> End[Program continues]
```

This diagram shows the flow from program start to banner printing and the eventual cleanup step.
