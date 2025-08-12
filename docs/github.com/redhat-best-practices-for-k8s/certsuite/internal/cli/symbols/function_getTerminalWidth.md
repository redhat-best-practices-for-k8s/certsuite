getTerminalWidth` (internal helper)

| Aspect | Detail |
|--------|--------|
| **Signature** | `func() int` |
| **Visibility** | Unexported – used only inside the `cli` package. |
| **Purpose** | Return the current terminal’s column width in characters. The value is used to wrap status‑update lines and ensure progress bars fit on the screen without wrapping. |
| **Inputs / Outputs** | No parameters; returns an `int`. If the terminal size cannot be determined, it falls back to a default of 80 columns (the classic terminal width). |
| **Key Dependencies** | *`golang.org/x/term.GetSize(int)`* – reads the terminal dimensions from the file descriptor. <br>*`os.Stdout.Fd()`* – obtains the stdout file descriptor for `GetSize`. <br>*`int(...)` conversion* – converts the returned width to an `int`. |
| **Side‑effects** | None. The function only queries the OS; it does not modify any global state or perform I/O beyond the terminal query. |
| **How it fits the package** |  
The CLI displays animated check‑result messages (e.g., “Running…”, “Pass”, “Fail”) and a progress bar that must adapt to the current terminal size. `getTerminalWidth` is called whenever the program needs to re‑calculate line lengths—for example, when handling terminal resize events or before printing status updates. It ensures that output remains tidy regardless of the user’s terminal window size. |

> **Mermaid diagram (optional)**  
> ```mermaid
> flowchart TD
>     A[CLI] -->|Print status| B[getTerminalWidth]
>     B --> C{GetSize}
>     C --> D[Return width or 80]
> ```
> This helper is a small but essential piece that keeps the user interface responsive and readable.
