isContainerCapabilitySet`

| Item | Detail |
|------|--------|
| **Package** | `accesscontrol` |
| **Signature** | `func (c *corev1.Capabilities, cap string) bool` |
| **Exported?** | No – helper used only within the test suite. |

### Purpose
Determines whether a specific Linux capability (`cap`) was explicitly added to a container’s security context via the `Capabilities.Add` list.

In Kubernetes tests this helper is used to verify that certain capabilities are present (or absent) when asserting policy compliance of workloads.

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `c` | `*corev1.Capabilities` | The capability list from a container’s security context. May be `nil`. |
| `cap` | `string` | Capability name to look for, e.g. `"NET_BIND_SERVICE"` or `"SYS_ADMIN"`. |

### Return Value
`bool`
- `true` if the given capability is found in `c.Add`.
- `false` otherwise (including when `c` is `nil`, has no `Add` entries, or the capability isn’t present).

### Implementation Flow

1. **Nil‑check** – If `c == nil`, immediately return `false`.  
2. **Check Add list** –  
   * If `len(c.Add) > 0`, call `StringInSlice(cap, c.Add)` to see if the exact string is listed.  
3. **Fallback for legacy names** –  
   * Some tests may use the older `Capability` enum type (`v1.Capability`). The function calls `capability.String()` (via the helper `Capability`) and checks again with `StringInSlice`. This ensures compatibility with both string literals and enum values.
4. **Result** – Return the boolean result of the lookup.

### Key Dependencies

| Dependency | Role |
|------------|------|
| `len` | To determine if the Add list is non‑empty before searching. |
| `StringInSlice` | Helper that checks membership of a string in a slice. Used twice for raw and enum forms. |
| `Capability` | Converts capability constants to their string representation for comparison. |

### Side Effects

None – pure function; only reads its arguments.

### Package Context

Within the `accesscontrol` test suite, this helper is called by various test cases that inspect container security contexts. It abstracts the logic of searching a capabilities list, keeping the tests concise and focused on policy assertions rather than repeated slice checks.
