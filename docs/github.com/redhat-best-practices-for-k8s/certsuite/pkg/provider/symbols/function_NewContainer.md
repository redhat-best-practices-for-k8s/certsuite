NewContainer`

### Purpose
Creates a new instance of the internal **`Container`** type used by the CertSuite provider.  
The function simply allocates the struct and initializes its fields with sane defaults so that callers can immediately start configuring or inspecting it.

> *Why a dedicated constructor?*  
> The package keeps `Container` fields unexported to enforce validation via helper methods. `NewContainer` provides a single place where all required defaults are set, reducing boilerplate in the rest of the codebase.

### Signature
```go
func NewContainer() *Container
```

- **Returns**: a pointer to an initialized `*Container`.

There are no parameters and no exported side‑effects; the function is pure.

### Dependencies & Globals
`NewContainer` does not reference any global variables or other package functions.  
It relies only on the definition of the `Container` struct (located in `containers.go`) and uses Go’s built‑in allocation.

### Typical Usage
```go
c := NewContainer()
c.SetImage("quay.io/example/image:v1")
```

After construction, callers usually set fields such as image, command, environment variables, or resource limits via the public API on `Container`.

### How It Fits the Package
- **Encapsulation**: Keeps the internal representation of a container hidden from consumers.  
- **Consistency**: Guarantees that every `Container` starts in a known state (e.g., empty env slice, default resource limits).  
- **Extensibility**: If later the package needs to inject global defaults (e.g., from environment variables), they can be added here without touching all call sites.

---

#### Mermaid Diagram – Container Creation Flow
```mermaid
flowchart TD
    A[Call NewContainer()] --> B{Allocate & init}
    B --> C[*Container{...} initialized]
    C --> D[Return pointer to caller]
```

---
