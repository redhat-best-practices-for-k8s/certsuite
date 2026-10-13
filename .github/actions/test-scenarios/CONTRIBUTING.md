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
  cleanup-workload.sh     shared cleanup: delete results dir and namespace
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
   it can be one line: `exec ../../deploy-workload.sh <namespace> <manifest-path>`.
3. Add `cleanup.sh` taking the results directory as `$1`, usually
   `exec ../../cleanup-workload.sh <namespace> "$1"`.
4. Put the manifests and a `certsuite-config.yaml` in `manifests/`. Each
   manifest defines a single Deployment named after the file, because
   `deploy-workload.sh` waits on `deployment/<manifest basename>`.
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

A matching state alone is weak evidence: a test can report `failed` for a
reason that has nothing to do with the workload. Assert an object whenever the
check produces one. `generate-readme.py` enforces this for every batch test and
for the IDs in its `OBJECT_ASSERTION_IDS` set.

## What the runner does per scenario

1. `deploy.sh <manifest>`.
2. `./certsuite run --label-filter=<joined test IDs> --config-file=<path>/manifests/certsuite-config.yaml --output-dir=<output_dir> --cleanup-probe=false`.
3. Validate each test ID in the entry against the claim.
4. On failure, grep `certsuite.log` for the test IDs and probe-related lines.
5. `cleanup.sh <output_dir>`, which also runs if the script exits early.

### The probe DaemonSet is shared between scenarios

The runner passes `--cleanup-probe=false`. The `certsuite-probe` DaemonSet in
namespace `cnf-suite` is created by the first scenario and then reused: deleting
it costs about 40s of waiting for pods to terminate, and recreating it about
11s, which dominated the total runtime. A scenario must therefore not assume a
freshly created probe, and must not delete the `cnf-suite` namespace in its
`cleanup.sh`.

## Cluster preconditions

The scenarios run against the partner Kind cluster created by
`.github/actions/setup-partner-cluster`. Two properties of that cluster are load
bearing:

- **At least two nodes labelled `node-role.kubernetes.io/worker`.** The partner
  workload's `make install` applies these labels. Without them `GetWorkerCount`
  returns fewer than two workers and `lifecycle-pod-high-availability`,
  `lifecycle-pod-scheduling` and `lifecycle-topology-spread-constraint` **skip**
  rather than pass or fail, so their scenarios report a state mismatch.
- **Two schedulable nodes.** Scenarios that spread replicas with required
  hostname anti-affinity will not become Available on a single-node cluster.

Running the suite against your own cluster means reproducing both.

## Debugging a failing scenario

```bash
# Run everything, as CI does. Requires jq, kubectl and a ./certsuite built from
# the current checkout.
make build-certsuite-tool
.github/actions/test-scenarios/run-scenarios.sh

# Raise the log level.
SMOKE_TESTS_LOG_LEVEL=debug .github/actions/test-scenarios/run-scenarios.sh
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
