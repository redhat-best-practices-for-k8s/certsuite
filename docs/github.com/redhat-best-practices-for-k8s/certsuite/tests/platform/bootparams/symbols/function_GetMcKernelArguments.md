GetMcKernelArguments`

**Location**

`github.com/redhat-best-practices-for-k8s/certsuite/tests/platform/bootparams/bootparams.go:69`

```go
func GetMcKernelArguments(env *provider.TestEnvironment, target string) map[string]string
```

---

## Purpose

`GetMcKernelArguments` extracts the kernel command‑line arguments for a specific **Machine Config (MC)** target from a test environment.  
The function is used by tests that need to verify or manipulate boot parameters that are stored in the `TestEnvironment`.

---

## Parameters

| Name   | Type                         | Description |
|--------|------------------------------|-------------|
| `env`  | `*provider.TestEnvironment` | The test environment that holds all boot‑parameter data. |
| `target` | `string` | Identifier of the MC whose arguments should be returned (e.g., `"worker"` or `"master"`). |

---

## Return Value

| Type | Description |
|------|-------------|
| `map[string]string` | A map where each key is a kernel argument name and its value is the associated string. If no arguments are found for the target, an empty map is returned. |

---

## Key Steps & Dependencies

1. **Lookup**  
   The function retrieves the raw command‑line string stored in `env.KernelArguments[target]`. This field is part of the `TestEnvironment` struct and holds a space‑separated list of arguments.

2. **Parsing**  
   It calls the helper `ArgListToMap` (defined elsewhere in the package) to split the raw string into individual `key=value` pairs and convert them into a Go map.

3. **Return**  
   The resulting map is returned directly; no further manipulation occurs.

---

## Side Effects

* None – the function performs only read‑only operations on the supplied environment.
* It does not modify global state or any fields of `env`.

---

## Package Context

`bootparams` provides utilities for handling kernel boot parameters in certsuite tests.  
Other functions in the package convert between string lists and maps, validate arguments, and expose constants like:

```go
const (
    grubKernelArgsCommand = "grub2-mkconfig" // not used here but part of the same domain
    kernelArgscommand     = "kernel-args"
)
```

`GetMcKernelArguments` is a central helper that other tests call when they need to examine or assert specific boot‑parameter settings for a given machine configuration.

---

### Suggested Mermaid Flow

```mermaid
flowchart TD
  A[Call GetMcKernelArguments(env, target)] --> B{Lookup env.KernelArguments[target]}
  B -->|Found| C[Raw string]
  C --> D[ArgListToMap(raw)]
  D --> E[Return map[string]string]
```

This diagram illustrates the linear path from input to output.
