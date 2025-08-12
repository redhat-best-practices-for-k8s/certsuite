createExceptionsConfiguration` – Interactive Configuration Builder

| Feature | Detail |
|---------|--------|
| **Signature** | `func()()` |
| **Visibility** | unexported (internal to the *config* command) |
| **Purpose** | Prompts the user for a series of exception lists that influence how CertSuite will run.  These lists are stored in the global `certsuiteConfig` structure and later written to a YAML configuration file by the `generate-config` sub‑command. |

---

#### How it Works

1. **Initial Prompt**  
   The function begins with a short explanatory message printed via `Printf`. It informs the user that the following questions will populate *exception* fields in the config.

2. **Gathering Exceptions**  
   For each exception category (kernel taints, Helm charts, protocol names, services, non‑scalable deployments and stateful sets) the function:
   - Loads a default list from a helper (`loadAcceptedKernelTaints`, `loadHelmCharts`, …).
   - Calls `getAnswer` to display the current list and allow the user to edit it.
   - Normalises the input by converting to lower‑case, trimming spaces, and removing duplicate entries.

3. **Storing Results**  
   The cleaned slice of strings is stored in the corresponding field of `certsuiteConfig`.  
   Example: `certsuiteConfig.KernelTaints = cleanedKernelTaints`

4. **Completion**  
   After all categories are processed, the function returns; control goes back to the command runner that will write the final configuration file.

---

#### Dependencies

| Dependency | Role |
|------------|------|
| `ReplaceAll`, `ToLower` (strings package) | Normalise user input. |
| `Contains` (strings) | Detect duplicates while filtering. |
| `Run` (`exec.Command`) | Executes an external command to load defaults (e.g., `kubectl`). |
| Helper loaders (`loadAcceptedKernelTaints`, `loadHelmCharts`, …) | Provide initial default lists for each exception type. |
| `getAnswer` | Generic prompt/validation routine that shows the current list and reads user input. |

---

#### Side Effects & Global State

* Modifies the **global** `certsuiteConfig` variable – all fields updated by this function are later persisted to disk.
* Uses `fmt.Printf` for console output; no other I/O side effects.
* No error handling is performed inside the function itself; it relies on the called helpers (`load…`, `getAnswer`) to report issues.

---

#### Context in the Package

The *config* sub‑command of CertSuite generates a YAML file that describes how the tool should behave.  
`createExceptionsConfiguration` is part of the interactive wizard that runs when the user executes:

```bash
certsuite generate config
```

It collects exception data from the user, stores it in `certsuiteConfig`, and then the `saveConfigCmd` command writes this struct to a file using Go's `yaml` marshaler. This function is therefore central to customizing CertSuite’s runtime behaviour through configuration.

---

#### Mermaid Diagram (suggested)

```mermaid
flowchart TD
  A[User runs "certsuite generate config"] --> B[interactive wizard]
  B --> C[createExceptionsConfiguration]
  C --> D{Load defaults}
  D -->|kernel taints| E[loadAcceptedKernelTaints]
  D -->|helm charts| F[loadHelmCharts]
  D -->|protocol names| G[loadProtocolNames]
  D -->|services| H[loadServices]
  D -->|non‑scalable deployments| I[loadNonScalableDeployments]
  D -->|non‑scalable stateful sets| J[loadNonScalableStatefulSets]
  E & F & G & H & I & J --> K[getAnswer (prompt)]
  K --> L[Normalize input]
  L --> M[Store in certsuiteConfig]
  M --> N[Return to wizard]
```

---
