NewClaimBuilder`

> **Location**: `pkg/claimhelper/claimhelper.go:111`  
> **Package**: `github.com/redhat-best-practices-for-k8s/certsuite/pkg/claimhelper`

## Overview
`NewClaimBuilder` is a constructor that creates and initializes a `ClaimBuilder`, which orchestrates the generation of test claims for CertSuite. It pulls configuration from environment variables, loads existing claim data if present, augments it with additional metadata (e.g., version information), and prepares the internal state required by other builder methods.

---

## Signature
```go
func NewClaimBuilder(env *provider.TestEnvironment) (*ClaimBuilder, error)
```

| Parameter | Type                    | Description                                    |
|-----------|------------------------|------------------------------------------------|
| `env`     | `*provider.TestEnvironment` | Reference to the test environment (e.g., OCP/OKD). The builder uses this to query cluster versions. |

| Return | Type           | Description                                      |
|--------|----------------|--------------------------------------------------|
| `(*ClaimBuilder, error)` | A fully‑initialized `ClaimBuilder` or an error if initialization fails. |

---

## Key Steps & Dependencies

1. **Read environment configuration**  
   ```go
   os.Getenv("CNF_FEATURE_VALIDATION_REPORT_KEY")
   ```
   *Used to determine the output file name (`CNFFeatureValidationJunitXMLFileName`).*

2. **Create claim root structure**  
   `CreateClaimRoot()`  
   *Provides a base `ClaimRoot` object that will hold all test claims.*

3. **Load existing configuration (if any)**  
   ```go
   if configBytes, err := os.ReadFile(file); err == nil {
       UnmarshalConfigurations(configBytes, claimRoot)
   }
   ```
   *Allows incremental builds or reuse of previous claim data.*

4. **Populate metadata**  
   * Adds timestamps (`DateTimeFormatDirective`) and cluster‑specific version strings via:*
   - `GetVersionOcClient()`
   - `GetVersionOcp()`
   - `GetVersionK8s()`

5. **Generate test nodes**  
   `GenerateNodes()` – populates the claim root with the actual test cases based on the environment.

6. **Persist state (optional)**  
   *If a configuration file already existed, its contents are merged; otherwise, a new one is created.*

---

## Side‑Effects

| Effect | Explanation |
|--------|-------------|
| **File I/O** | Reads from and may write to a claim configuration file located in the working directory. The filename is derived from `CNFFeatureValidationReportKey`. |
| **Environment Variable Dependency** | Relies on environment variables; missing or malformed values will cause errors. |
| **Cluster Queries** | Calls external commands (`oc`, `kubectl`) via helper functions to fetch cluster versions, which may fail if the client is not configured correctly. |

---

## How It Fits in the Package

`claimhelper` provides a façade for creating, reading, and writing test claim data used by CertSuite’s reporting tools.  
- **`NewClaimBuilder`** is the entry point for consumers that want to programmatically generate or update claims.  
- The returned `*ClaimBuilder` exposes methods (e.g., `GenerateNodes`, `MarshalConfigurations`) that operate on the internal `ClaimRoot`.  
- Once built, the claim data can be marshalled to JSON/YAML or serialized into JUnit XML for downstream consumption by CI pipelines.

---

## Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[NewClaimBuilder] --> B{Read Env Var}
    B --> C[CNF_FEATURE_VALIDATION_REPORT_KEY]
    C --> D[CreateClaimRoot()]
    D --> E{File Exists?}
    E -- Yes --> F[UnmarshalConfigurations()]
    E -- No  --> G[Skip]
    F --> H[Merge with ClaimRoot]
    H --> I[GenerateNodes()]
    I --> J[Add Metadata (timestamp, versions)]
    J --> K[Return ClaimBuilder]
```

---

**TL;DR:** `NewClaimBuilder` builds a fully‑wired claim object from environment variables and optional existing data, ready for generating test claims in CertSuite.
