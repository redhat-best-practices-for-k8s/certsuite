SetupFakeOlmClient`

```go
func SetupFakeOlmClient([]runtime.Object) func()
```

## Purpose

`SetupFakeOlmClient` is a helper used by the unit‑test suite to replace the real **Operator Lifecycle Manager (OLM)** client with a fake implementation.  
The function:

1. Builds an in‑memory Kubernetes client (`*fake.Clientset`) that knows about a set of pre‑created objects supplied by the caller.
2. Stores this fake client in the package‑level `clientsHolder` so that subsequent calls to `GetClient()` return the mock instead of contacting a real cluster.
3. Returns a **cleanup** function that restores the original client when the test finishes.

This allows tests to exercise OLM‑dependent logic without needing an actual Kubernetes API server or OLM installation.

## Parameters

| Name | Type | Description |
|------|------|-------------|
| `objects` | `[]runtime.Object` | A slice of Kubernetes objects (e.g., CRDs, ClusterServiceVersions) that should be pre‑loaded into the fake client. These are used by the mocked OLM interface to answer queries such as *“does this CSV exist?”*.

## Return Value

| Type | Description |
|------|-------------|
| `func()` | A closure that, when invoked, restores the original real client in `clientsHolder`. The caller typically defers this function in a test (`defer setup()`). |

## Key Dependencies & Calls

- **`k8s.io/client-go/kubernetes/fake.NewSimpleClientset`**  
  Creates the fake client from the supplied objects.  
- **`clientsholder.clientsHolder`** (unexported package variable)  
  Holds the current `*fake.Clientset`. The function writes to this variable and returns a closure that resets it.

## Side‑Effects

1. **Global state mutation** – The package’s internal client holder is overwritten, affecting any code that reads from `GetClient()` during the test.
2. **Test isolation** – Because the original client is restored via the returned cleanup function, tests can safely run in parallel without leaking state.

## Package Context

`clientsholder` provides a thin abstraction over the Kubernetes client used by CertSuite.  
Other parts of the code call `GetClient()` to perform OLM‑related operations (e.g., installing CRDs).  
During normal operation this returns a real client; during tests `SetupFakeOlmClient` swaps it out for a deterministic fake, enabling fast and reliable unit tests.

---

### Mermaid diagram (optional)

```mermaid
flowchart TD
    A[Test] --> B{Calls SetupFakeOlmClient}
    B --> C[Creates fake.Clientset from objects]
    C --> D[stores in clientsHolder]
    D --> E[Test code uses GetClient() → fake client]
    E --> F{Runs test logic}
    F --> G[Returns cleanup closure]
    G --> H[defer cleanup()] 
    H --> I[Restores original client in clientsHolder]
```

This diagram illustrates the lifecycle of the fake client during a test.
