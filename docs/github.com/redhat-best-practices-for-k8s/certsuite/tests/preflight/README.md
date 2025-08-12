## Package preflight (github.com/redhat-best-practices-for-k8s/certsuite/tests/preflight)



### Functions

- **LoadChecks** — func()()
- **ShouldRun** — func(string)(bool)

### Globals


### Call graph (exported symbols, partial)

```mermaid
graph LR
  LoadChecks --> Debug
  LoadChecks --> GetTestEnvironment
  LoadChecks --> WithBeforeEachFn
  LoadChecks --> NewChecksGroup
  LoadChecks --> testPreflightContainers
  LoadChecks --> IsOCPCluster
  LoadChecks --> Info
  LoadChecks --> testPreflightOperators
  ShouldRun --> GetTestEnvironment
  ShouldRun --> labelsAllowTestRun
  ShouldRun --> GetTestParameters
  ShouldRun --> Warn
```

### Symbol docs

- [function LoadChecks](symbols/function_LoadChecks.md)
- [function ShouldRun](symbols/function_ShouldRun.md)
