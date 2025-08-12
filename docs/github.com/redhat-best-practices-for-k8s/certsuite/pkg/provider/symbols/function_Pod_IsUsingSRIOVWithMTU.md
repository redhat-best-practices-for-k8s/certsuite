Pod.IsUsingSRIOVWithMTU`

| Aspect | Detail |
|--------|--------|
| **Receiver** | `Pod` – a struct that represents a Kubernetes pod and contains its metadata, spec and status information. |
| **Signature** | `func (p Pod) IsUsingSRIOVWithMTU() (bool, error)` |
| **Exported** | Yes – it is part of the public API of the `provider` package. |

### Purpose

The function determines whether *any* network interface attached to a pod uses the SR‑I OV plug‑in **and** has an MTU value explicitly set in its CNI configuration.

- **SR‑I OV detection** – It examines each NetworkAttachmentDefinition (NAD) that is referenced by the pod’s annotations.  
- **MTU check** – For each NAD, it inspects the embedded `K8sCniCncfIoV1` spec to see if an MTU field is present.

If at least one such interface exists, the function returns `true`.  If none are found, it returns `false`.  Errors may be returned when any of the underlying Kubernetes API calls fail or when a malformed annotation is encountered.

### Inputs & Outputs

| Input | Type | Description |
|-------|------|-------------|
| `p` (receiver) | `Pod` | The pod whose interfaces are inspected. |

| Output | Type | Description |
|--------|------|-------------|
| `bool` | Indicates whether an SR‑I OV interface with MTU is present. |
| `error` | Any error that occurred while retrieving NADs or parsing annotations. |

### Key Dependencies

| Dependency | Role |
|------------|------|
| `getCNCFNetworksNamesFromPodAnnotation` | Extracts the list of NAD names from the pod’s annotation (`k8s.v1.cni.cncf.io/networks`). |
| `GetClientsHolder()` | Provides a Kubernetes client set used to fetch NAD objects. |
| `NetworkAttachmentDefinitions()` | Returns a handle for querying NAD resources. |
| `K8sCniCncfIoV1()` | Parses the CNI spec embedded in an NAD to inspect fields like `MTU`. |
| `sriovNetworkUsesMTU(nad)` | Helper that checks whether a particular NAD is SR‑I OV and has MTU set. |

The function also logs diagnostic information via the package’s `Debug` logger, but these are side‑effects only for observability.

### Algorithm Overview

```text
1. Parse pod annotation to get list of NAD names.
2. For each name:
   a. Fetch the NAD object from Kubernetes.
   b. If it is an SR‑I OV network and has MTU set, return (true, nil).
3. If loop completes with no matches, return (false, nil).
```

### Side Effects

- **Logging** – Uses `Debug` to emit information about the discovery process; does not modify any state.
- **Kubernetes API calls** – Reads NAD objects but does not write or delete anything.

### Package Context

The `provider` package implements a set of helpers that allow CertSuite to introspect Kubernetes resources (nodes, pods, deployments, etc.).  `Pod.IsUsingSRIOVWithMTU` is one such helper that aids the certification logic for networks: it lets other parts of CertSuite determine whether SR‑I OV networking with explicit MTU configuration is in use, which may affect compliance checks or recommended best practices.
