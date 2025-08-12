SkipHelmChartList`

```go
type SkipHelmChartList struct {
    Name string // Identifier of the Helm chart that should be excluded.
}
```

### Purpose

`SkipHelmChartList` represents a single entry in a configuration file that tells CertSuite to **ignore** (skip) a specific Helm chart during its scanning process.  
The package `configuration` is responsible for parsing YAML/JSON config files that control which resources are evaluated. This struct is the building block for the *“skip‑helm‑chart”* list.

### Fields

| Field | Type   | Description |
|-------|--------|-------------|
| `Name` | `string` | The exact chart name (or a pattern) to exclude from scanning. |

> **Note**: Despite its plural name, each instance contains only one chart identifier. A slice of these structs is usually used in the top‑level configuration.

### Usage Flow

1. **Load Config** – The package reads a YAML/JSON file into a struct that includes a field like `SkipHelmChartList []configuration.SkipHelmChartList`.
2. **Populate List** – For every entry under the `skip-helm-chart` key, a new `SkipHelmChartList` is created with the chart name.
3. **Scan Decision** – When CertSuite iterates over discovered Helm charts, it checks whether the chart’s name matches any `Name` in this list. If so, the chart is skipped.

```
mermaid
graph TD;
    ConfigFile -->|parses to| SkipHelmChartList[];
    ScanProcess -->|checks| SkipHelmChartList[];
```

### Dependencies & Side‑Effects

- **Dependencies**: None beyond the Go standard library. It is a plain data holder.
- **Side‑Effects**: Instantiating this struct has no side effects. Its presence in the configuration file directly influences which charts are omitted from scanning.

### Integration into the Package

`SkipHelmChartList` lives inside `github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration`.  
It is part of the public API, allowing users to extend or modify the configuration format without touching the core logic. The struct itself does not perform any logic; it simply conveys user intent (skip these charts) to other components of CertSuite.

--- 

**Bottom line:**  
Use `SkipHelmChartList` when you want CertSuite to ignore certain Helm charts—add an entry with the chart’s name under the `skip-helm-chart` section of your configuration file.
