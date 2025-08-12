AddCatalogEntry`

| Aspect | Details |
|--------|---------|
| **Package** | `identifiers` (github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers) |
| **Signature** | `func AddCatalogEntry(id, title, description, category, subcategory, docLink string, optional bool, tags map[string]string, remediations ...string) claim.Identifier` |

### Purpose
Adds a test case to the global JUnit catalog (`Catalog`) and returns a `claim.Identifier`.  
The function normalises input strings, builds a full description (using `BuildTestCaseDescription`), and updates internal bookkeeping such as tags and remediations.

### Inputs

| Parameter | Type | Meaning |
|-----------|------|---------|
| `id` | string | Unique test case identifier (e.g. `"APICompatibilityWithNextOCPRelease"`) |
| `title` | string | Human‑readable title of the test |
| `description` | string | Short description; may contain tags like `@mandatory` |
| `category` | string | Main category (e.g., `"Security"`, `"Best Practices"`) |
| `subcategory` | string | Sub‑category (e.g., `"Container"`). May be empty. |
| `docLink` | string | URL to documentation or test source |
| `optional` | bool | Marks the test as optional – affects filtering logic |
| `tags` | map[string]string | Key/value pairs used for classification, e.g. `{"tag":"mandatory"}` |
| `remediations ...string` | variadic | Zero or more remediation identifiers that are associated with this test |

### Process

1. **Trim whitespace** – `id`, `title`, and `description` are passed through `strings.TrimSpace`.
2. **Length checks** – If any of the trimmed strings is empty, a panic occurs (`len(...) == 0`).  
   *This ensures all catalog entries contain required metadata.*
3. **Description construction** – Calls `BuildTestCaseDescription(id, title, description)` to produce a canonical description string used in the test output.
4. **Catalog entry creation** – Constructs a new `claim.Identifier` (internal struct) and appends it to the global slice `Catalog`.  
   The identifier contains all provided fields plus the constructed description.
5. **Return value** – Returns the created `claim.Identifier`, allowing callers to reference the ID in other parts of the test suite.

### Dependencies & Side‑Effects

| Dependency | Role |
|------------|------|
| `strings.TrimSpace` | Cleans input strings. |
| `len` | Validates non‑empty inputs. |
| `append` | Adds entry to `Catalog`. |
| `BuildTestCaseDescription` | Generates a formatted description string. |

*Side‑effects*:  
- Mutates the global `Catalog` slice.  
- No other package state is changed.

### Integration in the Package

All test identifiers are declared by calling `AddCatalogEntry`.  
For example, `TestAPICompatibilityWithNextOCPReleaseIdentifier` is created as:

```go
TestAPICompatibilityWithNextOCPReleaseIdentifier = AddCatalogEntry(
    "APICompatibilityWithNextOCPRelease",
    "...", // title
    "...", // description
    "Compliance", "OCP", "https://...", true,
    map[string]string{"tag":"mandatory"},
    APICompatibilityWithNextOCPReleaseRemediation)
```

The returned `claim.Identifier` is then used by the test framework to register the test case and link it with remediation logic.  
Thus, `AddCatalogEntry` is the central factory for building the JUnit test catalog and ensuring consistency across all identifiers.
