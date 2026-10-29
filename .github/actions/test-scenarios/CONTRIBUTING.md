# Writing a smoke test scenario

A scenario deploys a known-good or known-bad workload into an isolated
namespace, runs the certsuite against it, and asserts what the claim says. This
is how a catalog test earns a "Covered" row in [README.md](README.md).

`run-scenarios.sh` drives everything from `scenarios.json`. The pre-main
workflow runs it once per architecture after the partner cluster is up
(`.github/workflows/pre-main.yaml`, job `smoke-tests-local`).

## Layout

```text
test-scenarios/
  scenarios.json          every scenario, in run order
  run-scenarios.sh        the runner
  deploy-workload.sh      shared deploy: create namespace, apply, wait
  cleanup-workload.sh     shared cleanup: delete results dir, delete namespace
  generate-readme.py      regenerates README.md from CATALOG.md + scenarios.json
  <suite>/<test-name>/    one directory per scenario group
    deploy.sh             takes a manifest filename
    cleanup.sh            takes a results directory
    manifests/
      certsuite-config.yaml
      <workload>.yaml
  batches/                multi-test scenarios; see batches/README.md
```

## Adding a single-test scenario

1. Create a directory under the suite the test belongs to, e.g.
   `networking/<test-name>/`.
2. Add `deploy.sh` taking the manifest filename as `$1`. With no extra setup
   it can be one line:
   `exec ../../deploy-workload.sh "${SCENARIO_NAMESPACE:-<base>}" <manifest-path>`.
   The runner exports `SCENARIO_NAMESPACE` (see
   [Namespaces](#namespaces-are-per-entry)); the fallback is the base namespace
   from `certsuite-config.yaml`, which keeps the script runnable by hand.
3. Add `cleanup.sh` taking the results directory as `$1`, usually
   `exec ../../cleanup-workload.sh "${SCENARIO_NAMESPACE:-<base>}" "$1"`.
4. Put the manifests and a `certsuite-config.yaml` in `manifests/`.
   `deploy-workload.sh` waits for every Deployment in the namespace to become
   Available, so a manifest may hold more than one workload.
5. Add one entry per workload to `scenarios.json`:

   ```json
   {
     "name": "Human-readable name",
     "description": "What this workload is and why the result is expected.",
     "label_filter": "test-case-label",
     "path": "networking/<test-name>",
     "manifest": "<workload>.yaml",
     "expected_result": "passed|failed",
     "min_noncompliant": 2,
     "min_compliant": 2,
     "compliant_contains": "uses TLS",
     "output_dir": "<workload>-results"
   }
   ```

6. Regenerate the coverage inventory and commit it:

   ```bash
   python3 .github/actions/test-scenarios/generate-readme.py
   ```

   `make check-scenario-readme`, part of `make lint`, fails if it is stale.

### Entry fields

| Field | Required | Meaning |
|---|---|---|
| `name` | yes | Shown in the runner output and in `README.md`. |
| `description` | yes | Rendered in the `README.md` scenario list. Say what the workload is and why the result is expected. |
| `path` | yes | Scenario directory, relative to `test-scenarios/`. |
| `manifest` | yes | File in `<path>/manifests/`, passed to `deploy.sh`. |
| `output_dir` | yes | Results directory. Must be unique across all entries. |
| `label_filter` | single-test | The catalog test ID. Becomes `--label-filter`. |
| `expected_result` | single-test | `passed` or `failed`; compared to `.claim.results[<id>].state`. |
| `tests` | batch | Test ID to expectations map; replaces `label_filter` and the expectation fields. See [batches/README.md](batches/README.md). |
| `min_noncompliant` | no | Minimum `NonCompliantObjectsOut` entries. |
| `min_compliant` | no | Minimum `CompliantObjectsOut` entries. |
| `compliant_contains` | no | Only count compliant objects whose `ObjectFieldsValues` contain this string. |
| `noncompliant_contains` | no | Only count non-compliant objects containing this string, **and** fail if any non-compliant object does not. Implies `min_noncompliant: 1`. |
| `fixed_namespace` | no | `true` means the entry manages its own namespaces: the runner uses `certsuite-config.yaml` unchanged and does not export `SCENARIO_NAMESPACE`. Only the operator batches need this. |
| `lane` | no | What the entry may run beside; defaults to `namespaced`. See [Lanes](#lanes). |

A matching state alone is weak evidence: a test can report `failed` for a
reason that has nothing to do with the workload. Assert an object whenever the
check produces one. `generate-readme.py` enforces this for every batch test and
for the IDs in its `OBJECT_ASSERTION_IDS` set.

## Lanes

`run-scenarios.sh --jobs N` runs up to N lanes at once. Entries in the same lane
always run serially; entries in different lanes overlap. `--jobs 1` (the
default, and what you get without the flag) is the old strictly serial loop.

The constrained lanes are handed out whole. The `namespaced` lane is the one
that is safe to overlap freely, so its entries are spread round-robin over as
many workers as are left, which stops one long lane from leaving the others
idle. That is why `--jobs 4` runs seven units: `cluster-scoped`, `node`,
`operator` and `namespaced-1` … `namespaced-4`.

| Lane | Used by | Why it is kept apart |
|---|---|---|
| `namespaced` | everything that does not name a lane | Touches only its own namespace, so it is safe beside any other lane. |
| `cluster-scoped` | `batches/pod-spec` (RuntimeClass, ClusterRoleBinding, StorageClasses, PVs), `batches/crds` (one CRD), `access-control/namespace-crd-versions` and `access-control/namespace-crd-versions-noncompliant` (one CRD each, each with a retired API version) and `batches/service-mesh` (an `istio-system` namespace and `istiod` Deployment) | Creates objects that outlive the namespace and are visible to every other run. `batches/service-mesh` creates the two objects autodiscovery looks for when deciding a mesh is installed, which flips a discovered flag for the whole run. |
| `node` | `batches/workload-controllers` and `batches/lifecycle` | Judges cluster-wide node and zone counts, and pod recreation cordons and drains the nodes hosting the workloads under test, which also disrupts what the other lanes run there. `batches/lifecycle` also creates a CRD, which its cleanup deletes before the namespace. |
| `operator` | `batches/operators/*` | Certsuite's operator checks read every CSV, Subscription and CatalogSource in the cluster; and OLM installs are slow and CPU-hungry. |

Adding an entry without a `lane` puts it in `namespaced`, which is the right
answer for a scenario that only creates objects in its own namespace. If a new
scenario creates a CRD, a ClusterRoleBinding, a StorageClass, a PV, a
RuntimeClass or a Namespace, give it its own lane (or an existing one that owns
that kind of object) and say why in the table above.

The probe DaemonSet is created by whichever certsuite run gets there first, so
the runner gives the first lane a head start and starts the rest once
`cnf-suite/certsuite-probe` exists. That keeps a concurrent first run from
racing on the DaemonSet creation.

Lane output is written to a per-lane log and printed under
`=== Lane '<name>' output ===` when that lane finishes, and each lane's timings
are merged into the single table.

## What the runner does per scenario

1. `deploy.sh <manifest>`.
2. `./certsuite run --label-filter=<joined test IDs> --config-file=<path>/manifests/certsuite-config.yaml --output-dir=<output_dir> --create-xml-junit-file --omit-artifacts-zip-file --log-level=<SMOKE_TESTS_LOG_LEVEL> --cleanup-probe=false --quiet`. The `--quiet` flag keeps certsuite's own
   console output out of the job log; everything it suppresses is still written
   to `<output_dir>/certsuite.log`.
3. Validate each test ID in the entry against the claim.
4. On failure, grep `certsuite.log` for the test IDs and probe-related lines.
5. `cleanup.sh <output_dir>`, which also runs if the script exits early.

Per-phase timings are written to `--timings-file` (default
`${TMPDIR:-/tmp}/scenario-timings.tsv`) and printed as a table at the end.

### Namespaces are per entry

Several entries share one scenario directory, and therefore one base namespace
in `certsuite-config.yaml`. The runner derives a namespace per entry from that
base plus the entry's `output_dir` (for example
`lifecycle-probe-smoke-ns-probe-violations`) and renders a copy of the config
with the new name. The base is kept as a prefix because some checks match on it
— `access-control-namespace` treats `istio-` as a reserved prefix.

That gives each entry a clean namespace even when the previous entry's namespace
is still terminating. `generate-readme.py` fails if two entries derive the same
namespace. Scripts must read `SCENARIO_NAMESPACE` rather than hardcoding the
namespace; entries with `fixed_namespace` are the exception.

### Cleanup does not wait for finalization

`cleanup-workload.sh` deletes the namespace with `--wait=false` and appends it
to `$SCENARIO_REAP_FILE`. The runner waits for every registered namespace in
parallel once the suite is done, so a namespace that takes 10s to finalize costs
the suite 10s in total rather than 10s per scenario.

Pass `--wait` as the third argument when cluster-scoped objects cannot be
deleted until the namespace is gone. `batches/pod-spec` does this for the
entries whose statically bound PVs stay wedged on the `pvc-protection`
finalizer; it checks for PVCs first so its volume-free entries keep the fast
path.

### The probe DaemonSet is shared between scenarios

The runner passes `--cleanup-probe=false`. The `certsuite-probe` DaemonSet in
namespace `cnf-suite` is created by the first scenario and then reused: deleting
it costs about 40s of waiting for pods to terminate, and recreating it about
11s, which dominated the total runtime. A scenario must therefore not assume a
freshly created probe, and must not delete the `cnf-suite` namespace in its
`cleanup.sh`.

## Cluster preconditions

The scenarios run against the partner Kind cluster created by
`.github/actions/setup-partner-cluster`. Three properties of that cluster are load
bearing:

- **At least two nodes labelled `node-role.kubernetes.io/worker`.** The partner
  workload's `make install` applies these labels. Without them `GetWorkerCount`
  returns fewer than two workers and `lifecycle-pod-high-availability`,
  `lifecycle-pod-scheduling` and `lifecycle-topology-spread-constraint` **skip**
  rather than pass or fail, so their scenarios report a state mismatch.
- **Two schedulable nodes.** Scenarios that spread replicas with required
  hostname anti-affinity will not become Available on a single-node cluster.
- **Dual-stack pod and service networks.** The golden pod-spec Service sets
  `ipFamilyPolicy: PreferDualStack` and `networking-dual-stack-service` asserts
  it ends up with two cluster IPs. On a single-stack cluster the check errors
  out instead.

Running the suite against your own cluster means reproducing all three.

## Debugging a failing scenario

```bash
# Run everything, as CI does. Requires jq, kubectl and a ./certsuite built from
# the current checkout.
make build-certsuite-tool
.github/actions/test-scenarios/run-scenarios.sh --jobs 4

# Serial, which is the same thing with --jobs 1. Use it to tell a real failure
# apart from one caused by lanes overlapping.
.github/actions/test-scenarios/run-scenarios.sh

# Raise the log level.
SMOKE_TESTS_LOG_LEVEL=debug .github/actions/test-scenarios/run-scenarios.sh

# Run one scenario group: matches against the entry's name or path.
.github/actions/test-scenarios/run-scenarios.sh --only networking/tls-minimum-version
```

If a scenario fails in a way you cannot reproduce, check that the binary
matches the checkout — a stale `./certsuite` from another branch produces
results that do not match the source:

```bash
./certsuite version   # the git commit must be an ancestor of HEAD
```

A workload that is not Ready is left out of certsuite discovery, so an expected
failure silently turns into a pass. Every `deploy.sh` must wait for all of its
workloads.

A scenario whose workload is *expected* to be broken inverts the wait: deploy
waits for the broken state, not for Ready. The `batches/operators/hugepages`
entry deploys a pod that can never schedule (it requests hugepages the Kind
workers do not expose), so its `deploy.sh` waits for
`--for=condition=PodScheduled=false` — the pod existing but unschedulable is
what the scenario asserts, and waiting for anything stronger would fail the
deploy step.
