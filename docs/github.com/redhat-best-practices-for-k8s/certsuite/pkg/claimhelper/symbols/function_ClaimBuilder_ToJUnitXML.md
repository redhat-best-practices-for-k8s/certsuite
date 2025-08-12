ClaimBuilder.ToJUnitXML`

### Purpose
`ToJUnitXML` serialises a test‑result claim into the JUnit XML format and writes it to disk.  
It is used by the test harness when a claim has been built with `ClaimBuilder`, allowing the result to be consumed by CI tools that understand JUnit reports.

```go
// ToJUnitXML creates a JUnit‑compatible file from the current ClaimBuilder.
// The caller supplies:
//   * name      – base filename (e.g. “my‑test”)
//   * startTime – timestamp when the test started
//   * endTime   – timestamp when the test finished
//
// It returns a function that, when invoked, performs the write and logs
// progress/errors via the embedded logger in ClaimBuilder.
func (cb ClaimBuilder) ToJUnitXML(name string, startTime, endTime time.Time) func()
```

### Inputs

| Parameter | Type      | Description |
|-----------|-----------|-------------|
| `name`    | `string`  | Base name for the output file. The method appends `CNFFeatureValidationJunitXMLFileName` (a constant `"cnf-feature-validation-junit.xml"`) to produce the final filename. |
| `startTime` | `time.Time` | When the test began; used in the JUnit `<testcase>` element’s `timestamp`. |
| `endTime`   | `time.Time` | When the test finished; used to compute elapsed time for reporting. |

### Output
A **zero‑argument closure** (`func()`) that performs the following steps when called:

1. **Populate XML structure** – Calls `populateXMLFromClaim(cb.claim, startTime, endTime)` to transform the internal claim representation into a struct suitable for JUnit marshaling.
2. **Marshal to pretty JSON** – Uses `xml.MarshalIndent` (actually `encoding/xml`) to generate indented XML bytes from that struct.
3. **Write file** – Persists the bytes to disk under `<name>CNFFeatureValidationJunitXMLFileName</name>` with permissions set by the package‑private `claimFilePermissions`.
4. **Logging** – Uses `cb.logger.Info` to log a successful write and `cb.logger.Fatal` to abort on any error (file creation, marshaling, or writing).

If an error occurs at any point, the closure calls `Fatal`, terminating the process with an appropriate message.

### Dependencies & Side‑Effects

| Dependency | Role |
|------------|------|
| `populateXMLFromClaim` | Converts the claim into a JUnit struct. |
| `xml.MarshalIndent` | Serialises that struct to XML bytes. |
| `WriteFile` | Writes the XML to disk. |
| `logger.Info / logger.Fatal` | Emits logs and handles unrecoverable errors. |

The method itself is pure; all side‑effects happen only when the returned closure is executed.

### Package Context

- **Package**: `claimhelper`
- **Related constants**  
  - `CNFFeatureValidationJunitXMLFileName` – filename suffix.  
  - `CNFFeatureValidationReportKey` – key used in claim data (not directly in this function).  
  - `DateTimeFormatDirective`, `TestStateFailed`, `TestStateSkipped` – formatting and state helpers for the XML payload.

- **Typical flow**:  
  1. A test creates a `ClaimBuilder`.  
  2. The test populates the builder with assertions (`Assert...`).  
  3. After execution, it calls `builder.ToJUnitXML("test-name", start, end)()` to persist the result.

The function bridges the internal claim representation and external CI reporting by producing a standard JUnit XML file that can be consumed by tools such as Jenkins or GitHub Actions.
