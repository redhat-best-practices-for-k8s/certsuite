LabelsExprEvaluator` – Interface Overview

| Aspect | Details |
|--------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/labels` |
| **Location** | `pkg/labels/labels.go:13` |
| **Exported** | ✅ (public) |

### Purpose
The `LabelsExprEvaluator` interface abstracts the evaluation of label‑based expressions. In CertSuite, many tests and policy checks rely on evaluating whether a set of Kubernetes resource labels satisfies a user‑defined expression (e.g., `"app=web && tier!=frontend"`). By defining this contract as an interface, different implementations can be swapped in without changing consumer code.

### Core Method

```go
Eval(expr string, labels map[string]string) (bool, error)
```

| Parameter | Type | Meaning |
|-----------|------|---------|
| `expr` | `string` | The label expression to evaluate. Syntax is defined by the package’s parser (typically a simple subset of Go boolean expressions). |
| `labels` | `map[string]string` | Key‑value pairs representing the labels attached to a Kubernetes object. |

| Return | Type | Meaning |
|--------|------|---------|
| `bool` | Indicates whether the expression evaluates to true for the given label set. |
| `error` | Non‑nil if the expression is malformed or evaluation fails (e.g., unknown operator). |

### Dependencies & Side Effects
- **Dependencies**  
  - Relies on an underlying parser/evaluator implementation (often a struct that implements this interface).  
  - May import `strings`, `errors`, or custom parsing packages internally, but callers only interact via the interface.

- **Side Effects**  
  - None. The method is pure: it does not modify its inputs and produces no global state changes. It merely returns a boolean result and/or an error.

### Usage Context in the Package
1. **Policy Evaluation** – Policy objects contain label expressions; `Eval` determines if a resource satisfies that policy.  
2. **Test Filtering** – Test runners filter test cases based on labels; the evaluator decides inclusion/exclusion.  
3. **Dynamic Configuration** – Users can plug in custom evaluators (e.g., for extended syntax) by providing any type that implements `LabelsExprEvaluator`.

### Suggested Diagram
```mermaid
flowchart TD
    A[Caller] -->|Eval(expr, labels)| B[ConcreteEvaluator]
    B --> C{Result}
    C -->|true| D[Proceed]
    C -->|false| E[Skip]
```

This interface cleanly separates **what** is being evaluated from **how** it is performed, enabling extensibility and easier testing within the CertSuite codebase.
