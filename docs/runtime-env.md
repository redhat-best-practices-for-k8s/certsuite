<!-- markdownlint-disable line-length no-bare-urls -->
# Runtime environment variables

To run the test suite, some runtime environment variables are to be set.

## OCP >=4.12 Labels

The following labels need to be added to the namespace where the probe DaemonSet
is deployed (default `cnf-suite`; see `probeDaemonSetNamespace` in the config
file) if you are running OCP >=4.12:

```shell
pod-security.kubernetes.io/enforce: privileged
pod-security.kubernetes.io/enforce-version: latest
```

You can manually label the namespace with:

```shell
oc label namespace/cnf-suite pod-security.kubernetes.io/enforce=privileged
oc label namespace/cnf-suite pod-security.kubernetes.io/enforce-version=latest
```

## Preflight Integration

When running the `preflight` suite of tests, pass the dockerconfig path to
certsuite with `--preflight-dockerconfig`. Certsuite skips the preflight suite
if that flag is empty.

`PFLT_DOCKERCONFIG` is still consumed by the underlying
[Preflight library](https://github.com/redhat-openshift-ecosystem/openshift-preflight/blob/main/docs/CONFIG.md)
for pulling images. Set it to the same file when running as a standalone binary.
When running as a container, mount the docker config and pass
`--preflight-dockerconfig` to the `certsuite run` command.

Allow insecure connections to a private registry with self-signed certificates
using `--allow-preflight-insecure` (default: false).

Override the probe DaemonSet image with `--certsuite-probe-image`. The default
is `quay.io/redhat-best-practices-for-k8s/certsuite-probe:v0.0.42` (see
`debugTag` in `version.json`).

## Client Timeout

`CERTSUITE_CLIENT_TIMEOUT` (default: `10s`) sets the timeout for Kubernetes API
client operations such as resource discovery and API group listing. Increase
this value when running against remote or high-latency clusters where the
default 10-second timeout causes failures during startup.

```shell
export CERTSUITE_CLIENT_TIMEOUT=30s
```

Accepts any valid Go duration string (e.g., `15s`, `1m`, `90s`).
