FsDiff.RunTest`

```go
func (f FsDiff) RunTest() func(string)
```

`RunTest` is a **factory** that produces a test‑function for the *cnffsdiff* test suite.  
The returned function accepts the name of a target folder to compare and performs a full
diff between the container filesystem (mounted by Podman) and the expected state
stored under `targetFolders`. The method orchestrates mounting, diffing,
and cleanup while logging progress.

## Purpose

- **Mount** a custom Podman instance into a temporary directory.
- **Run** `podman diff` against the mounted image to capture the filesystem changes.
- **Compare** those changes with the expected files in one or more target folders.
- **Un‑mount** and clean up resources after each run.

This is used by integration tests that verify whether the container images
contain exactly the configuration files required by the platform.

## Inputs / Outputs

| Parameter | Type | Description |
|-----------|------|-------------|
| None (method receiver) | `FsDiff` | Holds configuration such as image name and mount paths. |

The method returns a **function** of type `func(string)`:

- The argument is the *name* of a folder inside `targetFolders` that holds
  the expected file layout.
- It does not return a value; errors are logged through the package’s logging
  helpers (`LogInfo`, `LogWarn`, etc.).

## Key Steps & Dependencies

1. **Mounting**  
   ```go
   installCustomPodman(f.image, tmpMountDestFolder)
   ```
   - Starts Podman with the specified image.
   - Errors are wrapped in a retry loop governed by `errorCode125RetrySeconds`.

2. **Diff Execution**  
   ```go
   runPodmanDiff(tmpMountDestFolder)
   ```
   - Invokes `podman diff` on the mounted container and stores JSON output.

3. **Target Folder Intersection**  
   ```go
   intersectTargetFolders(targetFolders, targetFolderName)
   ```
   - Computes which folders in the repository correspond to the test’s expectations.
   - If no match is found, a warning is logged and the function returns early.

4. **Comparison Logic**  
   - The JSON diff output is unmarshaled (`Unmarshal`).
   - For each expected file path in the target folder(s), `Contains`
     checks whether the diff includes that entry.
   - Missing or extra entries trigger an error via `Errorf`.

5. **Cleanup**  
   ```go
   unmountCustomPodman(tmpMountDestFolder)
   ```
   - Unmounts the container filesystem and removes temporary directories.

6. **Logging & Timing**  
   - Uses `LogInfo`, `LogWarn`, `LogDebug` for progress.
   - Sleeps between retries (`Sleep`) to give Podman time to recover from transient failures.

## Side Effects

- Creates a temporary mount directory under `/tmp/…`.
- Launches a Podman process; the process is terminated after the diff.
- Generates and deletes JSON files containing the diff output.
- Logs diagnostic messages that may be captured by the test harness.

## How It Fits in the Package

`FsDiff.RunTest` is the core driver for filesystem‑diff tests.  
Other parts of `cnffsdiff` provide helper functions (`installCustomPodman`,
`runPodmanDiff`, etc.) and constants that control retry behavior and paths.
The test suite calls `RunTest()` once per image, then iterates over all
target folders to validate the container’s contents.

---

### Mermaid Diagram (suggestion)

```mermaid
flowchart TD
  A[Call RunTest()] --> B{Return testFn}
  B --> C[testFn(targetFolderName)]
  C --> D[installCustomPodman]
  D --> E[podman diff -> JSON]
  E --> F[unmarshal diff]
  F --> G[intersectTargetFolders]
  G --> H[Check Contains for each expected file]
  H --> I{All matched?}
  I -- yes --> J[Log success]
  I -- no --> K[Errorf missing/extra files]
  J & K --> L[unmountCustomPodman]
  L --> M[Cleanup temp dirs]
```

This diagram visualises the flow of a single test execution.
