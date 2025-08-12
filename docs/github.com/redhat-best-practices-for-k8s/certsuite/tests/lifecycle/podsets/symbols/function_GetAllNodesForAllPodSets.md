GetAllNodesForAllPodSets`

| Aspect | Details |
|--------|---------|
| **Package** | `podsets` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/podsets`) |
| **Exported?** | ✅ |
| **Signature** | `func GetAllNodesForAllPodSets(pods []*provider.Pod) map[string]bool` |
| **Purpose** | Build a set of unique Kubernetes node names that host any pod belonging to the *pod sets* being tested. |

### How it works
1. **Input** – A slice of pointers to `provider.Pod`.  
   Each element represents a running pod with metadata such as its node name.
2. **Processing** – The function iterates over the slice and extracts the node name from each pod (`pod.Spec.NodeName`).  
   For every discovered node, it inserts an entry into a Go map where:
   * key = node name (string)
   * value = `true` (the boolean is unused; the map acts as a set).
3. **Return** – A map containing one entry per distinct node that hosts at least one pod from the input slice.

### Key dependencies
| Dependency | Role |
|------------|------|
| `make(map[string]bool)` | Initializes the result map. |
| `provider.Pod` (from an external package) | Provides the pod structure and the `Spec.NodeName` field used to fetch node names. |

No other functions, globals or side‑effects are involved.

### Where it fits
The test suite creates *pod sets* (ReplicaSets/StatefulSets).  
During lifecycle tests the framework needs to know **which nodes** are actively running those pods – e.g., for:
* Verifying that scaling occurs on all target nodes.
* Checking node‑specific health or configuration.

`GetAllNodesForAllPodSets` is a small helper used by higher‑level test logic to gather this information in one place, keeping the rest of the code clean and focused on orchestration.
