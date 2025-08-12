scaleDeploymentHelper`

| Item | Detail |
|------|--------|
| **Location** | `tests/lifecycle/scaling/deployment_scaling.go:79` |
| **Signature** | `func scaleDeploymentHelper(client typedappsv1.AppsV1Interface, deployment *appsv1.Deployment, desiredReplicas int32, timeout time.Duration, dryRun bool, logger *log.Logger) bool` |
| **Visibility** | Unexported (internal helper) |

### Purpose
Adjust the replica count of a single Deployment and optionally wait for it to reach the desired state.  
It is used by higher‑level scaling tests that need deterministic control over a Deployment’s size.

### Parameters

| Name | Type | Role |
|------|------|------|
| `client` | `typedappsv1.AppsV1Interface` | Kubernetes AppsV1 client for performing CRUD on Deployments. |
| `deployment` | `*appsv1.Deployment` | The Deployment object whose replica count is to be changed. The function reads its current spec and updates it. |
| `desiredReplicas` | `int32` | Target number of replicas after the update. |
| `timeout` | `time.Duration` | How long to wait for the Deployment to become ready if `dryRun` is false. |
| `dryRun` | `bool` | If true, only log the intended change; no API call or waiting occurs. |
| `logger` | `*log.Logger` | Logger used for informational and error messages. |

### Return Value
- **`true`** – The operation succeeded (either the Deployment was updated or the dry‑run simulation completed).  
- **`false`** – An unrecoverable error occurred during update, retry, or readiness wait.

### Key Steps

1. **Dry‑Run Mode**  
   - If `dryRun` is true: log intended replica change and return `true`.

2. **Update Flow (with Conflict Retry)**  
   - Wrap the update logic in `RetryOnConflict`, which retries on Kubernetes conflict errors.
   - Inside the retry loop:
     1. Fetch the latest Deployment using `client.Deployments(namespace).Get(...)`.
     2. Modify its `Spec.Replicas` to `desiredReplicas`.
     3. Call `client.Deployments(namespace).Update(...)`.

3. **Post‑Update Wait**  
   - If not a dry run, invoke `WaitForDeploymentSetReady` with the timeout to block until the Deployment’s pods are ready.
   - Log success or failure.

4. **Error Handling**  
   - All errors from Get/Update/Wait are logged via `logger.Error`.
   - The function returns `false` on any error that cannot be retried.

### Dependencies

| Function | Source Package | Role |
|----------|----------------|------|
| `RetryOnConflict` | internal test utilities | Handles retry logic for conflicts. |
| `Get`, `Update` | `client.Deployments(...)` | CRUD operations on Deployment resources. |
| `WaitForDeploymentSetReady` | scaling package (or shared utils) | Waits until the Deployment’s pods are ready or timeout occurs. |
| `log.Logger` methods (`Info`, `Error`) | standard library | Structured logging for debugging and audit. |

### Side Effects
- **Kubernetes API mutation**: updates the Deployment resource unless `dryRun` is true.
- **Blocking**: may wait up to `timeout` duration for readiness if not a dry run.

### How It Fits the Package

The `scaling` package contains end‑to‑end tests that exercise scaling behavior in various scenarios (horizontal autoscaling, manual scaling, etc.).  
`scaleDeploymentHelper` is the low‑level routine that:

1. Performs the actual replica count change,
2. Handles conflicts and retries, and
3. Optionally waits for stability.

Higher‑level test functions build on this helper to orchestrate sequences of scaling operations and validate resulting system behavior.
