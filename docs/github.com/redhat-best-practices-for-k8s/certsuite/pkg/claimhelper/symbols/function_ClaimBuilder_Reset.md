ClaimBuilder.Reset` – Package‑level Overview

| Item | Description |
|------|-------------|
| **Package** | `claimhelper` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/claimhelper`) |
| **Purpose** | Reinitialize a `ClaimBuilder` instance so it can be reused for building a new claim. The method clears the current state, sets fresh timestamps and prepares the builder for a clean start. |
| **Signature** | `func (cb ClaimBuilder) Reset() func()` |

---

### 1. What does `Reset` do?

`Reset` is an instance method on the `ClaimBuilder` struct that performs three main actions:

1. **Clear the claim payload** – it assigns a fresh, empty map to the builder’s internal storage.
2. **Set timestamps** – two fields are updated:
   * `cb.now`  ← `time.Now()` (local time)
   * `cb.utcNow` ← `time.Now().UTC()`
3. **Return a closure** that, when called, will reset the builder again.  
   The returned function simply invokes `Reset` on the same receiver (`cb.Reset()`), enabling chaining or deferred cleanup.

These steps ensure that subsequent calls to the builder start from a pristine state without residual data from previous builds.

---

### 2. Inputs & Outputs

| Parameter | Type | Notes |
|-----------|------|-------|
| Receiver (`cb`) | `ClaimBuilder` (value receiver) | The builder instance being reset. |

| Return Value | Type | Description |
|--------------|------|-------------|
| `func()` | Closure that calls `Reset` on the same `ClaimBuilder`. | Useful for patterns like `defer cb.Reset()()` or for explicit manual resetting after a build. |

---

### 3. Dependencies & Calls

- **Standard Library**  
  - `time.Now()` – obtains current local time.  
  - `time.Now().UTC()` – converts the same instant to UTC.

These are wrapped by internal helper functions `Format` and `UTC` in the same package, which are invoked indirectly by `Reset`. The exact implementations of these helpers are not shown here, but they format timestamps according to the constants defined at the top of the file (e.g., `DateTimeFormatDirective`).

---

### 4. Side Effects & State Changes

| Field | Old Value | New Value |
|-------|-----------|-----------|
| `cb.claimMap` | Previous claim data | `{}` (empty map) |
| `cb.now` | Last timestamp used | Current local time (`time.Now()`) |
| `cb.utcNow` | Last UTC timestamp | Current UTC time (`time.Now().UTC()`) |

No external resources are touched; the method only mutates the receiver’s internal fields.

---

### 5. Where it fits in the package

The `claimhelper` package provides utilities for constructing and validating "claims" (structured test reports).  
- `ClaimBuilder` is a fluent builder that accumulates claim data.  
- `Reset` allows the same builder instance to be reused safely, which can reduce allocations in high‑throughput testing scenarios or when building multiple claims sequentially.

Typical usage pattern:

```go
cb := NewClaimBuilder()
defer cb.Reset()()

// build first claim
cb.Add(...)

// reset and reuse
cb.Reset()()
// build second claim
```

---

### 6. Mermaid diagram (suggestion)

```mermaid
flowchart TD
    A[Call ClaimBuilder.Reset] --> B{Clear claimMap}
    B --> C[claimMap = {}]
    A --> D{Set timestamps}
    D --> E[now = time.Now()]
    D --> F[utcNow = time.Now().UTC()]
    A --> G[Return closure]
```

---
