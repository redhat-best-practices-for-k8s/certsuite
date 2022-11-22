# Runtime environment variables

To run the test suite, some runtime environment variables are to be set.

## OCP >=4.12 Labels
The following labels need to be added to your default namespace in your cluster if you are running OCP >=4.12:

```shell
pod-security.kubernetes.io/enforce: privileged
pod-security.kubernetes.io/enforce-version: latest
```

You can manually label the namespace with:
```shell
oc label namespace/default  pod-security.kubernetes.io/enforce=privileged
oc label namespace/default  pod-security.kubernetes.io/enforce-version=latest
```

## Disable intrusive tests
To skip intrusive tests which may disrupt cluster operations, issue the following:

```shell
export TNF_NON_INTRUSIVE_ONLY=true
```

Likewise, to enable intrusive tests, set the following:

```shell
export TNF_NON_INTRUSIVE_ONLY=false
```

## Disconnected environment
In a disconnected environment, only specific versions of images are mirrored to the local repo. For those environments,
the partner pod image `quay.io/testnetworkfunction/cnf-test-partner` and debug pod image `quay.io/testnetworkfunction/debug-partner` should be mirrored
and `TNF_PARTNER_REPO` should be set to the local repo, e.g.:

```shell
export TNF_PARTNER_REPO="registry.dfwt5g.lab:5000/testnetworkfunction"
```

Note that you can also specify the debug pod image to use with `SUPPORT_IMAGE` environment variable, default to `debug-partner:latest`.
