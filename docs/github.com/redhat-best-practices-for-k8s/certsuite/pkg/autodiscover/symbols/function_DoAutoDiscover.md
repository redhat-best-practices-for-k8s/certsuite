DoAutoDiscover`

```go
func DoAutoDiscover(cfg *configuration.TestConfiguration) DiscoveredTestData
```

`DoAutoDiscover` is the entry point for the **autodiscovery** subsystem of CertSuite.  
Its responsibility is to walk a Kubernetes cluster, collect all objects that are relevant for the test suite and return them in a `DiscoveredTestData` value.

| Section | Details |
|---------|---------|
| **Purpose** | Scan the target cluster (or local kubeconfig) for resources such as storage classes, namespaces, operators, pods, events, quotas, etc. The collected data is used by other packages to build test cases and verify certificate‑related behaviour. |
| **Parameters** | `cfg *configuration.TestConfiguration` – a pointer to the global test configuration that contains the kubeconfig path, discovery flags, verbosity settings, and optional filters. |
| **Return value** | `DiscoveredTestData` – a struct (defined elsewhere in the package) that holds slices/maps of discovered objects grouped by type. |

### High‑level flow

1. **Client set up**  
   - Calls `GetClientsHolder(cfg)` to create typed clients for all needed API groups (`corev1`, `storagev1`, `operatorsv1alpha1`, etc.).  
   - If the client creation fails, the function logs a fatal error and exits.

2. **Discover storage classes**  
   - `getAllStorageClasses` retrieves all StorageClass objects via `client.StorageV1()`.  

3. **Discover namespaces**  
   - `getAllNamespaces` pulls every namespace through `client.CoreV1()` and stores them in the result map.  

4. **Operator related discovery** (OpenShift only)  
   - `findSubscriptions`, `getAllInstallPlans`, `getAllCatalogSources`, and `getAllPackageManifests` collect information about Operator Lifecycle Manager objects (`Subscription`, `InstallPlan`, `CatalogSource`, `PackageManifest`).  
   - These calls are guarded by the presence of the `NonOpenshiftClusterVersion` constant; if the cluster is not OpenShift, the operator section is skipped.

5. **Pod discovery**  
   - Pods matching a set of label selectors (`probeHelperPodsLabelName/value`, `labelRegex`) are gathered with `FindPodsByLabels`.  
   - Their status counts are calculated by `CountPodsByStatus`.

6. **Event and quota checks**  
   - `findAbnormalEvents` looks for any non‑Normal events in the cluster.  
   - Resource quotas, PodDisruptionBudgets, NetworkPolicies, and Sriov resources (`SriovNetworkGVR`, `SriovNetworkNodePolicyGVR`) are fetched with their respective helper functions.

7. **Logging** – Throughout the process the function emits debug/info messages via `Debug()`, `Info()`, and error/fatal logs when something goes wrong.

8. **Return** – All gathered objects are stored in a single `DiscoveredTestData` value which is returned to the caller.

### Dependencies & side‑effects

| Dependency | Role |
|------------|------|
| `GetClientsHolder` | Provides typed Kubernetes clients. |
| `StorageV1`, `CoreV1`, `OperatorsV1alpha1`, `PolicyV1` | API group constructors for clientset. |
| `CreateLabels` | Builds label maps used when filtering pods. |
| Logging helpers (`Fatal`, `Debug`, `Info`) | Emit diagnostic output; fatal logs terminate the process. |
| Constants (`labelRegex`, `probeHelperPodsLabelName/value`, etc.) | Define label selectors and naming conventions for discovery. |

The function has **no external side‑effects** beyond logging; it only reads from the cluster. All discovered data is encapsulated in its return value.

### Integration with the package

`DoAutoDiscover` sits at the top of the autodiscover hierarchy.  
Other packages call it once during test initialization to obtain a snapshot of the current cluster state, which is then used by test runners to generate certificate tests and validate operator behaviour.

A concise diagram of its high‑level control flow:

```mermaid
flowchart TD
  A[Start] --> B{GetClientsHolder}
  B -->|ok| C[Discover StorageClasses]
  B -->|ok| D[Discover Namespaces]
  D --> E{OpenShift?}
  E -- yes --> F[Operator Discovery]
  E -- no --> G[Skip Operators]
  C & D & (F/G) --> H[Discover Pods]
  H --> I[Count Pod Statuses]
  I --> J[Find Abnormal Events]
  J --> K[Get Resource Quotas]
  K --> L[Get PDBs, NetPolys, Sriov]
  L --> M[Return DiscoveredTestData]
```

This function is the backbone of CertSuite’s automatic test configuration.
