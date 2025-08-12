SanitizeClaimFile`

**Package:** `github.com/redhat-best-practices-for-k8s/certsuite/pkg/claimhelper`  
**Signature**

```go
func SanitizeClaimFile(claimPath string, outputDir string) (string, error)
```

---

### Purpose
`SanitizeClaimFile` cleans and normalises a claim file before it is stored in the repository.  
It:

1. Reads an existing claim JSON/YAML file from `claimPath`.  
2. Parses the claim into an internal representation (`UnmarshalClaim`).  
3. Evaluates label expressions against test‑IDs to remove or keep only the relevant labels.  
4. Writes a sanitized version of the claim back to disk in `outputDir` (or overwrites the original if no output dir is supplied).  
5. Returns the path to the new file and any error that occurred.

The function is used by CLI commands that import user‑supplied claims, ensuring that only permitted fields remain before persisting them.

---

### Parameters

| Name | Type   | Description |
|------|--------|-------------|
| `claimPath` | `string` | Filesystem path to the claim file (JSON or YAML). |
| `outputDir` | `string` | Directory where the sanitized claim will be written. If empty, the original file is overwritten. |

---

### Returns

| Value | Type   | Description |
|-------|--------|-------------|
| `string` | Path of the newly created/overwritten claim file. |
| `error`  | Error if any step fails (file I/O, unmarshalling, label evaluation, etc.). |

---

### Key Dependencies & Calls

| Called Function | Purpose in this context |
|-----------------|-------------------------|
| `Info`, `Error` | Logging via the package’s logger. |
| `ReadClaimFile` | Reads raw bytes from disk. |
| `UnmarshalClaim` | Decodes JSON/YAML into a claim struct. |
| `NewLabelsExprEvaluator` | Builds an evaluator for label expressions in the claim. |
| `GetTestIDAndLabels` | Retrieves test ID and associated labels from the claim. |
| `Eval` | Applies label expression logic to determine which labels should be retained. |
| `delete` (built‑in) | Removes unneeded keys from the claim map. |
| `MarshalClaimOutput` | Serialises the sanitized claim back into JSON/YAML. |
| `WriteClaimOutput` | Writes the serialised data to `outputDir`. |

---

### Side Effects

- **File System**: Overwrites or creates a file in `outputDir` (or original location).  
- **Logging**: Emits informational and error messages.  
- **No global state mutation** beyond the file write.

---

### How It Fits the Package

The `claimhelper` package provides utilities for handling *claims* – structured representations of test results.  
`SanitizeClaimFile` is a core helper that prepares user‑supplied claim files for ingestion:

1. Ensures they conform to the internal schema (`UnmarshalClaim`).  
2. Removes extraneous or potentially unsafe data via label expression evaluation.  
3. Persists a clean, repository‑ready copy.

It is invoked by higher‑level CLI commands such as `claim import` and by automated pipelines that consume claims before storage.
