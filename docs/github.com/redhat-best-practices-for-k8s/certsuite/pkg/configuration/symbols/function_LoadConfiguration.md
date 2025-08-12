LoadConfiguration`

**Location:**  
`pkg/configuration/utils.go:34-49`

### Purpose
`LoadConfiguration` is the single point of entry for reading the test suite’s configuration file.  
* It loads the JSON/YAML (the format accepted by `encoding/json.Unmarshal`) once per program run.  
* Subsequent calls return the cached result to avoid repeated I/O and parsing.

### Signature
```go
func LoadConfiguration(path string) (TestConfiguration, error)
```

| Parameter | Type   | Description                      |
|-----------|--------|----------------------------------|
| `path`    | `string` | Filesystem path of the configuration file. |

| Return | Type              | Description                                 |
|--------|-------------------|---------------------------------------------|
| 1st    | `TestConfiguration` | The fully‑deserialized configuration object. |
| 2nd    | `error`           | Non‑nil if reading or unmarshalling failed. |

### Key Dependencies
| Dependency | Role in the function |
|------------|----------------------|
| `Debug`, `Info`, `Warn` | Logging helpers from the package’s logger (likely a wrapper around `logrus`). Used to trace progress and errors. |
| `ReadFile` | Reads the file contents into memory (`io/ioutil.ReadFile`). |
| `Unmarshal` | Deserialises JSON/YAML into the `TestConfiguration` struct. |

### Internal State
The function relies on three unexported package variables:

| Variable   | Initial Value | Usage |
|------------|---------------|-------|
| `configuration` | `{}` | Holds the parsed configuration once loaded. |
| `confLoaded` | `false` | Flag indicating whether a load has already occurred. |
| `parameters` | `{}` | (Unused in this snippet; likely holds CLI or environment parameters). |

The function checks `confLoaded`; if true it simply returns the cached `configuration`.  
Otherwise it reads the file, unmarshals into a temporary struct, copies relevant fields into `configuration`, sets `confLoaded = true`, and then returns.

### Side‑Effects
* **File I/O** – Reads the specified path once.
* **Global mutation** – On first call, writes to `configuration` and flips `confLoaded`.
* **Logging** – Emits debug/info/warn messages at key stages.

### Package Context
Within `github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration`, this helper centralises configuration handling. Other components (tests, runners, CLI) call `LoadConfiguration` to obtain settings such as test suites to run, output paths, or Kubernetes cluster details. Because the state is cached, repeated invocations are cheap and deterministic.

### Suggested Mermaid Flow
```mermaid
flowchart TD
    A[Call LoadConfiguration(path)] -->|confLoaded?=true?| B{Return cached}
    B -->|yes| C[Return configuration]
    B -->|no| D[ReadFile(path)]
    D --> E[Unmarshal(data, &tempConf)]
    E --> F[Copy tempConf to configuration]
    F --> G[Set confLoaded=true]
    G --> C
```

This diagram highlights the one‑time load logic and the early exit on subsequent calls.
