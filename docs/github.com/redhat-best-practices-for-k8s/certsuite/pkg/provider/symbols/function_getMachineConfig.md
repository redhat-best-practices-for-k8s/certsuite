getMachineConfig`

| Aspect | Details |
|--------|---------|
| **Signature** | `func(name string, mcs map[string]MachineConfig) (MachineConfig, error)` |
| **Visibility** | Unexported – used only inside the *provider* package. |

### Purpose
`getMachineConfig` retrieves a single OpenShift Machine‑Configuration object that matches a given name.  
The function first checks an in‑memory cache (`mcs`). If the requested configuration is present, it returns it immediately.  
If not cached, it uses the OpenShift Machine Configuration API to fetch the object from the cluster, unmarshals its JSON payload into a `MachineConfig` struct and stores it in the cache for future calls.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `name` | `string` | The Kubernetes name of the Machine‑Configuration to fetch. |
| `mcs` | `map[string]MachineConfig` | A map acting as a simple LRU/lookup cache keyed by configuration name. |

### Return values
| Value | Type | Description |
|-------|------|-------------|
| first return | `MachineConfig` | The requested configuration (either from the cache or freshly fetched). |
| second return | `error` | Non‑nil if the configuration cannot be found, the API call fails, or JSON unmarshalling errors. |

### Key Dependencies
* **Cluster clients** – Obtained via `GetClientsHolder()` and then `Get(MachineconfigurationV1())`.  
  These provide access to the OpenShift Machine Configuration API.
* **Machine‑Configuration clientset** – The returned `machineconfigurationsv1.MachineConfigurationClient` is used to call `MachineConfigs().Get(...)`.
* **JSON handling** – `json.Unmarshal()` deserialises the raw response into a `MachineConfig`.
* **Logging / error handling** – Errors are wrapped with `fmt.Errorf("...: %w", err)` for context.

### Side effects
* The function mutates the provided cache map by inserting the fetched configuration (`mcs[name] = mc`).
* No global state is changed; all side‑effects are confined to the supplied map and local variables.

### How it fits in the package
The *provider* package orchestrates test execution against an OpenShift cluster.  
Many tests need information from machine‑configuration objects (e.g., to verify CPU topology, huge pages).  
`getMachineConfig` is a lightweight helper that abstracts:

1. **Cache lookup** – prevents redundant API calls.
2. **Cluster retrieval** – hides the specifics of constructing the client and handling the response.

Other functions in this package call `getMachineConfig` whenever they need to inspect or validate a Machine‑Configuration, ensuring consistent error handling and reuse of cached data across tests.
