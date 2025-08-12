runTestSuite` – Run the Certsuite test suite

| Item | Details |
|------|---------|
| **Signature** | `func(*cobra.Command, []string) error` |
| **Visibility** | Unexported (`runTestSuite`) – used only inside the *run* command package. |
| **Purpose** | Orchestrates a single execution of Certsuite’s test suite. It parses CLI flags into internal parameters, starts any required server components, runs the tests, and performs cleanup. |

### Workflow

```mermaid
flowchart TD
    A[parse flags → initTestParamsFromFlags] --> B[GetTestParameters]
    B --> C{server needed?}
    C -- yes --> D[StartServer]
    D --> E[Startup]
    C -- no  --> F[skip server steps]
    E --> G[Run tests (Run)]
    G --> H[Shutdown]
    H --> I[log results & exit]
```

1. **Parameter extraction** – `initTestParamsFromFlags(cmd, args)` reads command‑line flags and populates a shared configuration structure (`GetTestParameters()`).
2. **Server handling**  
   * If the parameters indicate that a test server is required, `StartServer` launches it.  
   * `Startup` performs any additional initialization for the server or environment.
3. **Test execution** – `Run()` runs the actual test suite; it returns an error if tests fail or a system issue occurs.
4. **Cleanup** – Regardless of success, `Shutdown` stops the server and releases resources.
5. **Logging & exit** – Informational logs are emitted with `Info`, fatal errors terminate via `Fatal`. The function returns any error from `Run()` to propagate failure status.

### Inputs / Outputs

| Parameter | Type | Role |
|-----------|------|------|
| `cmd` | `*cobra.Command` | Cobra command context (contains flag values). |
| `args` | `[]string` | Positional arguments; currently unused but kept for Cobra compatibility. |

**Return value**

- `error`:  
  * `nil` – all tests passed and cleanup succeeded.  
  * non‑nil – indicates failure during test execution or setup/teardown.

### Dependencies

| Dependency | Role |
|------------|------|
| `initTestParamsFromFlags` | Parses CLI flags into internal config. |
| `GetTestParameters` | Provides the current configuration snapshot. |
| `StartServer`, `Shutdown` | Manage optional server lifecycle. |
| `Startup` | Additional environment setup (e.g., DB connections). |
| `Run` | Executes the test suite logic. |
| Logging helpers (`Info`, `Fatal`) | Emit progress and error messages. |

### Side Effects

* May start/stop an external test server.
* Writes logs via the package’s logging utilities.
* Modifies global state only through the shared configuration returned by `GetTestParameters`.

### Package Context

The `run` package implements the `certsuite run` CLI command.  
`runTestSuite` is bound to that command as its `RunE` handler, making it the core routine that drives a test execution from flag parsing to cleanup.
