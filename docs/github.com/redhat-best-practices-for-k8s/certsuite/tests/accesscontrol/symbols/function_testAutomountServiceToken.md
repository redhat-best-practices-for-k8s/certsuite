testAutomountServiceToken`

| Aspect | Details |
|--------|---------|
| **Purpose** | Validates that every pod in the cluster adheres to the “automount‑service‑token” policy:  
  * the pod must be bound to the default service account (`default`), **and**  
  * the pod spec must explicitly set `automountServiceAccountToken: false`. If either condition fails, the pod is considered non‑compliant. |
| **Inputs** | - `check *checksdb.Check`: mutable object representing the current compliance check being executed.  <br>- `env *provider.TestEnvironment`: execution context that provides access to Kubernetes clients and test metadata. |
| **Outputs / Side‑effects** | 1. **Result on `check`** – The function calls `SetResult()` once all pods are evaluated, marking the check as `Pass` if no violations were found or `Fail` otherwise. <br>2. **Report objects** – For each pod that is non‑compliant a `PodReportObject` (via `NewPodReportObject`) is appended to the global `check.ReportObjects` slice.  <br>3. **Logging** – Uses `LogInfo` for progress messages and `LogError` when Kubernetes API calls fail. |
| **Key Steps** | 1. Log start of evaluation.<br>2. Retrieve a client holder (`GetClientsHolder`) to access the core v1 API.<br>3. Call `EvaluateAutomountTokens()` – this helper walks all pods, returning two slices: compliant and non‑compliant pods (based on the conditions above).<br>4. For each pod in the *non‑compliant* slice, create a report object and append it to `check.ReportObjects`.<br>5. Log completion and set the final result on `check`. |
| **Dependencies** | - `corev1.CoreV1()` – Kubernetes API client for pods.<br>- `EvaluateAutomountTokens` – business logic that performs the actual pod inspection.<br>- `NewPodReportObject` – helper that formats a pod into a report entry.<br>- Logging utilities (`LogInfo`, `LogError`).<br>- `SetResult` – marks the check outcome. |
| **Package Context** | Part of the *accesscontrol* test suite for CertSuite.  The function is invoked by a Ginkgo test case (via a closure that receives the current `check` and `env`) to run during the automated compliance audit. It contributes to the overall assessment of whether the cluster correctly disables automatic mounting of service account tokens, which is critical for mitigating privilege‑escalation risks. |

### Mermaid Flow (Optional)

```mermaid
flowchart TD
    A[Start] --> B{Retrieve K8s client}
    B -->|Success| C[EvaluateAutomountTokens]
    C --> D{Non‑compliant pods?}
    D -->|Yes| E[Create & append report objects]
    D -->|No| F[Log success]
    E --> F
    F --> G[SetResult(Pass/Fail)]
    G --> H[End]
```

**Note:** The function is *read‑only* with respect to the cluster; it only queries pod information and records findings. No resources are mutated.
