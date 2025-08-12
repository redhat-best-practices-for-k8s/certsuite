EvaluateAutomountTokens`

| Feature | Description |
|---------|-------------|
| **Signature** | `func EvaluateAutomountTokens(client corev1typed.CoreV1Interface, pod *provider.Pod) (bool, string)` |
| **Exported?** | ✅ |

### Purpose
Checks that a Pod’s `automountServiceAccountToken` setting is correct.  
The function verifies two things:

1. **Pod‑level override** – if the Pod spec explicitly sets `automountServiceAccountToken`, it must be `true`.  
2. **Service‑account inheritance** – if no explicit value exists, the setting is inherited from the ServiceAccount that the Pod uses. The helper `IsAutomountServiceAccountSetOnSA` determines whether the SA has the token enabled.

If either check fails, the function reports a detailed error message.

### Parameters
| Name | Type | Role |
|------|------|------|
| `client` | `corev1typed.CoreV1Interface` | Kubernetes client used to fetch the Pod’s ServiceAccount. |
| `pod` | `*provider.Pod` | The Pod object being evaluated. |

> **Note**: The function does not modify the `Pod`; it only reads from the API.

### Return values
| Value | Type | Meaning |
|-------|------|---------|
| First (bool) | `true` if all checks pass, otherwise `false`. |
| Second (string) | An error description when validation fails; empty string on success. |

### Key dependencies
- **`Sprintf`** – Used for building human‑readable messages.
- **`IsAutomountServiceAccountSetOnSA`** – Determines the automount flag on a ServiceAccount.
- **`Error`** – Standard error constructor for formatting.

The function uses the client only to look up the associated ServiceAccount when needed; otherwise it relies solely on the Pod struct.

### Side effects
None. The function performs read‑only operations and returns diagnostic information.

### Package context
`rbac` contains utilities that validate RBAC‑related settings in a Kubernetes cluster.  
`EvaluateAutomountTokens` is one of several checks used by automated tests to ensure Pods are not inadvertently granted access through service‑account tokens, helping maintain least‑privilege security posture.
