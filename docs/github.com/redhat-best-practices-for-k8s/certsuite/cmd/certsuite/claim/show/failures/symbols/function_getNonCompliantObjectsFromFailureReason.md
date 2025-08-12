getNonCompliantObjectsFromFailureReason`

| Aspect | Details |
|--------|---------|
| **Package** | `failures` (`github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/show/failures`) |
| **Signature** | `func getNonCompliantObjectsFromFailureReason(reason string) ([]NonCompliantObject, error)` |
| **Exported?** | No – internal helper used by the *show failures* command. |

---

### Purpose
`getNonCompliantObjectsFromFailureReason` extracts the list of objects that caused a claim to fail from a JSON‑encoded failure reason string.

The claim’s test case contains a `checkDetails` field which holds a JSON array of **NonCompliantObject** entries (each representing an object that broke compliance).  
This function:

1. Decodes that JSON into Go structs.
2. Validates the decoded slice is non‑empty.
3. Returns either the populated slice or an error.

---

### Inputs

| Parameter | Type | Description |
|-----------|------|-------------|
| `reason` | `string` | Raw text from the claim’s `checkDetails`. It is expected to be a JSON array string (e.g., `[{"kind":"Deployment","namespace":"foo", ...}]`). |

---

### Outputs

| Return | Type | Description |
|--------|------|-------------|
| `[]NonCompliantObject` | slice of structs defined elsewhere in the package | The parsed non‑compliant objects. |
| `error` | error | Non‑nil if: <br>• JSON unmarshalling fails.<br>• Resulting slice is empty (indicating an unexpected failure format). |

---

### Key Steps & Dependencies

1. **JSON Unmarshal**  
   ```go
   var nonCompliant []NonCompliantObject
   err := json.Unmarshal([]byte(reason), &nonCompliant)
   ```
   Uses the standard `encoding/json` package.

2. **Empty‑slice check**  
   If `len(nonCompliant) == 0`, an error is returned via `fmt.Errorf`. This guards against malformed or empty failure reasons.

3. **Return**  
   On success, the slice is returned with a nil error.

---

### Side Effects & Assumptions

- No global state is modified; the function is pure.
- Relies on the `NonCompliantObject` type definition being present in the same package (not shown in the snippet).
- The caller must provide a valid JSON string; otherwise, an error will propagate up to the command’s output logic.

---

### Role in the Package

The *show failures* command reads a claim file (`claimFilePathFlag`) and filters failed test cases.  
For each failure it calls `getNonCompliantObjectsFromFailureReason` to turn the raw reason text into structured data, which is then formatted according to the chosen output format (`outputFormatJSON`, `outputFormatText`).  

Thus this helper bridges raw claim data (strings) with the richer, typed representation needed for user‑facing output.
