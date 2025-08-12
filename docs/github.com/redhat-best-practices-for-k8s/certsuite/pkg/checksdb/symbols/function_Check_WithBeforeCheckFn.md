## `WithBeforeCheckFn`

```go
func (c Check) WithBeforeCheckFn(fn func(*Check) error) *Check
```

### Purpose
Adds a **pre‑execution hook** to a `Check`.  
When the check is run, this function will be invoked **before** any of its main test logic. It allows callers to perform validation, setup or early exit logic that may influence whether the real check should proceed.

The returned value is the same `*Check` instance with the new hook registered, enabling fluent chaining:

```go
check := NewCheck("my-check").
    WithBeforeCheckFn(func(c *Check) error {
        // e.g. skip if a required annotation is missing
        return nil
    })
```

### Inputs
| Parameter | Type | Description |
|-----------|------|-------------|
| `fn` | `func(*Check) error` | A callback that receives the check itself and may return an error to abort execution. If the function returns a non‑nil error, the main check logic is skipped and the result will be marked as **aborted** (`CheckResultAborted`). |

### Outputs
| Return value | Type | Description |
|--------------|------|-------------|
| `*Check` | Pointer to the same `Check` instance | The method returns the receiver so that further modifiers can be chained. |

### Key Dependencies & Side Effects
- **No external global state** is touched; the function simply assigns the provided callback to a field in the `Check` struct (`beforeFn`).
- Subsequent calls to `Run()` (not shown here) will read this field and invoke it before performing any other logic.
- The error returned by `fn` propagates as an abort signal. It does **not** alter the check’s metadata or database registration.

### How it Fits the Package
The `checksdb` package defines a domain model for test checks (`Check`, `ChecksGroup`, etc.).  
`WithBeforeCheckFn` is part of the builder API that lets developers declaratively compose a check with optional behaviour:

- **Without** this method, all checks run their core logic unconditionally.
- With it, tests can short‑circuit, perform environment validation or enrich context before evaluation.

Because the package also exposes constants like `CheckResultAborted`, callers typically rely on these values to interpret the outcome after execution. This function is therefore a small but essential piece of the check configuration pipeline.
