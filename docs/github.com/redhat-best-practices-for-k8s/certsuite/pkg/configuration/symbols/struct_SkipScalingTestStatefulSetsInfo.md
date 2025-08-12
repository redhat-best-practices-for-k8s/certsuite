SkipScalingTestStatefulSetsInfo`

| Item | Details |
|------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration` |
| **File / Line** | `configuration.go:46` |
| **Exported** | ✅ |

### Purpose
`SkipScalingTestStatefulSetsInfo` is a lightweight value type used by the *scaling* test suite to identify StatefulSet resources that should be excluded from scaling operations.  
During a scaling test, the framework iterates over all `statefulsets` in the cluster and attempts to change their replica count. Some StatefulSets are known to break or behave unpredictably when scaled (e.g., due to init containers, volume claim handling, or application‑specific constraints).  The struct allows users to configure those exclusions declaratively.

### Fields
| Field | Type   | Role |
|-------|--------|------|
| `Name`      | `string` | The exact name of the StatefulSet. |
| `Namespace` | `string` | The namespace in which that StatefulSet resides. |

Both fields are required for an exclusion rule; a missing value results in the entry being ignored by the test logic.

### How It Is Used
1. **Configuration Loading**  
   The configuration package reads a YAML/JSON file (or other source) containing a slice of `SkipScalingTestStatefulSetsInfo`.  
2. **Lookup During Tests**  
   When the scaling test starts, it builds a set of keys in the form `<namespace>/<name>` from this slice.  
3. **Decision Point**  
   For each StatefulSet encountered, the test checks if its key is present in the skip set. If so, that instance is skipped and no scaling operation is performed.

```mermaid
graph TD
  ConfigFile[Configuration File] -->|Parse| SkipList[([]SkipScalingTestStatefulSetsInfo)]
  SkipList -->|Build Set| SkipSet{<ns>/<name> keys}
  TestRunner((Scaling Tests)) -->|Iterate StatefulSets| EachSS(SS)
  EachSS -->|Check SkipSet| ShouldSkip
  ShouldSkip -- Yes --> SkipAction[Do not scale]
  ShouldSkip -- No --> ScaleAction[Attempt scaling]
```

### Dependencies & Side‑Effects
- **Dependencies**  
  * The struct itself has no runtime dependencies.  
  * It is consumed by the test harness (`pkg/tests/scaling_test.go` or similar) which imports this package for configuration data.
- **Side‑Effects**  
  * No side effects are inherent to the struct; it merely holds data.  
  * The impact of using a `SkipScalingTestStatefulSetsInfo` entry is that the corresponding StatefulSet will not be modified during scaling tests, preserving its state.

### Integration in the Package
The `configuration` package aggregates various test configuration objects. `SkipScalingTestStatefulSetsInfo` lives alongside other structs such as `TestConfig`, `FeatureGateConfig`, etc., providing a unified source of truth for how tests should behave in a given cluster environment.  

By exposing this struct publicly, users can extend the skip list without modifying core test code—just add entries to the configuration file. This keeps the test logic decoupled from specific StatefulSet names or namespaces.

--- 

**TL;DR:**  
`SkipScalingTestStatefulSetsInfo` is a declarative entry used by scaling tests to whitelist StatefulSets that should *not* be scaled, preventing potential disruptions in known problematic workloads.
