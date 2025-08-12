Pod` – High‑level representation of a Kubernetes pod

| Field | Type | Purpose |
|-------|------|---------|
| `embedded:*corev1.Pod` | *corev1.Pod | The underlying Kubernetes API object.  All standard pod fields (metadata, spec, status…) are available through this embedded value. |
| `Containers []*Container` | slice of *Container | Convenience list of the pod’s containers.  Populated by `NewPod` via `getPodContainers`.  Used by many checks that operate on container‑level attributes. |
| `IsOperand bool` | flag | Indicates whether the pod belongs to an **operand** (a component installed by a higher‑level operator).  Influences filtering logic in the test environment. |
| `IsOperator bool` | flag | Marks pods that are part of an **operator** itself.  Used when adding operator pods to the test set (`addOperatorPodsToTestPods`). |
| `MultusNetworkInterfaces map[string]CniNetworkInterface` | mapping | Holds parsed Multus‑network interface data (name → details).  Filled by `NewPod` using the `CNFC` annotation and `GetClientsHolder`.  Needed for SR‑IOV checks. |
| `MultusPCIs []string` | slice | List of PCI device strings associated with the pod’s Multus interfaces.  Populated in `NewPod`. |
| `SkipMultusNetTests bool` | flag | If true, all network‑related tests that rely on Multus are skipped for this pod.  Set during construction based on operator/operand status or user configuration. |
| `SkipNetTests bool` | flag | Generic skip flag for any network test.  Useful when the pod is known to be a system component that shouldn’t be evaluated for networking compliance. |

### Construction

```go
func NewPod(pod *corev1.Pod) Pod
```

*Creates a `Pod` wrapper around a raw Kubernetes pod.*  
The function:

1. Initializes the embedded pod.
2. Calls helper functions to fetch:
   * IPs per network (`GetPodIPsPerNet`)
   * Annotations & labels (`GetAnnotations`, `GetLabels`)
   * PCI info (`GetPciPerPod`)
3. Builds container list via `getPodContainers`.
4. Detects operator/operand status and sets skip flags accordingly.

No state is mutated outside the returned struct; `NewPod` is pure with respect to global variables (only reads configuration).

### Key Methods & How They Use Fields

| Method | What it checks | Uses which fields |
|--------|----------------|-------------------|
| `IsPodGuaranteed()` | Checks if requested and limits are equal. | Reads `embedded.Spec.Resources`. |
| `IsPodGuaranteedWithExclusiveCPUs()` | Adds CPU‑isolation check to guarantee test. | Calls `AreCPUResourcesWholeUnits` & `AreResourcesIdentical`; uses pod resources. |
| `IsCPUIsolationCompliant()` | Verifies load balancing is disabled and runtime class set. | Uses `LoadBalancingDisabled`, `IsRuntimeClassNameSpecified`. |
| `HasHugepages()` | Detects any container with a hugepage resource name. | Iterates over `Containers`. |
| `GetRunAsNonRootFalseContainers()` | Returns containers that run as root (`runAsNonRoot=false` and user id 0). | Uses `Containers`, pod‑level defaults. |
| `IsUsingSRIOV()` / `IsUsingSRIOVWithMTU()` | Checks if any network interface is SR‑IOV (and optionally MTU set). | Reads `MultusNetworkInterfaces` and consults the NetworkAttachmentDefinition via `GetClientsHolder`. |
| `AffinityRequired()` | Parses the `affinity.required` annotation. | Uses `GetAnnotations`. |
| `IsShareProcessNamespace()` | Returns pod.Spec.ShareProcessNamespace flag. | Directly from embedded pod. |
| `String()` | Human‑readable representation (`<namespace>/<name>`). | Uses `embedded.Namespace`, `embedded.Name`. |

### Dependencies & Side Effects

* **External packages** – relies on `k8s.io/api/core/v1` for the core Pod type, and on helper functions in the same package (e.g., `GetClientsHolder`, `GetPodIPsPerNet`).  
* **Logging** – many methods call `Debug`, `Info`, or `Warn`; these write to the global logger but do not modify the pod itself.  
* **State changes** – none of the public methods mutate the Pod struct; they only read fields and return boolean results.

### How It Fits the Package

The `provider` package is responsible for translating raw Kubernetes objects into richer, test‑ready structures (`Pod`, `Container`, etc.) and providing a suite of compliance checks.  
`Pod` sits at the core:

* **Data aggregation** – it aggregates pod spec data with runtime information (PCI devices, Multus interfaces).  
* **Query interface** – the methods expose all the predicates needed by the test environment (`TestEnvironment`) to filter pods into categories such as guaranteed, using SR‑IOV, or requiring affinity.  
* **Extensibility** – new checks can be added simply by adding a method that inspects existing fields; no global state is required.

In short, `Pod` is the bridge between raw Kubernetes objects and the compliance logic used throughout CertSuite.
