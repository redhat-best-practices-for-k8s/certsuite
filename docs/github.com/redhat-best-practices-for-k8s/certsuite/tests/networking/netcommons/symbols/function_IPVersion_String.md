IPVersion.String` – String Representation of an IPv4/IPv6 Version

### Purpose
`IPVersion.String` converts a value of the internal `IPVersion` type to its human‑readable string form.  
The function is used wherever a textual representation of the IP version is needed, e.g.:

* Logging or debugging output.
* Building configuration files that reference IPv4/IPv6 by name.
* Returning values in API responses.

### Signature
```go
func (v IPVersion) String() string
```
* **Receiver** – `IPVersion` (`v`) holds one of the predefined constants:  
  * `Undefined` – unknown or not set.  
  * `IPv4` – IPv4 only.  
  * `IPv6` – IPv6 only.  
  * `IPv4v6` – both IPv4 and IPv6 are supported.
* **Returns** – a string describing the version.

### Implementation Overview
The method is a simple switch (or map lookup) that matches the receiver to one of the constants above and returns a corresponding constant string:

| Constant | Returned String |
|----------|-----------------|
| `Undefined` | `"undefined"` |
| `IPv4`      | `"ipv4"`       |
| `IPv6`      | `"ipv6"`       |
| `IPv4v6`    | `"ipv4/ipv6"`  |

The actual string literals are defined elsewhere in the package (`IPv4String`, `IPv6String`, etc.), ensuring consistency across the codebase.

### Dependencies & Side‑Effects
* **Dependencies** – None beyond the type definition of `IPVersion`.  
* **Side‑Effects** – None. The method is pure: it only reads its receiver and returns a value.

### Package Context
`netcommons` provides common networking utilities for Certsuite tests.  
The `IPVersion` type (and its constants) are used throughout the package to:

1. **Configure network interfaces** (`IFType`, e.g., `DEFAULT`, `MULTUS`).  
2. **Validate test environments** – ensuring that a node supports the required IP version before running a test.  
3. **Generate test data** – building request payloads for the networking API.

`IPVersion.String` is thus the canonical way to expose these enum values in text form, keeping callers free from hard‑coded strings and reducing the risk of typos.

### Example Usage
```go
v := IPv4v6
fmt.Println("Testing IP support:", v.String())
// Output: Testing IP support: ipv4/ipv6
```

This concise helper keeps all string representations in sync with the enum definition, simplifying maintenance and improving readability.
