hugepagesByNuma.String()` – Human‑readable representation of NUMA‑aware hugepage data

### Purpose
`String` implements the `fmt.Stringer` interface for the **`hugepagesByNuma`** type.  
It produces a compact, human‑friendly summary of how many hugepages of each size are allocated per NUMA node. The resulting string is intended to be used in debug or informational logs so that test results are easier to interpret.

### Receiver
```go
type hugepagesByNuma map[int]hugepagesSizeCounts
```
The receiver (`hugepagesByNuma`) maps a NUMA node index (int) to a `hugepagesSizeCounts` value, which itself is a mapping from hugepage size (in KiB) to the number of pages.

### Signature
```go
func (h hugepagesByNuma) String() string
```

- **Input:** none – it operates on the receiver.
- **Output:** a single `string`.

### Algorithm Overview

| Step | Operation |
|------|-----------|
| 1 | Initialise a buffer (`var buf bytes.Buffer`) to accumulate output. |
| 2 | Iterate over each NUMA node in ascending order (using `for numa, sizeCounts := range h`). |
| 3 | For each node, append the prefix `"NUMA <node>:"`. |
| 4 | Build a slice of strings describing each hugepage size and its count: for every `<size>:<count>` pair in `sizeCounts`, create a string formatted as `"<size>K:<count>"`. |
| 5 | Join these per‑node strings with commas, wrap them in square brackets, and append to the buffer. |
| 6 | After processing all nodes, return `buf.String()`. |

The implementation uses the standard library functions:

* `bytes.Buffer.WriteString` – for efficient string concatenation.
* `fmt.Sprintf` – to format numeric values into strings.
* The helper function `Ints` (not shown here) converts a slice of integers into a comma‑separated string, used when constructing the per‑node list.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `bytes.Buffer` | Accumulates output without repeated allocations. |
| `fmt.Sprintf` | Formats numbers and strings. |
| `Ints` (utility) | Turns integer slices into comma‑separated text. |

### Side Effects
* None – the function is pure: it reads from its receiver and writes only to a local buffer.

### How It Fits the Package

The **`hugepages`** package provides utilities for querying and asserting hugepage configuration on Linux hosts (e.g., reading `/proc/meminfo`, parsing kernel arguments).  
During tests, this package collects per‑NUMA node counts of hugepages and stores them in a `hugepagesByNuma`.  
When test results or diagnostics are printed, the `String` method is invoked implicitly by `fmt.Printf`/`log.Println`, yielding concise summaries like:

```
NUMA 0:[2K:8,4K:16]
NUMA 1:[2K:8,4K:16]
```

Thus, `hugepagesByNuma.String()` is the package’s “pretty‑printer” that translates internal data structures into readable log output.
