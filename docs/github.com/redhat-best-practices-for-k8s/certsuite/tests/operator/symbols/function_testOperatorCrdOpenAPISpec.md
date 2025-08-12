testOperatorCrdOpenAPISpec`

```go
func testOperatorCrdOpenAPISpec(check *checksdb.Check, env *provider.TestEnvironment)
```

## Purpose

Verifies that the **operator CustomResourceDefinition (CRD)** is defined with an **OpenAPI v3 schema**.  
The function records the result in a `check` object and logs progress for visibility.

## Inputs & Outputs

| Parameter | Type                            | Role |
|-----------|---------------------------------|------|
| `check`   | `*checksdb.Check`               | Receives status information (fields, result) that is later stored in the test report. |
| `env`     | `*provider.TestEnvironment`     | Supplies the environment context needed by helper functions such as `IsCRDDefinedWithOpenAPI3Schema`. |

The function has no return value; it mutates the supplied `check`.

## Key Dependencies

- **Logging** – `LogInfo` is used to emit informational messages.
- **Validation** – `IsCRDDefinedWithOpenAPI3Schema(env)` performs the actual CRD schema check and returns a boolean.
- **Report Construction** –  
  - `NewOperatorReportObject(check, "<field>")` creates a new report entry.  
  - `AddField(report, key, value)` attaches metadata (e.g., the operator name).  
  - `SetResult(report, result)` records the pass/fail outcome.

## Side Effects

- Mutates the passed `check` object by adding fields and setting its result.
- Emits log messages to the test output stream.

## Flow Summary

```mermaid
flowchart TD
    A[Start] --> B{IsCRDDefinedWithOpenAPI3Schema?}
    B -- Yes --> C[Create PASS report]
    B -- No  --> D[Create FAIL report]
    C --> E[Add operator field]
    D --> F[Add operator field]
    E --> G[SetResult(PASS)]
    F --> H[SetResult(FAIL)]
    G --> I[End]
    H --> I
```

## Package Context

`testOperatorCrdOpenAPISpec` is part of the **operator** test suite (`github.com/redhat-best-practices-for-k8s/certsuite/tests/operator`).  
It is invoked during the operator validation phase, ensuring that any CRD deployed by an operator adheres to the OpenAPI v3 schema requirements mandated by Kubernetes. This contributes to overall compliance and interoperability checks within CertSuite.
