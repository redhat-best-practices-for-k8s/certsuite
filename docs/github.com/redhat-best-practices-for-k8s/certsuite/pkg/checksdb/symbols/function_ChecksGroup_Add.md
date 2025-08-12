ChecksGroup.Add`

### Overview
`ChecksGroup.Add` is a method that registers a new `Check` instance into the current group.  
The method performs an in‑place append to the group's slice of checks while ensuring thread safety through a local mutex.

> **Why it matters** –  
> In the *checksdb* package, checks are organised into named groups (`ChecksGroup`).  Each group maintains its own list of `Check` objects that will later be executed by the test harness.  Adding a check to the correct group is therefore fundamental for both configuration and execution phases.

---

### Signature
```go
func (cg *ChecksGroup) Add(c *Check)
```

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `c`       | `*Check` | The check instance to be added. It is expected that the caller has already initialised all required fields of the struct. |

The function does **not** return a value; it mutates the receiver in place.

---

### Key Operations
1. **Locking**  
   ```go
   cg.Lock()
   ```
   - `ChecksGroup` embeds `sync.Mutex`, so this call acquires an exclusive lock on the group, preventing concurrent writes to its slice.

2. **Appending**  
   ```go
   cg.checks = append(cg.checks, c)
   ```
   - The new check is appended to the group's internal slice (`cg.checks`).  
   - No copy of `c` is made; the same pointer stored in the caller’s context is used.

3. **Unlocking**  
   ```go
   cg.Unlock()
   ```
   - Releases the mutex so other goroutines can access the group.

---

### Dependencies & Side‑Effects
- **Dependencies**
  * Uses only the standard `sync.Mutex` methods (`Lock`, `Unlock`) and Go’s built‑in `append`.
  * Relies on the `ChecksGroup` type having an embedded `sync.Mutex` field and a slice named `checks`.

- **Side‑Effects**
  * Mutates the receiver’s internal state by adding a new element to the slice.
  * Guarantees exclusive access during mutation, making it safe for concurrent use.

---

### How It Fits in the Package

```
ChecksGroup
├─ Add(*Check)          // <-- this method
├─ other group helpers
```

* `Add` is the primary way checks are registered into a group.  
* Other package functions (e.g., `RegisterCheck`, `ExecuteAll`) will later iterate over each group's slice to run tests or produce reports.
* Because groups can be accessed concurrently, `Add`'s locking logic keeps data races at bay.

---

### Usage Example

```go
group := NewChecksGroup("webserver")
check := &Check{ /* … */ }
group.Add(check) // safe even if other goroutines are adding to the same group
```

After this call, `group.checks` will contain a reference to `check`.

---
