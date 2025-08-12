Process` – Lightweight representation of a running process

| Element | Description |
|---------|-------------|
| **Package** | `crclient` (`github.com/redhat-best-practices-for-k8s/certsuite/internal/crclient`) |
| **File** | `crclient.go` (line 33) |

### Purpose
The `Process` struct is a minimal data model that captures the essential attributes of a process running inside a Kubernetes container.  
It is used by the package’s API functions to:

* Return information about processes when enumerating a container’s PID namespace (`GetContainerProcesses`, `GetPidsFromPidNamespace`).  
* Format a human‑readable string for logging or debugging via its `String()` method.

### Fields

| Field | Type | Meaning |
|-------|------|---------|
| `Args`  | `string` | The full command line (arguments) that launched the process. |
| `PPid`  | `int`    | Parent‑process ID – the PID of the process’s parent within the same namespace. |
| `Pid`   | `int`    | Process identifier in its PID namespace. |
| `PidNs` | `int`    | Identifier of the PID namespace that owns this process (useful when comparing processes across containers). |

### Methods

* **`String() string`** – Implements `fmt.Stringer`.  
  Returns a formatted description:  

  ```go
  fmt.Sprintf("Process %d (PPid=%d, Namespace=%d) Args=%s", p.Pid, p.PPid, p.PidNs, p.Args)
  ```

  This is primarily for diagnostics; no side effects.

### Usage Flow

1. **Discover PIDs** – `GetPidsFromPidNamespace` executes `ps` inside the probe pod to list all processes in a target PID namespace.
2. **Parse Output** – Each line of the command output is parsed into a `Process` instance (PID, PPID, Args).
3. **Collect Results** – The resulting slice (`[]*Process`) is returned by the higher‑level API (`GetContainerProcesses`), which then filters or aggregates as needed.

### Dependencies

| Dependency | Role |
|------------|------|
| `ExecCommandContainer` | Runs `ps` in the probe pod to fetch process listings. |
| `GetNodeProbePodContext`, `GetClientsHolder`, etc. | Provide the environment, context, and client objects required for executing commands. |

### Side Effects

* The struct itself has no side effects; it is a pure data holder.
* Creation of a `Process` instance involves parsing command output, which can fail – callers must handle the returned error from the surrounding functions.

### Integration in the Package

The `crclient` package focuses on interacting with containers via Kubernetes APIs and executing commands inside probe pods.  
`Process` is one of the core data types that represents runtime information extracted during those interactions. It bridges low‑level command execution (via `ps`) with higher‑level logic such as:

* Determining if a container is still alive (`GetContainerProcesses` may return an empty slice).
* Validating process trees for security checks.
* Logging detailed process lists when tests fail.

---

#### Mermaid diagram – Process extraction flow

```mermaid
flowchart TD
    A[Probe Pod] -->|ExecCommandContainer("ps")| B[Raw ps output]
    B --> C{Parse lines}
    C --> D[Process struct (Pid, PPid, Args, PidNs)]
    D --> E[Return []*Process to caller]
```

This concise model shows how the `Process` struct is populated and subsequently consumed within the `crclient` package.
