labelsExprParser`

`labelsExprParser` is an internal helper type that encapsulates a parsed labels‑expression tree and exposes a single public method, **`Eval`**, for evaluating the expression against a slice of label strings.

| Aspect | Details |
|--------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/labels` |
| **File** | `pkg/labels/labels.go` (line 17) |
| **Visibility** | Unexported – only the package itself can instantiate it. |
| **Fields** | `astRootNode ast.Expr` – the root of a Go‑style abstract syntax tree representing the parsed expression. |

### Purpose

When certsuite processes Kubernetes manifests, it often needs to decide whether a resource matches a set of labels described by a boolean expression (e.g., `"app=web && tier!=frontend"`).  
The `labelsExprParser` holds that parsed expression so that it can be reused for multiple label slices without reparsing.

### Key Method: `Eval`

```go
func (p *labelsExprParser) Eval(labels []string) bool
```

#### Inputs

| Parameter | Type | Meaning |
|-----------|------|---------|
| `labels` | `[]string` | The labels to test, each string formatted as `"key=value"` or just `"key"`. |

#### Outputs

* **bool** – `true` if the label slice satisfies the expression represented by `astRootNode`, otherwise `false`.

#### Internal Steps

1. **Normalize input**  
   A new slice is allocated (`make`) and each label string has all spaces removed using `ReplaceAll`.

2. **Tree traversal**  
   The method calls an unexported helper, `visit`, recursively walking the AST:
   * Each node type (e.g., `ast.BinaryExpr`, `ast.BasicLit`, etc.) is matched.
   * Comparisons (`==`, `!=`) and logical operators (`&&`, `||`, `!`) are evaluated against the normalized label slice.

3. **Error handling**  
   If an unexpected node type appears, `visit` triggers a fatal error via `log.Fatalf` (via the helper `Error`).  This is a side‑effect that terminates execution; therefore callers should be confident that only valid expressions reach this point.

#### Dependencies

* **Go AST package (`go/ast`)** – provides the node types used in `astRootNode`.
* **Standard library string functions** – for cleaning input and comparison.
* **Package‑internal helpers** – `visit` (recursive evaluator) and `Error`.

### How It Fits the Package

The `labels` package exposes a higher‑level API (`Parse`, `Matches`) that internally creates a `labelsExprParser`.  
Once created, callers repeatedly invoke `Eval` to test many label slices efficiently.  
Because the parser is unexported, its implementation can evolve without breaking external code; only the public functions need remain stable.

---

### Suggested Mermaid Diagram

```mermaid
flowchart TD
  A[Parse Expression] --> B[labelsExprParser{astRootNode}]
  B --> C(Eval(labels []string) bool)
  C --> D[visit(node, labels)]
  D --> E[Compare / Logical Ops]
```

This diagram visualises the flow from parsing to evaluation and highlights that `Eval` is the public gateway to the internal AST logic.
