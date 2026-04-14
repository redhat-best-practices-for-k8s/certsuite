<!-- markdownlint-disable line-length no-bare-urls -->
# Running the Certsuite in a Disconnected Environment

This guide explains how to run the certsuite against an air-gapped or
disconnected OpenShift cluster where nodes cannot pull container images from
external registries.

For general information on disconnected OpenShift environments, see the
[Red Hat disconnected environments documentation](https://docs.redhat.com/en/documentation/openshift_container_platform/4.21/html/disconnected_environments/about-installing-oc-mirror-v2).
The [oc-mirror GitHub repository](https://github.com/openshift/oc-mirror)
contains additional examples and source code.

## The Problem

The certsuite deploys a privileged DaemonSet called `certsuite-probe` onto every
node in the target cluster. This DaemonSet runs commands on the host that many
test cases depend on (e.g., ping, platform checks). By default the probe image
is pulled from `quay.io/redhat-best-practices-for-k8s/certsuite-probe:latest`.
In a disconnected cluster the nodes cannot reach quay.io, so the DaemonSet fails
to start and the certsuite cannot run.

The solution is to mirror the required images into a container registry that is
accessible from within the disconnected cluster, then tell the certsuite (or the
cluster itself) where to find them.

## Images Required

| Image | Purpose | When needed |
| --- | --- | --- |
| `quay.io/redhat-best-practices-for-k8s/certsuite-probe:latest` | Privileged DaemonSet for running platform-level test commands on cluster nodes | Always |
| `quay.io/redhat-best-practices-for-k8s/certsuite:latest` | Certsuite container image | Only when running the certsuite as a pod inside the cluster (see [cluster-deploy](cluster-deploy/README.md)) |
| `quay.io/redhat-best-practices-for-k8s/oct:latest` | Offline certification database for checking container, operator, and Helm chart certification status | Optional but recommended for fully disconnected environments |

## Approach 1: oc-mirror v2 (Recommended)

[oc-mirror](https://github.com/openshift/oc-mirror) is the Red Hat-supported
tool for mirroring container images to disconnected registries. It uses an
`ImageSetConfiguration` to declare which images to mirror and generates the
cluster resources (IDMS/ITMS) needed to redirect image pulls to the mirror.

A sample `ImageSetConfiguration` is provided in this repository at
[`examples/disconnected/imageset-config.yaml`](../examples/disconnected/imageset-config.yaml).

### Prerequisites

- The `oc-mirror` plugin v2 installed on your connected workstation.
  See the
  [oc-mirror documentation](https://docs.redhat.com/en/documentation/openshift_container_platform/4.17/html/disconnected_environments/mirroring-images-for-a-disconnected-installation-using-the-oc-mirror-plugin-v2)
  for installation instructions.
- A mirror registry accessible from both the connected workstation (for pushing)
  and the disconnected cluster nodes (for pulling).
- Registry credentials configured in `${XDG_RUNTIME_DIR}/containers/auth.json`
  (or `~/.docker/config.json`) for both the source (quay.io) and destination
  registries.

### Step 1: Create the ImageSetConfiguration

Copy the sample configuration or create your own. At minimum, include the probe
image:

```yaml
kind: ImageSetConfiguration
apiVersion: mirror.openshift.io/v2alpha1
mirror:
  additionalImages:
    - name: quay.io/redhat-best-practices-for-k8s/certsuite-probe:latest
    - name: quay.io/redhat-best-practices-for-k8s/certsuite:latest
    - name: quay.io/redhat-best-practices-for-k8s/oct:latest
```

Remove any images you do not need. The probe image is always required.

### Step 2: Mirror images to disk

On a workstation with internet access, run:

```shell
oc-mirror -c imageset-config.yaml file:///path/to/output-dir --v2
```

This downloads the images and writes them to disk as a portable archive.

### Step 3: Transfer the archive

Copy the output directory to the disconnected environment using whatever
transport is available (USB drive, secure file transfer, etc.).

### Step 4: Load images into the mirror registry

On a host that can reach the mirror registry, run:

```shell
oc-mirror -c imageset-config.yaml --from file:///path/to/output-dir \
  docker://<mirror-registry> --v2
```

Replace `<mirror-registry>` with the hostname and port of your mirror registry
(e.g., `registry.example.com:5000`).

### Step 5: Apply the generated cluster resources

oc-mirror generates `ImageDigestMirrorSet` (IDMS) and/or `ImageTagMirrorSet`
(ITMS) custom resources that tell the cluster to redirect image pulls from the
original registry to the mirror. Apply them:

```shell
oc apply -f oc-mirror-workspace/results-*/cluster-resources/
```

Verify the resources were created:

```shell
oc get imagedigestmirrorset
oc get imagetagmirrorset
```

After applying these resources, the cluster nodes will automatically pull images
from the mirror registry when a pod references the original quay.io path. This
means the certsuite can run with its default configuration and the probe
DaemonSet will pull its image from the mirror transparently.

### Step 6: Run the certsuite

With IDMS/ITMS in place, the certsuite can be run normally:

```shell
certsuite run -l <label-filter> -c <config-file> -k <kubeconfig> -o <output-dir>
```

If you prefer to be explicit, you can also override the probe image path
directly:

```shell
certsuite run \
  --certsuite-probe-image=<mirror-registry>/redhat-best-practices-for-k8s/certsuite-probe:latest \
  -l <label-filter> -c <config-file> -k <kubeconfig> -o <output-dir>
```

## Approach 2: Manual Mirroring with skopeo

If you already have a mirror registry configured and do not want to use
oc-mirror, you can copy images manually with
[skopeo](https://github.com/containers/skopeo).

### Step 1: Copy images to the mirror registry

On a connected workstation, copy each image:

```shell
skopeo copy \
  docker://quay.io/redhat-best-practices-for-k8s/certsuite-probe:latest \
  docker://<mirror-registry>/redhat-best-practices-for-k8s/certsuite-probe:latest

skopeo copy \
  docker://quay.io/redhat-best-practices-for-k8s/certsuite:latest \
  docker://<mirror-registry>/redhat-best-practices-for-k8s/certsuite:latest

skopeo copy \
  docker://quay.io/redhat-best-practices-for-k8s/oct:latest \
  docker://<mirror-registry>/redhat-best-practices-for-k8s/oct:latest
```

If the connected workstation cannot reach the mirror registry directly, use
`skopeo copy` to save images to a directory or tarball, transfer them, then push
from the disconnected side:

```shell
# On connected host
skopeo copy \
  docker://quay.io/redhat-best-practices-for-k8s/certsuite-probe:latest \
  dir:///path/to/certsuite-probe

# Transfer /path/to/certsuite-probe to the disconnected environment

# On disconnected host
skopeo copy \
  dir:///path/to/certsuite-probe \
  docker://<mirror-registry>/redhat-best-practices-for-k8s/certsuite-probe:latest
```

### Step 2: Run the certsuite with the mirrored probe image

Unlike oc-mirror, skopeo does not generate IDMS/ITMS resources. You must tell
the certsuite where to find the probe image using the `--certsuite-probe-image`
flag:

```shell
certsuite run \
  --certsuite-probe-image=<mirror-registry>/redhat-best-practices-for-k8s/certsuite-probe:latest \
  -l <label-filter> -c <config-file> -k <kubeconfig> -o <output-dir>
```

Alternatively, set the `SUPPORT_IMAGE` environment variable:

```shell
export SUPPORT_IMAGE=<mirror-registry>/redhat-best-practices-for-k8s/certsuite-probe:latest
certsuite run -l <label-filter> -c <config-file> -k <kubeconfig> -o <output-dir>
```

## Offline Certification Database

The certsuite checks whether container images, operators, and Helm charts are
Red Hat certified. In a disconnected environment, the certsuite cannot reach the
Red Hat catalog APIs. The [OCT tool](https://github.com/redhat-best-practices-for-k8s/oct)
provides an offline database for these checks.

### Creating the offline database

On a connected host, run the OCT container to dump the database:

```shell
mkdir -p offline-db

docker run -v $(pwd)/offline-db:/tmp/dump:Z \
  --user $(id -u):$(id -g) \
  --env OCT_DUMP_ONLY=true \
  quay.io/redhat-best-practices-for-k8s/oct:latest
```

If you have already mirrored the OCT image, pull it from your mirror registry
instead.

### Using the offline database

Transfer the `offline-db` directory to your disconnected jumpbox, then pass it
to the certsuite:

```shell
certsuite run \
  --offline-db=/path/to/offline-db \
  --certsuite-probe-image=<mirror-registry>/redhat-best-practices-for-k8s/certsuite-probe:latest \
  -l <label-filter> -c <config-file> -k <kubeconfig> -o <output-dir>
```

## Verification

### Verify images are in the mirror registry

```shell
skopeo inspect docker://<mirror-registry>/redhat-best-practices-for-k8s/certsuite-probe:latest
```

### Verify IDMS/ITMS is active (oc-mirror approach only)

```shell
oc get imagedigestmirrorset
oc get imagetagmirrorset
```

### Verify the probe DaemonSet starts

Run the certsuite with `--cleanup-probe=false` to keep the probe running after
tests complete, then verify the DaemonSet pods are running:

```shell
oc get daemonset -n certsuite
oc get pods -n certsuite
```
