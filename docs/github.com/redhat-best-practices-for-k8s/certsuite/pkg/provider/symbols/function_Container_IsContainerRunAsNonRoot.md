## `IsContainerRunAsNonRoot`

```go
func (c *Container) IsContainerRunAsNonRoot() (bool, string)
```

### Purpose  
Determines whether a container is configured to run as a non‑root user.  
The function inspects the container’s security context (`RunAsUser`, `RunAsGroup`, `RunAsNonRoot` and the image’s default UID/GID) and returns:

| Return value | Meaning |
|--------------|---------|
| `true`       | The container **does** run as a non‑root user. |
| `false`      | The container **does not** run as a non‑root user (runs as root or the check cannot be conclusively made). |

The second return value is a diagnostic string that explains *why* the decision was made (e.g., “RunAsNonRoot flag set to true”, “image has UID 1000”, etc.). This string is primarily used for logging and test output.

### Inputs  
| Parameter | Type | Notes |
|-----------|------|-------|
| `c` | `*Container` (receiver) | The container object whose security context is examined. |

No external arguments are required; the function relies solely on data stored in the receiver.

### Key dependencies  

| Dependency | Role |
|------------|------|
| `fmt.Sprintf` | Formats diagnostic messages. |
| `PointerToString` | Converts a pointer to its string representation for logging (used when reporting pointer values of optional fields). |

These are standard library or package helper functions; no additional side effects.

### Side effects  
The function is **pure**: it does not modify the container, global state, or any external resource. It only reads from the receiver and returns computed results.

### How it fits in the `provider` package  

* The `provider` package implements a Kubernetes test framework that validates best‑practice configurations.  
* Containers are a core abstraction; many tests examine their security settings (e.g., image scanning, pod spec validation).  
* `IsContainerRunAsNonRoot` is a utility used by higher‑level checks such as:
  * **Pod security policy** validations – ensuring containers do not run as root.
  * **Compliance reports** – generating user‑friendly messages about container privileges.

By encapsulating the logic in one method, other parts of the package can reuse it without duplicating the intricate logic around Kubernetes security contexts.
