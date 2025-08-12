testPodDisruptionBudgets`

| Aspect | Details |
|--------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/tests/observability` |
| **Signature** | `func(*checksdb.Check, *provider.TestEnvironment)` |
| **Visibility** | Unexported (internal test helper) |

#### Purpose
`testPodDisruptionBudgets` is a test helper that validates all Pod Disruption Budget (PDB) resources found in the current Kubernetes environment.  
It:

1. Iterates over every PDB discovered by `checksdb.Check`.
2. Uses the test environment (`provider.TestEnvironment`) to retrieve the real PDB objects.
3. Performs validation checks via `CheckPDBIsValid`.
4. Records any errors or warnings into a structured report object.

#### Inputs

| Parameter | Type | Role |
|-----------|------|------|
| `check` | `*checksdb.Check` | Holds all discovered test artifacts, including the list of PDBs (`check.PodsDisruptionBudgets`). |
| `env` | `*provider.TestEnvironment` | Provides Kubernetes client access and context needed to fetch live PDB objects. |

#### Workflow

1. **Logging**  
   Begins with a log entry: *“testing pod disruption budgets”*.

2. **Loop over discovered PDBs**  
   For each PDB in `check.PodsDisruptionBudgets`:
   - Log the namespace and name.
   - Convert its label selector to a `labels.Selector` via `LabelSelectorAsSelector`.
   - Attempt to retrieve the live object using `env.ClientSet().PolicyV1beta1().PodDisruptionBudgets(namespace).Get(...)`.

3. **Validation**  
   If retrieval succeeds, call `CheckPDBIsValid(pdb, selector)`:
   - On success → add a *pass* field to the report.
   - On failure → log an error and add a *fail* field.

4. **Reporting**  
   For each PDB, a new `ReportObject` is created (via `NewReportObject`) with fields such as `"name"`, `"namespace"`, `"valid"`, etc. These objects are appended to the check’s report list.

5. **Error handling**  
   Any Kubernetes API error or validation failure triggers `LogError` and is captured in the report.

#### Key Dependencies

| Dependency | Role |
|------------|------|
| `env.ClientSet()` | Provides access to the Kubernetes API. |
| `LabelSelectorAsSelector` | Converts string selectors into a usable `labels.Selector`. |
| `CheckPDBIsValid` | Encapsulates PDB validation logic (e.g., ensuring minAvailable/maxUnavailable constraints). |
| `NewReportObject`, `AddField` | Build structured test reports. |
| Logging helpers (`LogInfo`, `LogError`) | Emit diagnostic messages during the test run. |

#### Side‑Effects

* No state is modified outside of the provided `check` object; all changes are local to the report list.
* Logs are produced for each PDB processed.

#### Package Context

Within the *observability* test suite, this function complements other resource validators (e.g., service accounts, deployments). It ensures that any PDB defined in a cluster adheres to best‑practice constraints before the test completes. The results feed into the overall compliance report emitted by CertSuite.
