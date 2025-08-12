ClientsHolder` – Centralized Kubernetes/OCP Client Store

| Field | Type | Purpose |
|-------|------|---------|
| `KubeConfig` | `[]byte` | Raw kubeconfig used to create all other clients.  The config is read once when the holder is instantiated and stored for debugging or re‑creation. |
| `RestConfig` | `*rest.Config` | Parsed REST configuration derived from `KubeConfig`. It contains authentication, host, TLS, etc., and is reused by every client to avoid repeated parsing. |
| `K8sClient` | `kubernetes.Interface` | Standard clientset for core Kubernetes APIs (pods, services, configmaps…).  Used throughout the suite for common operations. |
| `APIExtClient` | `apiextv1.Interface` | Client for CustomResourceDefinitions (CRDs). Needed to inspect or create CRDs that may be added by operators. |
| `K8sNetworkingClient` | `networkingv1.NetworkingV1Interface` | Client for networking‑related resources such as NetworkPolicies and Ingresses. |
| `CNCFNetworkingClient` | `cncfNetworkAttachmentv1.Interface` | Client for the CNCF Network Attachment Definition CRD, used by OLM or other operators that create network attachments. |
| `MachineCfg` | `ocpMachine.Interface` | OCP‑specific machine configuration client (e.g., MachineSets, Machines).  Required when tests need to manipulate cluster nodes. |
| `OcpClient` | `clientconfigv1.ConfigV1Interface` | Client for the OCP `Config` CRD, which stores cluster‑wide configuration objects. |
| `OlmClient` | `olmClient.Interface` | Operator Lifecycle Manager client for generic OLM resources (ClusterServiceVersion, InstallPlan, etc.). |
| `OlmPkgClient` | `olmpkgclient.PackagesV1Interface` | OLM package client for accessing PackageManifest objects. |
| `DiscoveryClient` | `discovery.DiscoveryInterface` | Provides introspection of available API groups and resources; used during holder initialization to populate `GroupResources`. |
| `DynamicClient` | `dynamic.Interface` | Generic client that can operate on arbitrary unstructured resources, useful for tests or operators that create custom objects. |
| `ScalingClient` | `scale.ScalesGetter` | Client to retrieve scale sub‑resources (e.g., deployments, replicasets).  Used when the suite needs to adjust replica counts. |
| `GroupResources` | `[]*metav1.APIResourceList` | Cached list of all API resources discovered during initialization. This avoids repeated discovery calls and speeds up subsequent operations. |
| `ApiserverClient` | `apiserverscheme.Interface` | Client for the internal `apiserver` APIs (e.g., admission webhook configuration).  Mostly used by tests that manipulate admission policies. |
| `ready` | `bool` | Indicates whether the holder was successfully populated with all clients during construction.  Consumers should check this flag before using the holder to avoid nil‑pointer dereference. |

### How It Is Built

The singleton is created via:

1. **`newClientsHolder(...string)`** – Loads kubeconfig, builds a `*rest.Config`, and then constructs each client by calling the appropriate `NewForConfig` or `NewSimpleClientset` functions.
2. The function also performs discovery (`ServerPreferredResources`) to fill `GroupResources`.
3. If any client fails to instantiate, an error is returned; callers typically log fatally.

The public helpers `GetClientsHolder`, `GetNewClientsHolder`, and `GetTestClientsHolder` provide convenient ways to obtain the singleton or a test‑specific mock holder.

### Key Dependencies

* **k8s.io/client-go** – Core clientset, discovery, dynamic, rest, etc.
* **openshift-client-go** – OCP specific clients (`clientconfigv1`, `machine`, `apiserverscheme`).
* **cncf/kn-plugin-network-attachment-definition** – CNCF networking CRD client.
* **operator-framework/operator-lifecycle-manager** – OLM client.

These dependencies are imported at the package level, and the holder simply holds references to their interfaces.

### Side Effects & Usage

* The holder is **read‑only** after construction; it does not modify any global state.  
* All clients share the same `RestConfig`, ensuring consistent authentication and TLS across all operations.
* `ExecCommandContainer` (the only method on the struct) uses the Kubernetes client to run arbitrary commands inside a container, which is essential for tests that need to inspect logs or internal state.

### Where It Fits in the Package

The `clientsholder` package abstracts away the complexity of creating and configuring multiple Kubernetes/OCP clients.  Code throughout *certsuite* imports this single holder instead of repeatedly constructing individual clients, leading to:

1. **Consistency** – All parts use the same kubeconfig and REST settings.
2. **Testability** – The test helper `GetTestClientsHolder` replaces real clients with mocked ones for unit tests.
3. **Convenience** – Consumers can access any needed client through a single struct (`ClientsHolder`) without worrying about initialization details.

> **Note:** The holder is intentionally *not* thread‑safe beyond the initial construction; all fields are immutable after setup, so concurrent read access is safe.
