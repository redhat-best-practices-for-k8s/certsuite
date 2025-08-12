testTainted` – Platform Test Helper

## Purpose
`testTainted` is a private helper that runs the **kernel‑taint** test for a single node in a Kubernetes cluster.

The function:
1. Verifies that the workload under test has been deployed.
2. Creates a *taint tester* for the node.
3. Collects kernel‑taint information (bitmask, letters, modules).
4. Builds structured **node reports** and **taint reports** to be used by the rest of the test suite.

It is invoked from the `platform` test suite after a workload has been scheduled on a target node.

---

## Signature
```go
func testTainted(node *checksdb.Check, env *provider.TestEnvironment)
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `node` | `*checksdb.Check` | The node record that holds the node name and any pre‑existing check data. |
| `env`  | `*provider.TestEnvironment` | Shared test environment (contains clientsets, context, logger, etc.). |

The function has no return value; all results are written into the supplied `node` via helper report objects.

---

## High‑level Flow

```mermaid
flowchart TD
    A[Start] --> B{HasWorkloadDeployed?}
    B -- No --> C[LogInfo & Return]
    B -- Yes --> D[NewNodeTaintedTester]
    D --> E[GetKernelTaintsMask]
    E --> F[DecodeBitMask → letters]
    F --> G[Check for taints]
    G --> H{No Taints}
    H -- Yes --> I[Create NodeReportObject (empty)]
    H -- No  --> J[Add taint data to reports]
    J --> K[GetTainterModules]
    K --> L[Create TaintReportObject]
    L --> M[Finish]
```

---

## Detailed Steps

1. **Check Workload Status**  
   * `HasWorkloadDeployed(node)` verifies that the test workload is actually running on this node.  
   * If not, a log message is emitted and the function returns early.

2. **Create Tester Context**  
   * A new `context.Context` (`NewContext()`) isolates cancellation signals for this node.
   * `NewNodeTaintedTester(ctx, env)` builds a tester that can query kernel‑taint information from the node’s `/proc/sys/kernel/taints`.

3. **Kernel Taint Bitmask**  
   * `GetKernelTaintsMask()` returns an unsigned integer bitmask of active taints.
   * If the mask is zero → no taints; log info and record an empty report.

4. **Decode Mask to Human‑Readable Forms**  
   * `DecodeKernelTaintsFromBitMask(mask)` gives a slice of single‑letter codes (e.g., `"A"`, `"C"`).
   * `FormatUint(mask, 16)` converts the mask to a hex string for reporting.
   * The decoded letters are joined with commas and added to a node report via `AddField`.

5. **Tainter Modules**  
   * `GetTainterModules()` lists kernel modules that caused taints (e.g., `"usbcore"`).  
   * If present, each module name is appended to the node report.

6. **Node & Taint Reports**  
   * `NewNodeReportObject(node.Name)` creates a structured report container.
   * For every decoded taint letter, a separate `NewTaintReportObject` is built and attached to the node report.
   * All reports are stored in `node.Report`.

7. **Error Handling**  
   * Any error during querying or decoding results in:
     - Logging via `LogError`.
     - Appending an error field (`AddField("error", err.Error())`) to the relevant report.

---

## Dependencies

| Called Function | Responsibility |
|-----------------|----------------|
| `HasWorkloadDeployed` | Checks workload presence. |
| `NewContext`, `NewNodeTaintedTester` | Builds testing context and taint reader. |
| `GetKernelTaintsMask`, `DecodeKernelTaintsFromBitMask`, `DecodeKernelTaintsFromLetters` | Kernel‑taint extraction/decoding. |
| `GetTainterModules` | Retrieves modules responsible for taints. |
| Report helpers (`NewNodeReportObject`, `NewTaintReportObject`, `AddField`) | Builds structured test reports. |
| Logging utilities (`LogInfo`, `LogError`, `Error`) | Emits diagnostic messages. |

All these helpers reside in the same package or imported sub‑packages of the *platform* test suite.

---

## Side Effects

* **Node state** – None; only reads from `/proc` and writes to the supplied `node.Report`.
* **Logging** – Emits informational and error logs via the shared logger.
* **Context cancellation** – The created context is not cancelled within this function but can be externally.

---

## How It Fits the Package

The *platform* package orchestrates end‑to‑end tests of Kubernetes nodes.  
`testTainted` is part of the per‑node test pipeline:

1. A workload is deployed (`beforeEachFn`).
2. `testTainted` validates kernel taint status after deployment.
3. The results are aggregated into the overall test report.

Thus, it bridges low‑level node introspection with high‑level test reporting.
