DiscoveredTestData` – Autodiscovery Result

The **`DiscoveredTestData`** type is the central data holder returned by the package’s auto‑discovery routine (`DoAutoDiscover`).  
It aggregates all Kubernetes objects and runtime information that a test run might need to validate against.

| Section | Description |
|---------|-------------|
| **Purpose** | Collects every resource that can influence or be influenced by the test subject, together with environment metadata. The struct is serialised (or inspected) downstream in test suites, telemetry or reporting. |
| **Key Dependencies** | *Kubernetes client-go* types (`corev1`, `appsv1`, …), *Operator Lifecycle Manager* types (`olmv1Alpha`, `olmPkgv1`), OpenShift config types (`configv1`) and several vendor‑specific extensions (e.g., `nadClient`). The discovery logic pulls these via the configured client set. |
| **Side Effects** | None – the struct is a plain data container. Populating it may trigger API calls in the discovery routine, but once constructed no further mutation occurs unless callers modify its fields directly. |

## Field Overview

> Fields are grouped by resource type or metadata category.

### Kubernetes Objects

| Category | Representative Types | Typical Use |
|----------|----------------------|-------------|
| **Control‑Plane** | `ClusterOperators []configv1.ClusterOperator`<br>`Nodes *corev1.NodeList` | Inspect operator health and node status. |
| **RBAC** | `Roles []rbacv1.Role`<br>`RoleBindings []rbacv1.RoleBinding`<br>`ClusterRoleBindings []rbacv1.ClusterRoleBinding` | Validate access controls around the test subject. |
| **Workloads** | `Deployments []appsv1.Deployment`<br>`StatefulSet []appsv1.StatefulSet`<br>`Pods []corev1.Pod`<br>`OperandPods []*corev1.Pod`<br>`ProbePods []corev1.Pod`<br>`Hpas []*scalingv1.HorizontalPodAutoscaler` | Capture the full pod set that might be under test. |
| **Networking** | `Services []*corev1.Service`<br>`NetworkPolicies []networkingv1.NetworkPolicy`<br>`NetworkAttachmentDefinitions []nadClient.NetworkAttachmentDefinition` | Verify service endpoints and policy enforcement. |
| **Storage** | `PersistentVolumes []corev1.PersistentVolume`<br>`PersistentVolumeClaims []corev1.PersistentVolumeClaim`<br>`StorageClasses []storagev1.StorageClass` | Ensure volume configuration matches expectations. |
| **Operator Lifecycle Manager (OLM)** | `AllCsvs []*olmv1Alpha.ClusterServiceVersion`<br>`AllSubscriptions []olmv1Alpha.Subscription`<br>`AllInstallPlans []*olmv1Alpha.InstallPlan`<br>`AllCatalogSources []*olmv1Alpha.CatalogSource`<br>`AllPackageManifests []*olmPkgv1.PackageManifest` | Capture OLM‑managed components that may provide the test’s runtime. |
| **CRDs** | `AllCrds []*apiextv1.CustomResourceDefinition` | Detect custom resources in the cluster. |

### Custom / Vendor Resources

| Field | Type | Notes |
|-------|------|-------|
| `SriovNetworks []unstructured.Unstructured`<br>`SriovNetworkNodePolicies []unstructured.Unstructured` | `[]unstructured.Unstructured` | Generic holder for SR‑IOV CRs when API types are not compiled in. |
| `HelmChartReleases map[string][]*release.Release` | Helm release objects (from the `helm.sh/helm/v3/pkg/release` package) | Useful for tests that verify Helm‑deployed workloads. |

### Metadata & Runtime Information

| Field | Description |
|-------|-------------|
| `K8sVersion string`<br>`OpenShiftVersion string` | Cluster version strings extracted from the API server. |
| `OCPStatus string` | Overall OpenShift health status (e.g., “Ready”). |
| `CollectorAppEndpoint, CollectorAppPassword` | Credentials for sending telemetry back to CertSuite’s collector service. |
| `ConnectAPI*` fields | Configuration used by the Connect component (base URL, key, proxy). |
| `Env configuration.TestParameters` | Parameters passed into the test run (e.g., namespace list, ignore lists). |
| `ExecutedBy string` | User or CI job that triggered discovery. |
| `PartnerName string` | Optional partner identifier used for telemetry. |

### Derived / Helper Data

| Field | Purpose |
|-------|---------|
| `CSVToPodListMap map[types.NamespacedName][]*corev1.Pod` | Quick lookup of pods belonging to a particular CSV (useful for status checks). |
| `ScaleCrUnderTest []ScaleObject` | Scale objects that are being tested, if any. |
| `PodStates PodStates` | Summary counts of pod statuses (running, pending, etc.). |
| `AbnormalEvents []corev1.Event` | Events flagged as abnormal during discovery (e.g., failed starts). |

## Usage Flow

```go
// 1. Build a TestConfiguration
cfg := configuration.NewTestConfig(...)

// 2. Discover everything relevant
data := autodiscover.DoAutoDiscover(cfg)

// 3. Inspect or serialize the result
fmt.Println("Namespaces found:", len(data.AllNamespaces))
jsonBytes, _ := json.MarshalIndent(data, "", "  ")
```

The returned `DiscoveredTestData` can then be passed to validation functions, sent to a remote collector, or simply logged for debugging.

---

### Suggested Mermaid Diagram

```mermaid
graph TD
    A[DoAutoDiscover] --> B{Populate}
    B --> C[AllNamespaces]
    B --> D[AllPods]
    B --> E[ClusterOperators]
    B --> F[OLM Resources]
    B --> G[Metadata (K8sVersion, etc.)]
```

This diagram visualises how `DoAutoDiscover` drives the population of each major section within `DiscoveredTestData`.
