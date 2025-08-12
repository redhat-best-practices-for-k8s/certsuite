OkNok.String() string`

| Item | Detail |
|------|--------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/tests/accesscontrol/securitycontextcontainer` |
| **Receiver type** | `OkNok` (defined elsewhere in the same package) |
| **Signature** | `func (o OkNok) String() string` |
| **Exported?** | Yes |

### Purpose

The method implements the `fmt.Stringer` interface for the `OkNok` type.  
It converts an `OkNok` value into its human‑readable representation so that the value can be printed with standard formatting functions (`fmt.Println`, `%s`, etc.).

### Inputs & Outputs

| Parameter | Description |
|-----------|-------------|
| `o OkNok` | The receiver instance whose state determines the string to return. |

**Return**

- A single string containing either `"OK"` or `"NOK"`.  
  - If the internal value of `OkNok` matches the exported constant `OK`, the method returns `OKString`.
  - Otherwise it returns `NOKString`.

The constants `OKString` and `NOKString` are defined near the top of the file:

```go
const (
    OK        OkNok = iota // ok value
    NOK                    // not‑ok value
)

const (
    OKString  = "OK"
    NOKString = "NOK"
)
```

### Key Dependencies

- **Constants** `OK`, `NOK`, `OKString`, `NOKString` – provide the mapping between numeric values and strings.
- No other functions or globals are referenced.

### Side Effects

The method is pure: it performs no writes, does not modify any global state, and only reads from its receiver.  
It can safely be called concurrently on different instances of `OkNok`.

### How It Fits the Package

`securitycontextcontainer` contains a series of predefined security context categories (e.g., `Category1`, `Category2`, …).  
Each category is associated with an `OkNok` value that indicates whether the category satisfies certain security requirements.  

When tests run, they often need to output human‑readable diagnostics.  The `String()` method allows a test to print an `OkNok` directly:

```go
fmt.Printf("Category1 check: %s\n", Category1)
```

Because `Category1` is of type `OkNok`, the call to `String()` yields `"OK"` or `"NOK"`.  
Thus, this method serves as a lightweight formatter that bridges internal status codes with user‑friendly output.
