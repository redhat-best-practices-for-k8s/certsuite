getAllCatalogSources`

| Aspect | Details |
|--------|---------|
| **Purpose** | Collects every *CatalogSource* object that exists in a Kubernetes cluster and returns them as a slice of pointers to the Open‑Shift Operators `v1alpha1.CatalogSource` type. |
| **Signature** | `func getAllCatalogSources(client v1alpha1.OperatorsV1alpha1Interface) []*olmv1Alpha.CatalogSource` |
| **Parameters** | * `client`: An interface that provides access to the OpenShift Operators API (`v1alpha1`). It must support the `CatalogSources()` method, which in turn offers a `List(context.Context, metav1.ListOptions)` call. |
| **Return value** | A slice of pointers to `olmv1Alpha.CatalogSource`. If an error occurs while listing, the function logs it (via the package‑level `TODO` placeholder) and returns whatever has been collected up to that point—potentially an empty slice. |

### How it works

```text
client.CatalogSources().List(ctx, metav1.ListOptions{}) → List of CatalogSource objects
```

* The function calls `CatalogSources()` on the provided client, then immediately calls `List` with a context and default list options.
* On success, the returned `runtime.Object` is asserted to a `*olmv1Alpha.CatalogSourceList`. Each item in that list is appended to the result slice.
* If any step fails (client call or type assertion), an error message is emitted (`TODO: log the error`) and the function returns whatever has been collected so far.

### Dependencies

| Dependency | Role |
|------------|------|
| `v1alpha1.OperatorsV1alpha1Interface` | Provides access to OpenShift Operators resources. |
| `olmv1Alpha.CatalogSource`, `CatalogSourceList` | Types representing the catalog source objects returned by the API. |
| `TODO`, `Error` | Placeholder for proper error handling/logging (currently unimplemented). |

### Side effects

* **No state mutation** – The function only reads from the cluster and builds a new slice; it does not modify any cluster resources or package globals.
* **Logging** – Currently a `TODO` placeholder is used, so no real side effect occurs. In a production build this should be replaced with structured logging.

### Package context

In the *autodiscover* package, this helper underpins higher‑level discovery logic that needs to know which operator catalogs are available in the cluster (e.g., for determining which operators can be installed or validated). It is used by other functions such as `discoverOperators` and `getOperatorCatalogSources`. By abstracting away the raw API call, it keeps the rest of the package focused on business logic rather than low‑level client handling.
