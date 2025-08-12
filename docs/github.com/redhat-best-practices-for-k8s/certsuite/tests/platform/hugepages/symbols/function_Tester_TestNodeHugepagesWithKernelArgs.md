Tester.TestNodeHugepagesWithKernelArgs`

**Purpose**

`TestNodeHugepagesWithKernelArgs` validates that the amount of hugepages configured on a Kubernetes node matches what is requested via kernel‑boot arguments.  
It ensures:

| Hugepage size | Expected behaviour |
|---------------|--------------------|
| Size specified in `kernArgs` (`--hugepagesz`) | Total number of pages reported by the node must equal the value supplied in `--hugepages` |
| Any other hugepage sizes | No pages should be present (sum = 0) |

If either condition fails, the test reports an error.

---

### Signature

```go
func (t *Tester) TestNodeHugepagesWithKernelArgs() (bool, error)
```

* **Receiver** – `*Tester` from the same package.  
  The struct holds state for the test run; it is not modified by this method.
* **Returns** –  
  * `true` when the node’s hugepage configuration satisfies the kernel‑argument rules.  
  * `false, err` if a mismatch occurs or an internal error happens.

---

### Key Steps & Dependencies

1. **Parse Kernel Arguments**
   ```go
   pageSize, pages := getMcHugepagesFromMcKernelArguments()
   ```
   * `getMcHugepagesFromMcKernelArguments()` reads the node’s kernel‑boot parameters (e.g., `hugepagesz=2M hugepages=1024`) and returns:
     * `pageSize` – size string (`"2M"` or `"1G"`, etc.)
     * `pages`    – number of pages requested.

2. **Retrieve Node Hugepages**
   ```go
   n := t.GetNode()
   ```
   * `GetNode()` is a helper that fetches the node’s current hugepage allocation from the API (or via `/proc/meminfo`).

3. **Validate Requested Size**
   ```go
   if _, ok := n.Hugepages[pageSize]; !ok {
       return false, Errorf("kernel argument requested size %s but node has no such entry", pageSize)
   }
   ```
   * If the node does not expose the requested size, the test fails.

4. **Check Total Page Count**
   ```go
   if n.Hugepages[pageSize].Total != pages {
       return false, Errorf("requested %d hugepages of size %s but node reports %d",
           pages, pageSize, n.Hugepages[pageSize].Total)
   }
   ```
   * The reported total must equal the kernel‑argument value.

5. **Ensure Other Sizes Are Zero**
   ```go
   for sz, hp := range n.Hugepages {
       if sz != pageSize && hp.Total != 0 {
           return false, Errorf("unexpected hugepage size %s with count %d", sz, hp.Total)
       }
   }
   ```
   * Any other size must have a zero total.

6. **Success**
   ```go
   Info("Hugepages match kernel arguments")
   return true, nil
   ```

---

### Side Effects

* The function only performs reads; it does not modify the node or any configuration.
* It logs informational messages via `Info` and errors via `Errorf`.

---

### Package Context

The **hugepages** test package validates various aspects of hugepage handling on a Kubernetes cluster.  
`TestNodeHugepagesWithKernelArgs` is one of several tests that:

1. Read the node’s configuration (kernel args, `/proc/meminfo`, etc.).
2. Compare it against expected values derived from the platform’s defaults or user‑provided overrides.
3. Report compliance via the test harness.

This function specifically ties together kernel boot parameters and runtime hugepage allocation to ensure that the cluster behaves as intended for memory‑intensive workloads.
