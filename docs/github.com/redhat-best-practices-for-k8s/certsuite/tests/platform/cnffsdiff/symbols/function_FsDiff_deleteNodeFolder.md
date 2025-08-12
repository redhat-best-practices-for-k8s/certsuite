# `deleteNodeFolder` – Package **cnffsdiff**

> Internal helper used by the `FsDiff` type to clean up temporary
> directories that were created on a remote node during an FSDIFF
> operation.

## Purpose

When the test framework mounts a container’s filesystem into the host,
temporary folders are created inside the node’s `/tmp` hierarchy:

```
/tmp/.certsuite-fsdiff-<id>
  └─ mount‑point   (nodeTmpMountFolder)
```

`deleteNodeFolder` removes that temporary mount directory from the
remote node so that subsequent runs start with a clean state.

## Receiver

```go
func (f *FsDiff) deleteNodeFolder() error
```

- `*FsDiff`: holds information about the current test run and contains
  the remote node name (`node`) used in the command string below.  
  The function is **unexported** – only code inside this package calls it.

## Inputs / Outputs

| Parameter | Description |
|-----------|-------------|
| none     | The method uses fields of `FsDiff` (not shown here) and two
             global constants to build the command string. |

| Return value | Type | Meaning |
|--------------|------|---------|
| `error`      | `error` | `nil` if removal succeeded, otherwise an error produced by
              `execCommandContainer`. |

## Key Steps

1. **Build the remote command**  
   ```go
   cmd := fmt.Sprintf(
       "rm -rf %s/%s",
       nodeTmpMountFolder,
       f.node,
   )
   ```
   * `nodeTmpMountFolder` – global constant pointing to the base temp
     directory on the node.
   * `f.node` – the name of the remote host.

2. **Execute it inside a container**  
   ```go
   _, err := execCommandContainer(cmd)
   ```
   The helper runs the command in the test’s podman container that has
   access to the node’s filesystem, so the temporary directory is
   removed from the node itself.

3. **Return any error** – callers decide whether a failure should halt
   the test or be logged.

## Dependencies

| Dependency | Why it matters |
|-------------|----------------|
| `execCommandContainer` | Executes arbitrary shell commands on the target node via the test container. |
| `Sprintf` (from `fmt`) | Builds the command string. |

No other package-level variables are accessed directly; only the global
constants defined at the top of the file (`nodeTmpMountFolder`, etc.) are used.

## Side‑Effects

- Deletes files/directories from the remote node – irreversible for that
  test run.
- No changes to local state or configuration are made.

## Where It Fits in the Package

The `cnffsdiff` package orchestrates filesystem comparisons between a
container and its host. During a diff, temporary mount points are
created on the node; `deleteNodeFolder` is invoked (typically from a
cleanup routine) to ensure those directories do not persist across
tests or leak resources.

```
FsDiff
 ├─ createTempMount()   // sets up nodeTmpMountFolder/<node>
 └─ deleteNodeFolder()  // removes it after diff completes
```
