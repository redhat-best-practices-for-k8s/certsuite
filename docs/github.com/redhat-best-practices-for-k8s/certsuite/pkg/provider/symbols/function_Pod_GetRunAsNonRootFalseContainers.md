Pod.GetRunAsNonRootFalseContainers`

| Aspect | Detail |
|--------|--------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider` |
| **Receiver type** | `Pod` – represents a Kubernetes pod object used by CertSuite to evaluate security settings. |
| **Signature** | `func (p Pod) GetRunAsNonRootFalseContainers(podSC map[string]bool) ([]*Container, []string)` |

---

### Purpose
Identifies every container in the pod that is **explicitly or implicitly** configured to run as a non‑root user *but* still has `runAsUser` set to `0`.  
Such a configuration violates the Kubernetes security best practice that requires containers with `runAsNonRoot: true` to run under a UID other than 0.

The function returns:

1. **`containers`** – a slice of pointers to `Container` objects that match the rule.
2. **`warnings`** – a slice of human‑readable messages explaining why each container failed the check (used by CertSuite’s reporting layer).

---

### Inputs

| Parameter | Type | Meaning |
|-----------|------|---------|
| `podSC` | `map[string]bool` | Map containing default security context values for the pod (`runAsNonRoot`, `runAsUser`). The map keys are the string names of the fields; the value is its boolean representation. If a key is absent, the field is considered *unset* and defaults to Kubernetes’ behaviour (i.e., `false` for booleans). |

---

### Key Dependencies

| Called function | What it does |
|-----------------|--------------|
| `IsContainerRunAsNonRoot(container)` | Returns whether the container’s own `securityContext.runAsNonRoot` is set to `true`. |
| `IsContainerRunAsNonRootUserID(container)` | Returns whether the container’s `securityContext.runAsUser` equals zero (`0`). |

Both helpers inspect the container's own security context; if a field is missing, they fall back to the pod defaults passed in via `podSC`.

---

### Algorithm (pseudocode)

```go
for each container c in p.Spec.Containers:
    // 1. Determine runAsNonRoot value
    nonRoot := IsContainerRunAsNonRoot(c)
    if !nonRoot {            // default or unset -> treat as false
        if podSC["runAsNonRoot"] { nonRoot = true }
    }

    // 2. Determine runAsUser value
    isZeroUID := IsContainerRunAsNonRootUserID(c)
    if !isZeroUID {
        if podSC["runAsUser"] && podSC["runAsUser"] == "0" { isZeroUID = true }
    }

    // 3. If container should run as non‑root but UID==0 → record
    if nonRoot && isZeroUID:
        append c to containers list
        create warning message and append to warnings list
```

The function uses the standard Go `append` built‑in twice—once for each output slice.

---

### Side Effects

* None. The function only reads from the receiver (`Pod`) and the supplied map; it does not mutate any state.
* It returns newly allocated slices, leaving caller‑supplied data untouched.

---

### How it fits the package

`GetRunAsNonRootFalseContainers` is part of CertSuite’s **pod security checks**.  
During a pod audit, the `provider.Pod` object is populated from the Kubernetes API and then passed to this method along with any pod‑level security context defaults. The returned containers and warnings are fed into CertSuite’s reporting framework to flag non‑compliant workloads.

In the broader provider package:

- **Pods** → encapsulate spec + status.
- **Containers** → represent each container within a pod (including name, image, security context).
- **SecurityContext helpers** (`IsContainerRunAsNonRoot`, `IsContainerRunAsNonRootUserID`) provide reusable logic across multiple checks.

This function therefore plays a central role in validating the *run‑as‑non‑root* policy for containers.
