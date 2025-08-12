loadNamespaces`

```go
func loadNamespaces(namespaces []string) func()
```

| Aspect | Description |
|--------|-------------|
| **Purpose** | Builds a closure that, when executed, will add the supplied list of namespace names to the configuration object used by the CertSuite generator. The function is intended to be called lazily – the actual mutation of the global `certsuiteConfig` happens only when the returned function runs. |
| **Parameters** | `namespaces []string` – a slice containing the names of Kubernetes namespaces that should be included in the generated configuration. |
| **Return value** | A zero‑argument function (`func()`) that performs the side effect described above. The caller can decide when to invoke it, e.g., after all command line flags have been parsed or during a `Run` phase of a Cobra command. |
| **Key dependencies** | - **`certsuiteConfig`** (global in this package): the configuration struct into which namespaces are appended.<br> - The Go built‑in `append` function is used to extend slices. No other external packages are referenced. |
| **Side effects** | 1. Iterates over each element of the `namespaces` slice.<br>2. For each namespace, calls `certsuiteConfig.Namespaces = append(certsuiteConfig.Namespaces, ns)` (the exact field name is inferred; it may be something like `Namespaces`).<br>3. The mutation occurs only when the returned closure is invoked; calling `loadNamespaces` itself does **not** change any state. |
| **Where it's used** | While not shown in the snippet, such a pattern is typical for Cobra command flag handling: the flag value (`[]string`) is captured by `loadNamespaces`, and the resulting function is stored as a `PreRunE` or `RunE` callback of the generate sub‑command. This keeps flag parsing separate from configuration building. |
| **Example usage** (hypothetical)** | ```go\ncmd.Flags().StringSliceVar(&ns, \"namespace\", nil, \"Namespaces to include\")\ncmd.PreRunE = loadNamespaces(ns)\n``` |

### Summary

`loadNamespaces` is a helper that turns a list of namespace names into a deferred configuration mutation. It keeps flag parsing clean and allows the rest of the generator code to operate on a fully populated `certsuiteConfig`.
