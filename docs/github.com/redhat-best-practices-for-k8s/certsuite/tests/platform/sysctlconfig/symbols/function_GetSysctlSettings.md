GetSysctlSettings` – Package *sysctlconfig*

| Item | Detail |
|------|--------|
| **File** | `sysctlconfig.go:49` |
| **Signature** | `func GetSysctlSettings(env *provider.TestEnvironment, containerName string) (map[string]string, error)` |
| **Exported?** | Yes |

---

## Purpose

`GetSysctlSettings` retrieves the current sys‑ctls that are exposed to a container in a test environment.  
It runs the `sysctl -a` command inside the specified container, parses its output and returns a map where each key is a sys‑ctl name and the value is its current setting.

This helper is used by tests that need to verify kernel configuration changes (e.g., after applying a CIS benchmark or a custom policy).

---

## Parameters

| Name | Type | Role |
|------|------|------|
| `env` | `*provider.TestEnvironment` | Holds the overall test context: provider information, client sets and any other shared state. |
| `containerName` | `string` | The name of the container inside which the command is executed (e.g., `"certsuite"`, `"sysctl-test"`). |

---

## Return Values

| Value | Type | Description |
|-------|------|-------------|
| First return | `map[string]string` | Mapping from sys‑ctl names to their current values. Empty map on error. |
| Second return | `error` | Non‑nil if any step fails (client lookup, context creation, command execution or parsing). |

---

## Key Dependencies & Flow

```mermaid
flowchart TD
  A[GetSysctlSettings] --> B{Retrieve ClientsHolder}
  B --> C[provider.GetClientsHolder(env)]
  C --> D{Create Context}
  D --> E[client.NewContext()]
  E --> F{Exec sysctl -a}
  F --> G[ExecCommandContainer(ctx, containerName, "sysctl", "-a")]
  G --> H{Parse Output}
  H --> I[parseSysctlSystemOutput(stdout)]
  I --> J[Return map]
```

1. **ClientsHolder** – `GetClientsHolder` pulls the shared Kubernetes client set from the test environment.
2. **Context** – `NewContext` builds a context containing the Kubernetes client, namespace and container name; this is required by the command executor.
3. **Command Execution** – `ExecCommandContainer` runs `sysctl -a` inside the target container and streams back stdout/stderr.
4. **Parsing** – `parseSysctlSystemOutput` turns the raw string into a key/value map.

---

## Side‑Effects

* No state is mutated outside of local variables; all interactions are read‑only or temporary.
* The function performs I/O: it connects to the cluster, spawns a command inside a pod and reads its output.  
  Consequently, it may block until the command completes or fails.

---

## Usage Context in `sysctlconfig` Package

The package contains tests that:

1. Apply CIS‑style sysctl changes (via `ApplyCISConfig`).
2. Verify those changes using `GetSysctlSettings`.
3. Optionally compare against a baseline configuration.

`GetSysctlSettings` is the core helper that turns the container’s kernel state into a Go map for assertion logic elsewhere in the test suite.
