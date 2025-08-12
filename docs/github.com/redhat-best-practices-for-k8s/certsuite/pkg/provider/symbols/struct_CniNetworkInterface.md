CniNetworkInterface` – Representation of a CNI‑annotated network interface

| Feature | Description |
|---------|-------------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider` |
| **File / line** | `provider.go:162` |

### Purpose
`CniNetworkInterface` is a plain data holder that mirrors the structure of a single network interface as reported by Kubernetes’ CNI annotation `"k8s.v1.cni.cncf.io/networks-status"`.  
The provider package uses it to expose pod networking information to other parts of certsuite (e.g., policy validation, certificate issuance).  

### Fields

| Field | Type | Notes |
|-------|------|-------|
| `Interface` | `string` | The kernel name of the interface inside the pod (`eth0`, `ens3`, …). |
| `Name` | `string` | Human‑readable network name as defined in the CNI configuration. |
| `IPs` | `[]string` | All IP addresses assigned to this interface (IPv4 and/or IPv6). |
| `Default` | `bool` | Indicates whether this interface is the default route for the pod. |
| `DeviceInfo` | `deviceInfo` | Internal struct holding low‑level device details (e.g., MAC, MTU). The definition is in the same file but not exported. |
| `DNS` | `map[string]interface{}` | Optional DNS configuration that may accompany the interface (search domains, nameservers, etc.). |

> **Note**: All fields are exported so callers can freely read the parsed values.

### Key Dependencies

- The struct is instantiated by **`GetPodIPsPerNet`**, a public helper that parses the CNI annotation JSON into a map of `CniNetworkInterface`.  
  ```go
  func GetPodIPsPerNet(podName string) (map[string]CniNetworkInterface, error)
  ```
- The parsing routine uses standard library functions:
  - `encoding/json.Unmarshal` to decode the annotation.
  - `fmt.Errorf` for error handling.

### Side Effects

None.  
The struct is immutable once created; its fields are simple values or slices that callers may copy if mutation safety is required.

### How It Fits the Package

| Layer | Role |
|-------|------|
| **Provider** | Supplies pod networking data to higher‑level certsuite modules (e.g., network policy checks). |
| **CNI Annotation Parser** | Transforms raw annotation JSON → `map[string]CniNetworkInterface`. |
| **Consumer** | Uses the map to look up IPs per interface, decide default routes, or inspect DNS settings. |

### Example Usage

```go
// Retrieve all interfaces for a pod.
interfaces, err := provider.GetPodIPsPerNet("my-pod")
if err != nil {
    log.Fatalf("failed to get network info: %v", err)
}

// Inspect the first interface.
iface := interfaces["eth0"]
fmt.Printf("Interface %s (%s) has IPs: %v\n", iface.Interface, iface.Name, iface.IPs)

// Check if it is default route.
if iface.Default {
    fmt.Println("This is the pod's default network.")
}
```

### Suggested Mermaid Diagram

```mermaid
classDiagram
    class CniNetworkInterface {
        +string Interface
        +string Name
        +[]string IPs
        +bool Default
        +deviceInfo DeviceInfo
        +map[string]interface{} DNS
    }
    
    class GetPodIPsPerNet {
        +func(string) (map[string]CniNetworkInterface, error)
    }
    
    GetPodIPsPerNet --> CniNetworkInterface : parses JSON
```

This diagram visualises the direct relationship between the parsing function and the data structure.
