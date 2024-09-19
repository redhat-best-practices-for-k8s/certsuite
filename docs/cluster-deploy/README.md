<!-- markdownlint-disable line-length no-bare-urls no-emphasis-as-heading -->
# How to deploy the Cert Suite App inside a Kubernetes/Openshift cluster

This is a developer's guide to deploy a Pod in a kubernetes/Openshift cluster that runs the CNF Cert Suite app inside.

This folder contains two files:

* [./certsuite.yaml](certsuite.yaml)
* [./kustomization.yaml](kustomization.yaml)

## certsuite.yaml

This file contains all the kubernetes templates for deploying the CNF Cert Suite inside a Pod named "certsuite" in a namespace also named "certsuite". In order to deploy the pod, just write:

```console
oc apply -f k8s/certsuite.yaml
namespace/certsuite created
clusterrole.rbac.authorization.k8s.io/certsuite-cr created
clusterrolebinding.rbac.authorization.k8s.io/certsuite-crb created
configmap/certsuite-config created
secret/certsuite-preflight-dockerconfig created
pod/certsuite created
```

The first thing in that yaml is the namespace, so it's the first resource that will be created in the cluster. Then, a cluster role and its cluster role binding will be created. This cluster role is needed because the CNF Cert Suite needs access to all the resources in the whole cluster.

Then, there's a configMap with the whole config (certsuite_config.yaml) that will be used by the pod to create the certsuite_config.yaml file inside a volume folder. Also, there's a secret with the preflight's dockerconfig file content that will also be used by the CNF Cert Suitep pod.

The Cert Suite pod is the last resource defined in the certsuite.yaml file. It has only one container that uses the [quay.io/redhat-best-practices-for-k8s/certsuite:latest](latest) tag of the CNF Cert Suite. The command slice of this container has a hardcoded labels to run as many test cases as possible, excluding the intrusive ones.

## kustomization.yaml

This kustomization file allows the deployment of the CNF Cert Suite using this command:

```console
oc kustomize k8s/ | oc apply -f -
```

The `kustomization` tool used by `oc` will parse the content of the [./kustomization.yaml](kustomization.yaml) file, which consists of a set of "transformers" over the resources defined in [./certsuite.yaml](certsuite.yaml).

By default, that command will deploy the CNF Cert Suite Pod without any mutation: it will be deployed in the same namespace and with the same configuration than using the `oc apply -f k8s/certsuite.yaml`.

But there are the three example of modifications included in [./kustomization.yaml](kustomization.yaml) that can be used out of the box that can be handy:

1. The namespace and the prefix/suffix of each resource's name. By default, the [./certsuite.yaml](certsuite.yaml) uses the namespace "certsuite" to deploy all the reources (except the cluster role and the cluster role binding), but this can be changed uncommenting the line that starts with `namespace:`. It's highly recommended to uncomment at least one of suffixName/prefixName so unique cluster role & cluster role-bindings can be created for each CNF Cert Pod. This way, you could run more than one CNF Cert Pod in the same cluster!.
2. The (ginkgo) labels expression, in case you want to run different test cases. Uncomment the object that starts with "patches:". The commented example changes the command to use the "preflight" label only.
3. The value of the CERTSUITE_NON_INTRUSIVE_ONLY env var. Uncomment the last object that starts with "patches:". The commented example changes the CERTSUITE_NON_INTRUSIVE_ONLY to false, so all the intrusive TCs will run in case the lifecycle TCs are selected to run by the appropriate labels.

In case both (1) and (2) wants to be used, just create a list of patches like this:

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
          ./run-cnf-suites.sh -l 'preflight' ; sleep inf
  - target:
      version: v1
      kind: Pod
      name: certsuite
    patch: |
      - op: replace
        path: /spec/containers/0/env/0/value
        value: false
```
