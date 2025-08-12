ManagedDeploymentsStatefulsets`

```go
type ManagedDeploymentsStatefulsets struct {
    Name string
}
```

## Overview

`ManagedDeploymentsStatefulsets` is a lightweight configuration holder used by the **configuration** package of CertSuite.  
It represents a single *statefulset* that is managed by the system.  The sole piece of data stored is the name of the statefulset, which allows other components to refer to it without embedding Kubernetes API objects.

> **Why this struct?**  
> A dedicated type gives us compile‑time safety and clear intent when passing around statefulset names.  It also makes future extensions (e.g., adding a namespace or selector) straightforward without breaking existing code.

## Fields

| Field | Type   | Description |
|-------|--------|-------------|
| `Name` | `string` | The Kubernetes name of the statefulset to manage. |

*All fields are exported, enabling external packages to construct and inspect instances directly.*

## Functions / Methods

The struct has no methods defined in this package; it is purely a data container.

## Dependencies & Side‑Effects

- **Dependencies**: None beyond the standard library.
- **Side‑effects**: Instantiating or modifying an instance has no side‑effects.  
  The struct is immutable by convention—once created, callers should treat its value as read‑only unless they explicitly reassign a new instance.

## Usage in the Package

Within `pkg/configuration`, this type is typically used to:

1. **Load configuration** – e.g., parsed from YAML/JSON where each entry corresponds to a statefulset name.
2. **Pass around** – functions that need to operate on specific managed statefulsets receive a slice of `ManagedDeploymentsStatefulsets` instead of raw strings, improving readability.

Example (pseudo‑code):

```go
var sts []configuration.ManagedDeploymentsStatefulsets

// Load from file
sts = configuration.LoadStatefulsetConfig("statefulsets.yaml")

for _, s := range sts {
    err := controller.EnsureExists(ctx, s.Name)
}
```

## Diagram Suggestion

A small Mermaid diagram could illustrate the flow:

```mermaid
flowchart TD
  A[config.yaml] --> B[LoadStatefulsetConfig]
  B --> C{Array of } ManagedDeploymentsStatefulsets
  C --> D[Controller.EnsureExists(Name)]
```

---

**In summary**, `ManagedDeploymentsStatefulsets` is a minimal, self‑documenting way to reference Kubernetes statefulsets within CertSuite’s configuration logic. It has no behavior of its own but serves as the typed building block for higher‑level configuration handling.
