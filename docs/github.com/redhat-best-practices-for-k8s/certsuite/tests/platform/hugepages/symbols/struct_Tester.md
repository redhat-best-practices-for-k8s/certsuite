Tester` – Huge‑Page Verification Helper

The **`Tester`** type encapsulates all state and logic required to validate the huge‑page configuration of a Kubernetes node.  
It is used by the test suite that verifies whether the node’s huge pages are correctly configured via either:

* systemd units exposed by the *Memory‑Controller* (MC) component, or
* kernel command‑line arguments.

The struct is created in `NewTester`, populated with data from the node and MC, then executed with `Run`.

---

## Purpose

- **Collect**: Read the node’s current huge‑page allocation per NUMA node.
- **Compare**: Match those values against what MC reports or what the kernel arguments expect.
- **Report**: Return a boolean indicating success and an error if something went wrong.

It is a thin, read‑only façade over several helper functions (`getNodeNumaHugePages`, `TestNodeHugepagesWithMcSystemd`, etc.) that perform the actual comparisons.

---

## Fields

| Field | Type | Role |
|-------|------|------|
| `commander` | `clientsholder.Command` | Executes shell commands inside containers or on hosts. Used by `getNodeNumaHugePages`. |
| `context` | `clientsholder.Context` | Holds kube‑config and API client; required for any Kubernetes interaction (e.g., retrieving MC units). |
| `mcSystemdHugepagesByNuma` | `hugepagesByNuma` | Cached huge‑page numbers reported by MC’s systemd units. Filled in `NewTester`. |
| `node` | `*provider.Node` | The node under test. Provides the name and IP for command execution. |
| `nodeHugepagesByNuma` | `hugepagesByNuma` | Actual huge‑page numbers read from `/sys/devices/system/node/...`. Filled in `getNodeNumaHugePages`. |

> **Note**: All fields are set during construction; the struct is intentionally immutable thereafter.

---

## Construction – `NewTester`

```go
func NewTester(
    node *provider.Node,
    pod *corev1.Pod,
    cmd clientsholder.Command,
) (*Tester, error)
```

*Initialises* a `Tester` instance:
1. Creates a `Context` via `NewContext(pod)` (sets up kube‑config).
2. Reads MC systemd unit configuration with `getMcSystemdUnitsHugepagesConfig(context)`.
3. Stores the parsed values in `mcSystemdHugepagesByNuma`.

If any step fails, an error is returned and no tester is created.

---

## Core Methods

| Method | Signature | Responsibility |
|--------|-----------|----------------|
| **`HasMcSystemdHugepagesUnits()`** | `func (t *Tester) HasMcSystemdHugepagesUnits() bool` | Checks whether MC reported any huge‑page units (`len(t.mcSystemdHugepagesByNuma) > 0`). |
| **`Run()`** | `func (t *Tester) Run() error` | Orchestrates the test. If MC units exist, it calls `TestNodeHugepagesWithMcSystemd`; otherwise it falls back to `TestNodeHugepagesWithKernelArgs`. Reports errors via the testing framework’s logging helpers (`Info`, `Errorf`). |
| **`TestNodeHugepagesWithKernelArgs()`** | `func (t *Tester) TestNodeHugepagesWithKernelArgs() (bool, error)` | Reads kernel arguments (via helper), compares totals per size against what the node reports. Returns success flag and any error. |
| **`TestNodeHugepagesWithMcSystemd()`** | `func (t *Tester) TestNodeHugepagesWithMcSystemd() (bool, error)` | Compares node’s huge‑page values to those reported by MC systemd units. Emits warnings/errors for mismatches. |
| **`getNodeNumaHugePages()`** | `func (t *Tester) getNodeNumaHugePages() (hugepagesByNuma, error)` | Executes `cat /sys/devices/system/node/.../nr_hugepages` inside the node’s container to build a map of huge‑page counts per NUMA node. |

All public methods are **side‑effect free** except for logging; they only read from the node or MC.

---

## Dependencies & Side Effects

- **Kubernetes API** – via `clientsholder.Context` and `provider.Node`.
- **Shell commands** – executed on the node container to read `/sys/...` files.
- **MC component** – systemd unit configuration is fetched through a dedicated helper (`getMcSystemdUnitsHugepagesConfig`).
- **Logging helpers** (`Info`, `Errorf`) – used for test output; no state mutation occurs.

The struct does not modify the node or MC; it only reads data and reports inconsistencies.

---

## Role in the Package

The `hugepages` package provides end‑to‑end tests that a Kubernetes cluster correctly advertises huge‑page resources.  
`Tester` is the central orchestrator:

1. **Setup** – `NewTester` pulls together node info, MC config, and command execution capability.
2. **Execution** – `Run` decides which verification path to use and reports success/failure.
3. **Reporting** – Errors are surfaced via the testing framework’s logger.

Thus, `Tester` is the bridge between raw system data and high‑level test assertions.
