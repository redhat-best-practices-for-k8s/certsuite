# Networking Test Scenarios

Self-contained scenario tests for the networking test suite. Each scenario deploys a single workload with a known configuration, runs the certsuite against it, and validates the expected result.

## Scenarios

### tls-minimum-version

Validates that the `networking-tls-minimum-version` test correctly checks services against the cluster's TLS security profile. On non-OpenShift clusters (including CI), the default Intermediate profile is used (min TLS 1.2). Each workload is tested in isolation.

| Scenario | Manifest | TLS Config | Expected Test Result |
|---|---|---|---|
| TLS 1.3 Only (compliant) | `tls13-only.yaml` | TLS 1.3 only | `passed` |
| TLS 1.2 Allowed (compliant) | `tls12-allowed.yaml` | TLS 1.2 + 1.3, Intermediate ciphers | `passed` |
| Plain HTTP (compliant) | `plain-http.yaml` | No TLS | `passed` |

### unsecured-container-ports

Validates that `networking-unsecured-container-ports` probes plaintext TCP
listeners on every node, not only the node of a randomly chosen probe pod.
Smoke Kind clusters have two workers; required pod anti-affinity spreads two
replicas across them. The runner also requires at least two non-compliant
objects so one same-node plaintext hit plus a cross-node timeout cannot pass
as a generic FAIL.

| Scenario | Manifest | Workload | Expected Test Result |
|---|---|---|---|
| Plain HTTP spread across nodes | `plain-http-spread.yaml` | Two replicas, plaintext :8080, hostname anti-affinity | `failed` (min 2 non-compliant objects) |
| TLS spread across nodes | `tls-spread.yaml` | Two replicas, TLS :8443, hostname anti-affinity | `passed` (min 2 objects containing "uses TLS") |

## Adding a new scenario

1. Create a directory under the appropriate test suite (e.g., `networking/<test-name>/`).
2. Add a `deploy.sh` that accepts a manifest filename as `$1`, creates the namespace, applies shared resources, and deploys the single workload.
3. Add a `cleanup.sh` to tear down resources.
4. Place Kubernetes manifests and certsuite config in a `manifests/` subdirectory. Each manifest should define a single workload (Deployment + Service).
5. Add one entry per workload to `../scenarios.json`:
   ```json
   {
     "name": "Human-readable name",
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
   `min_noncompliant` and `min_compliant` are optional. `compliant_contains`
   limits the compliant-object count to entries whose details include that
   string (so an unreachable port cannot satisfy a TLS pass).
   The runner script (`../run-scenarios.sh`) picks up all entries automatically. Validation is data-driven: the runner checks that the test state in `claim.json` matches `expected_result`.
