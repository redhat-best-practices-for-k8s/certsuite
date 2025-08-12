# `GetWorkerCount` – Counting Worker Nodes in a Test Environment

```go
func (env *TestEnvironment) GetWorkerCount() int
```

> Returns the number of nodes that are classified as **worker** nodes in the current test environment.

---

## Purpose

In a Kubernetes cluster, nodes can be tagged with labels to indicate their role (master/control‑plane vs. worker).  
`GetWorkerCount` is used by tests that need to know how many workers exist before performing
operations such as workload placement, resource quota checks, or scaling validations.

The method simply iterates over the environment’s cached node list and counts those for which
`IsWorkerNode` returns `true`.

---

## Inputs

| Parameter | Type | Description |
|-----------|------|-------------|
| receiver (`env`) | `*TestEnvironment` | Holds the cluster state, including a slice of nodes. |

> **No function arguments** – all information comes from the receiver.

---

## Output

| Return value | Type | Meaning |
|--------------|------|---------|
| `int` | Number of worker nodes in `env.Nodes`. |

If no nodes exist or none match the worker label set, the method returns `0`.

---

## Key Dependencies

| Dependency | How it’s used |
|------------|---------------|
| `TestEnvironment.IsWorkerNode(node *v1.Node) bool` | Called for each node to decide if it is a worker. |
| `env.Nodes` (slice of `*v1.Node`) | The collection over which the method iterates. |

The logic of determining whether a node is a worker lives in `IsWorkerNode`, which checks that
the node has **any** label from `WorkerLabels`.  
`WorkerLabels` is a package‑wide variable defined as:

```go
var WorkerLabels = [][]string{
    {"node-role.kubernetes.io/worker"},
    // ... other aliases
}
```

---

## Side Effects

* None – the method only reads state; it does not modify `env`, nodes, or any global variables.

---

## Relationship to the Package

`GetWorkerCount` belongs to the **provider** package, which implements an abstraction over a Kubernetes cluster used by CertSuite tests.  
It is typically called after the environment has been populated (e.g., during `Setup()` or in test fixtures) and before actions that depend on knowing how many workers are available.

The function is part of a suite of helper methods that expose high‑level information about the test environment:

* `GetMasterCount`
* `GetNodeCount`
* `GetPodsByContainerName`

These helpers simplify writing tests by hiding low‑level label checks and Kubernetes API interactions.
