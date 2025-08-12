printTestCaseInfoBox`

The function draws a stylized information box for a single test case on the command‑line interface of *certsuite*.  
It is used by the `info` sub‑command when rendering a list of test cases.

```go
func printTestCaseInfoBox(tc *claim.TestCaseDescription) {
    ...
}
```

## Purpose

*Format and display the metadata of a single test case in a box that:
- shows the test’s name, description, and category,
- highlights key fields (e.g., ID, status),
- wraps long text to fit the terminal width.*

The function is intentionally **read‑only** – it never mutates its argument.

## Inputs

| Parameter | Type                     | Description |
|-----------|--------------------------|-------------|
| `tc`      | `*claim.TestCaseDescription` | The test case description struct. All of its fields are read for display.*

> *If the pointer is nil, the function panics – callers must guard against that.*

## Output

The function writes formatted text to `stdout`.  
It does not return a value.

## Key Dependencies (called helpers)

| Helper | Purpose |
|--------|---------|
| `Repeat`  | Generates a string consisting of repeated characters (used for borders). |
| `Println` | Writes a line followed by a newline. |
| `Printf`  | Formats a string and writes it to stdout. |
| `LineColor`, `LineAlignCenter`, `LineAlignLeft` | Apply ANSI colour codes and alignment helpers defined elsewhere in the package. |
| `WrapLines` | Splits long strings into multiple lines that fit within `lineMaxWidth`. |

These helpers are all local to the `info` package; they wrap standard library printing functions with colour/formatting logic.

## Side Effects

- **Console output** – the function writes directly to `stdout`; no other state is modified.
- **Terminal width handling** – it uses the global variable `lineMaxWidth` to wrap text appropriately. This variable is set elsewhere (e.g., during command initialisation) based on terminal size.

## Package Integration

The `info` package implements a sub‑command that lists all test cases in a human‑readable form.  
`printTestCaseInfoBox` is called for each element of the list, producing a separate box per case.  

A simplified call flow:

```
main() → infoCmd → run() → listTestCases()
   │
   └─> for each tc: printTestCaseInfoBox(tc)
```

The package also defines `infoCmd` (the Cobra command) and `lineMaxWidth`, but those are not directly manipulated by this function.

## Example Output

```
──────────────────────
│      Test Case 001    │
├──────────────────────┤
│ ID:   001             │
│ Name:  Sample Check   │
│ Category: Security   │
├──────────────────────┤
│ Description:
│ This test verifies that...
└──────────────────────┘
```

*The actual colours and alignment are produced by the helper functions.*

---

**Summary** – `printTestCaseInfoBox` is a pure output routine that formats a single `TestCaseDescription` into a bordered, colour‑enhanced box, using the package’s helper utilities for alignment, wrapping, and terminal width handling.
