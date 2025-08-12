LabelsMatch` – Policy Label Selector Helper  

**Package:** `github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/policies`  
**Location:** `tests/networking/policies/policies.go:60`

---

## Purpose
`LabelsMatch` determines whether a Kubernetes *LabelSelector* (`v1.LabelSelector`) fully matches a set of labels represented by a Go map.  
It is used in tests to validate that resources (e.g., Pods, Services) satisfy the label requirements defined by a policy.

## Signature
```go
func LabelsMatch(selector v1.LabelSelector, labels map[string]string) bool
```

| Parameter | Type                  | Description |
|-----------|-----------------------|-------------|
| `selector`| `v1.LabelSelector`    | The selector to evaluate.  It may contain `MatchLabels`, `MatchExpressions`, or be empty. |
| `labels`   | `map[string]string`   | A map of key‑value label pairs from a resource. |

### Returns
- **`bool`** – `true` if the provided labels satisfy *all* conditions in the selector; otherwise `false`.

## Key Dependencies

| Dependency | Role |
|------------|------|
| `v1.LabelSelector` (K8s API) | Defines label selection logic. |
| `Size` function | Used to check if any match expression is present. |

> **Note:** The implementation only checks the `MatchLabels` field; it does not support `MatchExpressions`. If `selector.MatchLabels` is empty, the function immediately returns `true`.

## Implementation Overview

```go
func LabelsMatch(selector v1.LabelSelector, labels map[string]string) bool {
    // Empty selector matches everything.
    if len(selector.MatchLabels) == 0 && Size(selector.MatchExpressions) == 0 {
        return true
    }

    // Each key/value in MatchLabels must exist and equal the corresponding value.
    for k, v := range selector.MatchLabels {
        if val, ok := labels[k]; !ok || val != v {
            return false
        }
    }

    // No support for MatchExpressions – they are ignored.
    return true
}
```

1. **Empty Selector Check**  
   If both `MatchLabels` and `MatchExpressions` are empty, the selector matches any label set.

2. **Label Matching Loop**  
   Iterates over all key/value pairs in `selector.MatchLabels`.  
   - If a key is missing from `labels`, or its value differs, return `false`.  
   - Otherwise continue checking.

3. **Return Success**  
   After the loop, if no mismatches were found, return `true`.

## Side Effects & Constraints
- The function performs only read‑only operations on its inputs; it has no side effects.
- It ignores any `MatchExpressions` (unsupported in this implementation).  
  If an expression is present, the result may be incorrect unless the selector contains only `MatchLabels`.
- The function assumes that the caller supplies a valid `v1.LabelSelector`.

## How It Fits the Package
Within the *policies* test package, `LabelsMatch` is a small utility used by higher‑level tests that verify whether a generated policy correctly selects intended resources. By abstracting label matching logic into this helper, test cases remain concise and focused on policy semantics rather than repetitive selector evaluation.

---
