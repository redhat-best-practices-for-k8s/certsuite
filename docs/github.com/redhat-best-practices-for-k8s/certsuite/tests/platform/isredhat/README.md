## Package isredhat (github.com/redhat-best-practices-for-k8s/certsuite/tests/platform/isredhat)



### Structs

- **BaseImageInfo** (exported) — 2 fields, 2 methods

### Functions

- **BaseImageInfo.TestContainerIsRedHatRelease** — func()(bool, error)
- **IsRHEL** — func(string)(bool)
- **NewBaseImageTester** — func(clientsholder.Command, clientsholder.Context)(*BaseImageInfo)

### Call graph (exported symbols, partial)

```mermaid
graph LR
  BaseImageInfo_TestContainerIsRedHatRelease --> runCommand
  BaseImageInfo_TestContainerIsRedHatRelease --> Info
  BaseImageInfo_TestContainerIsRedHatRelease --> IsRHEL
  IsRHEL --> MustCompile
  IsRHEL --> FindAllString
  IsRHEL --> len
  IsRHEL --> Info
  IsRHEL --> MustCompile
  IsRHEL --> FindAllString
  IsRHEL --> len
```

### Symbol docs

- [struct BaseImageInfo](symbols/struct_BaseImageInfo.md)
- [function BaseImageInfo.TestContainerIsRedHatRelease](symbols/function_BaseImageInfo_TestContainerIsRedHatRelease.md)
- [function IsRHEL](symbols/function_IsRHEL.md)
- [function NewBaseImageTester](symbols/function_NewBaseImageTester.md)
