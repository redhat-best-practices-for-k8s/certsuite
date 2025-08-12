SetTestK8sClientsHolder`

```go
func SetTestK8sClientsHolder(kubernetes.Interface)() 
```

| Aspect | Details |
|--------|---------|
| **Purpose** | Temporarily replace the package‑wide Kubernetes client holder with a mock or test client. The function is intended for unit tests that need to inject a custom `kubernetes.Interface` without touching production code. |
| **Parameters** | `client kubernetes.Interface` – the mock or test client to use while the test runs. |
| **Return value** | A *cleanup* closure of type `func()`. When called, it restores the original client holder that was in place before the call. This pattern guarantees that tests remain isolated and do not leak state into other tests. |
| **Key global dependency** | `clientsHolder` – an unexported variable that holds the current Kubernetes client implementation used by the rest of the package. The function assigns to this variable. |
| **Side effects** | 1. Overwrites `clientsHolder` with the supplied test client.<br>2. Captures the previous value so it can be restored later.<br>3. The returned closure, when invoked, resets `clientsHolder` back to its original value. No other global state is modified. |
| **How it fits in the package** | - The `clientsholder` package abstracts access to a Kubernetes client across the codebase.<br>- Production code calls a getter (e.g., `GetClient()`) which reads from `clientsHolder`.<br>- Tests use `SetTestK8sClientsHolder` to inject mock clients and automatically clean up after themselves, ensuring test isolation. |

### Typical Usage Pattern

```go
func TestSomething(t *testing.T) {
    // create a fake client (from controller-runtime or client-go mocks)
    fakeClient := kubernetes.NewSimpleClientset()

    // replace the global holder; get cleanup function
    cleanup := SetTestK8sClientsHolder(fakeClient)
    defer cleanup()          // restore original state after test

    // run code that depends on the client
    err := DoSomethingThatUsesClient()
    require.NoError(t, err)
}
```

### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[Test starts] --> B[SetTestK8sClientsHolder(fake)]
    B --> C{clientsHolder overwritten}
    C --> D[Code under test uses fake client]
    D --> E[Cleanup called at defer]
    E --> F[clientsHolder restored to original]
```

This function is the sole public entry point for injecting a test Kubernetes client into the `clientsholder` package, providing deterministic and isolated unit tests.
