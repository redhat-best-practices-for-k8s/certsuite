createCertSuiteResourcesConfiguration`

| | |
|---|---|
| **Package** | `config` (`github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/config`) |
| **Exported** | ❌ (private helper) |
| **Signature** | `func()()` |

### Purpose
Collect the user‑defined resource configuration for CertSuite and populate the global `certsuiteConfig` variable.  
The function drives an interactive prompt session that walks the user through a series of questions:

1. **Namespaces** – which namespaces to include.
2. **Pod labels** – label selectors for pods.
3. **Operator labels** – label selectors for operators.
4. **CRD filters** – optional filtering of Custom Resource Definitions.
5. **Managed deployments & stateful sets** – resources that CertSuite should monitor.

Each step loads the current cluster state (via `load…` helpers), presents a list, and records the user’s choice(s) in `certsuiteConfig`.  

The function returns no value; its side effect is updating the global configuration object for later use by the rest of the generate command.

### Flow & Key Dependencies
```mermaid
flowchart TD
    A[Start] --> B{loadNamespaces()}
    B -->|ok| C[getAnswer("namespaces")]
    C --> D{loadPodLabels()}
    D --> E[getAnswer("pods")]
    E --> F{loadOperatorLabels()}
    F --> G[getAnswer("operators")]
    G --> H{loadCRDfilters()}
    H --> I[getAnswer("crdFilters")]
    I --> J{loadManagedDeployments()}
    J --> K[getAnswer("managedDeployments")]
    K --> L{loadManagedStatefulSets()}
    L --> M[getAnswer("managedStatefulSets")]
    M --> N[Store in certsuiteConfig]
    N --> O[End]
```

| Step | Called function | What it does |
|------|-----------------|--------------|
| Load data | `loadNamespaces`, `loadPodLabels`, `loadOperatorLabels`, `loadCRDfilters`, `loadManagedDeployments`, `loadManagedStatefulSets` | Reads current cluster state (namespaces, labels, CRDs, etc.) and returns a list of options. |
| Prompt | `getAnswer(question string)` | Displays the question with the loaded options, captures the user’s selection(s). |
| Update config | Assignment to `certsuiteConfig.Resources.<field>` | Persists selections for later use by the generation logic. |

### Side Effects
- **Global state mutation**: writes to `certsuiteConfig`, a package‑level variable that holds all configuration data for the generate command.
- **Console I/O**: uses `fmt.Printf` and `getAnswer` (which in turn reads from stdin) to interact with the user.
- **External commands**: calls `kubectl` via the helper `Run` function inside each `load…` function, so it expects a working Kubernetes context.

### Interaction with Other Package Elements
* The function is invoked by the command‑line handler (`generateConfigCmd`) when the user chooses the “resources” submenu.  
* It relies on constants defined in `const.go` (e.g., `namespacesHelp`, `podsPrompt`, etc.) to build the prompts and help text presented to the user.

### Summary
`createCertSuiteResourcesConfiguration` is a purely interactive helper that gathers resource‑selection preferences from the user, updates the global configuration object, and has no return value. It ties together several cluster‑state loaders, prompt helpers, and constant strings to produce a ready‑to‑use `certsuiteConfig` for subsequent generation steps.
