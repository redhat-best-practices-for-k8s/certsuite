# Batch scenarios

A batch scenario checks many catalog tests in a single certsuite run. One
`--label-filter` carries every test ID in the entry, and the runner validates
each test separately against the same claim.

Batching exists because the per-run overhead dominates: 41 tests as 41
single-test entries cost 41 deploy/run/cleanup cycles, and as one batch they
cost one. The scenarios in here cover most of the pod-spec, workload
controller, process, CRD and operator catalog.

See [../CONTRIBUTING.md](../CONTRIBUTING.md) for the scenario framework and the
full `scenarios.json` field reference.

## Groups

| Group | Namespace | Pod label | Covers |
|---|---|---|---|
| `pod-spec` | `pod-spec-smoke-ns` | `scenario-pod-spec: target` | Pod and container spec checks: capabilities, host namespaces, security context, resources, lifecycle hooks, reserved ports, CPU pinning and isolation, service accounts, RBAC bindings, storage |
| `workload-controllers` | `controllers-smoke-ns` | `scenario-controllers: target` | Checks that judge the controller rather than the pod: high availability, PodDisruptionBudget, topology spread, AffinityRequired |
| `processes` | `processes-smoke-ns` | `scenario-processes: target` | Checks that inspect processes inside the container: process count per container, sshd daemons |
| `crds` | `crds-smoke-ns` | `scenario-crds: target` | Checks that inspect the CRDs matched by `targetCrdFilters` and the Roles that reference them |
| `operators` | `operators-golden-smoke-ns` and per-entry namespaces | per-entry `scenario-operators-*: target` | Operator and OLM checks: subscriptions, CSV status, privileges, skipRange, CRD ownership, semantic versioning, hugepages, bundle counts, tenant namespaces |

Each group has its own namespace and label so one group's workloads are invisible
to the other's runs.

### pod-spec entries

| Entry | Manifest | Expectation |
|---|---|---|
| Golden workloads | `golden.yaml` | 41 tests `passed`. A golden Deployment mounting a statically bound volume, a Guaranteed whole-CPU variant with the `certsuite-smoke` RuntimeClass and the CPU/IRQ load-balancing annotations, a ClusterIP Service, a ResourceQuota and the storage objects backing the volume. |
| One violation per workload | `violations.yaml` | 39 tests `failed`, each on its own workload. |
| Security context without fsGroup | `bad-security-context.yaml` | `access-control-security-context` `failed`. |
| Shared process namespace with SYS_PTRACE | `ptrace.yaml` | `access-control-sys-ptrace-capability` `passed`. |

### workload-controllers entries

| Entry | Manifest | Expectation |
|---|---|---|
| Golden workloads | `golden.yaml` | 4 tests `passed`. Two-replica Deployments with PodDisruptionBudgets: one with pod anti-affinity, one with hostname and zone topology spread constraints, one labelled `AffinityRequired` with node affinity. |
| One violation per workload | `violations.yaml` | 4 tests `failed`: one replica, a missing PodDisruptionBudget, a hostname-only spread constraint, and an `AffinityRequired` workload that keeps pod anti-affinity. |

### processes entries

| Entry | Manifest | Expectation |
|---|---|---|
| Golden workloads | `golden.yaml` | 2 tests `passed`. A plain nginx Deployment: one process per container, no sshd listening. |
| One violation per workload | `violations.yaml` | 2 tests `failed`: a Deployment whose container runs multiple processes, and an sshd Deployment listening in the foreground. |

### crds entries

| Entry | Manifest | Expectation |
|---|---|---|
| Golden objects | `golden.yaml` | 2 tests `passed`. A namespaced CRD with a status subresource and a Role granting read access to it. |
| One violation per object | `violations.yaml` | 2 tests `failed`: a CRD without a status subresource, and a Role that grants an extra pods rule next to its CRD access. |

The `crds` group deploys no pods; its label matches nothing and the checks read
the CRD and Role objects directly. The golden and violations runs share the CRD
name, so `crds/cleanup.sh` deletes the CRD before deleting the namespace —
otherwise the second run applies over the first one's CRD instead of replacing
it.

### operators entries

The operator group runs real OLM installs against an in-cluster File-Based
Catalog: each scenario's ConfigMap holds the FBC, an `opm registry serve`
Deployment serves it, and the CatalogSource points at that Service's ClusterIP
by address. The bundle has no image, so OLM resolves the InstallPlan inline
from the CSV embedded in the catalog and nothing is ever published — only the
operator deployment image (`registry.k8s.io/pause:3.10`) is pulled. The
Subscription is applied only after `deploy.sh` patches the CatalogSource
address, so the InstallPlan resolves on the first attempt.

| Entry | Manifest | Expectation |
|---|---|---|
| Golden OLM install | `golden/olm-golden.yaml` | 12 tests `passed`. A SingleNamespace operator with a subscription, skipRange, no clusterPermissions and one owned CRD with a structural schema, installed by an OperatorGroup targeting a tenant namespace. |
| One violation per object | `bad-csvs/bad-csvs.yaml` | 8 tests `failed`. Hand-written CSVs in a namespace with no OperatorGroup — OLM marks them Failed but never runs their install strategy, so their specs stay as written. |
| Unschedulable hugepages pod | `hugepages/hugepages.yaml` | `operator-pods-no-hugepages` `failed`. The operator pod requests hugepages-2Mi, which the Kind workers do not expose, so it stays Pending. |
| Oversized catalog | `bundle-count/big-catalog.yaml` | `operator-catalogsource-bundle-count` `failed`. 1002 bundles (one installable plus 1001 fillers) push the catalog past the 1000-bundle limit. |

