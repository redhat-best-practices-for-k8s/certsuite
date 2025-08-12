GetCsiDriver`

| Aspect | Details |
|--------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/diagnostics` |
| **Exported?** | Yes (`func GetCsiDriver() map[string]interface{}`) |
| **Purpose** | Retrieve a list of CSI (Container Storage Interface) drivers that are installed on the Kubernetes cluster and return them as a serialisable map. |

### High‑level flow

1. **Get client holder**  
   `GetClientsHolder()` returns a struct containing pre‑configured controller/clientsets for various API groups.

2. **List CSI driver resources**  
   Using the `clientset` from the holder, the function calls
   ```go
   csidrv.List(ctx, metav1.ListOptions{})
   ```
   where `csidrv` is obtained by:
   - Creating a new scheme (`NewScheme()`)
   - Adding the CSI CRD types to it (`AddToScheme(scheme)`).
   The list operation returns a Kubernetes API response containing all `CSIDriver` objects.

3. **Encode the result**  
   The raw API object is marshalled into JSON with:
   ```go
   codec := NewCodecFactory(legacyCodec)
   data, _ := codec.Encode(&obj, nil)
   ```
   (the legacy codec is used because CSI CRDs are versioned against `v1`).

4. **Return**  
   The JSON bytes are unmarshalled into a generic `map[string]interface{}` so that callers can consume the driver list in a language‑agnostic way.

### Dependencies

| Dependency | Role |
|------------|------|
| `GetClientsHolder()` | Supplies clientsets for interacting with the Kubernetes API. |
| `CSIDrivers` | Type representing the CSI driver CRD. |
| `StorageV1` | Provides access to the storage API group where CSI drivers live. |
| `NewScheme`, `AddToScheme` | Build a runtime scheme that knows about CSI types. |
| `LegacyCodec`, `NewCodecFactory` | Handle serialization/deserialization of Kubernetes objects in their legacy API version format. |
| `Encode`, `Unmarshal` | Convert between Go structs and JSON representation. |

### Side effects & error handling

* The function logs errors (via the package’s logger, not shown here) when any step fails but **does not** propagate them – instead it simply returns an empty map.
* No modification of cluster state; only read operations are performed.

### Usage context

`GetCsiDriver()` is part of the diagnostics toolkit that collects system information for certsuite.  
It allows other diagnostic modules to inspect which CSI drivers are present, enabling checks such as:

```go
drivers := diagnostics.GetCsiDriver()
for name := range drivers {
    // perform driver‑specific tests
}
```

This makes `GetCsiDriver` a key data provider in the overall diagnostic workflow.
