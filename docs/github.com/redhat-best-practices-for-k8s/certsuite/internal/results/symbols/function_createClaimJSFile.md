createClaimJSFile`

| Aspect | Details |
|--------|---------|
| **Purpose** | Generate a JavaScript file (`claimjson.js`) that embeds the raw JSON of a claim for consumption by the HTML result viewer. |
| **Signature** | `func createClaimJSFile(claimJSONPath, outputDir string) (string, error)` |
| **Inputs** | * `claimJSONPath` – absolute or relative path to the source `claim.json`. <br>* `outputDir` – directory where the generated JS file should be written. |
| **Outputs** | * The full path of the created JS file.<br>* An error if any step fails. |
| **Key Steps** | 1. Read the claim JSON from `claimJSONPath`. <br>2. Construct a JS module string that assigns the parsed JSON to a global variable (see `jsClaimVarFileName`). <br>3. Write this string to `<outputDir>/claimjson.js` with `writeFilePerms` permissions. |
| **Dependencies** | • `ioutil.ReadFile` – reads source JSON.<br>• `filepath.Join` – builds the output path.<br>• `os.WriteFile` – writes the JS file.<br>• Error formatting via `fmt.Errorf`. |
| **Side‑effects** | Creates/overwrites a file on disk. No in‑memory state is modified outside the returned path. |
| **Error handling** | Returns wrapped errors: <br>`cannot read claim JSON`, `cannot write claim JS` – each includes the underlying error for debugging. |
| **Package role** | Part of the `results` internal package that prepares assets for the HTML test‑report viewer. This function is called during result packaging to ensure the JavaScript consumer has access to the claim data without needing an extra network request. |

### Usage Flow

```
claimJSON := "/tmp/claim.json"
outDir    := "./html"

jsPath, err := createClaimJSFile(claimJSON, outDir)
// jsPath == "./html/claimjson.js"
// The file now contains:
   // export const claim = <parsed JSON>;
```

### Mermaid diagram (optional)

```mermaid
flowchart TD
  A[Input: claim.json] --> B[ReadFile]
  B --> C{Success?}
  C -- Yes --> D[Format JS string]
  D --> E[WriteFile to outputDir/claimjson.js]
  E --> F[Return path]
  C -- No --> G[Error "cannot read claim JSON"]
```

> **Note**: The function is unexported, intended for internal use only. It relies on the package‑level constant `jsClaimVarFileName` (typically `"claim.json"`), but this name is not directly referenced in the code snippet provided; it would be used when composing the JS string.

---

This concise documentation explains what `createClaimJSFile` does, how it interacts with other parts of the `results` package, and its impact on the filesystem.
