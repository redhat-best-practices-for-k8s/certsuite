Check.SetAbortChan`

| | |
|-|-|
| **Package** | `checksdb` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb`) |
| **Receiver type** | `Check` (a struct that represents a single check definition) |
| **Signature** | `func(chan string)() ` |
| **Exported** | ✅ |

### Purpose

`SetAbortChan` attaches an *abort channel* to the current `Check`.  
When the check is executed, it can listen on this channel for a signal that
terminates its execution prematurely. The method returns a function that,
when called, closes the channel and signals any goroutine waiting on it.

This pattern allows callers (typically test harnesses or orchestrators) to:
1. Create an unbuffered `chan string`.
2. Pass it into `SetAbortChan` to register it with the check.
3. Keep a reference to the returned closure, which they can invoke when
   they want to abort the check (e.g., on timeout or cancellation).

### Inputs & Outputs

| Parameter | Type | Description |
|-----------|------|-------------|
| `abortCh` | `chan string` | The channel that will receive an abort signal. It is expected to be unbuffered and of type `string`, but the method does not enforce this; it merely stores the reference. |

| Return value | Type | Description |
|--------------|------|-------------|
| `func()` | Closure that closes `abortCh` when called | This function must be invoked by the caller when aborting is desired. It guarantees that the channel is closed exactly once, preventing multiple closures or panics. |

### Key Dependencies

* **Check struct** – Holds an internal field (likely named `abortChan`) where this channel will be stored. The actual struct definition isn’t shown in the snippet but is implied by the receiver type.
* **No global state** – The method does not read or write any of the package‑wide globals (`dbByGroup`, `dbLock`, etc.). It operates purely on the receiver instance.

### Side Effects

1. **State mutation** – Stores the provided channel in the check’s internal field, replacing any previously set abort channel.
2. **Resource cleanup** – The returned closure ensures that the channel is closed when invoked, avoiding resource leaks and ensuring that any goroutine blocked on a receive will unblock.

No other side effects (e.g., logging, metrics) are performed by this method.

### How It Fits the Package

`checksdb` manages a collection of `Check` objects used to validate Kubernetes resources.  
During execution:

1. A test runner creates an abort channel for each check it runs.
2. The runner calls `check.SetAbortChan(abortCh)` before launching the check.
3. If the runner needs to terminate the check (e.g., after a timeout), it calls the returned closure, which closes the channel and signals the running goroutine.

This design decouples the abort logic from the check implementation while keeping the API minimal: the check only exposes `SetAbortChan`, and the caller controls when to abort via the returned function.
