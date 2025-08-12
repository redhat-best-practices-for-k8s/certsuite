cliCheckLogSniffer` – Internal Slog Writer

| Feature | Details |
|---------|---------|
| **Visibility** | Unexported (`cliCheckLogSniffer`) – only used within the `internal/cli` package. |
| **Role** | Acts as a lightweight implementation of `io.Writer` that feeds check‑output into the CLI logger (`slog`). It is passed to `slog.NewHandler(..., slog.HandlerOptions{Writer: &cliCheckLogSniffer{}})` so that every log record emitted by the checks ends up on the terminal. |
| **State** | No exported or unexported fields – the struct is *stateless*. The receiver exists only to give a type name for the `Write` method, enabling Go’s interface implementation without storing any configuration. |

### `Write(p []byte) (int, error)`

* **Purpose** – Convert raw log bytes into user‑friendly output on the console.
* **Inputs**
  * `p`: byte slice produced by the check logic or slog formatter.
* **Outputs**
  * Returns the number of bytes written (`len(p)`) and an `error` (always `nil` in the current implementation).
* **Key dependencies & side‑effects**
  * Calls `isTTY()` to determine if the current output stream is a terminal.  
    * If true, it may prepend ANSI colour codes or format the text for better readability.
  * Uses standard library functions:  
    * `len(p)` – count of bytes processed.  
    * `string(p)` – conversion from byte slice to string for printing.  
    * Another `len` call is used when writing to the underlying stream (e.g., `os.Stdout.WriteString(...)`).
  * Writes directly to the terminal (`os.Stdout`) – side effect of producing visible output.
* **Behavioural notes**
  * Because the struct has no fields, all state‑related decisions (like colour support) are derived from external helpers (`isTTY`).  
  * The implementation is intentionally minimal; it does not buffer or modify the log payload beyond terminal detection.

### Relationship to the rest of the package

* **`cli.go`** – The file where `cliCheckLogSniffer` is declared and used. It supplies this writer to the logging pipeline that drives the command‑line interface.
* **Checks execution** – When a check runs, it emits logs via slog; those logs are routed through `cliCheckLogSniffer.Write`, which surfaces them on the user’s terminal.
* **No persistence** – The struct does not write to files or network streams; it is strictly for CLI visibility.

---

#### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[Checks] -->|log via slog| B[slog Handler]
    B -->|uses Writer| C[cliCheckLogSniffer]
    C -->|Write() | D[stdout (TTY)]
```

This diagram illustrates the data flow from a check to the terminal through `slog` and our custom writer.
