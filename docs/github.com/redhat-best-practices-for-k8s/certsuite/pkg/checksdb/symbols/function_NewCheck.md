NewCheck`

```go
func NewCheck(name string, tags []string) *Check
```

### Purpose  
Creates a new **check** instance that can be registered in the checks database (`checksdb`).  
A check represents a single validation rule that will later be executed against a
Kubernetes cluster.  The function is the canonical entry point for creating all
checks; every other constructor or helper ultimately calls `NewCheck`.

### Parameters

| Name | Type   | Description |
|------|--------|-------------|
| `name` | `string` | Human‑readable identifier of the check (e.g., `"KubeAudit/valid-namespace"`).  The name is used as a key when storing the check in its group. |
| `tags` | `[]string` | Optional list of labels that can be used to filter or categorize checks at runtime. |

### Returns

* `*Check` – A fully initialised check object ready for registration and execution.

The returned check has:

- a logger obtained from `GetMultiLogger`, so log output is routed through the
  shared logging infrastructure.
- default values for all optional fields (`Result`, `Description`, etc.).
- an empty `SkipReason` and `skipMode`.

### Key Dependencies

| Dependency | How it’s used |
|------------|---------------|
| `With(name string, tags []string)` (method of `Check`) | Sets the check’s name and tags during construction. |
| `GetMultiLogger()` | Supplies a logger that writes to all configured outputs; this logger is stored in `check.Logger`. |

No other global state or side‑effects are touched directly by `NewCheck`; it merely prepares a struct for later insertion into the checks database (`dbByGroup`) via `Register`.

### Side Effects

`NewCheck` itself has no persistent side effects.  It does not modify any package globals, lock mutexes, or alter the checks database.  
All observable changes happen when the returned check is subsequently registered.

### How it Fits in the Package

- **Creation → Registration**: `NewCheck` creates a check; other functions (`Register`, `AddToGroup`) add it to the global map.
- **Logging**: By using `GetMultiLogger`, all checks share the same logging configuration, simplifying debugging and output aggregation.
- **Tagging & Filtering**: The `tags` slice allows runtime filtering (e.g., run only checks with a particular label).  These tags are stored in the check’s `Tags` field for later evaluation.

### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[NewCheck(name, tags)] --> B{Create Check}
    B --> C[Set name & tags via With]
    B --> D[Assign logger via GetMultiLogger]
    D --> E[*Check ready]
```

This diagram shows the two main steps performed by `NewCheck`: setting metadata and attaching a shared logger.
