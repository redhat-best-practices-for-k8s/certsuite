testNoSSHDaemonsAllowed`

| Aspect | Details |
|--------|---------|
| **Package** | `accesscontrol` – test helpers for Kubernetes access‑control checks |
| **File**   | `suite.go` (line 908) |
| **Signature** | `func(*checksdb.Check, *provider.TestEnvironment)` |
| **Exported?** | No – used internally by the test suite |

### Purpose
The function implements compliance check **“No SSH daemons allowed”**.  
It inspects every pod in the cluster and verifies that none of them expose an SSH service (typically on port 22). The result of the check is recorded in the provided `*checksdb.Check` object.

> **Why?**  
> Exposing SSH inside a pod is a security risk because it can give attackers lateral movement or persistence channels. This test enforces that policy.

### Inputs
| Parameter | Type | Usage |
|-----------|------|-------|
| `c` | `*checksdb.Check` | Holds the current compliance check metadata; results are written back via `SetResult`. |
| `env` | `*provider.TestEnvironment` | Supplies the runtime environment, notably the list of all pods to examine (`env.Pods`). |

### Key Steps
1. **Initialisation** – empty slices for *compliant* and *non‑compliant* pod reports.
2. **Iterate over every pod** in `env.Pods`.
   1. For each container, call `GetSSHDaemonPort` to find a listening port that matches the SSH protocol (`sshServicePortProtocol`).
   2. If an error occurs while retrieving the port, log it and skip further checks for that pod.
   3. If no port is found → **compliant**; create a `NewPodReportObject` with status “passed”.
   4. If a port is found:
      - Resolve the port number (may be named or numeric) using `ParseInt`.
      - Retrieve all listening ports of that container via `GetListeningPorts`.
      - Find the exact listening address (`ListenAddr`) matching the SSH port.
      - Record a **non‑compliant** pod report with details (pod name, namespace, container name, offending port).
3. After scanning all pods:
   * If any non‑compliant pods were found → `SetResult` to “failed” and attach both compliant & non‑compliant lists to the check result.
   * Otherwise → set result to “passed”.

### Dependencies
| Called Function | Role |
|-----------------|------|
| `LogInfo`, `LogError` | Logging for debugging / error reporting. |
| `GetSSHDaemonPort` | Returns the SSH daemon port (if any) for a container. |
| `ParseInt` | Converts string port to integer when needed. |
| `GetListeningPorts` | Lists all ports that a container is listening on. |
| `NewPodReportObject` | Builds a report entry for a pod/container pair. |
| `SetResult` | Finalises the compliance check with status and attached data. |

### Side Effects
* No state mutation outside of `c.SetResult`.  
* Logs are written to the test environment’s logger.

### How it fits the package
The **accesscontrol** package contains a suite of tests that validate Kubernetes cluster security posture.  
`testNoSSHDaemonsAllowed` is one such test, grouped with others under the same file (`suite.go`). It is invoked by the test harness (likely through `beforeEachFn`) when running the full compliance assessment.

---

#### Suggested Mermaid diagram

```mermaid
flowchart TD
  A[All Pods] --> B{For each pod}
  B --> C[GetSSHDaemonPort]
  C -- error --> D[LogError & skip]
  C -- no port --> E[Mark compliant]
  C -- port found --> F[ParseInt]
  F --> G[GetListeningPorts]
  G --> H[Find ListenAddr]
  H --> I[Mark non‑compliant]
  E & I --> J[Collect reports]
  J --> K{Any non‑compliant?}
  K -- yes --> L[SetResult(failed)]
  K -- no --> M[SetResult(passed)]
```

This diagram visualises the decision flow for each pod and how results are aggregated.
