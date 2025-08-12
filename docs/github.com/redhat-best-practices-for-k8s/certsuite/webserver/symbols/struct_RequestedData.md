RequestedData` – Structured payload for the CertSuite web‑server

The **`RequestedData`** type is a plain Go struct that aggregates all of the configuration parameters supplied by the user when launching a test run through the web‑server API.  
It lives in the `webserver` package (`github.com/redhat-best-practices-for-k8s/certsuite/webserver`) and is used only for data transport – no business logic is attached to it.

> **Why is this struct needed?**  
> The web‑server receives JSON (or other encoded) payloads from clients. These payloads describe *what* tests should be run, *where* they should be executed, and various runtime options. Rather than passing a long list of arguments around, the server decodes the request into a single strongly typed struct that is then passed to helper functions such as `updateTnf`. This makes the API easier to evolve: adding a new field simply requires updating the struct definition.

## Field Overview

| Field | Type | Typical content | Notes |
|-------|------|-----------------|-------|
| `AcceptedKernelTaints` | `[]string` | List of kernel taints that the test runner should ignore. | Used by the test harness to filter nodes. |
| `CollectorAppEndPoint` | `[]string` | Endpoint(s) for a collector application (e.g., metrics sink). | Usually one URL. |
| `CollectorAppPassword` | `[]string` | Password(s) used to authenticate to the collector. | Should be kept secret; not logged. |
| `ConnectAPIBaseURL` | `[]string` | Base URL of the Connect API that provides policy or test data. | Can contain multiple URLs for fail‑over. |
| `ConnectAPIKey` | `[]string` | Authorization key for the Connect API. | Sensitive – avoid printing. |
| `ConnectAPIProxyPort` | `[]string` | Port number of an HTTP proxy to use when talking to Connect. | Optional; empty slice means no proxy. |
| `ConnectAPIProxyURL` | `[]string` | Proxy URL. | Paired with `ConnectAPIProxyPort`. |
| `ConnectProjectID` | `[]string` | Identifier for the project in Connect that owns this test run. | Used to scope data retrieval. |
| `ExecutedBy` | `[]string` | Name or ID of the user/CI system that triggered the run. | Helps with audit logs. |
| `ManagedDeployments` | `[]string` | Names of deployments that should be managed during tests. | e.g., scaling tests. |
| `ManagedStatefulsets` | `[]string` | Names of statefulsets to manage. | Similar to deployments. |
| `OperatorsUnderTestLabels` | `[]string` | Kubernetes labels that identify operators under test. | Used for discovery. |
| `PartnerName` | `[]string` | Name of the partner providing the operator or workload. | Optional metadata. |
| `PodsUnderTestLabels` | `[]string` | Labels used to select pods that are part of the test subject. | Enables selective probing. |
| `ProbeDaemonSetNamespace` | `[]string` | Namespace where the probe daemonset should be deployed. | Usually a fixed namespace. |
| `SelectedOptions` | `[]string` | Arbitrary flags or options chosen by the user (e.g., “run‑slow”). | Handled by switch logic elsewhere. |
| `Servicesignorelist` | `[]string` | Services that should be excluded from certain checks. | May contain regexes. |
| `SkipHelmChartList` | `[]string` | Helm charts that should not be processed. | Useful for large repos. |
| `SkipScalingTestDeploymentsname` | `[]string` | Names of deployments to skip during scaling tests. | Prevents interference with critical workloads. |
| `SkipScalingTestDeploymentsnamespace` | `[]string` | Namespaces of the above deployments. | Allows cross‑namespace filtering. |
| `SkipScalingTestStatefulsetsname` | `[]string` | Names of statefulsets to skip. | |
| `SkipScalingTestStatefulsetsnamespace` | `[]string` | Namespaces of the above statefulsets. | |
| `TargetCrdFiltersnameSuffix` | `[]string` | Suffixes that filter CRD names (e.g., “‑operator”). | |
| `TargetCrdFiltersscalable` | `[]string` | Indicates whether a CRD is scalable (`true/false`). | Used for scaling tests. |
| `TargetNameSpaces` | `[]string` | Namespaces to target for tests. | |
| `ValidProtocolNames` | `[]string` | List of network protocols that are considered valid (e.g., “TCP”, “UDP”). | Validates test definitions. |

> **All fields are slices** – this mirrors how JSON arrays are unmarshaled into Go. Even if a field contains a single value, the slice abstraction keeps the code simple and consistent.

## How `RequestedData` is used

1. **Deserialization**  
   The web‑server receives an HTTP request (typically POST) with a JSON body.  
   ```go
   var req RequestedData
   err := json.NewDecoder(r.Body).Decode(&req)
   ```
2. **Validation & Normalization**  
   Functions like `updateTnf` take the raw slice values, perform sanity checks, and sometimes convert them into internal maps or configuration structs.
3. **Propagation to Test Runner**  
   After validation, the struct is passed down the call stack (often as a pointer) until it reaches the test harness that will execute the requested tests.

> The design intentionally keeps `RequestedData` immutable after construction – no methods mutate it. This guarantees thread‑safety when multiple goroutines inspect the same request.

## Dependencies & Side Effects

- **Dependencies**:  
  - Standard library packages (`encoding/json`, `net/http`).  
  - No external dependencies beyond those required by the rest of the webserver package.
- **Side effects**:  
  - The struct itself has no side effects.  
  - Functions that consume it may log values, write to a database, or trigger network calls (e.g., `updateTnf` uses `json.Marshal`/`Unmarshal`, and may call `log.Fatal` on error).

## Placement in the Package

Within `webserver/webserver.go`, the struct sits near other request‑handling helpers. It is one of the first types defined, making it readily visible to developers working on API endpoints or documentation generators. All request‑processing functions reference this type by value or pointer; thus it acts as the central contract between HTTP clients and the internal test execution engine.

```mermaid
flowchart TD
    subgraph Client
        A[HTTP POST /run] -->|JSON body| B[webserver handler]
    end

    subgraph Server
        B --> C{decode JSON}
        C --> D[RequestedData struct]
        D --> E{validation & mapping}
        E --> F[updateTnf (or other runner)]
        F --> G[Execute tests]
    end
```

### TL;DR

- **Purpose**: Carries all user‑supplied configuration for a CertSuite test run.  
- **Inputs/Outputs**: Populated from JSON request body, read by downstream functions.  
- **Key Dependencies**: Only the Go standard library for decoding/encoding.  
- **Side Effects**: None inherent; used solely as data transport.  
- **Package Role**: Serves as the API contract between HTTP clients and the test execution engine within `webserver`.
