FailureReasonOutTestString`

```go
func FailureReasonOutTestString(f FailureReasonOut) string
```

### Purpose  
`FailureReasonOutTestString` converts a `FailureReasonOut` value into a human‑readable test string.  
The function is used by the test suite to generate deterministic output when a failure reason
is reported, so that test results can be compared against expected snapshots.

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `f`  | `FailureReasonOut` | The struct instance containing the failure details. |

> **Note**: The definition of `FailureReasonOut` is in the same package but not shown here; it
> typically holds fields such as a message, reason code, and optional metadata.

### Return value

| Type | Description |
|------|-------------|
| `string` | A formatted string that represents the contents of `f`. |

The string follows the pattern:

```
{FailureReasonOut: <ReportObjectTestStringPointer(f.Message)>, Reason: <ReportObjectTestStringPointer(f.Reason)>}
```

where `<ReportObjectTestStringPointer>` is a helper that safely dereferences pointer fields
and returns an empty string if the pointer is `nil`.

### Dependencies

| Dependency | Kind | Notes |
|------------|------|-------|
| `fmt.Sprintf` | function | Used twice to build the final string. |
| `ReportObjectTestStringPointer` | function | Handles optional pointer dereferencing for string fields. |

These helpers are defined elsewhere in `testhelper`.

### Side effects

* None. The function is pure; it only reads from its argument and returns a new string.

### Package context

`FailureReasonOutTestString` lives in the **testhelper** package, which provides
utilities for generating deterministic test output across the CertSuite project.
It is part of the public API (`exported: true`) so that other packages can call it when they need to log or compare failure reasons during tests.
