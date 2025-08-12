Check.WithCheckFn`

> **Location**: `pkg/checksdb/check.go` – line 118  
> **Signature**

```go
func (c *Check) WithCheckFn(fn func(check *Check) error) *Check
```

## Purpose

`WithCheckFn` is a convenience method that lets you attach an arbitrary validation function to a `Check`.  
It returns the same `*Check` instance, so calls can be chained.  The function receives the check itself, allowing it to inspect or modify its fields before execution.

Typical use:

```go
check := NewCheck("my-check").
    WithDescription("Ensures something…").
    WithCheckFn(func(c *Check) error {
        // perform custom validation
        if c.SomeField == "" {
            return errors.New("field missing")
        }
        return nil
    })
```

## Parameters

| Name | Type | Description |
|------|------|-------------|
| `fn` | `func(check *Check) error` | A closure that takes the check pointer and returns an error if validation fails. The error is used as the check’s result (`Failed`, `Error`, etc.). |

> **Note**: The function should not modify the check’s public state in ways that affect other checks, unless intentionally doing so.

## Return Value

| Type | Description |
|------|-------------|
| `*Check` | The same pointer that was used to invoke the method, enabling fluent chaining. |

## Side‑Effects & Dependencies

1. **No global mutation** – The function only captures the passed closure; it does not touch package globals (`dbByGroup`, `resultsDB`, etc.).  
2. **Execution context** – The attached function is executed later by the framework (e.g., when a check is run). At that point, any changes to the check’s fields persist in the shared `Check` instance.  
3. **Error handling** – The returned error is interpreted by the checks runner:
   * `nil` → result stays as previously set (`Passed`, `Skipped`, etc.).  
   * non‑nil → the check is marked `Failed` or `Error` depending on how the caller interprets it.

## Relationship to the Package

- **Checks** are stored in a global registry (`dbByGroup`) and may belong to groups (`ChecksGroup`).  
- The `WithCheckFn` method is part of the public API that allows test authors to enrich a check with custom logic without modifying the core engine.  
- It complements other builder methods (`WithDescription`, `WithLabels`, etc.) for configuring a check before it is registered.

## Example Usage

```go
// Define a new check that verifies a resource exists.
check := NewCheck("resource-exists").
    WithDescription("Ensures the specified CRD is present.").
    WithCheckFn(func(c *Check) error {
        // Imagine `c.Context` holds the Kubernetes client.
        if _, err := c.Client.Get(context.Background(), metav1.ObjectMeta{Name: c.Params["name"]}); err != nil {
            return fmt.Errorf("resource not found: %w", err)
        }
        return nil
    })

// Register the check in its group.
Group("crd").Add(check)
```

In this example, `WithCheckFn` attaches a custom validation that interacts with Kubernetes. The returned `*Check` is then added to the `"crd"` group for later execution.

---

**Key Takeaway:**  
`WithCheckFn` lets you inject bespoke logic into a check while preserving a fluent API and keeping package state untouched until the check is actually run.
