GetReconciledResults`

**Package:** `checksdb`  
**Signature**

```go
func GetReconciledResults() map[string]claim.Result
```

### Purpose

`GetReconciledResults` aggregates the current state of all checks that belong to a **Claim**.  
Because the certsuite‑claim Go client only exposes results as `map[string]interface{}`, this helper converts those raw values into the strongly‑typed `claim.Result` type defined in the external package.

### Inputs / Outputs

| Parameter | Type | Description |
|-----------|------|-------------|
| *none*    | –    | The function operates on the global state of the checks database. |

| Return value | Type | Description |
|--------------|------|-------------|
| `map[string]claim.Result` | Map keyed by check ID, with each value being a fully‑deserialized result struct. | Represents every check that has been executed for the current Claim, ready for serialization or further processing. |

### Key Dependencies

* **Global state** – The function reads from the package’s internal `resultsDB`, which is populated elsewhere in the checks lifecycle.
* **`claim.Result` type** – Imported from `github.com/redhat-best-practices-for-k8s/certsuite/pkg/claim`.  
  No other external packages are used.

### Side Effects

None. The function only reads data; it does not modify any global or local state.

### How It Fits the Package

* The `checksdb` package maintains an in‑memory database of checks (`dbByGroup`, `resultsDB`, etc.).  
* During a Claim run, individual check executions update `resultsDB`.  
* When the Claim needs to report its outcome (e.g., during serialization or API response), it calls `GetReconciledResults()` to obtain a clean map that can be marshalled into JSON or sent over the network.

This helper abstracts away the conversion from raw interface maps to typed results, simplifying client code and ensuring consistency across the system.
