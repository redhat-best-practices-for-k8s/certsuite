testContainerCertification`

```go
func testContainerCertification(
    ci provider.ContainerImageIdentifier,
    validator certdb.CertificationStatusValidator,
) bool
```

## Purpose

`testContainerCertification` is a helper used by the certification test suite to determine whether a specific container image satisfies the current certification policy.  
It abstracts the common logic of:

1. Querying the certificate database for a container’s certification status.
2. Applying a supplied validator that may impose additional constraints (e.g., operator type, online/offline status).

The function is **not exported**; it lives only inside the `certification` test package and is invoked by higher‑level tests.

## Parameters

| Name | Type | Description |
|------|------|-------------|
| `ci` | `provider.ContainerImageIdentifier` | A value that uniquely identifies a container image (registry, repo, tag). It is passed to the certificate database lookup. |
| `validator` | `certdb.CertificationStatusValidator` | A callback that receives a certification status and returns a boolean indicating whether the status meets custom rules. This allows tests to inject different validation logic without modifying this helper. |

## Return Value

- **`bool`** –  
  *`true`* if the container image is certified **and** passes the supplied validator; otherwise *false*.  
  The function never panics and propagates any errors from `IsContainerCertified` as a failed test case (converted to `false`).

## Key Dependencies

| Dependency | Role |
|------------|------|
| `certdb.IsContainerCertified` | Core lookup that checks the certificate database for a given image. Returns `(status, err)`. |
| `validator` | Applied to the returned status; may enforce rules such as operator type (`CertifiedOperator`) or network mode (`Online`). |

## Side Effects

- No global state is mutated.
- The function performs I/O via `IsContainerCertified`, which contacts the certificate database (or mock in tests).
- Errors from `IsContainerCertified` are logged by the caller; this helper simply returns `false`.

## How It Fits the Package

The `certification` package orchestrates end‑to‑end tests that validate whether operators and container images meet Red Hat’s certification criteria.  
Typical test flow:

```text
test case -> build image identifier ->
  testContainerCertification(id, validator) ->
    IsContainerCertified(id) ->
      validator(status)
```

The helper centralizes the pattern so individual test functions can focus on preparing inputs and asserting outcomes.

### Mermaid Flow Diagram (suggested)

```mermaid
flowchart TD
    A[Image Identifier] --> B{IsContainerCertified?}
    B -- Yes --> C[Certification Status]
    C --> D[Validator(status)]
    D -- Pass --> E[Return true]
    D -- Fail --> F[Return false]
    B -- No / Error --> F
```

This visual aid can be inserted into the package README to illustrate the certification check pipeline.
