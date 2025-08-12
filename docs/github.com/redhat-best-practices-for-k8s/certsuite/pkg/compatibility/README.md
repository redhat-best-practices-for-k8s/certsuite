## Package compatibility (github.com/redhat-best-practices-for-k8s/certsuite/pkg/compatibility)



### Structs

- **VersionInfo** (exported) — 5 fields, 0 methods

### Functions

- **BetaRHCOSVersionsFoundToMatch** — func(string, string)(bool)
- **DetermineOCPStatus** — func(string, time.Time)(string)
- **FindMajorMinor** — func(string)(string)
- **GetLifeCycleDates** — func()(map[string]VersionInfo)
- **IsRHCOSCompatible** — func(string, string)(bool)
- **IsRHELCompatible** — func(string, string)(bool)

### Globals


### Call graph (exported symbols, partial)

```mermaid
graph LR
  BetaRHCOSVersionsFoundToMatch --> FindMajorMinor
  BetaRHCOSVersionsFoundToMatch --> FindMajorMinor
  BetaRHCOSVersionsFoundToMatch --> StringInSlice
  BetaRHCOSVersionsFoundToMatch --> StringInSlice
  DetermineOCPStatus --> IsZero
  DetermineOCPStatus --> Split
  DetermineOCPStatus --> GetLifeCycleDates
  DetermineOCPStatus --> IsZero
  DetermineOCPStatus --> Before
  DetermineOCPStatus --> Equal
  DetermineOCPStatus --> After
  DetermineOCPStatus --> Before
  FindMajorMinor --> Split
  IsRHCOSCompatible --> BetaRHCOSVersionsFoundToMatch
  IsRHCOSCompatible --> FindMajorMinor
  IsRHCOSCompatible --> GetLifeCycleDates
  IsRHCOSCompatible --> NewVersion
  IsRHCOSCompatible --> Error
  IsRHCOSCompatible --> NewVersion
  IsRHCOSCompatible --> Error
  IsRHCOSCompatible --> GreaterThanOrEqual
  IsRHELCompatible --> GetLifeCycleDates
  IsRHELCompatible --> len
  IsRHELCompatible --> NewVersion
  IsRHELCompatible --> NewVersion
  IsRHELCompatible --> GreaterThanOrEqual
```

### Symbol docs

- [struct VersionInfo](symbols/struct_VersionInfo.md)
- [function BetaRHCOSVersionsFoundToMatch](symbols/function_BetaRHCOSVersionsFoundToMatch.md)
- [function DetermineOCPStatus](symbols/function_DetermineOCPStatus.md)
- [function FindMajorMinor](symbols/function_FindMajorMinor.md)
- [function GetLifeCycleDates](symbols/function_GetLifeCycleDates.md)
- [function IsRHCOSCompatible](symbols/function_IsRHCOSCompatible.md)
- [function IsRHELCompatible](symbols/function_IsRHELCompatible.md)
