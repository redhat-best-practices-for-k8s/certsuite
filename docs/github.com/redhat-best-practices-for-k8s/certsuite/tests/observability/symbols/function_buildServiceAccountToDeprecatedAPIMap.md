buildServiceAccountToDeprecatedAPIMap`

| Item | Detail |
|------|--------|
| **Signature** | `func(buildServiceAccountToDeprecatedAPIMap([]apiserv1.APIRequestCount, map[string]struct{}) map[string]map[string]string)` |
| **Exported?** | No – internal helper used by the test suite. |

### Purpose
The function constructs a two‑level lookup that associates *workload service accounts* with the APIs they are about to deprecate and the release in which the API will be removed.

- **Outer key** – Service account name (`string`).
- **Inner map** – Key: API path (`string`).  
  Value: The `removedInRelease` string that indicates when the API will be retired.

The map is used by tests to verify that workloads are not using APIs that will soon disappear.

### Inputs

| Parameter | Type | Role |
|-----------|------|------|
| `apiCounts` | `[]apiserv1.APIRequestCount` | A slice of API request summaries. Each element contains an `APIRequestCount.Status.RemovedInRelease` field that may be empty or contain the release name. |
| `serviceAccounts` | `map[string]struct{}` | A set (empty struct values) representing the workload service accounts to consider. Only requests from these SAs are kept in the result. |

### Output

- `map[string]map[string]string` – As described above, mapping each relevant SA to its deprecated APIs and removal releases.

### Algorithm Overview

```text
1. Allocate the outer map: saToAPI := make(map[string]map[string]string)
2. For every apiReq in apiCounts:
   a. Skip if RemovedInRelease is empty.
   b. Extract the service account from apiReq.ServiceAccount (may contain namespace prefix).
      - Split by '/' and take the last element; this normalises values like
        "ns/sa-name" → "sa-name".
   c. If the extracted SA exists in the `serviceAccounts` set:
      i. Ensure saToAPI[sa] map is initialized.
     ii. Set saToAPI[sa][apiReq.APIPath] = apiReq.Status.RemovedInRelease
3. Return saToAPI.
```

### Key Dependencies

| Dependency | Where used |
|------------|-----------|
| `make` (map) | Creates the outer and inner maps. |
| `strings.Split` | Normalises service account names by discarding any namespace prefix. |
| `len` | Checks that a split produced at least one component before accessing the last element. |

### Side‑effects

- No global state is modified; the function is pure.
- The only external influence is the `env.ServiceAccounts` set, accessed indirectly via the `serviceAccounts` argument.

### Integration in Package

Within the **observability** test suite (`suite.go`), this helper is called during test setup to generate a reference table of deprecated APIs per workload service account. Subsequent tests consume that map to assert that workloads do not invoke APIs slated for removal, ensuring backward‑compatibility and policy compliance.

---

#### Suggested Mermaid Diagram (optional)

```mermaid
graph TD
  A[apiCounts] --> B{filter: removedInRelease != ""}
  B --> C{extract SA}
  C --> D{SA in serviceAccounts?}
  D -- yes --> E[initialize saToAPI[sa]]
  E --> F[saToAPI[sa][apiPath] = removedInRelease]
```

This diagram visualises the decision flow from raw API request data to the final lookup map.
