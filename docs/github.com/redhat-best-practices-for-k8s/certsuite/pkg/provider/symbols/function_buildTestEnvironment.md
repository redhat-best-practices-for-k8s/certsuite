buildTestEnvironment`

```go
func buildTestEnvironment() func()
```

### Purpose  
`buildTestEnvironment` is a helper that prepares the test environment for CertSuite’s provider tests. It performs the following:

1. **Initialises configuration** – loads global settings via `LoadConfiguration`.
2. **Sets up runtime data** – creates the internal state (`providerState`) with timestamps, operator lists and node information.
3. **Deploys a monitoring DaemonSet** – calls `deployDaemonSet` to install an agent that collects metrics during tests.
4. **Auto‑discovers operators** – discovers existing Operators in the cluster using `DoAutoDiscover`, then enriches them with OperatorGroups via `GetAllOperatorGroups`.
5. **Creates operator and operand pods** – builds a list of test pods for each Operator and its operands, adding any operator‑specific containers.
6. **Initialises node information** – populates the nodes slice by calling `createNodes`.

The returned closure is intended to be executed at the end of the test suite (usually via `defer`) to clean up or log final state.

### Inputs / Outputs  
| Direction | Value | Description |
|-----------|-------|-------------|
| **Input** | None – the function relies on package‑level globals (`env`, configuration files, etc.). |
| **Output** | A `func()` closure. When called, it performs any required teardown or final logging (not shown in the snippet). |

### Key Dependencies  

| Dependency | Role |
|------------|------|
| `Now` | Records the current time for timestamps. |
| `GetTestParameters`, `LoadConfiguration` | Load configuration and test parameters into global state. |
| `Fatal`, `Debug`, `Error`, `Info` | Logging helpers that report progress or failures. |
| `deployDaemonSet` | Installs a DaemonSet used by tests. |
| `DoAutoDiscover` | Discovers all Operators present in the cluster. |
| `GetAllOperatorGroups` | Retrieves OperatorGroup objects for discovered operators. |
| `createOperators`, `getSummaryAllOperators` | Build and summarise operator data structures. |
| `createNodes` | Populates node information used by tests. |
| `NewEvent`, `NewPod` | Helper constructors for event and pod objects that are added to the test state. |
| `addOperatorPodsToTestPods`, `addOperandPodsToTestPods` | Append operator/operand pods to the global pod list. |
| `getPodContainers` | Extract container information from a pod definition. |

### Side‑Effects  
* Modifies package‑level globals (`env`, `loaded`, and the internal `providerState`).
* Writes logs via the various logging functions.
* Deploys resources (DaemonSet) to the cluster.

### How It Fits the Package  

The `provider` package orchestrates interactions with an OpenShift/Kubernetes cluster for CertSuite.  
`buildTestEnvironment` is called early in a test run (e.g., inside `TestMain`) to bootstrap all required state and resources. The returned cleanup closure ensures that any side effects are reversed or reported after tests complete. This function encapsulates the heavy lifting of environment preparation, keeping individual tests focused on validation logic rather than setup.
