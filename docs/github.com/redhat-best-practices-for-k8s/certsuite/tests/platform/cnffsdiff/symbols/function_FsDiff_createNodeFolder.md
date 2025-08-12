FsDiff.createNodeFolder`

| | |
|---|---|
| **Package** | `cnffsdiff` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/platform/cnffsdiff`) |
| **Receiver** | `f *FsDiff` – a value that holds state needed to perform filesystem diffing between a pod and the host. |
| **Signature** | `func (f *FsDiff) createNodeFolder() error` |
| **Exported?** | No – helper used only inside the package. |

---

## Purpose

`createNodeFolder` is an internal helper that prepares a temporary directory on the node (the machine where the test pod runs).  
It creates a folder structure that mirrors the location of the pod’s filesystem within the container runtime so that subsequent diff operations can read files from the host side.

The method:

1. Builds the destination path (`nodeTmpMountFolder`) inside the node.
2. Calls `execCommandContainer` to run `mkdir -p` inside the container, creating the directory on the host via the pod’s mount point.
3. Returns any error that occurs during command execution.

If the folder already exists, `mkdir -p` is harmless; the function simply propagates success.

---

## Inputs / Outputs

| Parameter | Type | Description |
|-----------|------|-------------|
| `f *FsDiff` | receiver | Holds configuration such as pod name, container runtime details, and environment variables. |

| Return | Type | Meaning |
|--------|------|---------|
| `error` | `nil` on success; otherwise an error describing why the folder could not be created (e.g., permission denied, command failure). |

---

## Key Dependencies

| Dependency | How it is used |
|------------|----------------|
| `execCommandContainer` | Executes a shell command inside the pod’s container. The function is passed the container name and a `mkdir -p` command built with `Sprintf`. |
| `fmt.Sprintf` | Formats the command string: `"mkdir -p %s"` where `%s` is `nodeTmpMountFolder`. A second call to `Sprintf` builds an error message template. |

These dependencies are imported from other parts of the `cnffsdiff` package and standard library (`fmt`). No external packages are invoked directly in this function.

---

## Side‑Effects

* Creates a directory tree on the node (via container runtime mount) at the path stored in the global variable `nodeTmpMountFolder`.
* Does **not** modify any state inside the `FsDiff` struct; it only performs an I/O operation.
* Errors are propagated upward; callers must handle them.

---

## How It Fits the Package

The `cnffsdiff` package compares filesystem snapshots between a Kubernetes pod and the host.  
During initialization, it needs a staging area on the node where files can be extracted from the container’s mount points.  

- **Step 1 – Node folder creation** (`createNodeFolder`)  
  Prepares that staging directory.

- **Step 2 – Extraction** (`extractFiles` – not shown here)  
  Copies files into this temporary folder.

- **Step 3 – Diffing** (`diffFolders`)  
  Compares the extracted contents against a baseline.

Thus, `createNodeFolder` is the first concrete action that makes subsequent diff operations possible. It encapsulates the platform‑specific logic of creating a directory inside a container’s namespace while keeping the rest of the code clean and testable.

---

### Suggested Mermaid Flow

```mermaid
flowchart TD
    A[FsDiff.createNodeFolder] --> B{Is nodeTmpMountFolder set?}
    B -- No --> C[Build mkdir command]
    C --> D[execCommandContainer(container, cmd)]
    D --> E{Success?}
    E -- Yes --> F[Return nil]
    E -- No  --> G[Return error]
```

This diagram visualises the simple control flow of the helper.
