isMultiNamespacedOperator`

| Item | Detail |
|------|--------|
| **Package** | `operator` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/operator`) |
| **Visibility** | Unexported (used only inside this package) |
| **Signature** | `func isMultiNamespacedOperator(operatorName string, nsList []string) bool` |
| **Position in source** | `/Users/deliedit/dev/certsuite/tests/operator/helper.go:114` |

---

## Purpose

`isMultiNamespacedOperator` determines whether a given operator (identified by its name) is configured to run across multiple Kubernetes namespaces.  
The function returns `true` only when:

1. The operator’s name appears in the list of *multi‑namespaced* operators, **and**
2. The supplied namespace list contains more than one entry.

This helper is used during test setup to decide whether to create a dedicated operator deployment per namespace or to deploy a single instance that spans all namespaces.

---

## Parameters

| Name | Type | Role |
|------|------|------|
| `operatorName` | `string` | The name of the operator under inspection. |
| `nsList` | `[]string` | Slice of Kubernetes namespaces that are relevant for the test scenario. |

---

## Return Value

| Type | Meaning |
|------|---------|
| `bool` | `true` if the operator is multi‑namespaced *and* more than one namespace is provided; otherwise `false`. |

---

## Key Dependencies

1. **`StringInSlice`**  
   A package‑private helper that checks whether a string exists in a slice of strings. It is used to look up `operatorName` in the list of known multi‑namespaced operators.

2. **`len()`** (built‑in)  
   Counts elements in `nsList`. The function requires at least two namespaces for a multi‑namespace deployment.

3. **Global variable `env`** (not directly used here but part of the package’s context) – provides test configuration, potentially influencing the list of known operators (via other helpers).

---

## Side Effects

None.  
The function is pure: it only reads its inputs and performs deterministic checks; it does not modify global state or interact with external systems.

---

## How It Fits in the Package

Within `tests/operator`, many tests need to know whether an operator should be instantiated once per namespace or once globally. The list of multi‑namespaced operators is maintained elsewhere (often in a constant slice).  
`isMultiNamespacedOperator` encapsulates this logic so that test code remains concise:

```go
if isMultiNamespacedOperator(opName, nsList) {
    // Deploy one instance that covers all namespaces
} else {
    // Deploy separate instances per namespace
}
```

Because it is unexported, the function is intentionally used only by helper routines in this package and keeps the public API minimal.
