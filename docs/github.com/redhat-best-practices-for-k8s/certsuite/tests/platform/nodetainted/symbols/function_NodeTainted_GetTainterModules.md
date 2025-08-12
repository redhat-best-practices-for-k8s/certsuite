# `GetTainterModules`

**Package:** `nodetainted`  
**Receiver type:** `NodeTainted`  
**Signature**

```go
func (n NodeTainted) GetTainterModules(allowlist map[string]bool) (map[string]string, map[int]bool, error)
```

---

## Purpose

`GetTainterModules` discovers which kernel modules running on a node are responsible for setting *kernel taint bits*.  
It returns:

| Return value | Meaning |
|--------------|---------|
| `tainters`   | `map[moduleName]string` – for each module, the letters that represent the taint bits it has set. The letters are decoded from the kernel’s `/proc/sys/kernel/tainted_modules` output via `DecodeKernelTaintsFromLetters`. Modules present in the supplied `allowlist` are **excluded** from this map. |
| `taintBits`  | `map[int]bool` – a set of all taint‑bit positions that have been set by any module, including those on the allowlist. The bits correspond to the kernel’s bitmask representation. |
| `error`      | Non‑nil if any step fails (command execution, parsing, or decoding). |

The function is typically used in tests to verify that no *unapproved* modules are tainting a node.

---

## Inputs

| Parameter | Type | Description |
|-----------|------|-------------|
| `allowlist` | `map[string]bool` | A set of module names that should be ignored when reporting tainters. The map’s keys are module names; the value is unused (`true`). |

The receiver (`NodeTainted`) holds connection details and a command‑execution helper (`runCommand`).

---

## Workflow

1. **Collect all modules**  
   Calls `getAllTainterModules()` to run a shell command on the node that lists every loaded module along with its taint letters (e.g., `modprobe -l`). Errors are wrapped with `Errorf`.

2. **Decode taint letters**  
   For each module, it uses `DecodeKernelTaintsFromLetters` to translate the human‑readable letter string into a bitmask (`uint64`). If decoding fails, an error is returned.

3. **Debug logging**  
   Two `Debug()` calls provide runtime diagnostics:
   * First logs the raw mapping from modules to letters.
   * Second logs the decoded bit masks for each module.

4. **Aggregate bits**  
   Calls `GetTaintedBitsByModules` (which internally aggregates bit positions across all modules) and merges them into `taintBits`.

5. **Filter allowlisted modules**  
   Builds the final `tainters` map by iterating over the decoded results and excluding any module whose name appears in `allowlist`.

6. **Return**  
   Returns `(tainters, taintBits, nil)` on success; otherwise returns an error.

---

## Key Dependencies

| Dependency | Role |
|------------|------|
| `getAllTainterModules` | Executes node command to retrieve raw module‑taint data. |
| `DecodeKernelTaintsFromLetters` | Converts taint letter strings into bit masks. |
| `GetTaintedBitsByModules` | Aggregates bits across modules for the final set of taint bits. |
| `runCommand` (global) | Low‑level helper to run shell commands on the node. |
| `Debug`, `Errorf` | Logging utilities from the test framework. |

---

## Side Effects

* No state is modified in `NodeTainted`; all operations are read‑only.
* The function logs diagnostic information via `Debug`.
* It may return an error if any command fails or data cannot be parsed.

---

## Integration Context

Within the **nodetainted** test suite, this method allows tests to assert that:

1. No *unknown* modules are tainting a node (`tainters` should be empty after applying the allowlist).  
2. The overall set of taint bits matches expectations (`taintBits` can be compared against a known mask).

It is typically called from higher‑level test helpers such as `CheckTainters()` or `ValidateNodeState()`.
