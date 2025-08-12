SetTestClientGroupResources`

```go
func SetTestClientGroupResources(resources []*metav1.APIResourceList) func()
```

| Aspect | Detail |
|--------|--------|
| **Purpose** | Provides a test helper that injects a custom list of API resource groups into the package‑wide `clientsHolder`. This allows unit tests to simulate different Kubernetes discovery responses without contacting a real cluster. |
| **Parameters** | `resources` – a slice of pointers to `metav1.APIResourceList`, each describing one API group (e.g., `"apps/v1"`, `"batch/v1beta1"`). The list is what the fake client will report when it performs discovery. |
| **Return value** | A *zero‑argument* function (`func()`) that, when invoked, applies the supplied `resources` to the global `clientsHolder`.  This design lets tests defer applying the mock until they are ready (e.g., after creating a fake client). |
| **Key side effects** | • Mutates the package’s internal `clientsHolder` variable. <br>• Replaces any previously cached discovery information so that subsequent calls to the holder will use the new list. |
| **Dependencies** | * `metav1.APIResourceList` – from `k8s.io/apimachinery/pkg/apis/meta/v1`. <br>* The unexported global `clientsHolder` defined in this package (see line 82 of `clientsholder.go`). |
| **How it fits the package** | The `clientsholder` package centralises all Kubernetes client instances used by Certsuite. Tests need to control what API groups are reported by those clients; `SetTestClientGroupResources` is the canonical way to do that. It’s typically called in a test’s setup phase, e.g.:

```go
func TestSomething(t *testing.T) {
    // Prepare fake discovery data.
    resources := []*metav1.APIResourceList{{
        GroupVersion: "v1",
        APIResources: []metav1.APIResource{{Name:"pods", Kind:"Pod"}},
    }}

    // Install the mock into the holder.
    apply := clientsholder.SetTestClientGroupResources(resources)
    defer apply() // Apply when ready.

    // ... run test that uses the clients …
}
```

This helper keeps tests isolated from a real Kubernetes API server while still exercising code paths that depend on discovery data.
