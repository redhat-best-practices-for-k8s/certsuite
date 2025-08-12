Check.WithTimeout`

| Aspect | Details |
|--------|---------|
| **Signature** | `func (c Check) WithTimeout(d time.Duration) *Check` |
| **Receiver type** | `Check` – a struct that represents an individual test/check in the checks database. |
| **Return value** | A new pointer to a `Check` instance with its timeout field set. |

### Purpose
`WithTimeout` is a convenience constructor used when registering or modifying a check to specify how long it may run before being aborted.  
In practice, callers create a `Check`, then chain `.WithTimeout(duration)` to attach a timeout value that will later be enforced by the test runner.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `d`  | `time.Duration` | The maximum duration the check is allowed to run. A zero or negative value indicates “no timeout”. |

### Return Value
A pointer to a **new** `Check` instance (deep‑copy of the receiver) whose `timeout` field equals `d`.  
The original `Check` remains unchanged, enabling immutable style chaining.

### Side Effects & Dependencies
* No global state is modified.  
* The function only touches the `timeout` field of the `Check` struct; no other fields are altered.  
* Because it returns a pointer to a copy, subsequent modifications do not affect previously returned checks.  

### How It Fits in the Package
- **Checks registration** – When adding checks to the database (`checksdb.Add()` or similar), callers often need to set a timeout for long‑running tests. `WithTimeout` is part of that fluent API.  
- **Test execution** – The runner reads this field to enforce timeouts (via context deadlines or timers).  
- **Immutability pattern** – Many other `Check` methods (`WithTags`, `WithLabels`, etc.) follow the same pattern; `WithTimeout` provides consistency across the API.

### Example Usage

```go
// Create a new check and set a 30‑second timeout.
c := NewCheck("my-check").
        WithDescription("Checks something important").
        WithTimeout(30 * time.Second)

// Register it in the checks database.
checksdb.Add(c)
```

In this example, `WithTimeout` simply records the desired duration; later, when the test runner executes the check, it will enforce a 30‑second limit.

---

**Key takeaway:** `Check.WithTimeout` is an immutable setter that attaches a timeout to a check, enabling callers to configure execution limits without affecting other parts of the checks database.
