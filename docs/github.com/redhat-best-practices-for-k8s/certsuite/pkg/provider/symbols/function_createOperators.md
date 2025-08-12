createOperators`

**File:** `pkg/provider/operators.go` (line 154)  
**Package:** `provider`

---

### Purpose
`createOperators` aggregates data from the Operator Lifecycle Manager (OLM) resources and builds a slice of lightweight **Operator** objects that are used by Cert‑Suite to validate operator‑related requirements in an OpenShift cluster.

The function:

1. Deduplicates ClusterServiceVersions (CSV) per name.  
2. Determines which namespaces each operator targets based on the subscription.  
3. Collects installation plans, catalog sources and package manifests for a complete operator picture.  
4. Returns the constructed `[]*Operator` slice for further processing by the provider.

---

### Inputs
| Parameter | Type | Description |
|-----------|------|-------------|
| `csvList` | `[]*olmv1Alpha.ClusterServiceVersion` | All CSVs discovered in the cluster. |
| `subscriptions` | `[]olmv1Alpha.Subscription` | Subscription objects that indicate which operator is installed and where it should run. |
| `packageManifests` | `[]*olmpkgv1.PackageManifest` | Metadata about available packages (operators). |
| `installPlans` | `[]*olmv1Alpha.InstallPlan` | Objects describing the actual installation steps for operators. |
| `catalogSources` | `[]*olmv1Alpha.CatalogSource` | Source of operator images and metadata. |
| `isLocalCluster` | `bool` | Flag indicating whether the cluster is a local test environment (affects label handling). |
| `hasNamespacePrefix` | `bool` | Indicates if subscription names contain a namespace prefix. |

---

### Output
```go
[]*Operator
```
A slice of pointers to **Operator** structs, each representing an operator found in the cluster with all relevant OLM metadata attached.

---

### Key Dependencies (called functions)

| Called Function | Role |
|-----------------|------|
| `getUniqueCsvListByName` | Removes duplicate CSVs per name. |
| `SplitN`, `len` | String and slice utilities for parsing subscription names. |
| `Debug`, `Info`, `Warn`, `Error` | Logging helpers from the provider's logger. |
| `String` | Converts various values to string for logging. |
| `getAtLeastOneSubscription` | Finds a subscription that targets at least one namespace. |
| `getOperatorTargetNamespaces` | Parses target namespaces from a subscription. |
| `getAtLeastOneInstallPlan` | Retrieves the first install plan for an operator. |
| `append` | Builds the result slice incrementally. |

---

### Side‑Effects
* **Logging** – The function emits debug/info/warn/error logs during processing.
* No global state is mutated; it only reads its inputs and returns a new slice.

---

### How It Fits the Package

The `provider` package orchestrates all Cert‑Suite checks against an OpenShift cluster.  
`createOperators` sits in the **operator discovery** phase:

1. **Discovery** – Other functions gather raw OLM resources (`csvList`, `subscriptions`, …).
2. **Transformation** – `createOperators` converts those raw objects into a uniform `Operator` model.
3. **Validation** – Subsequent provider logic (e.g., health checks, version validation) consumes the returned slice.

By isolating this transformation logic, the package keeps discovery, modeling, and validation concerns separate, making the code easier to test and maintain.
