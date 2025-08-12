LoadInternalChecksDB`

**Purpose**

`LoadInternalChecksDB` is the package‑wide initializer that loads all built‑in (internal) security checks into CertSuite’s in‑memory database.  
It does **not** perform any I/O itself; instead it returns a closure that, when invoked, will load each check type by calling `LoadChecks`.  The returned function has signature `func()` and is intended to be executed during package initialization (e.g., via `init()` or a user‑triggered bootstrap routine).

---

### Signature

```go
func LoadInternalChecksDB() func()
```

* **Returns** – A no‑argument closure that, when called, performs the actual loading of checks.

---

### How it works

1. **Closure construction**  
   The function builds an anonymous `func()` that captures no external state (other than the call to `LoadChecks`).  

2. **Loading sequence**  
   Inside the closure it calls `LoadChecks` repeatedly—once for each check type supported by CertSuite (e.g., pod‑security, image‑scanning, network‑policy).  
   The repeated calls are generated at compile time; the function simply forwards to the same helper without any additional logic.

3. **No side effects until invoked**  
   Importantly, the outer `LoadInternalChecksDB` itself does *not* perform I/O or modify global state.  Only when the returned closure is executed will the checks be read from disk (or embedded resources) and inserted into CertSuite’s internal registry.

---

### Dependencies

| Dependency | Role |
|------------|------|
| `LoadChecks` | Helper that loads a single check type into the database. Called nine times within the closure to cover all built‑in check categories. |

No other globals or types are referenced directly by this function.

---

### Usage Pattern

```go
// During package init (or any bootstrap step)
var loadInternalChecks = certsuite.LoadInternalChecksDB()

func init() {
    // Actual loading happens here, after the package is fully initialized.
    loadInternalChecks()
}
```

The closure allows for deferred execution, which can be useful when the database should only be populated after certain runtime conditions are met (e.g., configuration files are available).

---

### Fit in the Package

`LoadInternalChecksDB` resides in `certsuite/certsuite.go`, a central file that defines package‑wide constants and bootstrap logic.  
It is the gateway through which all internal checks become available to CertSuite’s scanning engine, ensuring that every supported check type is registered before any scan runs.

---
