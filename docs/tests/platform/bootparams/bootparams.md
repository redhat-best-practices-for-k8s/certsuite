# Package bootparams

**Path**: `tests/platform/bootparams`

## Table of Contents

- [Overview](#overview)
- [Exported Functions](#exported-functions)
  - [GetMcKernelArguments](#getmckernelarguments)
  - [TestBootParamsHelper](#testbootparamshelper)
- [Local Functions](#local-functions)
  - [getCurrentKernelCmdlineArgs](#getcurrentkernelcmdlineargs)
  - [getGrubKernelArgs](#getgrubkernelargs)

## Overview

The bootparams package provides utilities for verifying that the kernel arguments specified in a node’s MachineConfig match those actually present on the running system, including both the container command line and GRUB configuration. It is intended for use within CertSuite test environments to detect misconfigurations.

### Key Features

- GetMcKernelArguments parses a MachineConfig string of kernel arguments into a key‑value map for easy lookup.
- TestBootParamsHelper compares the expected kernel args against the current container and GRUB values, emitting warnings or debug logs when mismatches occur.
- Internal helpers getCurrentKernelCmdlineArgs and getGrubKernelArgs execute commands inside probe pods to capture live kernel argument states.

### Design Notes

- Assumes that executing grub commands in a probe pod reflects the host’s boot configuration.
- Functions return detailed errors on command execution failures, which may surface as test failures.
- Best practice: run TestBootParamsHelper early in a test suite after MachineConfig changes to catch drift before deployment.

### Exported Functions Summary

| Name | Purpose |
|------|----------|
| [func GetMcKernelArguments(env *provider.TestEnvironment, nodeName string) (aMap map[string]string)](#getmckernelarguments) | Converts the list of kernel arguments (`[]string`) from a node’s MachineConfig into a key‑value map for easy lookup. |
| [func TestBootParamsHelper(env *provider.TestEnvironment, cut *provider.Container, logger *log.Logger) error](#testbootparamshelper) | Compares expected kernel arguments from MachineConfig (`GetMcKernelArguments`) against the current command‑line arguments in the container and GRUB configuration, emitting warnings or debug logs for mismatches. |

### Local Functions Summary

| Name | Purpose |
|------|----------|
| [func getCurrentKernelCmdlineArgs(env *provider.TestEnvironment, nodeName string) (map[string]string, error)](#getcurrentkernelcmdlineargs) | Executes the `grubKernelArgsCommand` inside a probe pod to capture the current kernel command‑line arguments and returns them as a map of key/value pairs. |
| [func getGrubKernelArgs(env *provider.TestEnvironment, nodeName string) (aMap map[string]string, err error)](#getgrubkernelargs) | Executes `grub2-editenv list` inside a probe pod to obtain the current GRUB kernel command line arguments and returns them as a key‑value map. |

## Exported Functions

### GetMcKernelArguments

**GetMcKernelArguments** - Converts the list of kernel arguments (`[]string`) from a node’s MachineConfig into a key‑value map for easy lookup.

Retrieve the kernel argument map defined in a node’s MachineConfig.

---

#### Signature (Go)

```go
func GetMcKernelArguments(env *provider.TestEnvironment, nodeName string) (aMap map[string]string)
```

---

#### Summary Table

| Aspect | Details |
|--------|---------|
| **Purpose** | Converts the list of kernel arguments (`[]string`) from a node’s MachineConfig into a key‑value map for easy lookup. |
| **Parameters** | `env *provider.TestEnvironment` – test environment containing node data.<br>`nodeName string` – name of the target node. |
| **Return value** | `map[string]string` – mapping of kernel argument names to their values (empty string if no value). |
| **Key dependencies** | Calls `arrayhelper.ArgListToMap` from `github.com/redhat-best-practices-for-k8s/certsuite/pkg/arrayhelper`. |
| **Side effects** | None. Pure function; only reads data from the supplied environment. |
| **How it fits the package** | Provides a helper for tests that need to compare MachineConfig kernel arguments against other sources (e.g., runtime cmdline, GRUB). |

---

#### Internal workflow (Mermaid)

```mermaid
flowchart TD
  GetMcKernelArguments --> ArgListToMap
```

---

#### Function dependencies (Mermaid)

```mermaid
graph TD
  func_GetMcKernelArguments --> func_ArgListToMap
```

---

#### Functions calling `GetMcKernelArguments` (Mermaid)

```mermaid
graph TD
  func_testSysctlConfigs --> func_GetMcKernelArguments
  func_TestBootParamsHelper --> func_GetMcKernelArguments
```

---

#### Usage example (Go)

```go
// Minimal example invoking GetMcKernelArguments
env := &provider.TestEnvironment{ /* … populate as needed … */ }
nodeName := "worker-0"

kernelArgs, err := bootparams.GetMcKernelArguments(env, nodeName)
if err != nil {
    // Handle error if the environment is malformed (not expected in current signature)
}
for key, val := range kernelArgs {
    fmt.Printf("Kernel arg %q = %q\n", key, val)
}
```

---

### TestBootParamsHelper

**TestBootParamsHelper** - Compares expected kernel arguments from MachineConfig (`GetMcKernelArguments`) against the current command‑line arguments in the container and GRUB configuration, emitting warnings or debug logs for mismatches.

Validates that the kernel command line arguments specified in a node’s MachineConfig match those actually present in the running container and in GRUB, logging any discrepancies.

```go
func TestBootParamsHelper(env *provider.TestEnvironment, cut *provider.Container, logger *log.Logger) error
```

| Aspect | Details |
|--------|---------|
| **Purpose** | Compares expected kernel arguments from MachineConfig (`GetMcKernelArguments`) against the current command‑line arguments in the container and GRUB configuration, emitting warnings or debug logs for mismatches. |
| **Parameters** | `env *provider.TestEnvironment` – test environment context.<br>`cut *provider.Container` – container under test.<br>`logger *log.Logger` – logger used for reporting. |
| **Return value** | `error` – non‑nil if the probe pod is missing or any helper call fails; otherwise nil. |
| **Key dependencies** | • `GetMcKernelArguments(env, nodeName)`<br>• `getCurrentKernelCmdlineArgs(env, nodeName)`<br>• `getGrubKernelArgs(env, nodeName)`<br>• `fmt.Errorf` for error construction |
| **Side effects** | Writes log entries via the supplied logger; no state mutation. |
| **How it fits the package** | Serves as the core check used by higher‑level tests (e.g., `testUnalteredBootParams`) to ensure that boot parameters are not unintentionally altered on a node. |

#### Internal workflow

```mermaid
flowchart TD
  A["Start"] --> B{"Probe pod exists?"}
  B -- No --> C["Return error"]
  B -- Yes --> D["Retrieve MachineConfig args"]
  D --> E["Get current container kernel args"]
  E --> F{"Error?"}
  F -- Yes --> G["Return error"]
  F -- No --> H["Get GRUB kernel args"]
  H --> I{"Error?"}
  I -- Yes --> J["Return error"]
  I -- No --> K["Iterate over MachineConfig keys"]
  K --> L{"Key present in current args?"}
  L -- Yes --> M{"Values match?"}
  M -- No --> N["Warn mismatch (current)"]
  M -- Yes --> O["Debug match (current)"]
  L -- No --> P["Skip"]
  K --> Q{"Key present in GRUB args?"}
  Q -- Yes --> R{"Values match?"}
  R -- No --> S["Warn mismatch (GRUB)"]
  R -- Yes --> T["Debug match (GRUB)"]
  Q -- No --> U["Skip"]
  O & N & T & S --> V["Finish loop"]
  V --> W["Return nil"]
```

#### Function dependencies

```mermaid
graph TD
  func_TestBootParamsHelper --> func_GetMcKernelArguments
  func_TestBootParamsHelper --> func_getCurrentKernelCmdlineArgs
  func_TestBootParamsHelper --> func_getGrubKernelArgs
  func_TestBootParamsHelper --> fmt_Errorf
```

#### Functions calling `TestBootParamsHelper`

```mermaid
graph TD
  func_testUnalteredBootParams --> func_TestBootParamsHelper
```

#### Usage example

```go
// Minimal example invoking TestBootParamsHelper
env := provider.NewTestEnvironment(...)
cut := &provider.Container{NodeName: "node-1", ...}
logger := log.New(os.Stdout, "", log.LstdFlags)

if err := bootparams.TestBootParamsHelper(env, cut, logger); err != nil {
    fmt.Printf("Boot params check failed: %v\n", err)
} else {
    fmt.Println("Boot parameters are consistent")
}
```

---

## Local Functions

### getCurrentKernelCmdlineArgs

**getCurrentKernelCmdlineArgs** - Executes the `grubKernelArgsCommand` inside a probe pod to capture the current kernel command‑line arguments and returns them as a map of key/value pairs.

#### Signature (Go)

```go
func getCurrentKernelCmdlineArgs(env *provider.TestEnvironment, nodeName string) (map[string]string, error)
```

#### Summary Table

| Aspect | Details |
|--------|---------|
| **Purpose** | Executes the `grubKernelArgsCommand` inside a probe pod to capture the current kernel command‑line arguments and returns them as a map of key/value pairs. |
| **Parameters** | `env *provider.TestEnvironment` – test environment containing probe pods; <br> `nodeName string` – name of the node whose probe pod is queried. |
| **Return value** | `map[string]string` – parsed kernel arguments; `error` if execution fails or output cannot be parsed. |
| **Key dependencies** | • `clientsholder.GetClientsHolder()`<br>• `clientsholder.NewContext(...)`<br>• `o.ExecCommandContainer(ctx, kernelArgscommand)`<br>• `strings.Split`, `strings.TrimSuffix`<br>• `arrayhelper.ArgListToMap` |
| **Side effects** | No state mutation; performs I/O by executing a command inside a container. |
| **How it fits the package** | Provides low‑level data needed for boot parameter validation in the *bootparams* test suite. |

#### Internal workflow (Mermaid)

```mermaid
flowchart TD
  A["GetClientsHolder"] --> B["NewContext"]
  B --> C["ExecCommandContainer"]
  C --> D["TrimSuffix & Split"]
  D --> E["ArgListToMap"]
```

#### Function dependencies (Mermaid)

```mermaid
graph TD
  func_getCurrentKernelCmdlineArgs --> func_GetClientsHolder
  func_getCurrentKernelCmdlineArgs --> func_NewContext
  func_getCurrentKernelCmdlineArgs --> func_ExecCommandContainer
  func_getCurrentKernelCmdlineArgs --> func_Split
  func_getCurrentKernelCmdlineArgs --> func_TrimSuffix
  func_getCurrentKernelCmdlineArgs --> func_ArgListToMap
```

#### Functions calling `getCurrentKernelCmdlineArgs` (Mermaid)

```mermaid
graph TD
  func_TestBootParamsHelper --> func_getCurrentKernelCmdlineArgs
```

#### Usage example (Go)

```go
// Minimal example invoking getCurrentKernelCmdlineArgs
env := &provider.TestEnvironment{ /* populated elsewhere */ }
nodeName := "worker-node-01"

args, err := getCurrentKernelCmdlineArgs(env, nodeName)
if err != nil {
    log.Fatalf("failed to retrieve kernel args: %v", err)
}
fmt.Printf("Kernel arguments for %s: %+v\n", nodeName, args)
```

---

### getGrubKernelArgs

**getGrubKernelArgs** - Executes `grub2-editenv list` inside a probe pod to obtain the current GRUB kernel command line arguments and returns them as a key‑value map.

#### Signature (Go)

```go
func getGrubKernelArgs(env *provider.TestEnvironment, nodeName string) (aMap map[string]string, err error)
```

#### Summary Table

| Aspect | Details |
|--------|---------|
| **Purpose** | Executes `grub2-editenv list` inside a probe pod to obtain the current GRUB kernel command line arguments and returns them as a key‑value map. |
| **Parameters** | `env *provider.TestEnvironment` – test environment holding probe pods.<br>`nodeName string` – name of the node whose probe pod will be queried. |
| **Return value** | `aMap map[string]string` – mapping of GRUB kernel argument names to values (empty string if no value).<br>`err error` – any execution or parsing error. |
| **Key dependencies** | • `clientsholder.GetClientsHolder()` – obtains Kubernetes client holder.<br>• `clientsholder.NewContext(...)` – builds context for pod, namespace and container.<br>• `ExecCommandContainer(ctx, grubKernelArgsCommand)` – runs command inside the pod.<br>• `strings.Split`, `strings.HasPrefix` – parse output.<br>• `arrayhelper.FilterArray`, `arrayhelper.ArgListToMap` – filter & convert list to map. |
| **Side effects** | No state mutation; performs I/O by executing a container command and parsing its stdout. |
| **How it fits the package** | Provides GRUB‑level kernel parameters used in boot‑parameter validation tests within `bootparams`. |

#### Internal workflow (Mermaid)

```mermaid
flowchart TD
  A["Start"] --> B["Get Kubernetes client holder"]
  B --> C["Build pod context with node’s probe pod"]
  C --> D["Execute grub command inside container"]
  D --> E{"Success?"}
  E -- No --> F["Return error"]
  E -- Yes --> G["Split output by newline"]
  G --> H["Filter lines starting with options"]
  H --> I{"Exactly one line?"}
  I -- No --> J["Return error"]
  I -- Yes --> K["Split options line into args"]
  K --> L["Discard first empty element"]
  L --> M["Convert arg list to map"]
  M --> N["Return map"]
```

#### Function dependencies (Mermaid)

```mermaid
graph TD
  func_getGrubKernelArgs --> clientsholder.GetClientsHolder
  func_getGrubKernelArgs --> clientsholder.NewContext
  func_getGrubKernelArgs --> ExecCommandContainer
  func_getGrubKernelArgs --> arrayhelper.FilterArray
  func_getGrubKernelArgs --> arrayhelper.ArgListToMap
```

#### Functions calling `getGrubKernelArgs` (Mermaid)

```mermaid
graph TD
  TestBootParamsHelper --> getGrubKernelArgs
```

#### Usage example (Go)

```go
// Minimal example invoking getGrubKernelArgs
env := &provider.TestEnvironment{ /* initialized elsewhere */ }
nodeName := "worker-0"

grubArgs, err := getGrubKernelArgs(env, nodeName)
if err != nil {
    log.Fatalf("failed to get GRUB args: %v", err)
}
fmt.Printf("GRUB kernel arguments for %s: %+v\n", nodeName, grubArgs)
```

---
