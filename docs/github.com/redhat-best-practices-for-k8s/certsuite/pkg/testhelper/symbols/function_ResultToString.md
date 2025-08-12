ResultToString`

```go
func ResultToString(result int) string
```

### Purpose  
`ResultToString` translates a numeric test‑result code into its human‑readable string form.  
It is used by the testing framework to convert internal integer constants (`SUCCESS`, `FAILURE`, `ERROR`) into a format that can be logged, reported or displayed in UI tools.

### Parameters  

| Name   | Type | Description |
|--------|------|-------------|
| `result` | `int` | The numeric result code produced by a test case. |

> **Note**: The function accepts any integer; only the predefined constants are recognised.

### Return Value  

* `string`:  
  * `"SUCCESS"` when `result == SUCCESS`.  
  * `"FAILURE"` when `result == FAILURE`.  
  * `"ERROR"`   when `result == ERROR`.  
  * An empty string (`""`) for any other value.

### Dependencies

| Dependency | Role |
|------------|------|
| `SUCCESS`, `FAILURE`, `ERROR` constants | Integer values that represent the three possible test outcomes. These are defined in the same file (lines ~30‑33). |

No external packages or global variables are accessed, so the function is pure and thread‑safe.

### Side Effects  

None – it only reads its argument and returns a value. It does **not** modify any global state or perform I/O.

### Usage Context  

`ResultToString` lives in the `testhelper` package, which contains helper utilities for running tests against Kubernetes clusters.  
Typical usage:

```go
status := runTest(...)
fmt.Println("Test status:", ResultToString(status))
```

The function is called whenever a numeric test result needs to be presented as text, such as in logs or test reports.

### Summary  

* **Input**: an `int` representing a test outcome.  
* **Output**: a string describing that outcome (`"SUCCESS"`, `"FAILURE"`, `"ERROR"` or empty).  
* **No side effects** – pure function.  
* **Key dependency**: the three integer constants defined in the same package.

This helper keeps code that consumes test results clean and maintainable by centralising the mapping logic.
