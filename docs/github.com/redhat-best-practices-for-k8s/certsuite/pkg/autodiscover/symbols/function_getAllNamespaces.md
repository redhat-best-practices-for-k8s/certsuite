getAllNamespaces`

| Aspect | Details |
|--------|---------|
| **Signature** | `func (corev1client.CoreV1Interface) ([]string, error)` |
| **Visibility** | Unexported – used only within the *autodiscover* package. |

### Purpose
`getAllNamespaces` is a small helper that retrieves the names of every namespace present in a Kubernetes cluster and returns them as a slice of strings.  
The function is primarily used by other discovery routines that need to iterate over all namespaces (e.g., when collecting resources or inspecting operator deployments).

### Parameters
| Name | Type | Role |
|------|------|------|
| `corev1client.CoreV1Interface` | `corev1client.CoreV1Interface` | A Kubernetes client capable of performing Core V1 operations. The caller supplies the client; this function does **not** create or close it.

### Return Values
| Position | Type | Meaning |
|----------|------|---------|
| 1 | `[]string` | Ordered list of namespace names (`metadata.Name`). May be empty if no namespaces exist. |
| 2 | `error` | Non‑nil if the Kubernetes API call fails or if any other error occurs while processing the list. The caller should handle this error.

### Key Steps & Dependencies
1. **List Namespaces**  
   Calls `client.Namespaces().List(ctx, metav1.ListOptions{})`. This is a standard client-go pattern for fetching all Namespace resources.

2. **Error Handling**  
   If the API call returns an error, it’s wrapped with `fmt.Errorf` and returned immediately (`return nil, fmt.Errorf(...)`).

3. **Collect Names**  
   Iterates over the resulting list (`namespaceList.Items`) and appends each namespace’s name to a slice.

4. **Return Slice**  
   Returns the populated slice and a `nil` error on success.

### Side Effects
- No mutation of global state; all work is confined to local variables.
- The only observable effect is the returned data or an error.

### How It Fits the Package
The *autodiscover* package contains logic for discovering various Kubernetes resources (operators, CRDs, etc.) across a cluster. Many discovery paths require knowledge of which namespaces exist. `getAllNamespaces` centralizes that lookup so other functions can simply call it without duplicating client‑creation or error handling code.

```mermaid
flowchart TD
    A[Caller] -->|provides| B[getAllNamespaces]
    B --> C{List Namespaces}
    C --> D[List Result]
    D --> E{Error?}
    E -- Yes --> F[Return nil, fmt.Errorf(...)]
    E -- No --> G[Iterate Items]
    G --> H[Append name to slice]
    H --> I[Return []string, nil]
```

### Example Usage
```go
namespaces, err := getAllNamespaces(clientset.CoreV1())
if err != nil {
    log.Fatalf("cannot list namespaces: %v", err)
}
for _, ns := range namespaces {
    // do something with each namespace
}
```

---

**Note:** The function itself is straightforward; its primary value lies in providing a clean, reusable abstraction for the rest of the discovery logic.
