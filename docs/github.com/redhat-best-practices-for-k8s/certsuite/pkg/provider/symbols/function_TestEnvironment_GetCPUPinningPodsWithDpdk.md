TestEnvironment.GetCPUPinningPodsWithDpdk`

| Aspect | Details |
|--------|---------|
| **Package** | `provider` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider`) |
| **Exported?** | Yes |
| **Receiver** | `env TestEnvironment` – the environment that holds Kubernetes client and cached resources. |
| **Signature** | `func (env TestEnvironment) GetCPUPinningPodsWithDpdk() []*Pod` |
| **Purpose** | Return all Pods in the cluster that satisfy *both* of the following conditions:<br>1. They use **CPU pinning** – i.e., their QoS class is *Guaranteed* and they have an `exclusive-cpu-ids` annotation or equivalent CPU allocation.<br>2. They are running with **DPDK** (Data Plane Development Kit) enabled, which is detected by the presence of a DPDK‑specific container or environment variable inside the pod. |
| **Key Dependencies** | 1. `filterDPDKRunningPods()` – filters pods that have a DPDK container/setting.<br>2. `GetGuaranteedPodsWithExclusiveCPUs()` – returns pods that are Guaranteed and use exclusive CPU IDs (i.e., CPU‑pinned).<br>The function composes these two filters: it first obtains the guaranteed‑with‑exclusive‑CPU set, then keeps only those also present in the DPDK set. |
| **Side Effects** | None – purely read‑only; does not modify state or interact with external systems beyond reading cached pod data via `TestEnvironment`. |
| **Return Value** | Slice of pointers to `Pod` objects that meet both criteria. The slice may be empty if no such pods exist. |
| **Usage in the Package** |  
  * Used by higher‑level tests that need to assert behaviour on CPU‑pinned DPDK workloads (e.g., network performance, isolation checks).<br>  * Provides a convenient filter for test suites that want to skip or focus on these pods without reimplementing the logic. |
| **How it Works** | ```go\nfunc (env TestEnvironment) GetCPUPinningPodsWithDpdk() []*Pod {\n    // Step 1: get all DPDK‑running pods.\n    dpdkPods := filterDPDKRunningPods(env.GetAllPods())\n    // Step 2: get all Guaranteed pods with exclusive CPUs.\n    cpuPinnedPods := env.GetGuaranteedPodsWithExclusiveCPUs()\n    // Step 3: intersect the two sets.\n    return intersectPods(dpdkPods, cpuPinnedPods)\n}\n```<br>Where `intersectPods` is a helper that retains only pods present in both slices. The actual implementation relies on set‑like maps keyed by pod UID for efficient intersection. |
| **Mermaid Flow** | ```mermaid\nflowchart TD\n  A[GetAllPods] --> B[filterDPDKRunningPods]\n  C[GetGuaranteedPodsWithExclusiveCPUs] --> D\n  B & D --> E[Intersection]\n  E --> F[Return slice of *Pod]\n```\n*The flow shows the two independent paths that are merged to produce the final result.* |

**Unknowns / Missing Context**

- The internal structure of `Pod` (e.g., fields for annotations, labels) is not visible here; we assume standard Kubernetes API objects.
- Exact detection logic inside `filterDPDKRunningPods` and how exclusive CPU IDs are stored in a pod is not shown; the documentation assumes typical patterns used by certsuite.
