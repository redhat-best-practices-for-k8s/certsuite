NewLabelsExprEvaluator`

| Aspect | Details |
|--------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/labels` |
| **Signature** | `func NewLabelsExprEvaluator(expr string) (LabelsExprEvaluator, error)` |

### Purpose
Creates a runtime evaluator for label expressions.  
The function takes an expression string that may contain placeholder variables and returns an object implementing the `LabelsExprEvaluator` interface which can later be invoked to compute concrete labels based on supplied context.

### Inputs & Outputs
| Parameter | Type | Description |
|-----------|------|-------------|
| `expr` | `string` | A label‑expression template, e.g. `"app={{ .AppName }}-{{ .Version }}"`. |

| Return | Type | Meaning |
|--------|------|---------|
| first return value | `LabelsExprEvaluator` | An evaluator capable of generating labels from the parsed expression. The concrete type is hidden behind an interface; callers interact only via that interface. |
| second return value | `error` | Non‑nil if parsing or preprocessing fails (e.g., syntax errors, unsupported placeholders). |

### Key Dependencies
* **`strings.ReplaceAll`** – Used twice to perform preliminary string substitutions on the raw expression before parsing.
* **`ParseExpr`** – Parses the sanitized expression into an internal representation suitable for evaluation. Likely comes from a templating or expression‑parsing package (e.g., `text/template`, `expr`).
* **`fmt.Errorf`** (`Errorf`) – Wraps any errors encountered during parsing with contextual information.

### Side Effects
The function is pure: it does not modify global state, external files, or other packages. It only allocates memory for the evaluator and its parsed representation.

### How It Fits in the Package
* **Package Responsibility** – The `labels` package provides mechanisms to define and evaluate label expressions that can be applied to Kubernetes resources.
* **Integration Flow**  
  1. Call `NewLabelsExprEvaluator` with a template string when initializing a component that needs dynamic labels.  
  2. Store the returned `LabelsExprEvaluator`.  
  3. When the actual label values are known (e.g., at runtime), invoke methods on the evaluator to produce concrete key‑value pairs.

This function is the entry point for turning user‑defined expression strings into executable evaluators, bridging static configuration and dynamic label generation.
