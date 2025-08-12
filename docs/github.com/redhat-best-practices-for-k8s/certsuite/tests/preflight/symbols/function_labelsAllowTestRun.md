labelsAllowTestRun` – Internal Helper

| Aspect | Details |
|--------|---------|
| **Package** | `preflight` (tests for the pre‑flight suite) |
| **Signature** | `func(labels string, allowList []string) bool` |
| **Visibility** | Unexported – used only inside this package |

### Purpose
`labelsAllowTestRun` determines whether a test should be executed given a label string that may contain multiple comma‑separated labels and an optional *allow list* of labels.

- If `allowList` is empty, every label passes (`true`).  
- Otherwise the function checks if **any** label in the comma‑separated list exists in `allowList`.  
  - It uses the standard library helper `strings.Contains`, but the call in the code base refers to a custom `Contains` function (likely a wrapper over `strings.Contains`).

### Parameters

| Name | Type | Role |
|------|------|------|
| `labels` | `string` | A comma‑separated list of labels attached to a test. |
| `allowList` | `[]string` | List of labels that are permitted for execution. |

> **Note**: The function does *not* trim spaces, so callers must provide clean input or the comparison will be case‑sensitive and whitespace sensitive.

### Return Value

- `bool` – `true` if at least one label from `labels` is present in `allowList`; otherwise `false`.  
  - When `allowList` is empty, the function always returns `true`.

### Key Dependencies
* **`Contains`**: A helper (likely a thin wrapper over `strings.Contains`) that checks for substring presence.  
* No global variables or other state are accessed; the function is pure.

### Side‑Effects
None – purely functional.

### Integration in the Package
The pre‑flight tests use label filtering to decide which test cases should run in a given environment:

1. **Test Definition**: Each test case declares one or more labels.
2. **Execution Phase**: Before running, `labelsAllowTestRun` is invoked with those labels and an optional allow list (derived from command line flags or configuration).
3. **Result**: If the function returns `false`, the test is skipped; otherwise it proceeds.

This mechanism allows selective execution of tests without changing the test code itself.

### Suggested Mermaid Diagram
```mermaid
flowchart TD
    A[Labels string] -->|split by ","| B{Split labels}
    B --> C[Array of individual labels]
    C --> D{allowList empty?}
    D -- Yes --> E[Return true]
    D -- No --> F[Check each label with Contains]
    F --> G{Found match?}
    G -- Yes --> H[Return true]
    G -- No --> I[Return false]
```

The diagram illustrates the decision flow inside `labelsAllowTestRun`.
