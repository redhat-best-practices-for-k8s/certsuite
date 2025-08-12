## Package operatingsystem (github.com/redhat-best-practices-for-k8s/certsuite/tests/platform/operatingsystem)



### Functions

- **GetRHCOSMappedVersions** — func(string)(map[string]string, error)
- **GetShortVersionFromLong** — func(string)(string, error)

### Globals


### Call graph (exported symbols, partial)

```mermaid
graph LR
  GetRHCOSMappedVersions --> make
  GetRHCOSMappedVersions --> Split
  GetRHCOSMappedVersions --> TrimSpace
  GetRHCOSMappedVersions --> Split
  GetRHCOSMappedVersions --> TrimSpace
  GetRHCOSMappedVersions --> TrimSpace
  GetShortVersionFromLong --> GetRHCOSMappedVersions
```

### Symbol docs

- [function GetRHCOSMappedVersions](symbols/function_GetRHCOSMappedVersions.md)
- [function GetShortVersionFromLong](symbols/function_GetShortVersionFromLong.md)
