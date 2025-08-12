FsDiff.unmountCustomPodman`

```go
func (f *FsDiff) unmountCustomPodman() func()
```

#### Purpose  
`unmountCustomPodman` performs cleanup after a test that has mounted a custom Podman image directory into the node’s filesystem.  
It returns an empty function so it can be used as a deferred clean‑up routine:

```go
defer f.unmountCustomPodman()   // no-op return value, just for symmetry with other helpers
```

#### Inputs / Outputs  
* **Receiver** – `f *FsDiff` (no arguments).  
* **Return value** – an anonymous function of type `func()` that does nothing.  
  The returned function is only a placeholder; the real work happens immediately when `unmountCustomPodman` is invoked.

#### Key steps
| Step | Action | Notes |
|------|--------|-------|
| 1 | Log “Unmounting custom podman folder” | Uses `LogInfo`. |
| 2 | Call `unmountProbePodmanFolder()` | Detaches the previously mounted probe directory (`partnerPodmanFolder`). |
| 3 | Log “Cleaning up temporary mount destination folder” | Uses `LogInfo`. |
| 4 | Delete node‑side temp folder (`nodeTmpMountFolder`) via `deleteNodeFolder` | Removes any leftover files that were created during the mount. |

#### Dependencies
* **`LogInfo`** – simple logger used throughout the package for test diagnostics.
* **`unmountProbePodmanFolder`** – helper that unmounts the probe directory from the node; this function is defined elsewhere in `fsdiff.go`.
* **`deleteNodeFolder`** – removes a specified directory tree on the target node.

#### Side effects
* The probe folder (`partnerPodmanFolder`) is detached from the node’s filesystem.
* Temporary directories under `nodeTmpMountFolder` are deleted, freeing disk space and ensuring subsequent tests start with a clean state.

#### How it fits in the package  
`cnffsdiff` implements file‑system diff logic used by CertSuite platform tests.  
During those tests, custom Podman images may be mounted into a node to provide test data.  
`unmountCustomPodman` is part of the teardown phase that guarantees these mounts do not persist beyond the test’s lifetime.

---

#### Suggested Mermaid diagram (optional)

```mermaid
flowchart TD
    A[FsDiff.unmountCustomPodman] --> B{Log “Unmounting custom podman folder”}
    B --> C[unmountProbePodmanFolder]
    C --> D{Log “Cleaning up temporary mount destination folder”}
    D --> E[deleteNodeFolder(nodeTmpMountFolder)]
```
This diagram visualises the linear sequence of actions performed by the function.
