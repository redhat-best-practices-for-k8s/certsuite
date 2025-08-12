## Package hugepages (github.com/redhat-best-practices-for-k8s/certsuite/tests/platform/hugepages)



### Structs

- **Tester** (exported) — 5 fields, 5 methods

### Functions

- **NewTester** — func(*provider.Node, *corev1.Pod, clientsholder.Command)(*Tester, error)
- **Tester.HasMcSystemdHugepagesUnits** — func()(bool)
- **Tester.Run** — func()(error)
- **Tester.TestNodeHugepagesWithKernelArgs** — func()(bool, error)
- **Tester.TestNodeHugepagesWithMcSystemd** — func()(bool, error)
- **hugepagesByNuma.String** — func()(string)

### Call graph (exported symbols, partial)

```mermaid
graph LR
  NewTester --> NewContext
  NewTester --> Info
  NewTester --> getNodeNumaHugePages
  NewTester --> Errorf
  NewTester --> Info
  NewTester --> getMcSystemdUnitsHugepagesConfig
  NewTester --> Errorf
  Tester_HasMcSystemdHugepagesUnits --> len
  Tester_Run --> HasMcSystemdHugepagesUnits
  Tester_Run --> Info
  Tester_Run --> TestNodeHugepagesWithMcSystemd
  Tester_Run --> Errorf
  Tester_Run --> Info
  Tester_Run --> TestNodeHugepagesWithKernelArgs
  Tester_Run --> Errorf
  Tester_TestNodeHugepagesWithKernelArgs --> getMcHugepagesFromMcKernelArguments
  Tester_TestNodeHugepagesWithKernelArgs --> Errorf
  Tester_TestNodeHugepagesWithKernelArgs --> Errorf
  Tester_TestNodeHugepagesWithKernelArgs --> Info
  Tester_TestNodeHugepagesWithKernelArgs --> Errorf
  Tester_TestNodeHugepagesWithMcSystemd --> Warn
  Tester_TestNodeHugepagesWithMcSystemd --> Errorf
  Tester_TestNodeHugepagesWithMcSystemd --> Errorf
  Tester_TestNodeHugepagesWithMcSystemd --> Errorf
  Tester_TestNodeHugepagesWithMcSystemd --> Errorf
  Tester_TestNodeHugepagesWithMcSystemd --> Errorf
  hugepagesByNuma_String --> append
  hugepagesByNuma_String --> Ints
  hugepagesByNuma_String --> WriteString
  hugepagesByNuma_String --> Sprintf
  hugepagesByNuma_String --> WriteString
  hugepagesByNuma_String --> Sprintf
  hugepagesByNuma_String --> String
```

### Symbol docs

- [struct Tester](symbols/struct_Tester.md)
- [function NewTester](symbols/function_NewTester.md)
- [function Tester.HasMcSystemdHugepagesUnits](symbols/function_Tester_HasMcSystemdHugepagesUnits.md)
- [function Tester.Run](symbols/function_Tester_Run.md)
- [function Tester.TestNodeHugepagesWithKernelArgs](symbols/function_Tester_TestNodeHugepagesWithKernelArgs.md)
- [function Tester.TestNodeHugepagesWithMcSystemd](symbols/function_Tester_TestNodeHugepagesWithMcSystemd.md)
- [function hugepagesByNuma.String](symbols/function_hugepagesByNuma_String.md)
