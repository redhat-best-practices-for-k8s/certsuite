GetPciPerPod`

```go
func GetPciPerPod(name string) ([]string, error)
```

| Aspect | Description |
|--------|-------------|
| **Purpose** | Retrieves a list of PCI device IDs that should be attached to the pod named *name*. The function is used by the provider when preparing a workload that requires specific hardware resources. |
| **Input** | `name` – the name of the pod for which PCI information is requested. |
| **Output** | A slice of strings (`[]string`) containing PCI device identifiers, or an error if parsing fails. |
| **Key dependencies** | * `json.Unmarshal` – parses a JSON‑encoded string into a Go value.<br>* `fmt.Errorf` – constructs descriptive error messages.<br>* Built‑in `append` – builds the result slice. |
| **Side effects** | None. The function is pure: it does not modify global state, mutate arguments, or perform I/O beyond decoding data that was supplied to it. |
| **How it fits the package** | Within the `provider` package, many functions assemble pod specifications for testing or deployment. Some workloads need explicit PCI device binding (e.g., SR-IOV). `GetPciPerPod` is called by higher‑level logic that gathers the list of PCI IDs from a JSON string stored in an environment variable or configuration file, then injects them into the pod spec’s `volumeMounts`/`deviceRequests`. This keeps the PCI extraction logic isolated and reusable. |

### Typical usage flow

1. A caller passes a pod name to `GetPciPerPod`.
2. The function obtains a JSON string (the source of this string is outside the shown code – likely from an environment variable or configuration file).
3. It unmarshals that JSON into a slice of strings.
4. If unmarshalling fails, it returns an error; otherwise it returns the slice.

### Example

```go
pciList, err := provider.GetPciPerPod("my‑pod")
if err != nil {
    log.Fatalf("cannot get PCI devices: %v", err)
}
fmt.Printf("PCI devices for pod: %v\n", pciList)
```

### Note on implementation details

The actual source of the JSON string is not visible in the provided snippet. The function therefore relies on external code (e.g., an environment variable or a configuration map) to supply the data it needs to parse. This design allows the provider to remain agnostic about where PCI device lists originate while still providing a clear API for consuming that information.
