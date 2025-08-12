getGrubKernelArgs` – Boot‑Parameter Extraction for GRUB

**File:** `tests/platform/bootparams/bootparams.go`  
**Package:** `bootparams`

---

## Purpose
`getGrubKernelArgs` extracts the kernel command line arguments that were passed to a Linux system via the **GRUB** boot loader.  
The function is used by tests that need to verify that particular kernel options are present (or absent) after a container image has been booted.

It works by:

1. Running `grub2-editenv list` inside the target container to get GRUB’s environment variables.
2. Parsing the output for a line beginning with the key defined in `grubKernelArgsCommand`.
3. Converting that line into a map of argument names → values (e.g., `"root=UUID=1234" -> {"root":"UUID=1234"}`).

The resulting map is returned to the caller, along with any error encountered.

---

## Signature

```go
func getGrubKernelArgs(env *provider.TestEnvironment, container string) (map[string]string, error)
```

| Parameter | Type                 | Description |
|-----------|----------------------|-------------|
| `env`     | `*provider.TestEnvironment` | Holds the test context and a client holder that can run commands inside containers. |
| `container` | `string`              | The name (or ID) of the container in which GRUB’s environment should be inspected. |

| Return value | Type            | Description |
|--------------|-----------------|-------------|
| `map[string]string` | Map of kernel argument names to their values. |
| `error` | Non‑nil if any step fails (e.g., command execution error, parsing failure). |

---

## Key Dependencies

| Dependency | Role |
|------------|------|
| `GetClientsHolder()` | Retrieves the holder that provides access to containers for executing commands. |
| `NewContext(env)` | Builds a context object used by `ExecCommandContainer`. |
| `ExecCommandContainer(ctx, container, cmd…)` | Runs a shell command inside the specified container and returns its stdout/stderr. |
| `Errorf(...)` | Formats errors returned to callers. |
| `Split(s, sep)` | Splits strings (used for parsing lines). |
| `FilterArray(arr, prefix)` | Filters an array of strings that start with a given prefix (`HasPrefix`). |
| `ArgListToMap(args []string)` | Converts a slice like `["root=foo", "rw"]` into a map. |

---

## Algorithm (in prose)

1. **Run GRUB command**  
   ```bash
   grub2-editenv list
   ```
   inside the target container.

2. **Locate kernel‑args line**  
   Search the output for a line that starts with `grubKernelArgsCommand` (`"GRUB_CMDLINE_LINUX_DEFAULT="`).  
   If none found → return an empty map and no error.

3. **Extract argument list**  
   * Remove the prefix and surrounding quotes.
   * Split on spaces to get individual arguments.

4. **Map arguments**  
   Convert the slice into a `map[string]string` where each key is the part before `=` (or the whole string if no `=`).

5. **Return map & error**  

---

## Side Effects

* Executes an external command inside a container; may block until the command finishes.
* No state is mutated outside of the returned map; all side effects are confined to I/O.

---

## Integration in the Package

`bootparams` contains helpers for retrieving boot‑parameters from various boot loaders (GRUB, UEFI, etc.).  
`getGrubKernelArgs` is the implementation specific to GRUB and is used by higher‑level tests such as:

```go
func TestRootUUID(t *testing.T) {
    env := setupTestEnv()
    args, err := getGrubKernelArgs(env, "mycontainer")
    // assert that "root" key exists, etc.
}
```

The function keeps the package logic decoupled from container orchestration details by delegating command execution to `provider.TestEnvironment`.

---

## Mermaid Diagram (suggestion)

```mermaid
flowchart TD
  A[Call getGrubKernelArgs] --> B[ExecCommandContainer("grub2-editenv list")]
  B --> C{Output contains "GRUB_CMDLINE_LINUX_DEFAULT="?}
  C -- No --> D[Return empty map]
  C -- Yes --> E[Extract args string]
  E --> F[Split on spaces → args slice]
  F --> G[ArgListToMap → result map]
  G --> H[Return map, nil error]
```

---
