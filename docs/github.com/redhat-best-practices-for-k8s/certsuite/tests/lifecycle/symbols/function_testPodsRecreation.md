testPodsRecreation`

| Item | Details |
|------|---------|
| **Package** | `lifecycle` (github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle) |
| **Signature** | `func (*checksdb.Check, *provider.TestEnvironment)` |
| **Exported** | No – used only inside the test suite. |
| **Purpose** | Verify that when a node disappears from the cluster, all pods that belong to deployments or statefulsets are automatically recreated on other nodes and reach the *Ready* state within the configured timeout window. |

---

### Inputs

| Parameter | Type | Role |
|-----------|------|------|
| `c` | `*checksdb.Check` | The check record that is being evaluated; the function writes results into this object. |
| `env` | `*provider.TestEnvironment` | Provides access to the Kubernetes API, test configuration and helper utilities (e.g., logging, timeouts). |

---

### Key Steps & Dependencies

1. **Pre‑check**  
   * Calls `skipIfNoPodSetsetsUnderTest(c)` – aborts early if there are no relevant pod sets in this check.

2. **Initial state capture**  
   * Collects all nodes that currently host pods of the target deployments/statefulsets via `GetAllNodesForAllPodSets`.
   * Builds a *deployment* and *statefulset* report object for each, attaching node selectors or runtime class information if present.
   * Creates per‑node reports (`NewNodeReportObject`) summarizing how many pods belong to that node.

3. **Simulate node loss**  
   * Uses `CordonHelper` to cordon the target node(s) (making them unschedulable).
   * Waits for the Kubernetes control plane to react and deletes all pods scheduled on those nodes.
   * Records any errors through `LogError`.

4. **Wait for pod recreation**  
   * Invokes `WaitForAllPodSetsReady` with a per‑pod timeout (`timeoutPodRecreationPerPod`) and an overall set‑ready timeout (`timeoutPodSetReady`).  
   * After the wait, counts pods that were deleted and needed recreation via `CountPodsWithDelete`.

5. **Result aggregation**  
   * If any pod failed to reach ready state or was not recreated, marks the check as **Failed** using `SetResult`.  
   * Otherwise sets the result to **Passed**.  
   * Calls `SetNeedsRefresh` if a node was cordoned and later uncordoned, signalling that subsequent checks should re‑evaluate the environment.

6. **Cleanup**  
   * Invokes `CordonCleanup` to restore node schedulability after the test.

---

### Outputs

* The function mutates the supplied `*checksdb.Check`:
  * Sets result status (`Passed`, `Failed`) and optional failure reason.
  * Adds detailed report objects for deployments, statefulsets, pods, and nodes (via `NewDeploymentReportObject`, `NewStatefulSetReportObject`, etc.).
* Logs diagnostic information throughout execution.

---

### Side Effects

* **Cluster state changes** – cordoning/uncordoning nodes and allowing Kubernetes to delete/recreate pods.
* **Logging** – uses the environment’s logger (`LogDebug`, `LogInfo`, `LogError`).
* **Timeout handling** – respects package‑level timeout constants:
  * `timeout` (overall test timeout)
  * `timeoutPodRecreationPerPod`
  * `timeoutPodSetReady`

---

### How It Fits the Package

The `lifecycle` package contains a battery of tests that validate Kubernetes lifecycle behaviours.  
`testPodsRecreation` is one of those checks, specifically focusing on **pod resilience** when nodes fail.  

It relies on several helper functions defined elsewhere in the test suite:

| Helper | Responsibility |
|--------|----------------|
| `CordonHelper`, `CordonCleanup` | Manage node schedulability |
| `GetAllNodesForAllPodSets` | Retrieve current pod‑to‑node mapping |
| `WaitForAllPodSetsReady` | Wait for pods to become ready |
| Reporting helpers (`NewDeploymentReportObject`, etc.) | Build structured test results |

By orchestrating these pieces, the function demonstrates that Kubernetes can recover from node loss without manual intervention.
