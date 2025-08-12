## `installCustomPodman` – FsDiff helper

```go
func (d *FsDiff) installCustomPodman() error
```

| Item | Description |
|------|-------------|
| **Receiver** | `*FsDiff` – the diff‑collector that owns the temporary node directories. |
| **Parameters** | None. |
| **Returns** | An `error`. A non‑nil value indicates a failure to prepare the custom Podman environment for comparison. |

### Purpose
`installCustomPodman` prepares an isolated “custom‑podman” installation inside the test node’s temporary mount area so that subsequent filesystem diffs can compare the real and the custom installations side‑by‑side.

It performs the following steps:

1. **Create a dedicated folder**  
   Calls `createNodeFolder(d, "custom-podman")` to allocate a fresh directory under the node's temp root (`nodeTmpMountFolder`). This isolates the custom Podman files from any existing system data.

2. **Mount the custom podman source**  
   Invokes `mountProbePodmanFolder()` which mounts the pre‑packaged probe folder (see constant `partnerPodmanFolder`) into the freshly created node directory. The mount is performed at runtime, ensuring that only the required files are present for diffing.

3. **Clean up**  
   Whether mounting succeeds or fails, it calls `deleteNodeFolder(d, "custom-podman")` to remove the temporary directory and unmount any filesystem layers.

4. **Error handling**  
   Any error from the above steps is wrapped with context via `Errorf` (e.g., `"failed to install custom podman: %w"`). A successful run returns `nil`.

### Key Dependencies

| Dependency | Role |
|------------|------|
| `LogInfo` | Emits informational logs about progress. |
| `createNodeFolder` | Allocates a temporary directory for the custom Podman installation. |
| `mountProbePodmanFolder` | Performs the bind‑mount of the probe source into the temp directory. |
| `deleteNodeFolder` | Removes the temporary directory and cleans up mounts. |
| `Errorf` | Formats errors with contextual messages. |

### Side Effects

* Creates a new directory under the node’s temporary mount root (`nodeTmpMountFolder/custom-podman`).  
* Performs a bind‑mount of the probe source into that directory.  
* On completion (success or failure) the directory and its mounts are removed.

These side effects are intentional; the function is meant to be run once per test run and leaves no residue on the host system.

### How it fits the package

`FsDiff` orchestrates filesystem comparisons between a standard Podman installation and a custom one.  
`installCustomPodman` is the preparatory step that materialises the custom installation in a sandboxed location so that later diff logic can operate on two comparable directory trees:

```
nodeTmpMountFolder/
├─ podman          ← real system installation
└─ custom-podman   ← mounted probe installation
```

The resulting structure allows `FsDiff` to invoke its generic comparison routines without affecting the underlying host environment.
