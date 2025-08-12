## Package resources (github.com/redhat-best-practices-for-k8s/certsuite/tests/accesscontrol/resources)



### Functions

- **HasExclusiveCPUsAssigned** — func(*provider.Container, *log.Logger)(bool)
- **HasRequestsSet** — func(*provider.Container, *log.Logger)(bool)

### Call graph (exported symbols, partial)

```mermaid
graph LR
  HasExclusiveCPUsAssigned --> Cpu
  HasExclusiveCPUsAssigned --> Memory
  HasExclusiveCPUsAssigned --> IsZero
  HasExclusiveCPUsAssigned --> IsZero
  HasExclusiveCPUsAssigned --> Debug
  HasExclusiveCPUsAssigned --> AsInt64
  HasExclusiveCPUsAssigned --> Debug
  HasExclusiveCPUsAssigned --> AsInt64
  HasRequestsSet --> len
  HasRequestsSet --> Error
  HasRequestsSet --> IsZero
  HasRequestsSet --> Cpu
  HasRequestsSet --> Error
  HasRequestsSet --> IsZero
  HasRequestsSet --> Memory
  HasRequestsSet --> Error
```

### Symbol docs

- [function HasExclusiveCPUsAssigned](symbols/function_HasExclusiveCPUsAssigned.md)
- [function HasRequestsSet](symbols/function_HasRequestsSet.md)
