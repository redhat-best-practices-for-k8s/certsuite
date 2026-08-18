<!-- markdownlint-disable line-length no-bare-urls no-emphasis-as-heading -->
# How to deploy the Cert Suite App inside a Kubernetes/Openshift cluster

This is a developer's guide to deploy a Pod in a kubernetes/Openshift cluster that runs the Cert Suite app inside.

This folder contains two files:

* [./certsuite.yaml](certsuite.yaml)
* [./kustomization.yaml](kustomization.yaml)

## certsuite.yaml

This file contains all the kubernetes templates for deploying the Cert Suite inside a Pod named "certsuite" in a namespace also named "certsuite". From the repository root:

```console
oc apply -f docs/cluster-deploy/certsuite.yaml
namespace/certsuite created
clusterrole.rbac.authorization.k8s.io/certsuite-cr created
clusterrolebinding.rbac.authorization.k8s.io/certsuite-crb created
configmap/certsuite-config created
secret/certsuite-preflight-dockerconfig created
pod/certsuite created
```

Or from this directory: `oc apply -f certsuite.yaml`.

The first resource in that yaml is the namespace. Then a cluster role and cluster role binding are created. This cluster role is needed because the Cert Suite needs access to resources across the cluster.

A ConfigMap holds `certsuite_config.yml` and is mounted at `/usr/certsuite/config`. A secret holds the preflight dockerconfig and is mounted at `/usr/certsuite/config/preflight`. The pod command passes `--config-file` and `--preflight-dockerconfig` explicitly (certsuite does not read `CERTSUITE_CONFIGURATION_PATH` or `CERTSUITE_ALLOW_PREFLIGHT_INSECURE`).

The Cert Suite pod uses the [quay.io/redhat-best-practices-for-k8s/certsuite:latest](https://quay.io/repository/redhat-best-practices-for-k8s/certsuite) image. The command runs a broad label filter with `--intrusive=false`. In-cluster kubeconfig is detected automatically.

## kustomization.yaml

This kustomization file allows the deployment of the Cert Suite using this command from the repository root:

```console
oc kustomize docs/cluster-deploy/ | oc apply -f -
```

Or from this directory: `oc kustomize . | oc apply -f -`.

The `kustomization` tool used by `oc` will parse the content of the [./kustomization.yaml](kustomization.yaml) file, which consists of a set of "transformers" over the resources defined in [./certsuite.yaml](certsuite.yaml).

By default, that command will deploy the Cert Suite Pod without any mutation: it will be deployed in the same namespace and with the same configuration as `oc apply -f docs/cluster-deploy/certsuite.yaml`.

There are three example modifications in [./kustomization.yaml](kustomization.yaml):

1. The namespace and the prefix/suffix of each resource's name. By default, [./certsuite.yaml](certsuite.yaml) uses the namespace "certsuite" (except the cluster role and cluster role binding). Uncomment the line that starts with `namespace:`. Uncomment at least one of namePrefix/nameSuffix so unique cluster role and cluster role-bindings can be created for each CertSuite Pod. This way, you can run more than one CertSuite Pod in the same cluster.
2. The label expression, in case you want to run different test cases. Uncomment the object that starts with "patches:". The commented example changes the command to use the "preflight" label only.
3. Configuring the `--intrusive` flag. Uncomment the last object that starts with "patches:". The commented example removes `--intrusive=false`, so intrusive test cases will run if lifecycle tests are selected.

If both (1) and (2) are needed, create a list of patches like this:

```console
patches:
  - target:
      version: v1
      kind: Pod
      name: certsuite
    patch: |
      - op: replace
        path: /spec/containers/0/args/1
        value: |
          certsuite run --config-file=/usr/certsuite/config/certsuite_config.yml --preflight-dockerconfig=/usr/certsuite/config/preflight/preflight_dockerconfig.json -l 'preflight' ; sleep inf
  - target:
      version: v1
      kind: Pod
      name: certsuite
    patch: |
      - op: replace
        path: /spec/containers/0/args/1
        value: |
          certsuite run --config-file=/usr/certsuite/config/certsuite_config.yml --preflight-dockerconfig=/usr/certsuite/config/preflight/preflight_dockerconfig.json -l '!affiliated-certification-container-is-certified-digest && !access-control-security-context' ; sleep inf
```
