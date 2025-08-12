InitCatalog` – Building the Test‑Case Catalog

#### 📌 Function Signature
```go
func InitCatalog() map[claim.Identifier]claim.TestCaseDescription
```

| Element | Description |
|---------|-------------|
| **Return type** | A mapping from a test identifier (`claim.Identifier`) to its full description (`claim.TestCaseDescription`). |

---

### 🔍 Purpose

`InitCatalog` is the *single source of truth* for all JUnit‑style test cases that belong to the `identifiers` package.  
When the test framework starts it calls this function once, receives a map and uses it to:

1. **Register** every available test case with its human‑readable description.
2. Make sure each identifier is linked to its impact statement (`ImpactMap`) and any associated documentation links.

The catalog drives the test discovery phase – without it the framework cannot know which tests exist or how they should be reported.

---

### ⚙️ Inputs & Outputs

| Input | Output |
|-------|--------|
| **None** | `map[claim.Identifier]claim.TestCaseDescription` – a read‑only map of all test cases. |

The function does *not* take any arguments; it builds the catalog from constants defined in this package.

---

### 📚 Key Dependencies

| Dependency | Role |
|------------|------|
| `AddCatalogEntry` | Helper that inserts an entry into the global `Catalog`. It is invoked for every test case. |
| **Constants** (`Test*Identifier`, `*Remediation`, `ImpactMap`) | Provide the identifiers, remediation strings and impact statements used to build each `TestCaseDescription`. |
| Global variables `Catalog` & `Classification` | The map returned by `InitCatalog` is stored in `Catalog`; the classification data structure is also initialized here. |

> **Note:**  
> `AddCatalogEntry` itself updates `Catalog`, but `InitCatalog` only calls it; all side‑effects are confined to that global map.

---

### 🧩 How It Fits into the Package

```
identifiers/
├─ identifiers.go   // holds constants, test identifiers and InitCatalog
└─ impact.go        // holds ImpactMap mapping IDs → impact strings
```

1. **`InitCatalog`** is called by the test harness during package initialization.
2. It populates `Catalog`, which the harness uses to:
   - Generate JUnit reports (`TestCaseDescription` objects contain name, description, tags, etc.).
   - Resolve impacts via `ImpactMap`.
3. The catalog remains read‑only thereafter; no mutation occurs after `InitCatalog` completes.

---

### 📦 Example Usage

```go
// In the test runner
catalog := identifiers.InitCatalog()
for id, desc := range catalog {
    fmt.Printf("Running %s: %s\n", id, desc.Name)
}
```

The above is a simplified illustration; in reality the harness consumes the map to build test suites and write XML output.

---

### 📌 Side Effects

- **Global mutation** – `Catalog` (and implicitly `Classification`) are written once during initialization.
- No external state changes or I/O occur inside `InitCatalog`; it purely constructs data structures.

---

### 💡 Summary

`InitCatalog` is the *catalog builder* for the Certsuite test framework.  
It aggregates all test identifiers, their descriptions, impacts and documentation links into a single, immutable map that powers discovery, execution, and reporting of all JUnit test cases in the `identifiers` package.