Two checks cannot have a failing scenario on this cluster, so they are
covered on the passing side only:

- `operator-crd-openapi-schema` — apiextensions v1 requires a structural
  schema on every served CRD version since Kubernetes 1.16, so the check's
  non-compliant branch (a served CRD without `openAPIV3Schema`) cannot exist
  on any apiserver.
- `operator-semantic-versioning` — CSV `spec.version` is decoded by OLM's
  `version.OperatorVersion`, which parses through `semver.ParseTolerant` and
  re-serializes the normalized form, so a CSV carrying an invalid version
  cannot survive on the cluster.

The `bad-csvs` entry relies on OLM reconciling CSVs in a namespace without an
OperatorGroup to `phase=Failed` (reason `NoOperatorGroup`) within seconds
without running the install strategy; `deploy.sh` waits for exactly that
state. The `hugepages` and `bundle-count` entries' workloads come from a real
Subscription, so their CSVs carry the `olm.operatorNamespace` and
`olm.targetNamespaces` annotations OLM sets on install.

## Isolating violations

Every test in a batch sees every workload in the namespace. Two rules keep one
workload's violation from satisfying another test.

**One changed field per workload.** Each violating workload is the golden
template with exactly one difference, named after what it breaks
(`bad-host-pid`, `bad-no-requests`, `bad-ha`). A test that flags anything else
is flagging a workload it should not.

**`noncompliant_contains` as the guard.** The runner fails the test if *any*
non-compliant object lacks the string, not just if none contain it. So
`"noncompliant_contains": "bad-host-pid"` asserts both that
`access-control-pod-host-pid` failed and that it failed on exactly the intended
workload. Pair every `failed` expectation in a batch with it; pair every
`passed` expectation with `min_compliant`.

**Some checks cannot share a run.** When a check would flag most of the other
workloads, it gets its own entry rather than a loosened assertion:

- `access-control-security-context` — most violations add a capability or a
  host namespace, which also drops them out of security context category 1. Its
  failing case is the separate `bad-security-context.yaml`.
- `access-control-sys-ptrace-capability` — the compliant workload has to add
  `SYS_PTRACE`, which fails the security context check. Its passing case is the
  separate `ptrace.yaml`.
- High availability, PDB, topology spread and AffinityRequired — a one-replica
  violation trips high availability for every workload in the run, so these
  live in their own group with their own namespace.

If a new test cross-talks, split it into its own entry. Do not relax the
assertion.

## Editing the manifests

The manifests are generated. **Do not edit the YAML**; edit the generator and
regenerate:

```bash
python3 .github/actions/test-scenarios/batches/generate-manifests.py
```

The `operators` group has its own generator for its FBC catalogs and OLM
objects:

```bash
python3 .github/actions/test-scenarios/batches/operators/generate-manifests.py
```

`make check-scenario-readme`, part of `make lint`, fails if the checked-in YAML
no longer matches either generator.

`generate-manifests.py` holds one golden pod spec and a table mapping each test
ID to a workload name and the single mutation that breaks it, so a reviewer
reads the one-line difference instead of 2000 lines of YAML. To add a test to a
batch:

1. Add the test ID, workload name and mutation to `POD_SPEC_VIOLATIONS` or
   `CONTROLLER_VIOLATIONS`, or build the workloads directly in
   `processes_manifests()` / `crds_manifests()` when they are not a mutation of
   the shared pod spec template.
2. Confirm the golden template already satisfies the new check. If it does not,
   change the template and re-verify the other entries, because every workload
   inherits it.
3. Regenerate the manifests.
4. Add the test ID to both the golden and the violations entry in
   `scenarios.json`, with `min_compliant` and `noncompliant_contains`.
5. Regenerate the README: `python3 .github/actions/test-scenarios/generate-readme.py`.

## Cluster-scoped objects

Deleting a scenario namespace does not remove cluster-scoped objects, so every
group's `cleanup.sh` deletes them explicitly:

- `pod-spec` applies a `certsuite-smoke` RuntimeClass for the CPU-pinned
  workloads, a `bad-crb-binding` ClusterRoleBinding, and StorageClasses and
  PersistentVolumes backing statically bound volumes. The volumes use
  `reclaimPolicy: Delete` but survive the namespace in `Released` state with a
  stale `claimRef`, so the cleanup deletes them **after** the namespace: the PV
  binder refuses a PV whose `claimRef` still points at the deleted claim, and a
  leftover PV from a previous run would leave the new PVC `Pending`.
- `crds` applies the `scenariocrds.scenario-smoke.example.com` CRD, which both
  of its runs overwrite; the cleanup deletes it so the next run starts from a
  clean schema.
- `operators` applies the `goldencrds.scenario-smoke.example.com` CRD (inline
  in the golden operator's bundle) and the
  `badcrds.scenario-smoke.example.com` CRD (by hand in `bad-csvs`); each
  entry's cleanup deletes its own CRD before deleting the namespaces. The
  entries also span several namespaces (operator, catalog and tenant), so they
  delete them directly instead of the shared single-namespace
  `cleanup-workload.sh`.

Each PersistentVolume sits behind its own StorageClass, so the static binder
cannot cross-bind a PVC to another scenario's PV. Any future cluster-scoped
object needs the same treatment.
