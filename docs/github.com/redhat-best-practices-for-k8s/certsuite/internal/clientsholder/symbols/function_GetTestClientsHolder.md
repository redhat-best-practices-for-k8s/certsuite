GetTestClientsHolder`

```go
func GetTestClientsHolder([]runtime.Object) *ClientsHolder
```

## Purpose

`GetTestClientsHolder` is a helper used exclusively in unit tests to replace the global **client holder** with a mock implementation that only exposes pure Kubernetes interfaces.  
It takes a slice of `runtime.Object`s (Kubernetes API objects) and builds an in‑memory clientset from them, then stores this mocked holder in the package’s internal variable so that other parts of the test can use it transparently.

## Parameters

| Name | Type | Description |
|------|------|-------------|
| `objs` | `[]runtime.Object` | A list of pure Kubernetes objects (e.g. Pods, Services) that will be pre‑seeded into the fake clientset. Objects belonging to non‑K8s APIs (OLM, CRDs, etc.) should **not** be passed; they must be mocked with a dedicated clientset builder elsewhere.

## Return Value

| Type | Description |
|------|-------------|
| `*ClientsHolder` | A pointer to the newly created mock holder. The caller can inspect or use it directly, but normally the function just mutates the package‑level variable `clientsHolder`.

## Key Dependencies & Calls

- **`clientset.NewSimpleClientset`** – called three times to create fake clientsets for different API groups (core, apps, batch).  
  These are seeded with the supplied objects.
- **Multiple `append` calls** – build slices of clients and informers that will be stored in the holder.
- The function mutates the package‑level variable `clientsHolder`, overwriting any existing real client holder.

## Side Effects

1. **Global state mutation**: `clientsHolder` is replaced with a new instance, affecting all code paths that read this variable after the call.
2. **No external I/O**: The function works purely in memory; no network or disk operations occur.
3. **No error handling**: It assumes the objects are valid and that clientset construction succeeds.

## Package Context

`clientsholder` provides a singleton that holds all Kubernetes clients, informers, and other related utilities used throughout CertSuite. In production code this singleton is populated with real `k8s.io/client-go` clients.  
For unit tests, however, the test suite cannot rely on an actual cluster, so it uses `GetTestClientsHolder` to supply a fake clientset that mimics Kubernetes behavior.

```mermaid
flowchart TD
    Test[Unit Test] -->|Calls| GetTestClientsHolder
    GetTestClientsHolder -->|Creates| FakeClientSet
    FakeClientSet -->|Populates| clientsHolder (global)
    clientsHolder -->|Used by| other packages (e.g., controllers, reconcilers)
```

## Summary

- **What it does**: Builds a fake Kubernetes clientset from given objects and installs it as the global client holder for tests.  
- **When to use**: In any test that requires Kubernetes API interactions but should not hit a real cluster.  
- **Caveats**: Only pure K8s objects are accepted; other APIs must be mocked separately. The function overwrites global state, so it should be called early in the test setup.
