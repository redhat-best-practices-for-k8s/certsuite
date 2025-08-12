getClusterRestConfig`

### Purpose
`getClusterRestConfig` resolves a Kubernetes **REST configuration** that can be used by the rest of the CertSuite client code to talk to a cluster.  
It accepts an arbitrary number of kube‑config file paths (`...string`). The function tries the following in order:

1. If no path is supplied, it falls back to **in‑cluster** configuration (i.e., the service account that the Pod runs under).
2. If a single explicit path is provided, it loads that kube‑config.
3. If multiple paths are supplied, it attempts to merge them by reading each file and combining the resulting `clientcmdapi.Config` objects.

The returned `*rest.Config` can be passed directly to any Kubernetes client (`kubernetes.NewForConfig`, controller runtime, etc.).

### Inputs
| Parameter | Type | Meaning |
|-----------|------|---------|
| `...string` | `[]string` (variadic) | Paths to kube‑config files. An empty slice means “use in‑cluster config”. |

> **Note:** The function treats the slice as a list of file paths; it does not support other forms of configuration such as inline data or environment variables.

### Outputs
| Return value | Type | Meaning |
|--------------|------|---------|
| `*rest.Config` | pointer to `k8s.io/client-go/rest.Config` | The resolved REST config that can be used by Kubernetes clients. |
| `error` | error | Non‑nil if configuration could not be loaded or merged. |

### Key Dependencies & Calls
The function relies on the standard Kubernetes client‑go helpers:

- **`InClusterConfig()`** – obtains a REST config from the Pod’s service account.
- **`NewNonInteractiveDeferredLoadingClientConfig(...)`** – loads kube‑config files lazily, merging multiple sources when needed.
- **`GetClientConfigFromRestConfig(cfg *rest.Config)`** – converts a `*rest.Config` into a `clientcmd.ClientConfig` (needed for merging).
- **`createByteArrayKubeConfig([]byte) []byte`** – helper that creates an in‑memory kube‑config from raw bytes; used when merging.
- **`RawConfig()`** – extracts the underlying `clientcmdapi.Config`.
- **`NewDefaultClientConfigLoadingRules()`** – sets up default file search paths if none are supplied.

Logging helpers (`Info`, `Errorf`) provide diagnostic output but do not affect the returned config.

### Side Effects
* None beyond returning values and logging.  
* The function does **not** modify global state or mutate the input slice.

### How It Fits in the Package
`clientsholder` is responsible for keeping a singleton Kubernetes client that can be reused across CertSuite components. `getClusterRestConfig` is a low‑level helper used by the package’s public API (e.g., `GetClient`) to create or refresh this client when configuration changes are detected.

In practice:

```go
cfg, err := getClusterRestConfig(kubeCfgPaths...)
if err != nil {
    // handle error
}
clientset, err := kubernetes.NewForConfig(cfg)
```

The function is deliberately unexported because callers should use the higher‑level `GetClient` or similar abstractions that hide the intricacies of merging multiple kube‑config files.
