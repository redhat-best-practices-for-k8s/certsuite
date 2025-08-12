getAnswer`

```go
func getAnswer(prompt string, help string, defaultValue string) []string
```

### Purpose

`getAnswer` is a small helper that interacts with the user through standard input/output while collecting configuration values during the *certsuite generate* command.  
It prints a prompt (styled in cyan), optionally displays help text, reads a line of input from `os.Stdin`, splits it into tokens and returns those tokens as a slice of strings.

The function is intentionally simple – it does **not** perform validation or persistence; it merely captures raw user input for further processing by the caller.

### Inputs

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `prompt`  | `string` | Text shown to the user before the input field. Typically a short description of the configuration option (e.g., `"Enter namespace:"`). |
| `help`    | `string` | Optional help text that is displayed immediately after the prompt. If empty, no help is printed. |
| `defaultValue` | `string` | A string that will be appended to the input line in a lighter colour, showing the user what value would be used if they press *Enter* without typing anything. It is **not** automatically inserted into the returned slice; the caller must handle defaults. |

### Output

A `[]string` containing the tokens read from the user’s input.  
The string is split on whitespace using `strings.Split`.  
If the user simply presses Enter, the slice will be empty.

```go
// Example: user types "foo bar"
[]string{"foo", "bar"}
```

### Key Dependencies

| Called function | Role |
|-----------------|------|
| `HiCyanString`, `CyanString`, `WhiteString` | ANSI‑styled output for the prompt, help and default value. |
| `Print`, `Printf` | Emit formatted text to stdout. |
| `bufio.NewScanner(os.Stdin)` + `Scan()` | Read a single line from the terminal. |
| `Split`, `TrimSpace`, `Text` | Parse the raw input into tokens. |

The function does **not** depend on any global state; it only uses standard library packages.

### Side‑effects

* Writes to standard output (`os.Stdout`) – the prompt, optional help and default value.
* Reads from standard input (`os.Stdin`) – blocks until a line is entered or EOF occurs.
* No modification of package variables or configuration objects.

### How it fits the package

The `config` sub‑package implements an interactive wizard for creating a CertSuite configuration file.  
Each step in the wizard asks the user for one or more values (e.g., namespace, operator labels).  
Those steps call `getAnswer` to present a prompt and capture the input. The returned slice is then processed by higher‑level logic that validates, defaults, and ultimately writes the configuration to disk.

Because `getAnswer` is pure and has no hidden side effects, it can be unit‑tested easily by mocking `os.Stdin`/`os.Stdout`. It also keeps the wizard code focused on business logic while delegating all terminal I/O to this helper.
