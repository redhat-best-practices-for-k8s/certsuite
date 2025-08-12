## Package certsuite (github.com/redhat-best-practices-for-k8s/certsuite/pkg/certsuite)



### Functions

- **LoadChecksDB** — func(string)()
- **LoadInternalChecksDB** — func()()
- **Run** — func(string, string)(error)
- **Shutdown** — func()()
- **Startup** — func()()

### Call graph (exported symbols, partial)

```mermaid
graph LR
  LoadChecksDB --> LoadInternalChecksDB
  LoadChecksDB --> ShouldRun
  LoadChecksDB --> LoadChecks
  LoadInternalChecksDB --> LoadChecks
  LoadInternalChecksDB --> LoadChecks
  LoadInternalChecksDB --> LoadChecks
  LoadInternalChecksDB --> LoadChecks
  LoadInternalChecksDB --> LoadChecks
  LoadInternalChecksDB --> LoadChecks
  LoadInternalChecksDB --> LoadChecks
  LoadInternalChecksDB --> LoadChecks
  Run --> GetTestParameters
  Run --> Println
  Run --> Print
  Run --> GetTestEnvironment
  Run --> Info
  Run --> Now
  Run --> RunChecks
  Run --> Error
  Shutdown --> CloseGlobalLogFile
  Shutdown --> Fprintf
  Shutdown --> Exit
  Startup --> GetTestParameters
  Startup --> InitLabelsExprEvaluator
  Startup --> Fprintf
  Startup --> Exit
  Startup --> CreateGlobalLogFile
  Startup --> Fprintf
  Startup --> Exit
  Startup --> Warn
```

### Symbol docs

- [function LoadChecksDB](symbols/function_LoadChecksDB.md)
- [function LoadInternalChecksDB](symbols/function_LoadInternalChecksDB.md)
- [function Run](symbols/function_Run.md)
- [function Shutdown](symbols/function_Shutdown.md)
- [function Startup](symbols/function_Startup.md)
