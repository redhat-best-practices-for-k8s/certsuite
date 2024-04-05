<!-- markdownlint-disable line-length no-bare-urls -->
# Runtime environment variables

To run the test suite, some runtime environment variables are to be set.

## OCP >=4.12 Labels

The following labels need to be added to your default namespace in your cluster
if you are running OCP >=4.12:

```shell
pod-security.kubernetes.io/enforce: privileged
pod-security.kubernetes.io/enforce-version: latest
```

You can manually label the namespace with:

```shell
oc label namespace/default pod-security.kubernetes.io/enforce=privileged
oc label namespace/default pod-security.kubernetes.io/enforce-version=latest
```

## Disable intrusive tests

To skip intrusive tests which may disrupt cluster operations, issue the
following:

```shell
export TNF_NON_INTRUSIVE_ONLY=true
```

Likewise, to enable intrusive tests, set the following:

```shell
export TNF_NON_INTRUSIVE_ONLY=false
```

Intrusive tests are enabled by default.

## Preflight Integration

When running the `preflight` suite of tests, there are a few environment variables that
will need to be set:

`PFLT_DOCKERCONFIG` is a required variable for running the preflight test suite. This
provides credentials to the underlying preflight library for being able to pull/manipulate
images and image bundles for testing.

When running as a container, the docker config is mounted to the container via volume mount.

When running as a standalone binary, the environment variables are consumed directly from your local machine.

See more about this variable [here](https://github.com/redhat-openshift-ecosystem/openshift-preflight/blob/main/docs/CONFIG.md).

`TNF_ALLOW_PREFLIGHT_INSECURE` (default: false) is required set to `true` if you are running
against a private container registry that has self-signed certificates.

## Disconnected environment

In a disconnected environment, only specific versions of images are mirrored to
the local repo. For those environments, the partner pod image
`quay.io/testnetworkfunction/cnf-test-partner` and debug pod image
`quay.io/testnetworkfunction/debug-partner` should be mirrored and
`TNF_PARTNER_REPO` should be set to the local repo, e.g.:

```shell
export TNF_PARTNER_REPO=registry.dfwt5g.lab:5000/testnetworkfunction
```

Note that you can also specify the debug pod image to use with `SUPPORT_IMAGE`
environment variable, default to `debug-partner:5.0.6`.
