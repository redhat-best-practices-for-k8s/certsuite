GetNewClientsHolder` – Factory for a Clients Holder

**File:** `internal/clientsholder/clientsholder.go`  
**Package:** `clientsholder`

---

## Purpose
Creates and returns a new instance of **`*ClientsHolder`**, which is the central runtime container that holds all client objects used by CertSuite (e.g., Kubernetes, OpenShift, OIDC).  
The function also performs an early sanity check to ensure that the holder was created correctly; if not, it terminates the program with a fatal error.

---

## Signature
```go
func GetNewClientsHolder(name string) *ClientsHolder
```

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `name`    | `string` | A human‑readable identifier for the holder. It is stored inside the resulting struct for debugging/logging purposes. |

**Return value**

- `*ClientsHolder`: a fully initialised object ready to be used by callers.

---

## Key Dependencies

| Called Function | Purpose |
|-----------------|---------|
| `newClientsHolder(name string)` | Constructs a new `ClientsHolder` instance (internal helper). |
| `Fatal(message string, args ...interface{})` | From the package’s own logging/exit wrapper; prints an error and exits if holder creation fails. |

*No external packages are invoked directly by this function.*

---

## Side Effects

1. **Logging / Exit**  
   - If `newClientsHolder` returns `nil`, `Fatal` is called, causing the program to log a critical message and terminate immediately.
2. **State Change**  
   - None beyond returning the new holder; the function does not modify global state.

---

## How It Fits in the Package

The *clientsholder* package encapsulates all logic for creating, configuring, and accessing runtime clients. `GetNewClientsHolder` is the public entry point used by higher‑level modules (e.g., command line tools or test harnesses) to obtain a fresh holder instance.  
Once returned, callers can:

- Access the embedded clients via the holder’s fields/methods.
- Store the holder in package‑wide variables if needed for shared use.

```go
holder := clientsholder.GetNewClientsHolder("certsuite")
```

This pattern centralises error handling and ensures that every part of CertSuite works with a consistently initialised set of clients.
