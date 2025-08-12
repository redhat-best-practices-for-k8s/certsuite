LoadChecks` – Test Suite Initialisation

```go
func LoadChecks() func()
```

## Purpose

`LoadChecks` is the bootstrap function for the *pre‑flight* test suite in the CertSuite project.  
It creates and registers a set of checks that will be executed when the tests run. The function returns a `func()` which is intended to be used as a Ginkgo *BeforeEach* hook or similar entry point.

## Parameters & Return

| Name | Type | Description |
|------|------|-------------|
| –    | `() func()` | A closure that, when called, performs the suite‑level setup. No arguments are required and no values are returned by the closure itself; side effects are applied to package globals and registered test groups.

## Key Steps (in order)

1. **Debug Logging**  
   Calls `Debug` to log entry into the function.

2. **Environment Capture**  
   - `env = GetTestEnvironment()` obtains a global `provider.TestEnvironment`.  
   - The environment is stored in the package‑level variable `env`, making it accessible to other helpers that need cluster context.

3. **BeforeEach Hook Registration**  
   - `beforeEachFn` (package‑level var) is assigned a function that will run before each individual test.  
   - This function calls `WithBeforeEachFn` to register the closure with Ginkgo’s testing framework.

4. **Check Group Creation**  
   - `NewChecksGroup("preflight")` creates a container for all pre‑flight checks, grouping them under the name “preflight”.

5. **Add Container Checks**  
   - Calls `testPreflightContainers()` to generate checks that verify container‑level prerequisites (e.g., image pull policies, resource limits).  
   - These checks are added to the group created in step 4.

6. **Conditional Operator Checks**  
   - If the cluster is identified as an OpenShift Platform (`IsOCPCluster(env)`), `testPreflightOperators()` is called to add operator‑specific checks (e.g., Operator Lifecycle Manager presence).  
   - Each of these branches logs its action with `Info`.

7. **Return Closure**  
   The returned closure contains the logic described above and can be invoked by the test runner.

## Dependencies

| Function | Role |
|----------|------|
| `Debug` | Logging at debug level. |
| `GetTestEnvironment` | Provides cluster details used for conditional logic. |
| `WithBeforeEachFn` | Registers a function to run before each test case. |
| `NewChecksGroup` | Creates a logical grouping of checks. |
| `testPreflightContainers` | Generates container‑related checks. |
| `IsOCPCluster` | Determines if the environment is OpenShift. |
| `testPreflightOperators` | Generates operator‑specific checks for OCP clusters. |

## Side Effects

* Sets package variables `env` and `beforeEachFn`.  
* Registers a Ginkgo `BeforeEach` hook via `WithBeforeEachFn`.  
* Populates the pre‑flight check group with test cases.

These effects are intentional: the function is designed to be called once during suite initialisation, after which the registered checks will run automatically for each test case.

## Package Context

The `preflight` package contains all tests that verify a Kubernetes cluster meets prerequisites before CertSuite can execute its certification tests.  
`LoadChecks` ties together environment discovery and check registration, forming the backbone of this pre‑flight validation phase.
