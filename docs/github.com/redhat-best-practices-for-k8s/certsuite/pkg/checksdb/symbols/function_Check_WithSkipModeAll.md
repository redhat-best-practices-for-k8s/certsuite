Check.WithSkipModeAll`

### Overview
`WithSkipModeAll` is a method on the `Check` type that returns a new *copy* of the receiver with its **skip mode** set to `SkipModeAll`.  
In the checks database, each check can be configured to skip itself under certain conditions. The skip mode determines how those conditions are evaluated:

| Skip Mode | Meaning |
|-----------|---------|
| `SkipModeAny`  | The check is skipped if *any* of its skip expressions evaluate true. |
| **`SkipModeAll`** | The check is skipped only if *all* of its skip expressions evaluate true. |

Using `WithSkipModeAll` allows callers to override the default (`SkipModeAny`) for a particular check without mutating the original definition in the database.

---

### Function Signature
```go
func (c Check) WithSkipModeAll() *Check
```

| Component | Details |
|-----------|---------|
| **Receiver** | `c Check` – the original check instance. The receiver is passed by value, so the method works on a copy of the struct. |
| **Return Value** | `*Check` – a pointer to a new `Check` that has the same fields as `c`, except its `skipMode` field is set to `SkipModeAll`. |

---

### Dependencies & Context

- **Types**
  - `Check`: The core structure representing an individual test/check. It contains fields such as `ID`, `Name`, `Description`, `skipExprs []string`, and the internal `skipMode skipMode`.
  - `skipMode`: An enum-like type defined in `check.go`. The two possible values are `SkipModeAny` and `SkipModeAll`.

- **Constants**
  - `SkipModeAll` (exported): Used as the new mode.
  
- **Package Structure**
  - The method resides in `pkg/checksdb/check.go`, which also defines other utilities for managing checks, such as adding/removing from the global DB (`dbByGroup`) and evaluating skip expressions via `labelsExprEvaluator`.

---

### Side Effects & Mutability

- **No side effects** on the global state.  
  The method only creates a new local copy of the `Check` struct; it does not modify `c`, nor any shared data structures.
- The returned pointer can be stored or used independently, allowing callers to chain configuration calls like:
  ```go
  customCheck := originalCheck.WithSkipModeAll()
  ```

---

### Usage Pattern

```go
// Load a check from the database (returns a copy)
chk := checksdb.GetByID("example-check")

// Override skip mode for this instance only
customChk := chk.WithSkipModeAll()

// Run the check with the new configuration
result, err := customChk.Run(ctx, input)
```

---

### How It Fits the Package

`WithSkipModeAll` is part of a family of *fluent* helper methods that allow callers to modify specific aspects of a `Check` without mutating the canonical definition stored in `checksdb`.  
Other similar helpers (not shown) might include:

- `WithID(id string)`
- `WithName(name string)`
- `WithSkipExprs(exprs []string)`

These helpers enable a builder‑style API, improving readability when configuring checks for dynamic test runs.

---

### Summary

| Aspect | Detail |
|--------|--------|
| **Purpose** | Return a copy of a check with its skip mode set to *All* (skip only if all conditions are true). |
| **Inputs** | Receiver `Check` (by value). |
| **Outputs** | Pointer to a new `Check`. |
| **Side Effects** | None on global state. |
| **Dependencies** | `Check`, `SkipModeAll` constant, internal `skipMode` type. |
| **Package Role** | Provides fluent configuration for individual checks within the checks database. |

---
