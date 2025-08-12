## Package crclient (github.com/redhat-best-practices-for-k8s/certsuite/internal/crclient)



### Structs

- **Process** (exported) — 4 fields, 1 methods

### Functions

- **ExecCommandContainerNSEnter** — func(string, *provider.Container)(string, error)
- **GetContainerPidNamespace** — func(*provider.Container, *provider.TestEnvironment)(string, error)
- **GetContainerProcesses** — func(*provider.Container, *provider.TestEnvironment)([]*Process, error)
- **GetNodeProbePodContext** — func(string, *provider.TestEnvironment)(clientsholder.Context, error)
- **GetPidFromContainer** — func(*provider.Container, clientsholder.Context)(int, error)
- **GetPidsFromPidNamespace** — func(string, *provider.Container)([]*Process, error)
- **Process.String** — func()(string)

### Call graph (exported symbols, partial)

```mermaid
graph LR
  ExecCommandContainerNSEnter --> GetTestEnvironment
  ExecCommandContainerNSEnter --> GetNodeProbePodContext
  ExecCommandContainerNSEnter --> Errorf
  ExecCommandContainerNSEnter --> GetClientsHolder
  ExecCommandContainerNSEnter --> GetPidFromContainer
  ExecCommandContainerNSEnter --> Errorf
  ExecCommandContainerNSEnter --> Itoa
  ExecCommandContainerNSEnter --> ExecCommandContainer
  GetContainerPidNamespace --> GetNodeProbePodContext
  GetContainerPidNamespace --> Errorf
  GetContainerPidNamespace --> GetPidFromContainer
  GetContainerPidNamespace --> Errorf
  GetContainerPidNamespace --> Debug
  GetContainerPidNamespace --> Sprintf
  GetContainerPidNamespace --> ExecCommandContainer
  GetContainerPidNamespace --> GetClientsHolder
  GetContainerProcesses --> GetContainerPidNamespace
  GetContainerProcesses --> Errorf
  GetContainerProcesses --> GetPidsFromPidNamespace
  GetNodeProbePodContext --> Errorf
  GetNodeProbePodContext --> NewContext
  GetPidFromContainer --> Debug
  GetPidFromContainer --> Errorf
  GetPidFromContainer --> GetClientsHolder
  GetPidFromContainer --> ExecCommandContainer
  GetPidFromContainer --> Errorf
  GetPidFromContainer --> Errorf
  GetPidFromContainer --> Atoi
  GetPidFromContainer --> TrimSuffix
  GetPidsFromPidNamespace --> GetTestEnvironment
  GetPidsFromPidNamespace --> GetNodeProbePodContext
  GetPidsFromPidNamespace --> Errorf
  GetPidsFromPidNamespace --> ExecCommandContainer
  GetPidsFromPidNamespace --> GetClientsHolder
  GetPidsFromPidNamespace --> Errorf
  GetPidsFromPidNamespace --> GetPodName
  GetPidsFromPidNamespace --> MustCompile
  Process_String --> Sprintf
```

### Symbol docs

- [struct Process](symbols/struct_Process.md)
- [function ExecCommandContainerNSEnter](symbols/function_ExecCommandContainerNSEnter.md)
- [function GetContainerPidNamespace](symbols/function_GetContainerPidNamespace.md)
- [function GetContainerProcesses](symbols/function_GetContainerProcesses.md)
- [function GetNodeProbePodContext](symbols/function_GetNodeProbePodContext.md)
- [function GetPidFromContainer](symbols/function_GetPidFromContainer.md)
- [function GetPidsFromPidNamespace](symbols/function_GetPidsFromPidNamespace.md)
- [function Process.String](symbols/function_Process_String.md)
