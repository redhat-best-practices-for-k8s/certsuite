testContainersFsDiff`

| Feature | Description |
|---------|-------------|
| **Package** | `platform` (`tests/platform`) |
| **Visibility** | Unexported (used only inside the test suite) |
| **Signature** | `func(*checksdb.Check, *provider.TestEnvironment)` |
| **Purpose** | Verify that a container image under test did not install any new packages when compared to its base image. The check is performed by diffing the file‑system layers of the image and asserting that no additional package files are present in the “current” snapshot. |

### Parameters

| Name | Type | Role |
|------|------|------|
| `c` | `*checksdb.Check` | Holds metadata for the current test run (e.g., ID, status). The function updates its result via `SetResult`. |
| `env` | `*provider.TestEnvironment` | Test harness context that provides access to Docker clients, image references, and configuration needed to perform the diff. |

### Workflow

1. **Logging**  
   * Uses `LogInfo`/`LogError` from the test environment to report progress.

2. **Collecting Images**  
   - Calls `GetClientsHolder(env)` to obtain a client that can pull images.
   - Invokes `RunTest(ctx, env, imageRefs)` (via `NewFsDiffTester`) to perform the actual diff on all containers defined in the test environment.

3. **Processing Results**  
   - Retrieves raw diff results with `GetResults()`.
   - For each container:
     * Builds a report object (`NewContainerReportObject`).
     * Adds fields such as container name, image reference, and whether new packages were found.
     * If any container shows an unexpected package installation, logs an error.

4. **Aggregating**  
   - Combines per‑container reports into a final report list.
   - Calls `c.SetResult(report)` to store the outcome in the test database.

### Key Dependencies

| Dependency | Role |
|------------|------|
| `NewFsDiffTester` | Creates a tester that compares file system layers of images. |
| `GetClientsHolder` | Provides Docker client access needed for pulling images. |
| `RunTest`, `GetResults` | Execute the diff and return raw data. |
| `NewContainerReportObject`, `AddField` | Build structured reports for each container. |
| `LogInfo`, `LogError` | Emit test progress and error messages. |

### Side Effects

- **No state mutation**: The function only reads from the environment and writes the result back to the passed‑in `Check`.  
- **I/O**: Pulls images, performs layer diffs, and logs output.

### Placement in the Package

`testContainersFsDiff` is one of several helper functions that implement individual test cases for the *Platform* suite. It is invoked by the high‑level `RunSuite` logic when the suite’s configuration specifies a file‑system diff check. The function encapsulates all logic required to compare container layers, so other parts of the package can remain agnostic about the underlying Docker API calls.

---

#### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[Call testContainersFsDiff] --> B{Retrieve Clients}
    B --> C[Run FsDiffTester]
    C --> D{Collect Results}
    D --> E[Build Container Report]
    E --> F[Log Outcomes]
    F --> G[c.SetResult(report)]
```

This diagram illustrates the data flow from the test invocation through to result storage.
