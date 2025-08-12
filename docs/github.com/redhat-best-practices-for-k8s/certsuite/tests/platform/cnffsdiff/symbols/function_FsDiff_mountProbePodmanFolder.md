FsDiff.mountProbePodmanFolder`

| Item | Details |
|------|---------|
| **Package** | `cnffsdiff` – part of the CertSuite test harness for container‑file system diffing. |
| **Receiver** | `f *FsDiff` – the struct that orchestrates file‑system comparisons between two containers. |
| **Signature** | `func (f *FsDiff) mountProbePodmanFolder() error` |
| **Exported?** | No – internal helper used during test setup. |

---

## Purpose

During a CNF (Container Network Function) test run, the framework needs to access the Podman‑managed image layers on the test node.  
`mountProbePodmanFolder` mounts the Podman's storage directory (`/var/lib/containers/storage`) into a temporary location under `/tmp/certsuite-mounts/<node>`.  
The function is called once per test run; it ensures that the mount point exists and can be read by subsequent diff logic.

---

## Inputs / Outputs

| Input | Description |
|-------|-------------|
| None – all data comes from the receiver (`f`) and package globals. |

| Output | Type | Meaning |
|--------|------|---------|
| `error` | `nil` on success, non‑nil if mounting fails (e.g., permission denied, mount already exists). |

---

## Key Dependencies

| Dependency | How it is used |
|------------|----------------|
| `execCommandContainer` | Executes a shell command inside the test container. It runs: <br>`mount -t tmpfs none /tmp/certsuite-mounts/<node>` or similar to create a temporary mount point. |
| `Sprintf` (from `fmt`) | Builds the full mount destination path and the command string with dynamic node name (`f.node`). |

---

## Side‑Effects

1. **Creates** a directory under `/tmp/certsuite-mounts/<node>` if it does not already exist.
2. **Mounts** the Podman storage filesystem into that directory (via `execCommandContainer`).
3. **Leaves** the mount point in place for the lifetime of the test; cleanup is performed elsewhere (`unmountProbePodmanFolder`).  
4. **Logs** errors through returned `error`; no internal logging in this function.

---

## How It Fits the Package

The `cnffsdiff` package implements a file‑system diff engine that compares two container images (the CNF under test and its reference).  
Mounting the Podman storage is a prerequisite for accessing the underlying layers.  

Typical call sequence:

```go
fd := &FsDiff{node: "worker1"}
if err := fd.mountProbePodmanFolder(); err != nil {
    return err
}
defer fd.unmountProbePodmanFolder()
...
```

The function is deliberately lightweight; it only prepares the environment for the more expensive diffing logic that follows.  

---

## Suggested Mermaid Diagram

```mermaid
graph TD
  A[FsDiff.mountProbePodmanFolder] --> B{Check /tmp/certsuite-mounts/<node>}
  B -- not exists --> C[Create directory]
  C --> D[execCommandContainer("mount ...")]
  D --> E[Mount successful?]
  E -- yes --> F[Return nil]
  E -- no --> G[Return error]
```

---
