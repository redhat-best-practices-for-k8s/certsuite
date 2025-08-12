# `WithSkipCheckFn`

```go
func (c *Check) WithSkipCheckFn(fns ...func() (skip bool, reason string)) *Check
```

## Purpose

`WithSkipCheckFn` attaches one or more *skip‑functions* to a `Check`.  
A skip function is invoked at runtime before the check’s main logic runs.  
If any skip function returns `true`, the check is marked **skipped** and its
reason is recorded.

This mechanism allows checks to be conditionally disabled based on dynamic
runtime state (e.g., missing prerequisites, platform incompatibilities,
or custom user settings) without hard‑coding those conditions into the check
implementation itself.

## Inputs

| Parameter | Type                                   | Description |
|-----------|----------------------------------------|-------------|
| `fns`     | `...func() (skip bool, reason string)` | A variadic list of zero or more functions. Each function must return a boolean indicating whether to skip and an explanatory string. |

> **Note**: The receiver `c *Check` is the check instance being configured.

## Output

| Return Value | Type   | Description |
|--------------|--------|-------------|
| `*Check`     | pointer to the same `Check` instance | Allows method chaining (builder pattern). |

## Key Dependencies & Side‑Effects

1. **Internal State Modification**  
   The function appends each supplied skip function to the check’s internal
   slice of skip functions:
   ```go
   c.skipFns = append(c.skipFns, fns...)
   ```
   This mutates the `Check` object in place.

2. **No External Global Interaction**  
   The method does not read or modify any package‑level globals (`dbByGroup`,
   `dbLock`, etc.). Its effect is limited to the check instance.

3. **Concurrency Considerations**  
   Since a `Check` may be accessed concurrently by multiple goroutines
   (e.g., when running checks in parallel), callers should ensure that
   skip functions are added before any execution begins, or use external
   synchronization mechanisms if mutation occurs at runtime.

## How It Fits Into the Package

The `checksdb` package defines a registry of certificate‑suite checks.
Each check is represented by the `Check` struct, which includes:

- The actual test logic (`RunFunc`).
- Metadata (name, description, tags).
- A slice of skip functions (`skipFns`) that decide whether the check
  should run.

`WithSkipCheckFn` is part of the fluent API that lets package users
configure checks declaratively. Typical usage pattern:

```go
check := NewCheck("my-check").
    WithDescription("Example check").
    WithTags([]string{"k8s", "security"}).
    WithRunFunc(myTestFunction).
    WithSkipCheckFn(
        func() (bool, string) { return os.Getenv("SKIP") == "1", "SKIP env set" },
        anotherCondition,
    )
```

During execution, the runner will iterate over `check.skipFns` and
short‑circuit to a **skipped** result if any function returns true.

Thus, `WithSkipCheckFn` is essential for making checks adaptive and
environment‑aware without cluttering the core test logic.
