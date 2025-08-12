SubSlice` – Slice‑containment helper

**Package:** `stringhelper`  
**File:** `pkg/stringhelper/stringhelper.go` (line 41)  
**Exported:** yes  

---

### Purpose
`SubSlice` determines whether *every* element of a candidate slice is present in a reference slice.  
It is essentially a set‑subset test for string slices.

> **Use case** – e.g., checking that all required labels exist on a Kubernetes object.

---

### Signature

```go
func SubSlice(candidate []string, reference []string) bool
```

| Parameter | Type        | Meaning                                          |
|-----------|-------------|--------------------------------------------------|
| `candidate` | `[]string` | Slice whose elements we want to verify are in `reference`. |
| `reference` | `[]string` | Slice that may contain all or some of the candidate’s elements. |

**Return value**

- `true` – every element of `candidate` is found somewhere in `reference`.
- `false` – at least one element of `candidate` is missing from `reference`.

---

### Key dependencies

| Dependency | Role |
|------------|------|
| `StringInSlice` (internal helper) | Performs a single‑element membership test. It returns `true` if the given string exists in a slice, otherwise `false`. |

The function loops over `candidate`, calling `StringInSlice(candidate[i], reference)` for each element.  
If any call returns `false`, the loop aborts early and the function returns `false`.

---

### Side effects

- **None** – The function only reads from its arguments; it does not modify either slice or any global state.

---

### How it fits the package

`stringhelper` provides small, reusable utilities for manipulating string slices.  
`SubSlice` complements other helpers such as:

| Helper | Typical task |
|--------|--------------|
| `StringInSlice` | Check single element membership |
| `StringJoin` | Concatenate slice elements |
| `UniqueStrings` | Remove duplicates |

Together they allow callers to perform common set‑operations without pulling in a heavy external library.

---

### Example

```go
required := []string{"app", "env"}
labels   := []string{"app", "env", "tier"}

if SubSlice(required, labels) {
    fmt.Println("All required labels are present")
}
```

The call returns `true` because every string in `required` is found within `labels`.

---

### Mermaid diagram (optional)

```mermaid
graph TD
  A[SubSlice(candidate, reference)]
  B[StringInSlice(element, reference)]
  A -->|for each element| B
  B -- true --> A
  B -- false --> A
```

This visualises the linear scan performed by `SubSlice`.
