## Package claimhelper (github.com/redhat-best-practices-for-k8s/certsuite/pkg/claimhelper)



### Structs

- **ClaimBuilder** (exported) — 1 fields, 3 methods
- **FailureMessage** (exported) — 3 fields, 0 methods
- **SkippedMessage** (exported) — 2 fields, 0 methods
- **TestCase** (exported) — 8 fields, 0 methods
- **TestSuitesXML** (exported) — 8 fields, 0 methods
- **Testsuite** (exported) — 12 fields, 0 methods

### Functions

- **ClaimBuilder.Build** — func(string)()
- **ClaimBuilder.Reset** — func()()
- **ClaimBuilder.ToJUnitXML** — func(string, time.Time, time.Time)()
- **CreateClaimRoot** — func()(*claim.Root)
- **GenerateNodes** — func()(map[string]interface{})
- **GetConfigurationFromClaimFile** — func(string)(*provider.TestEnvironment, error)
- **MarshalClaimOutput** — func(*claim.Root)([]byte)
- **MarshalConfigurations** — func(*provider.TestEnvironment)([]byte, error)
- **NewClaimBuilder** — func(*provider.TestEnvironment)(*ClaimBuilder, error)
- **ReadClaimFile** — func(string)([]byte, error)
- **SanitizeClaimFile** — func(string, string)(string, error)
- **UnmarshalClaim** — func([]byte, *claim.Root)()
- **UnmarshalConfigurations** — func([]byte, map[string]interface{})()
- **WriteClaimOutput** — func(string, []byte)()

### Call graph (exported symbols, partial)

```mermaid
graph LR
  ClaimBuilder_Build --> Now
  ClaimBuilder_Build --> Format
  ClaimBuilder_Build --> UTC
  ClaimBuilder_Build --> GetReconciledResults
  ClaimBuilder_Build --> MarshalClaimOutput
  ClaimBuilder_Build --> WriteClaimOutput
  ClaimBuilder_Build --> Info
  ClaimBuilder_Reset --> Format
  ClaimBuilder_Reset --> UTC
  ClaimBuilder_Reset --> Now
  ClaimBuilder_ToJUnitXML --> populateXMLFromClaim
  ClaimBuilder_ToJUnitXML --> MarshalIndent
  ClaimBuilder_ToJUnitXML --> Fatal
  ClaimBuilder_ToJUnitXML --> Info
  ClaimBuilder_ToJUnitXML --> WriteFile
  ClaimBuilder_ToJUnitXML --> Fatal
  CreateClaimRoot --> Now
  CreateClaimRoot --> Format
  CreateClaimRoot --> UTC
  GenerateNodes --> GetNodeJSON
  GenerateNodes --> GetCniPlugins
  GenerateNodes --> GetHwInfoAllNodes
  GenerateNodes --> GetCsiDriver
  GetConfigurationFromClaimFile --> ReadClaimFile
  GetConfigurationFromClaimFile --> Error
  GetConfigurationFromClaimFile --> Printf
  GetConfigurationFromClaimFile --> UnmarshalClaim
  GetConfigurationFromClaimFile --> Marshal
  GetConfigurationFromClaimFile --> Errorf
  GetConfigurationFromClaimFile --> Unmarshal
  MarshalClaimOutput --> MarshalIndent
  MarshalClaimOutput --> Fatal
  MarshalConfigurations --> GetTestEnvironment
  MarshalConfigurations --> Marshal
  MarshalConfigurations --> Error
  NewClaimBuilder --> Getenv
  NewClaimBuilder --> CreateClaimRoot
  NewClaimBuilder --> Debug
  NewClaimBuilder --> MarshalConfigurations
  NewClaimBuilder --> Errorf
  NewClaimBuilder --> UnmarshalConfigurations
  NewClaimBuilder --> CreateClaimRoot
  NewClaimBuilder --> GenerateNodes
  ReadClaimFile --> ReadFile
  ReadClaimFile --> Error
  ReadClaimFile --> Info
  SanitizeClaimFile --> Info
  SanitizeClaimFile --> ReadClaimFile
  SanitizeClaimFile --> Error
  SanitizeClaimFile --> UnmarshalClaim
  SanitizeClaimFile --> NewLabelsExprEvaluator
  SanitizeClaimFile --> Error
  SanitizeClaimFile --> GetTestIDAndLabels
  SanitizeClaimFile --> Eval
  UnmarshalClaim --> Unmarshal
  UnmarshalClaim --> Fatal
  UnmarshalConfigurations --> Unmarshal
  UnmarshalConfigurations --> Fatal
  WriteClaimOutput --> Info
  WriteClaimOutput --> WriteFile
  WriteClaimOutput --> Fatal
  WriteClaimOutput --> string
```

### Symbol docs

- [struct ClaimBuilder](symbols/struct_ClaimBuilder.md)
- [struct FailureMessage](symbols/struct_FailureMessage.md)
- [struct SkippedMessage](symbols/struct_SkippedMessage.md)
- [struct TestCase](symbols/struct_TestCase.md)
- [struct TestSuitesXML](symbols/struct_TestSuitesXML.md)
- [struct Testsuite](symbols/struct_Testsuite.md)
- [function ClaimBuilder.Build](symbols/function_ClaimBuilder_Build.md)
- [function ClaimBuilder.Reset](symbols/function_ClaimBuilder_Reset.md)
- [function ClaimBuilder.ToJUnitXML](symbols/function_ClaimBuilder_ToJUnitXML.md)
- [function CreateClaimRoot](symbols/function_CreateClaimRoot.md)
- [function GenerateNodes](symbols/function_GenerateNodes.md)
- [function GetConfigurationFromClaimFile](symbols/function_GetConfigurationFromClaimFile.md)
- [function MarshalClaimOutput](symbols/function_MarshalClaimOutput.md)
- [function MarshalConfigurations](symbols/function_MarshalConfigurations.md)
- [function NewClaimBuilder](symbols/function_NewClaimBuilder.md)
- [function ReadClaimFile](symbols/function_ReadClaimFile.md)
- [function SanitizeClaimFile](symbols/function_SanitizeClaimFile.md)
- [function UnmarshalClaim](symbols/function_UnmarshalClaim.md)
- [function UnmarshalConfigurations](symbols/function_UnmarshalConfigurations.md)
- [function WriteClaimOutput](symbols/function_WriteClaimOutput.md)
