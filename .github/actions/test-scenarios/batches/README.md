# Batch scenarios

A batch scenario checks many catalog tests in a single certsuite run. One
`--label-filter` carries every test ID in the entry, and the runner validates
each test separately against the same claim.

Batching exists because the per-run overhead dominates: 34 tests as 34
single-test entries cost 34 deploy/run/cleanup cycles, and as one batch they
cost one. The scenarios in here cover most of the pod-spec and workload
controller catalog.

See [../CONTRIBUTING.md](../CONTRIBUTING.md) for the scenario framework and the
full `scenarios.json` field reference.

## Groups

| Group | Namespace | Pod label | Covers |
|---|---|---|---|
| `pod-spec` | `pod-spec-smoke-ns` | `scenario-pod-spec: target` | Pod and container spec checks: capabilities, host namespaces, security context, resources, lifecycle hooks, reserved ports, CPU pinning and isolation |
| `workload-controllers` | `controllers-smoke-ns` | `scenario-controllers: target` | Checks that judge the controller rather than the pod: high availability, PodDisruptionBudget, topology spread, AffinityRequired |

Each group has its own namespace and label so one group's workloads are invisible
to the other's runs.

### pod-spec entries

| Entry | Manifest | Expectation |
|---|---|---|
| Golden workloads | `golden.yaml` | 34 tests `passed`. A golden Deployment, a Guaranteed whole-CPU variant with the `certsuite-smoke` RuntimeClass and the CPU/IRQ load-balancing annotations, and a ClusterIP Service. |
| One violation per workload | `violations.yaml` | 34 tests `failed`, each on its own workload. |
| Security context without fsGroup | `bad-security-context.yaml` | `access-control-security-context` `failed`. |
| Shared process namespace with SYS_PTRACE | `ptrace.yaml` | `access-control-sys-ptrace-capability` `passed`. |

### workload-controllers entries

| Entry | Manifest | Expectation |
|---|---|---|
| Golden workloads | `golden.yaml` | 4 tests `passed`. Two-replica Deployments with PodDisruptionBudgets: one with pod anti-affinity, one with hostname and zone topology spread constraints, one labelled `AffinityRequired` with node affinity. |
| One violation per workload | `violations.yaml` | 4 tests `failed`: one replica, a missing PodDisruptionBudget, a hostname-only spread constraint, and an `AffinityRequired` workload that keeps pod anti-affinity. |

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

`make check-scenario-readme`, part of `make lint`, fails if the checked-in YAML
no longer matches the generator.

`generate-manifests.py` holds one golden pod spec and a table mapping each test
ID to a workload name and the single mutation that breaks it, so a reviewer
reads the one-line difference instead of 2000 lines of YAML. To add a test to a
batch:

1. Add the test ID, workload name and mutation to `POD_SPEC_VIOLATIONS` or
   `CONTROLLER_VIOLATIONS`.
2. Confirm the golden template already satisfies the new check. If it does not,
   change the template and re-verify the other entries, because every workload
   inherits it.
3. Regenerate the manifests.
4. Add the test ID to both the golden and the violations entry in
   `scenarios.json`, with `min_compliant` and `noncompliant_contains`.
5. Regenerate the README: `python3 .github/actions/test-scenarios/generate-readme.py`.

## Cluster-scoped objects

`pod-spec` applies a cluster-scoped `certsuite-smoke` RuntimeClass for the
CPU-pinned workloads. Deleting the namespace does not remove it, so
`pod-spec/cleanup.sh` deletes it explicitly. Any future cluster-scoped object
needs the same treatment.
