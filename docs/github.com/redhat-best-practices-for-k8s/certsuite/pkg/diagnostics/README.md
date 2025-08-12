## Package diagnostics (github.com/redhat-best-practices-for-k8s/certsuite/pkg/diagnostics)



### Structs

- **NodeHwInfo** (exported) — 4 fields, 0 methods

### Functions

- **GetCniPlugins** — func()(map[string][]interface{})
- **GetCsiDriver** — func()(map[string]interface{})
- **GetHwInfoAllNodes** — func()(map[string]NodeHwInfo)
- **GetNodeJSON** — func()(map[string]interface{})
- **GetVersionK8s** — func()(string)
- **GetVersionOcClient** — func()(string)
- **GetVersionOcp** — func()(string)

### Call graph (exported symbols, partial)

```mermaid
graph LR
  GetCniPlugins --> GetTestEnvironment
  GetCniPlugins --> GetClientsHolder
  GetCniPlugins --> make
  GetCniPlugins --> NewContext
  GetCniPlugins --> ExecCommandContainer
  GetCniPlugins --> Error
  GetCniPlugins --> String
  GetCniPlugins --> Unmarshal
  GetCsiDriver --> GetClientsHolder
  GetCsiDriver --> List
  GetCsiDriver --> CSIDrivers
  GetCsiDriver --> StorageV1
  GetCsiDriver --> TODO
  GetCsiDriver --> Error
  GetCsiDriver --> NewScheme
  GetCsiDriver --> AddToScheme
  GetHwInfoAllNodes --> GetTestEnvironment
  GetHwInfoAllNodes --> GetClientsHolder
  GetHwInfoAllNodes --> make
  GetHwInfoAllNodes --> getHWJsonOutput
  GetHwInfoAllNodes --> Error
  GetHwInfoAllNodes --> Error
  GetHwInfoAllNodes --> getHWJsonOutput
  GetHwInfoAllNodes --> Error
  GetNodeJSON --> GetTestEnvironment
  GetNodeJSON --> Marshal
  GetNodeJSON --> Error
  GetNodeJSON --> Unmarshal
  GetNodeJSON --> Error
  GetVersionK8s --> GetTestEnvironment
  GetVersionOcp --> GetTestEnvironment
  GetVersionOcp --> IsOCPCluster
```

### Symbol docs

- [struct NodeHwInfo](symbols/struct_NodeHwInfo.md)
- [function GetCniPlugins](symbols/function_GetCniPlugins.md)
- [function GetCsiDriver](symbols/function_GetCsiDriver.md)
- [function GetHwInfoAllNodes](symbols/function_GetHwInfoAllNodes.md)
- [function GetNodeJSON](symbols/function_GetNodeJSON.md)
- [function GetVersionK8s](symbols/function_GetVersionK8s.md)
- [function GetVersionOcClient](symbols/function_GetVersionOcClient.md)
- [function GetVersionOcp](symbols/function_GetVersionOcp.md)
