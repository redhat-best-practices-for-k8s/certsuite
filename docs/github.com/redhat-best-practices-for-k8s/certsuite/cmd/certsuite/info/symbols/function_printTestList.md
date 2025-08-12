printTestList` – Package‑Level Helper

| Item | Detail |
|------|--------|
| **Package** | `info` (`github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/info`) |
| **Visibility** | Unexported (used only inside this package) |
| **Signature** | `func([]string)()` – accepts a slice of test names and returns nothing. |

## Purpose

`printTestList` formats and writes the list of available tests to standard output.  
It is called by the *info* command when the user requests a listing of all tests that can be run with CertSuite.

## Parameters & Return Value

| Parameter | Type | Description |
|-----------|------|-------------|
| `tests` | `[]string` | Slice containing the names (identifiers) of tests. The slice is expected to already be sorted by the caller; this function only formats it. |

The function does **not** return a value – its sole side effect is printing.

## Behaviour & Output Format

1. **Header**  
   Prints a header line (`"Available Tests:"`) followed by a blank line for readability.

2. **Test Names**  
   Iterates over `tests`, printing each name on its own line using `fmt.Println`.  

3. **Footer**  
   After the list, prints a second blank line and then a formatted summary:  
   ```
   Total tests: <count>
   ```  
   where `<count>` is the number of elements in the slice.

The output is deliberately simple; it relies on standard library `fmt` functions (`Println`, `Printf`) and does not perform any further formatting (no padding or column alignment).

## Dependencies

- **Standard Library**  
  - `fmt.Print*` for console output.  

- **Package‑Level Globals**  
  None – the function is intentionally isolated from global state, making it easy to test.

## Side Effects

The only observable side effect is writing to `os.Stdout`. No internal state or globals are modified, so repeated calls produce deterministic output.

## Relationship in the Package

Within `info`, this helper supports the *info* command’s ability to list tests. The command handler gathers available tests (likely from a registry), then invokes `printTestList` before exiting. Because it is unexported, other packages cannot call it directly; they must use the public command interface.

---

### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[infoCmd] --> B{list-tests?}
    B -- yes --> C[collectTests()]
    C --> D[printTestList(tests)]
```

This visualizes how `printTestList` fits into the command flow.
