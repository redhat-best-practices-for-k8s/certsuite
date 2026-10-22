<!-- markdownlint-disable line-length -->
# Security Policy

## Supported versions

Security fixes land on `main` and in the most recent tagged release. Older
releases are not maintained; upgrade to the latest release before reporting an
issue that may already be fixed.

- Latest tagged release: supported.
- `main`: supported, as the development branch.
- Older releases: not supported.

## Reporting a vulnerability

Do not open a public issue for a security problem.

Report it privately through GitHub:
[Security tab → Report a vulnerability](https://github.com/redhat-best-practices-for-k8s/certsuite/security/advisories/new).
If that form is unavailable, contact the
[CNF Cert team](https://github.com/orgs/redhat-best-practices-for-k8s/teams/cnfcert/members)
directly rather than filing a public issue.

Please include:

- The affected version, commit or container image tag.
- The command line or configuration used, and the cluster type (OpenShift or
  Kubernetes) and version.
- Reproduction steps, and the impact you believe the issue has.
- Any suggested fix, if you have one.

## Scope

In scope: the `certsuite` binary, its container image, its CI workflows, and the
claim file it writes.

Out of scope: the workloads under test, the offline certification database
fetched by `make get-db`, and findings that a test reports about a cluster — a
non-compliant result from a test case is the suite working as intended, not a
vulnerability in the suite.

## Disclosure

We will acknowledge a report within five business days, work with you on a
timeline, and credit you in the advisory unless you ask us not to.
