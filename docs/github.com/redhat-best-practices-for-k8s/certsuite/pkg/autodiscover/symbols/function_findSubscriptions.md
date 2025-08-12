findSubscriptions`

```go
func (client v1alpha1.OperatorsV1alpha1Interface, namespace []string) []olmv1Alpha.Subscription
```

### Purpose  
`findSubscriptions` discovers all **Operator Lifecycle Manager (OLM)** subscriptions that exist in the supplied Kubernetes namespaces.  
The returned slice contains the `Subscription` objects that will later be inspected for operator CSVs and labels.

> The function is intentionally unexported; it is used internally by the autodiscover package to gather operator data before other discovery routines run.

### Parameters  

| Name | Type | Description |
|------|------|-------------|
| `client` | `v1alpha1.OperatorsV1alpha1Interface` | A typed OLM client that exposes the `Subscriptions()` lister. It is created by a Kubernetes rest‑config in the calling code. |
| `namespace` | `[]string` | List of namespace names to search for subscriptions. The function iterates over each entry and queries OLM for all subscriptions inside it. |

### Return Value  

* `[]olmv1Alpha.Subscription` – A slice containing every subscription found across the supplied namespaces.  
  * If no namespaces are provided or none contain subscriptions, an empty slice is returned.
  * Errors encountered during listing are logged but **not** surfaced to the caller; instead they are swallowed after printing to the log.

### Key Dependencies  

| Called Function | Purpose |
|-----------------|---------|
| `Debug` / `Info` | Logging helpers from the package’s logger. Used for tracing and debugging. |
| `client.Subscriptions(namespace).List(...)` | OLM API call that fetches all subscriptions in a namespace. |
| `append` | Builds the result slice. |

The function relies on the **v1alpha1** client from the OLM SDK; therefore it requires a Kubernetes cluster with OLM installed and the relevant RBAC permissions to list subscriptions.

### Side Effects  

* Emits log entries for each namespace processed, successful queries, and errors.
* Does **not** modify any external state or the returned objects. The slice is a shallow copy of the API responses.

### How it Fits the Package

`findSubscriptions` is a helper used by higher‑level autodiscover logic (e.g., `discoverOperators`) to gather operator data.  
Once subscriptions are collected, other functions in the package:

1. Resolve each subscription’s CSV name.
2. Inspect the CSV for labels and annotations that indicate whether the operator should be included in the certsuite run.
3. Derive target namespaces or deployments from those labels.

By isolating the subscription discovery step into its own function, the autodiscover package keeps a clear separation between **data acquisition** (this function) and **business logic** (label parsing, policy evaluation).
