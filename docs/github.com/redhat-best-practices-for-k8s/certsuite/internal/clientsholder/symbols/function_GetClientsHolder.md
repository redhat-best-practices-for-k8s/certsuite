GetClientsHolder`

| Attribute | Details |
|-----------|---------|
| **Package** | `clientsholder` (`github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder`) |
| **Exported** | Yes |
| **Signature** | `func GetClientsHolder(...string) *ClientsHolder` |
| **Purpose** | Provide a globally‑available, lazily‑initialized instance of `ClientsHolder`. |

---

### Purpose

The function implements the classic *singleton* pattern for the `ClientsHolder` type.  
A single `ClientsHolder` is created on first use and reused throughout the program.  
This guarantees that all callers share the same set of Kubernetes clients and related
configuration, preventing accidental duplication or state inconsistency.

---

### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `...string` | variadic `string` | Optional configuration values used only on first creation. They are passed straight to `newClientsHolder`. Subsequent calls ignore these arguments because the instance is already created. |

> **Note:** The function accepts a variable number of strings but typically callers pass either none or one value (e.g., a kubeconfig path).  
> If more than one argument is supplied, only the first is used by `newClientsHolder`; the rest are ignored.

---

### Return Value

* `*ClientsHolder` – A pointer to the singleton instance.  
  The returned object contains all client sets and configuration that were created during the first call.

---

### Key Dependencies & Calls

| Called Function | Description |
|-----------------|-------------|
| `newClientsHolder` | Constructs a fresh `ClientsHolder` from the supplied arguments. This is executed only on the very first invocation of `GetClientsHolder`. |
| `Fatal` (from a logger package) | If initialization fails, the error is logged as fatal and the program terminates. |

The function does **not** read or modify any global variables directly; it uses the unexported
package‑level variable `clientsHolder`, which holds the singleton instance.

---

### Side Effects

1. **First Call Only**  
   * Instantiates a new `ClientsHolder`.  
   * May perform I/O (e.g., reading kubeconfig files, creating REST clients).  
   * If an error occurs during creation, it is reported via `Fatal` and the process exits.

2. **Subsequent Calls**  
   * Return the previously created instance without side effects.

---

### Interaction with the Package

- The package defines a private variable `clientsHolder` that stores the singleton.
- `GetClientsHolder` is the public API for retrieving this instance; other files in the same package (e.g., `clientsholder.go`) use it to access clients.
- Because the function is thread‑safe by virtue of being called only once, no additional synchronization primitives are required.

---

### Suggested Mermaid Diagram

```mermaid
graph TD;
    A[Program Start] --> B{clientsHolder set?}
    B -- No --> C[newClientsHolder(...)]
    C --> D[Set clientsHolder]
    D --> E[Return instance]
    B -- Yes --> F[Return existing instance]
```

This diagram visualizes the one‑time initialization and reuse pattern employed by `GetClientsHolder`.
