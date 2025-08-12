NewChecksGroup`

### Overview
`NewChecksGroup` creates (or retrieves) a **checks group** identified by the supplied name.  
A *checks group* is a logical container that holds multiple checks and their execution state. The function guarantees that only one instance per name exists in the global database.

### Signature
```go
func NewChecksGroup(name string) *ChecksGroup
```

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `name`    | `string` | Identifier for the group (e.g., `"k8s", "openshift"`). |

| Return | Type             | Description |
|--------|------------------|-------------|
| `*ChecksGroup` | Pointer to the created or existing group | The caller receives a reference to the singleton group instance. |

### Key Dependencies
- **Global registry**:  
  - `dbByGroup map[string]*ChecksGroup` – holds all groups by name.  
  - `dbLock sync.Mutex` – protects concurrent access to `dbByGroup`.  
- **Thread‑safety**: The function locks `dbLock` while checking/creating the entry and unlocks it afterward.

### Behaviour
1. Acquire `dbLock` (`Lock`) to serialize access.  
2. Look up `name` in `dbByGroup`.
   - If found, return the existing instance.
3. If not found:
   1. Instantiate a new `ChecksGroup` (the constructor is defined elsewhere).
   2. Store it in `dbByGroup` under `name`.
4. Release `dbLock` (`Unlock`).  
5. Return the pointer to the group.

No other global variables are modified. The function has no side effects beyond populating the registry and ensuring safe concurrent creation.

### Integration with the package
- **Centralised access**: All code that needs a checks group should call `NewChecksGroup`. This guarantees a single source of truth for each group name.
- **Lazy initialization**: Groups are created on demand; the first request triggers allocation, subsequent requests reuse the same object.
- **Thread safety**: By guarding `dbByGroup` with a mutex, the package remains safe when multiple goroutines request groups concurrently.

### Example
```go
// Retrieve or create the "k8s" checks group
group := NewChecksGroup("k8s")

// Now `group` can be used to add or run checks
```

---

**Mermaid diagram (suggestion)**

```mermaid
flowchart TD
    A[Caller] --> B{Lock dbLock}
    B --> C{Check dbByGroup[name]}
    C -- found --> D[Return existing ChecksGroup]
    C -- not found --> E[Create new ChecksGroup]
    E --> F[dbByGroup[name] = new Group]
    F --> G[Unlock dbLock]
    D & G --> H[Return pointer]
```

This function is a foundational piece of the checks database, enabling safe and consistent group management across the application.
