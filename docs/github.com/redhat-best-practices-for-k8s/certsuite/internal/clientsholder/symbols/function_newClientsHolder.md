newClientsHolder`

| Aspect | Details |
|--------|---------|
| **Package** | `clientsholder` (`github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder`) |
| **Visibility** | Unexported (internal helper) |
| **Signature** | `func(...string) (*ClientsHolder, error)` |
| **Purpose** | Create a fully‑initialized `ClientsHolder` for an OpenShift cluster. The function accepts one or more kubeconfig file paths and returns either a populated holder or an error if any client could not be constructed. |

### How it works

1. **Logging & Setup**  
   - Logs the start of client creation with `Info`.  

2. **REST configuration**  
   - Calls `getClusterRestConfig` (which merges kubeconfig paths and sets defaults).  

3. **Client construction**  
   For each API group required by CertSuite, it calls `clientcmd.NewForConfig` or the appropriate factory function to obtain a typed client:
   * OpenShift OAuth (`NewForConfig`)
   * OpenShift Projects (`NewForConfig`)
   * OpenShift Route (`NewForConfig`)
   * OpenShift ImageStreams (`NewForConfig`)
   * OpenShift BuildConfigs, Deployments, etc.  
   Errors are wrapped with `Errorf` and returned immediately if any step fails.

4. **Discovery & Mapper**  
   - Builds a discovery client via `NewDiscoveryClientForConfig`.
   - Uses the discovery API to fetch server‑preferred resources (`ServerPreferredResources`) and resolves scale kinds with `NewDiscoveryScaleKindResolver`.
   - Creates an `RESTMapper` with `NewDiscoveryRESTMapper`.

5. **Populate `ClientsHolder`**  
   The resulting clients, discovery client, mapper, and a list of known API groups are stored in a new `ClientsHolder` struct and returned.

### Inputs

| Parameter | Type | Description |
|-----------|------|-------------|
| `...string` | `[]string` | One or more paths to kubeconfig files. The first non‑empty path is used; subsequent ones can override values or add contexts. |

### Outputs

| Return | Type | Description |
|--------|------|-------------|
| `*ClientsHolder` | *struct* | Holds all typed clients, discovery client, REST mapper, and known API groups. |
| `error` | `error` | Non‑nil if any step (config loading, client creation, discovery) fails. |

### Key Dependencies

- **Kubernetes & OpenShift client-go**: `NewForConfig`, `NewDiscoveryClientForConfig`, etc.
- **kubeclient utilities**: `getClusterRestConfig`.
- **Logging helpers**: `Info`, `Errorf`.

### Side Effects

- No global state is mutated; all data lives in the returned `ClientsHolder`.
- The function performs network calls only during discovery (via `ServerPreferredResources`).  
  It can be considered *side‑effect‑free* from a pure code‑analysis standpoint but will contact the API server.

### Package Role

`clientsholder.newClientsHolder` is the core factory for all CertSuite operations that require cluster interaction. The package encapsulates client creation logic so other packages can simply call `GetClientsHolder()` (which internally invokes this function) and receive a ready‑to‑use holder without worrying about configuration or discovery details.

---

#### Suggested Mermaid diagram

```mermaid
flowchart TD
  A[Input kubeconfig paths] --> B[getClusterRestConfig]
  B --> C{Build REST Config}
  C --> D[Create typed clients (OAuth, Projects, Route, …)]
  D --> E[Discovery client]
  E --> F[Discover resources & scale kinds]
  F --> G[RESTMapper]
  G --> H[ClientsHolder struct]
```
