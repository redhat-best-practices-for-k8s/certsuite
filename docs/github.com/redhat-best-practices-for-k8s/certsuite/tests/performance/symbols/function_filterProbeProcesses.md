filterProbeProcesses`

| | |
|-|-|
| **File** | `tests/performance/suite.go` (line 364) |
| **Signature** | `func filterProbeProcesses(processes []*crclient.Process, container *provider.Container) ([]*crclient.Process, []*testhelper.ReportObject)` |
| **Visibility** | unexported |

### Purpose
During a performance test the controller manager runs multiple exec probes in each pod.  
`filterProbeProcesses` isolates the processes that belong to those probes and returns

1. **Filtered list of probe processes** – only the exec‑probe containers’ processes are kept.
2. **Report objects** – one `testhelper.ReportObject` per probe process, containing the
   command line and a reference to the owning container.

The returned report objects are later used by the test framework to collect metrics (e.g., CPU usage).

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `processes` | `[]*crclient.Process` | All processes discovered in the host namespace of the node. Each `Process` contains fields such as `PID`, `CommandLine`, etc. |
| `container` | `*provider.Container` | The container that was started for a probe. Contains metadata (`Name`, `Namespace`, `PodName`) used to match processes and create report objects. |

### Return Values

| Name | Type | Description |
|------|------|-------------|
| `[]*crclient.Process` | List of processes that belong to the probe containers only. All other processes are filtered out. |
| `[]*testhelper.ReportObject` | One report object per remaining process, populated with command line fields and container identifiers. |

### How It Works

1. **Get expected exec‑probe commands**  
   ```go
   cmds := getExecProbesCmds(container)
   ```
   `getExecProbesCmds` returns a slice of strings that represent the full command lines executed by each probe in the given container.

2. **Iterate over all processes**  
   For every process in `processes`:
   - The process’s command line (`proc.CommandLine`) is split into fields.
   - The first field (the executable) is compared against the list of expected probe commands.
   - If a match is found, the process is kept; otherwise it is discarded.

3. **Build report objects**  
   For each matched process:
   ```go
   obj := testhelper.NewContainerReportObject(container.PodName)
   obj.AddField("command", strings.Join(proc.CommandLine, " "))
   ```
   Two fields are added to the report object:  
   - `"command"` – the full command line of the probe.  
   - `"containerName"` – the name of the container that owns the process.

4. **Return**  
   The function returns:
   - A slice of the matched `*crclient.Process` objects.
   - A slice of the constructed `*testhelper.ReportObject`s.

### Dependencies & Side‑Effects

| Dependency | Description |
|------------|-------------|
| `getExecProbesCmds` | Provides the list of expected probe commands for a container. |
| `strings.Join`, `strings.Fields` | Used to manipulate command line strings. |
| `testhelper.NewContainerReportObject`, `AddField` | Construct report objects; no global state is modified. |

The function has **no side‑effects** beyond returning new slices and objects. It does not modify the input slices or any global variables.

### Role in the Package

In the *performance* test suite, after a container is started, the framework collects all host‑namespace processes (via `crclient.Process`).  
`filterProbeProcesses` is called to:

- **Isolate probe processes** – essential for accurate performance metrics.
- **Generate per‑process reports** – these objects are later aggregated and evaluated against thresholds.

Thus, this function acts as a bridge between raw process discovery and the test framework’s reporting infrastructure.
