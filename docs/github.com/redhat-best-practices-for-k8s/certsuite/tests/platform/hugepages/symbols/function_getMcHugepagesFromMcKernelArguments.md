getMcHugepagesFromMcKernelArguments`

| | |
|-|-|
| **Package** | `hugepages` (github.com/redhat-best-practices-for-k8s/certsuite/tests/platform/hugepages) |
| **File** | `tests/platform/hugepages/hugepages.go` |
| **Line** | 297 |

### Purpose
Extracts the *HugePages* configuration that is embedded in a MachineConfig’s kernel arguments string and returns two pieces of information:

1. A map where each key is a huge‑page size (in KiB) and the value is the number of pages requested for that size.
2. The default huge‑page size (also in KiB) that will be used when a page request does not specify an explicit size.

The function is only used internally by tests that need to compare the MachineConfig’s desired huge‑pages state with what would actually be applied on the node.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `mc` | `*provider.MachineConfig` | The MachineConfig object from which kernel arguments are read. |

### Returns
| Value | Type | Meaning |
|-------|------|---------|
| `map[int]int` | map of huge‑page size → page count | Represents the parsed `hugepagesz=<size>` / `hugepages=<count>` pairs found in the kernel args. |
| `int` | default huge‑page size (KiB) | The value that will be used as the “default” if a request omits an explicit size. |

### Algorithm
1. **Split** the MachineConfig’s `KernelArguments` string on spaces to obtain individual key/value pairs.
2. Iterate over each pair:
   * Split the pair on `=`; only pairs with exactly two fields are processed.
   * If the key is `hugepagesz`, convert its value to an integer (KiB) using `Atoi`. This becomes the current “default” size for subsequent requests that don’t specify a size.
   * If the key is `hugepages`, parse the numeric part of the value. The associated size comes from the most recent `hugepagesz` seen; if none was set, use `DefaultHugepagesz`.
   * Store the count in the map under the resolved page‑size key.
3. **Logging**:  
   * If any pair is malformed (wrong split length) or a conversion fails, a warning is logged via `Warn`.  
   * After parsing, the final mapping and default size are logged with `logMcKernelArgumentsHugepages`.
4. Return the constructed map and default size.

### Dependencies
* **`strings.Split`** – to break the kernel‑argument string into key/value pairs.
* **`strconv.Atoi`** – numeric conversion of strings.
* **`hugepageSizeToInt`** – helper that turns a page‑size string (e.g., `"2M"`) into an integer number of KiB.
* **`Warn`** – logs warning messages when parsing fails.
* **`logMcKernelArgumentsHugepages`** – records the parsed mapping for debugging.

### Side Effects
The function is read‑only: it does not modify the MachineConfig or any global state. All output is returned to the caller; the only observable side effect is diagnostic logging.

### Role in the Package
`hugepages` contains logic that validates and tests huge‑page configuration on nodes.  
* `getMcHugepagesFromMcKernelArguments` supplies a low‑level parser that turns raw kernel arguments into structured data, which other test helpers then compare against the actual runtime state (e.g., `/proc/meminfo`, `sysctl`).  

By centralising parsing here, the package can consistently interpret huge‑page settings across all tests.
