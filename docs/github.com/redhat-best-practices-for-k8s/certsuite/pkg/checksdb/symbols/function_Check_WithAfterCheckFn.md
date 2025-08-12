Check.WithAfterCheckFn`

> **Location**: `pkg/checksdb/check.go` – line 136  
> **Package**: `checksdb`

## Overview
`WithAfterCheckFn` is a convenience method that attaches an *after‑check* callback to a `Check`.  
The callback receives the finished `Check` instance and can perform any post‑processing (logging, metrics, cleanup, etc.).  
The method returns the modified `Check`, enabling fluent chaining.

## Signature
```go
func (c Check) WithAfterCheckFn(fn func(check *Check) error) *Check
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `fn` | `func(*Check) error` | A function to be invoked once the check has finished. It receives a pointer to the same `Check` instance and may return an error if the callback fails. |

| Return value | Type | Description |
|--------------|------|-------------|
| `*Check` | Pointer to the modified `Check` | Allows chaining of further modifiers or direct use in registrations. |

## How it Works
1. The receiver `c` is passed by value (copy of the struct).  
2. A new copy (`c`) is mutated: its field `afterFn` is set to the supplied function.  
3. The pointer to this modified copy is returned.

No global state or synchronization is touched; the method only mutates the local copy.  

## Dependencies
- **Data structure**: `Check` (defined in the same file). It contains an `afterFn func(*Check) error` field used by the check execution engine to call after a check finishes.
- **Package context**: Part of the *checksdb* registry that stores all checks and groups. The method itself does not interact with the global maps (`dbByGroup`, `resultsDB`) or locks.

## Side Effects
- None beyond setting the `afterFn` field on the returned copy.
- The original `Check` instance passed as a receiver remains unchanged (because of value semantics).

## Usage Pattern
```go
check := NewCheck("my-check").
    WithDescription("Validates something").
    WithAfterCheckFn(func(c *Check) error {
        log.Infof("Check %s finished with status %s", c.ID, c.Status)
        return nil
    })
```

The returned `*Check` can then be registered in the database or used directly.

## Fit within `checksdb`
- **Fluent API**: `WithAfterCheckFn` is one of several builder methods (`WithDescription`, `WithTag`, etc.) that let callers compose a `Check` declaratively.
- **Execution flow**: During check execution, after the main logic completes, the engine calls the stored `afterFn`. This allows post‑check actions without cluttering the core logic.
- **Isolation**: By returning a new instance, it preserves immutability of the original `Check`, which is important for concurrent registrations.

## Summary
`WithAfterCheckFn` is a lightweight helper that attaches a callback to be run after a check completes. It influences only the local copy of the `Check` and integrates seamlessly into the fluent API used throughout the *checksdb* package.
