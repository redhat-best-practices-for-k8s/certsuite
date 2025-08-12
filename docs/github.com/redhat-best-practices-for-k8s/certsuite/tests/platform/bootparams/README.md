## Package bootparams (github.com/redhat-best-practices-for-k8s/certsuite/tests/platform/bootparams)



### Functions

- **GetMcKernelArguments** — func(*provider.TestEnvironment, string)(map[string]string)
- **TestBootParamsHelper** — func(*provider.TestEnvironment, *provider.Container, *log.Logger)(error)

### Call graph (exported symbols, partial)

```mermaid
graph LR
  GetMcKernelArguments --> ArgListToMap
  TestBootParamsHelper --> Errorf
  TestBootParamsHelper --> GetMcKernelArguments
  TestBootParamsHelper --> getCurrentKernelCmdlineArgs
  TestBootParamsHelper --> Errorf
  TestBootParamsHelper --> getGrubKernelArgs
  TestBootParamsHelper --> Errorf
  TestBootParamsHelper --> Warn
  TestBootParamsHelper --> Debug
```

### Symbol docs

- [function GetMcKernelArguments](symbols/function_GetMcKernelArguments.md)
- [function TestBootParamsHelper](symbols/function_TestBootParamsHelper.md)
