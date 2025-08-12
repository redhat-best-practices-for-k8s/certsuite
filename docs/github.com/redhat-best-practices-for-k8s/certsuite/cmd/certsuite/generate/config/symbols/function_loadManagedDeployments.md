loadManagedDeployments`

```go
func loadManagedDeployments([]string) func()
```

#### Purpose  
`loadManagedDeployments` is a helper that creates a closure used by the command‑line
configuration wizard to capture a list of *managed deployments* entered by the user.
The returned function appends those values into the global `certsuiteConfig`
structure so they can be persisted later (e.g., when the user selects **Save**).

#### Parameters  
| Name | Type | Description |
|------|------|-------------|
| `[]string` | slice of strings | The list of deployment names supplied by the wizard. |

> **Note:** The parameter is unnamed in the source because it is only used inside
> the closure; its value is captured and never referenced directly.

#### Return Value  
A zero‑argument function (`func()`) that, when invoked, appends the captured slice
to `certsuiteConfig.ManagedDeployments`.  
The returned function has no return values itself.

#### Key Operations  
1. **Capture** the supplied slice in a closure.
2. On invocation, use Go’s built‑in `append` to add each element to
   `certsuiteConfig.ManagedDeployments`.

```go
return func() {
    certsuiteConfig.ManagedDeployments = append(
        certsuiteConfig.ManagedDeployments,
        args...,
    )
}
```

#### Dependencies & Side Effects  
* **Global state** – modifies the package‑level variable `certsuiteConfig`.
* No other global variables or external packages are accessed.
* The function has no side effects beyond mutating that slice; it does not
  perform I/O, logging, or error handling.

#### How It Fits the Package  
`config` is responsible for building an in‑memory configuration from user input.
The wizard collects various sections (e.g., namespaces, services, probes).
For each section, a dedicated `loadXxx` function returns a closure that records
the collected data.  
`loadManagedDeployments` follows this pattern for the *managed deployments* menu,
ensuring that whatever values the user enters are persisted in the final
configuration struct before it is written to disk or displayed.
