NewTester`

### Purpose
Creates a **`*Tester`** instance that can verify huge‑page configuration on a Kubernetes node.  
The function collects the current state of the node (NUMA huge‑pages, systemd units) and stores it in the tester context for later comparison against expected values.

### Signature
```go
func NewTester(node *provider.Node, pod *corev1.Pod, exec clientsholder.Command) (*Tester, error)
```

| Parameter | Type                      | Description |
|-----------|---------------------------|-------------|
| `node`    | `*provider.Node`          | The node on which the test will run. |
| `pod`     | `*corev1.Pod`             | The pod that owns the huge‑page resources (may be nil). |
| `exec`    | `clientsholder.Command`   | Interface used to execute commands inside the node or pod. |

### Return Values
- `*Tester`: Holds the gathered data and context needed for the test.
- `error`: Non‑nil if any step fails while collecting information.

### Key Steps & Dependencies

1. **Context Creation**  
   ```go
   ctx := NewContext(node, pod)
   ```
   Builds a testing context that stores node/pod references and command executor.

2. **Gather NUMA Huge‑Pages**  
   ```go
   numahp, err := getNodeNumaHugePages(ctx)
   ```
   Reads `/proc/meminfo` or similar to determine huge‑page allocation per NUMA node.
   - On failure, logs via `Info()` and returns the error.

3. **Store NUMA Data**  
   ```go
   ctx.Data["numahp"] = numahp
   ```

4. **Gather Systemd Units Config**  
   ```go
   unitsConfig, err := getMcSystemdUnitsHugepagesConfig(ctx)
   ```
   Parses the `systemd` unit files that configure huge‑page sizes/amounts.
   - On failure, logs via `Info()` and returns the error.

5. **Store Systemd Config**  
   ```go
   ctx.Data["units"] = unitsConfig
   ```

6. **Return Tester**  
   ```go
   return &Tester{Ctx: ctx}, nil
   ```

### Side‑Effects

- Emits informational logs (`Info`) to trace progress.
- Fails fast if any data collection step cannot complete, propagating the error up the call chain.

### Integration in the Package

The `hugepages` package provides a suite of tests that validate huge‑page settings on Kubernetes nodes.  
`NewTester` is the entry point for setting up a test run:

```go
tester, err := hugepages.NewTester(node, pod, exec)
if err != nil { /* handle error */ }

err = tester.Run() // performs assertions against ctx.Data
```

The returned `*Tester` holds all collected information in its context (`ctx.Data`).  
Subsequent test functions (e.g., `TestHugePagesConfiguration`) use this data to assert that the node’s huge‑page configuration matches expectations.

---

#### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[Node, Pod] --> B{NewContext}
    B --> C[*Tester]
    C --> D[getNodeNumaHugePages]
    D --> E[ctx.Data["numahp"]]
    C --> F[getMcSystemdUnitsHugepagesConfig]
    F --> G[ctx.Data["units"]]
```

This diagram visualizes the data flow from inputs to the populated tester context.
