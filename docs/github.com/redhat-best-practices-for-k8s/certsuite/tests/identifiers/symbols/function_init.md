init()` – Package‑level initialisation

| Element | Details |
|---------|---------|
| **File** | `identifiers.go` (line 55) |
| **Signature** | `func init() {}` |
| **Exported?** | No – package‑private, automatically executed by the Go runtime. |

### Purpose
The function is the standard *package initialization* routine that populates the test catalogue for the **identifiers** package.  
When any file in this package is imported, `init()` runs before `main` or any other code. It ensures that the global data structures (`Catalog`, `Classification`, etc.) are fully populated and ready for use by tests.

### Inputs / Outputs
- **Inputs:** None (no parameters).  
- **Outputs:** None directly returned; side‑effects only.

### Key Dependencies & Side‑Effects
| Dependency | Effect |
|------------|--------|
| `InitCatalog` | Called once to build the internal test catalogue. The function is defined elsewhere in the package and fills global variables such as `Catalog`, `Classification`, and potentially other lookup tables. |

The routine relies on the *global* state of the package; after execution, any code that imports this package can safely query:

```go
identifiers.Catalog      // map[string]TestCase
identifiers.Classification // map[ClaimType][]TestCase
```

### How it Fits the Package

- **Centralised Bootstrap:** All test identifiers, doc links, and impact strings are defined as exported constants/variables. `init()` ties them together by invoking `InitCatalog`, which registers each identifier in the catalogue.
- **Read‑Only Public API:** The package exposes only read‑only data structures; consumers do not modify them. `init()` guarantees that these structures are fully initialised before first use, preventing race conditions or nil references.
- **Testing Infrastructure:** The rest of the code base (e.g., test runners, assertion helpers) expects the catalogue to be ready. Without this `init()`, lookups would fail.

---

#### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[Package Import] --> B[Go Runtime]
    B --> C{Run init functions}
    C --> D[identifiers.init()]
    D --> E[InitCatalog() populates globals]
    E --> F[Global Catalog & Classification ready]
```

This diagram illustrates the order of execution when the package is first imported.
