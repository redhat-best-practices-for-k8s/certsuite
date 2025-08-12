## Package tolerations (github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/tolerations)



### Functions

- **IsTolerationDefault** — func(corev1.Toleration)(bool)
- **IsTolerationModified** — func(corev1.Toleration, corev1.PodQOSClass)(bool)

### Globals


### Call graph (exported symbols, partial)

```mermaid
graph LR
  IsTolerationDefault --> Contains
  IsTolerationModified --> IsTolerationDefault
  IsTolerationModified --> int64
```

### Symbol docs

- [function IsTolerationDefault](symbols/function_IsTolerationDefault.md)
- [function IsTolerationModified](symbols/function_IsTolerationModified.md)
