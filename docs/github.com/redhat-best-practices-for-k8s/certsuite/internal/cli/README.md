## Package cli (github.com/redhat-best-practices-for-k8s/certsuite/internal/cli)



### Structs

- **cliCheckLogSniffer**  — 0 fields, 1 methods

### Functions

- **LineAlignCenter** — func(string, int)(string)
- **LineAlignLeft** — func(string, int)(string)
- **LineColor** — func(string, string)(string)
- **PrintBanner** — func()()
- **PrintCheckAborted** — func(string, string)()
- **PrintCheckErrored** — func(string)()
- **PrintCheckFailed** — func(string)()
- **PrintCheckPassed** — func(string)()
- **PrintCheckRunning** — func(string)()
- **PrintCheckSkipped** — func(string, string)()
- **PrintResultsTable** — func(map[string][]int)()
- **WrapLines** — func(string, int)([]string)
- **cliCheckLogSniffer.Write** — func([]byte)(int, error)

### Globals

- **CliCheckLogSniffer**: 

### Call graph (exported symbols, partial)

```mermaid
graph LR
  LineAlignCenter --> Sprintf
  LineAlignCenter --> Sprintf
  LineAlignCenter --> len
  LineAlignLeft --> Sprintf
  PrintBanner --> Print
  PrintCheckAborted --> stopCheckLineGoroutine
  PrintCheckAborted --> Print
  PrintCheckErrored --> stopCheckLineGoroutine
  PrintCheckErrored --> Print
  PrintCheckFailed --> stopCheckLineGoroutine
  PrintCheckFailed --> Print
  PrintCheckPassed --> stopCheckLineGoroutine
  PrintCheckPassed --> Print
  PrintCheckRunning --> make
  PrintCheckRunning --> make
  PrintCheckRunning --> isTTY
  PrintCheckRunning --> Print
  PrintCheckRunning --> updateRunningCheckLine
  PrintCheckSkipped --> stopCheckLineGoroutine
  PrintCheckSkipped --> Print
  PrintResultsTable --> Printf
  PrintResultsTable --> Println
  PrintResultsTable --> Printf
  PrintResultsTable --> Println
  PrintResultsTable --> Printf
  PrintResultsTable --> Println
  PrintResultsTable --> Printf
  WrapLines --> Split
  WrapLines --> make
  WrapLines --> len
  WrapLines --> len
  WrapLines --> append
  WrapLines --> Fields
  WrapLines --> len
  WrapLines --> len
  cliCheckLogSniffer_Write --> isTTY
  cliCheckLogSniffer_Write --> len
  cliCheckLogSniffer_Write --> string
  cliCheckLogSniffer_Write --> len
```

### Symbol docs

- [function LineAlignCenter](symbols/function_LineAlignCenter.md)
- [function LineAlignLeft](symbols/function_LineAlignLeft.md)
- [function LineColor](symbols/function_LineColor.md)
- [function PrintBanner](symbols/function_PrintBanner.md)
- [function PrintCheckAborted](symbols/function_PrintCheckAborted.md)
- [function PrintCheckErrored](symbols/function_PrintCheckErrored.md)
- [function PrintCheckFailed](symbols/function_PrintCheckFailed.md)
- [function PrintCheckPassed](symbols/function_PrintCheckPassed.md)
- [function PrintCheckRunning](symbols/function_PrintCheckRunning.md)
- [function PrintCheckSkipped](symbols/function_PrintCheckSkipped.md)
- [function PrintResultsTable](symbols/function_PrintResultsTable.md)
- [function WrapLines](symbols/function_WrapLines.md)
- [function cliCheckLogSniffer.Write](symbols/function_cliCheckLogSniffer_Write.md)
