NodeTainted.getAllTainterModules`

| Item | Details |
|------|---------|
| **Package** | `nodetainted` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/platform/nodetainted`) |
| **Receiver type** | `NodeTainted` – the struct that represents a node under test. |
| **Signature** | `func (nt NodeTainted) getAllTainterModules() (map[string]string, error)` |

### Purpose
Collects all *taint‑module* names installed on the target node and maps each module to its current kernel taint value.  
The function is used by the test harness when verifying that a node’s taints match expected values.

### Inputs / Outputs

| Direction | Parameter | Type | Notes |
|-----------|-----------|------|-------|
| **Input** | *none* | – | The method relies solely on the state of the receiver (`nt`) and system commands. |
| **Output** | `map[string]string` | map module → taint value | A dictionary where keys are taint‑module names (e.g., `"intel_turbo"`) and values are their corresponding kernel taint strings (e.g., `"TINT"`). |
| | `error` | error | Non‑nil if the command fails, parsing errors occur, or expected output is missing. |

### Key Steps & Dependencies

1. **Command execution**  
   * Calls the helper `runCommand(nt.ctx, "taint", "-l")`.  
     - `runCommand` runs an external binary (`taint`) in the node’s context and returns its stdout as a string or an error.

2. **Parsing the output**  
   The command produces lines of the form:  
   ```
   <module> <value>
   ```  
   * Each line is split on whitespace using `strings.Split`.  
   * Lines with fewer than two parts are ignored (robustness against empty lines).

3. **Building the map**  
   For every valid pair, an entry `<module> → <value>` is added to a newly created map.

4. **Error handling**  
   * If the command itself fails, `Errorf` wraps and returns that error.  
   * If the output contains no module entries (`len(modules) == 0`), an explicit error is returned indicating the absence of taint‑modules.

### Side Effects

- No global state is modified; only local variables are created.
- The function may block while waiting for the external command to finish.

### Integration with the Package

`getAllTainterModules` is a private helper used by other public test functions (e.g., `NodeTainted.checkTaintValues`) to:

1. Retrieve current taint information from the node.
2. Compare it against expected taints defined in the test suite.

Because it depends on the external `taint` binary, its correctness is crucial for accurate taint verification across nodes of different kernel versions or custom taint modules.

---

#### Suggested Mermaid diagram (optional)

```mermaid
flowchart TD
  A[runCommand] --> B[stdout string]
  B --> C{split lines}
  C --> D[parse <module> <value>]
  D --> E[map[module]=value]
  E --> F{return map, nil}
```

This diagram visualizes the flow from executing `taint -l` to producing the final module‑to‑taint mapping.
