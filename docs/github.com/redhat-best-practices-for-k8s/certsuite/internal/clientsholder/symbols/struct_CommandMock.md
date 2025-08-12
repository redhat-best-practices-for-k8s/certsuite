CommandMock` – A Test Double for the `Command` Interface

| Element | Type | Purpose |
|---------|------|---------|
| **Struct** `CommandMock` | `struct{ ExecCommandContainerFunc func(context Context, s string) (string,string,error); calls struct{ExecCommandContainer []struct{Context Context; S string}}; lockExecCommandContainer sync.RWMutex }` | Provides a controllable stand‑in for the real `Command` implementation used by code that interacts with containers. |

## Overview

The package **clientsholder** contains an interface called `Command`.  
`CommandMock` implements this interface so tests can:

* Supply deterministic responses to command executions.
* Inspect how many times and with what arguments the command was invoked.

It is a classic *mock object* generated (or hand‑crafted) for unit testing. The struct’s fields are public so that test code can configure behaviour and later verify calls.

---

## Field Details

| Field | Type | Description |
|-------|------|-------------|
| `ExecCommandContainerFunc` | `func(context Context, s string) (string,string,error)` | **User‑supplied function** that is called whenever the mock’s `ExecCommandContainer` method runs. Tests set this to a closure that returns controlled outputs or panics. |
| `calls.ExecCommandContainer` | `[]struct{Context Context; S string}` | Stores each invocation of `ExecCommandContainer`. Each element records the arguments passed. |
| `lockExecCommandContainer` | `sync.RWMutex` | Protects concurrent access to the `calls` slice and ensures thread‑safe reads/writes when tests run in parallel or when the mock is used concurrently. |

---

## Methods

### `ExecCommandContainer`

```go
func (m *CommandMock) ExecCommandContainer(ctx Context, s string) (string, string, error)
```

* **What it does**  
  - Locks `lockExecCommandContainer` for writing (`Lock`).  
  - Records the call arguments into `calls.ExecCommandContainer`.  
  - Unlocks.  
  - Delegates to `ExecCommandContainerFunc`, returning its result.

* **Inputs**  
  - `ctx Context` – context for command execution (opaque type from this package).  
  - `s string` – container name or identifier.

* **Outputs**  
  - `(stdout string, stderr string, err error)` – whatever the supplied function returns.

* **Side effects**  
  - Thread‑safe recording of call arguments.  
  - Potential panic if the supplied function panics (tests should catch this).

---

### `ExecCommandContainerCalls`

```go
func (m *CommandMock) ExecCommandContainerCalls() []struct{Context Context; S string}
```

* **What it does**  
  - Reads the slice of recorded calls in a thread‑safe manner (`RLock`).  
  - Returns a copy of that slice.

* **Usage**  
  ```go
  if len(mock.ExecCommandContainerCalls()) == 0 { t.Fatal("expected call") }
  ```

---

## How It Fits Into `clientsholder`

The real code expects an object that satisfies the `Command` interface. During tests, instead of invoking actual container commands (which would be slow, flaky, or require a running cluster), the test creates:

```go
mock := &CommandMock{
    ExecCommandContainerFunc: func(ctx Context, s string) (string,string,error) {
        // deterministic mock response
        return "out", "", nil
    },
}
```

The production code uses `mock` wherever it needs a `Command`. After the test runs, assertions can be made on:

* The outputs returned by the mock function.
* The number and arguments of calls via `ExecCommandContainerCalls`.

---

## Suggested Mermaid Diagram

```mermaid
classDiagram
    class CommandMock {
        +ExecCommandContainerFunc func(context Context, s string) (string,string,error)
        +calls struct{ ExecCommandContainer []struct{Context Context; S string} }
        -lockExecCommandContainer sync.RWMutex
        +ExecCommandContainer(ctx Context, s string) (string,string,error)
        +ExecCommandContainerCalls() []struct{Context Context; S string}
    }

    CommandMock --> "1" Context : uses in methods
```

This diagram visualises the mock’s fields and its relationship with the `Context` type.

---

### Summary

* **Purpose** – Provide a lightweight, configurable stand‑in for container command execution during tests.  
* **Inputs/Outputs** – Mirrors the real interface; all behaviour is driven by the supplied function field.  
* **Key Dependencies** – Relies on `sync.RWMutex` for concurrency safety and on the package’s own `Context` type.  
* **Side Effects** – Records call history safely, enabling verification of interactions in unit tests.
