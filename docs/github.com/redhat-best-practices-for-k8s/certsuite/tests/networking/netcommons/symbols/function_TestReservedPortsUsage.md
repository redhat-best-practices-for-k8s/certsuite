TestReservedPortsUsage`

```go
func TestReservedPortsUsage(
    env *provider.TestEnvironment,
    reservedPorts map[int32]bool,
    podNamespace string,
    logger *log.Logger) []*testhelper.ReportObject
```

### Purpose
`TestReservedPortsUsage` validates that none of the *reserved* ports are in use by any Pod running inside a given Kubernetes namespace.  
The function is part of the `netcommons` test suite, which checks network‑related invariants for CertSuite.

1. **Input** –  
   - `env`: a `provider.TestEnvironment` instance that gives access to the K8s API server and helper methods.  
   - `reservedPorts`: a map whose keys are the port numbers that must *not* be in use (e.g., Istio side‑car ports). The boolean value is ignored; the map simply represents a set.  
   - `podNamespace`: namespace to scan for rogue pods.  
   - `logger`: optional logger used for debugging output.

2. **Output** –  
   Returns a slice of pointers to `testhelper.ReportObject`. Each element describes a *rogue* Pod that is listening on one or more ports from the reserved set. If no such Pods are found, an empty slice is returned.

3. **Key Dependency** –  
   The function internally calls `findRoguePodsListeningToPorts`, which does the heavy lifting: it iterates over all pods in `podNamespace`, discovers open TCP/UDP sockets, and flags those that match any port from `reservedPorts`.

4. **Side‑Effects** –  
   - No state is mutated outside of the returned slice; the function is read‑only with respect to the test environment.  
   - It logs diagnostic information if `logger` is non‑nil.

5. **Package Context**  
   In the `netcommons` package, `TestReservedPortsUsage` is one of several *test helper* functions that other CertSuite tests invoke to validate network configuration. The reserved ports are defined in the global variable `ReservedIstioPorts`, which is a pre‑populated map for common Istio side‑car ports.

---

### Flow (high level)

```mermaid
graph LR
  A[Start] --> B{Find rogue pods}
  B --> C{Any found?}
  C -- yes --> D[Create ReportObject(s)]
  C -- no --> E[Return empty slice]
```

1. **Call `findRoguePodsListeningToPorts`**  
   - Arguments: `env`, `reservedPorts`, `podNamespace`.  
   - Returns a list of pods that are listening on any reserved port.

2. **If the returned list is non‑empty**  
   - For each pod, create a `testhelper.ReportObject` summarizing the violation and append it to the result slice.

3. **Return the result slice** – may be empty if no violations were detected.

---

### Usage Example

```go
reserved := map[int32]bool{
    15001: true, // Istio gateway port
    15443: true, // Istio mTLS port
}
reports := netcommons.TestReservedPortsUsage(env, reserved, "istio-system", log.Default())
if len(reports) > 0 {
    fmt.Println("Found rogue pods:", reports)
}
```

This snippet checks the `istio-system` namespace for any pod listening on the two typical Istio ports and prints a report if violations exist.

---

### Summary

- **What**: Detects Pods misusing reserved network ports.  
- **Where**: In `netcommons.go` of CertSuite's networking tests.  
- **Why**: Ensures that critical ports (e.g., Istio side‑car) remain free for their intended service mesh traffic.
