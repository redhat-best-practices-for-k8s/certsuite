FsDiff – File‑System Diff Tester
`FsDiff` is the central type for the **cnffsdiff** test package.  
It orchestrates a comparison between two container file‑system snapshots (the *probe* pod and the target pod) by invoking `podman diff` inside a probe pod, mounting the probe’s rootfs into the host, and analysing the resulting output.

| Field | Type | Purpose |
|-------|------|---------|
| `ChangedFolders []string` | slice of folder paths | Directories that changed between snapshots. Populated after `RunTest`. |
| `DeletedFolders []string` | slice of folder paths | Directories removed in the target snapshot. |
| `Error error` | error | Stores the first fatal error encountered during a test run. |
| `check *checksdb.Check` | database entry | Metadata for the check being executed (ID, description, etc.). Used mainly for logging. |
| `clientHolder clientsholder.Command` | command executor | Interface to execute commands inside the probe pod (`ExecCommandContainer`). |
| `ctxt clientsholder.Context` | context | Runtime context passed to command execution. |
| `result int` | int | Exit status of the diff operation; 0 on success, non‑zero otherwise. |
| `useCustomPodman bool` | flag | Indicates whether a custom `podman` binary is mounted into the probe pod for diffing. |

### Constructor
```go
NewFsDiffTester(check *checksdb.Check,
                cmd clientsholder.Command,
                ctx clientsholder.Context,
                testID string) *FsDiff
```
* Determines if a custom `podman` should be used via `shouldUseCustomPodman`.  
* Creates an `FsDiff` instance, wiring the provided command executor and context.

### Core workflow – `RunTest`

1. **Setup** – If `useCustomPodman` is true:  
   * `installCustomPodman()` mounts a custom podman binary into the probe pod.  

2. **Mount host view** – `mountProbePodmanFolder()` mounts the probe pod’s rootfs on the host (e.g., `/var/lib/kubelet/pods/<id>/rootfs`).

3. **Diff execution** –  
   * Calls `runPodmanDiff(targetPath)` which runs `podman diff <target>` inside the probe pod and returns JSON output.

4. **Result handling** –  
   * Parses JSON via `Unmarshal` into a slice of map entries, extracting `Path` values.  
   * Filters paths that belong to target folders using `intersectTargetFolders`.  
   * Populates `ChangedFolders` or `DeletedFolders` accordingly.  

5. **Cleanup** –  
   * Unmounts the probe pod folder (`unmountCustomPodman()` → `unmountProbePodmanFolder`).  
   * Removes any temporary node directories created for custom podman (`deleteNodeFolder`).  

6. **Return status** – `result` is set to 0 on success; non‑zero if any step fails.

### Helper methods

| Method | Role |
|--------|------|
| `execCommandContainer(cmd, errorStr)` | Generic wrapper that runs a command inside the probe pod and aggregates stdout/stderr with an optional prefix. |
| `createNodeFolder()` / `deleteNodeFolder()` | Create/delete temporary directories on the host used to hold the custom podman binary. |
| `mountProbePodmanFolder()` / `unmountProbePodmanFolder()` | Mount/unmount the probe pod’s rootfs so that host commands can inspect it. |
| `installCustomPodman()` / `unmountCustomPodman()` | Install or remove a custom podman binary into the probe pod for diffing. |
| `intersectTargetFolders(folders []string)` | Filters a list of paths, keeping only those under any folder listed in `targetFolders`. Logs a warning if a path is outside expected locations. |
| `runPodmanDiff(target string)` | Executes `podman diff` inside the probe pod and returns raw JSON output or an error. |

### Public API

* **`GetResults() int`** – Returns the exit status (`result`) of the last run.  
  The caller can use this to decide if the test passed (0) or failed (>0).

* **`RunTest(testID string)`** – Executes the entire diffing workflow for a specific target pod identified by `testID`.  
  Side effects include:
  * Logging at various levels (`LogInfo`, `LogDebug`, `LogWarn`).  
  * Mounting/unmounting of host directories.  
  * Creation/deletion of temporary folders on the host.  

### Package role

`FsDiff` implements the **File System Diff** check in CertSuite’s platform tests.  
It is invoked by a higher‑level test harness that passes the relevant `Check` object and command executor.  
The results (`ChangedFolders`, `DeletedFolders`) are later aggregated into a test report.

---

#### Mermaid diagram (suggestion)

```mermaid
graph TD
  A[RunTest] --> B{useCustomPodman?}
  B -- yes --> C[installCustomPodman]
  C --> D[mountProbePodmanFolder]
  D --> E[runPodmanDiff(target)]
  E --> F[Parse JSON]
  F --> G[intersectTargetFolders]
  G --> H{folders found}
  H --> I[Populate Changed/Deleted]
  B -- no --> D
  I --> J[unmountCustomPodman]
  J --> K[cleanup temp dirs]
```

This diagram visualises the main decision points and side‑effect steps in a single flow.
