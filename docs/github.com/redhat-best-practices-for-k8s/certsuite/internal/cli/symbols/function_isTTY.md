isTTY`

```go
func isTTY() bool
```

### Purpose  
`isTTY` determines whether the current process is attached to a terminal (TTY).  
The function is used by the CLI package to decide if it should emit ANSI escape codes, progress spinners, or other terminal‑specific output.  When `false`, the library falls back to plain text logs.

### Inputs / Outputs  

| Direction | Type | Description |
|-----------|------|-------------|
| **Return** | `bool` | `true` if standard input is a TTY; otherwise `false`. |

The function takes no arguments and returns only a boolean flag.

### Key Dependencies

* `os.Stdin.Fd()` – obtains the file descriptor for standard input.  
* `golang.org/x/crypto/ssh/terminal.IsTerminal(fd int)` (imported as `IsTerminal`) – checks if that file descriptor refers to an interactive terminal.

The code:

```go
fd := os.Stdin.Fd()
return IsTerminal(int(fd))
```

### Side Effects

None. The function is read‑only; it merely interrogates the operating system state and returns a value.

### How It Fits in the `cli` Package

* **User Experience** – Other parts of the CLI package call `isTTY()` before printing colored output (e.g., using the `Red`, `Green`, etc. constants) or before starting a progress ticker (`tickerPeriodSeconds`).  
* **Logging** – The `CliCheckLogSniffer` component uses this flag to decide whether to format logs with ANSI colors or plain text.

### Usage Example

```go
if isTTY() {
    fmt.Println(Red + "Running in terminal" + Reset)
} else {
    fmt.Println("Non‑interactive mode")
}
```

### Mermaid Diagram (Optional)

```mermaid
flowchart TD
    A[isTTY()] --> B{os.Stdin.Fd()}
    B --> C{IsTerminal(fd)}
    C -- true --> D[return true]
    C -- false --> E[return false]
```

> **Note**: The function is intentionally unexported because it’s an implementation detail of the CLI rendering logic.
