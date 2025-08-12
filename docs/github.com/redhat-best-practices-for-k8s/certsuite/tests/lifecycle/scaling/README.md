## Package scaling (github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/scaling)



### Functions

- **CheckOwnerReference** — func([]apiv1.OwnerReference, []configuration.CrdFilter, []*apiextv1.CustomResourceDefinition)(bool)
- **GetResourceHPA** — func([]*scalingv1.HorizontalPodAutoscaler, string, string, string)(*scalingv1.HorizontalPodAutoscaler)
- **IsManaged** — func(string, []configuration.ManagedDeploymentsStatefulsets)(bool)
- **TestScaleCrd** — func(*provider.CrScale, schema.GroupResource, time.Duration, *log.Logger)(bool)
- **TestScaleDeployment** — func(*appsv1.Deployment, time.Duration, *log.Logger)(bool)
- **TestScaleHPACrd** — func(*provider.CrScale, *scalingv1.HorizontalPodAutoscaler, schema.GroupResource, time.Duration, *log.Logger)(bool)
- **TestScaleHpaDeployment** — func(*provider.Deployment, *v1autoscaling.HorizontalPodAutoscaler, time.Duration, *log.Logger)(bool)
- **TestScaleHpaStatefulSet** — func(*appsv1.StatefulSet, *v1autoscaling.HorizontalPodAutoscaler, time.Duration, *log.Logger)(bool)
- **TestScaleStatefulSet** — func(*appsv1.StatefulSet, time.Duration, *log.Logger)(bool)

### Call graph (exported symbols, partial)

```mermaid
graph LR
  CheckOwnerReference --> HasSuffix
  TestScaleCrd --> Error
  TestScaleCrd --> GetClientsHolder
  TestScaleCrd --> GetName
  TestScaleCrd --> GetNamespace
  TestScaleCrd --> scaleCrHelper
  TestScaleCrd --> Error
  TestScaleCrd --> scaleCrHelper
  TestScaleCrd --> Error
  TestScaleDeployment --> GetClientsHolder
  TestScaleDeployment --> Info
  TestScaleDeployment --> scaleDeploymentHelper
  TestScaleDeployment --> AppsV1
  TestScaleDeployment --> Error
  TestScaleDeployment --> scaleDeploymentHelper
  TestScaleDeployment --> AppsV1
  TestScaleDeployment --> Error
  TestScaleHPACrd --> Error
  TestScaleHPACrd --> GetClientsHolder
  TestScaleHPACrd --> GetNamespace
  TestScaleHPACrd --> HorizontalPodAutoscalers
  TestScaleHPACrd --> AutoscalingV1
  TestScaleHPACrd --> int32
  TestScaleHPACrd --> GetName
  TestScaleHPACrd --> Debug
  TestScaleHpaDeployment --> GetClientsHolder
  TestScaleHpaDeployment --> HorizontalPodAutoscalers
  TestScaleHpaDeployment --> AutoscalingV1
  TestScaleHpaDeployment --> int32
  TestScaleHpaDeployment --> Debug
  TestScaleHpaDeployment --> scaleHpaDeploymentHelper
  TestScaleHpaDeployment --> Debug
  TestScaleHpaDeployment --> scaleHpaDeploymentHelper
  TestScaleHpaStatefulSet --> GetClientsHolder
  TestScaleHpaStatefulSet --> HorizontalPodAutoscalers
  TestScaleHpaStatefulSet --> AutoscalingV1
  TestScaleHpaStatefulSet --> int32
  TestScaleHpaStatefulSet --> int32
  TestScaleHpaStatefulSet --> Debug
  TestScaleHpaStatefulSet --> scaleHpaStatefulSetHelper
  TestScaleHpaStatefulSet --> Debug
  TestScaleStatefulSet --> GetClientsHolder
  TestScaleStatefulSet --> StatefulSets
  TestScaleStatefulSet --> AppsV1
  TestScaleStatefulSet --> Debug
  TestScaleStatefulSet --> int32
  TestScaleStatefulSet --> Debug
  TestScaleStatefulSet --> scaleStatefulsetHelper
  TestScaleStatefulSet --> Error
```

### Symbol docs

- [function CheckOwnerReference](symbols/function_CheckOwnerReference.md)
- [function GetResourceHPA](symbols/function_GetResourceHPA.md)
- [function IsManaged](symbols/function_IsManaged.md)
- [function TestScaleCrd](symbols/function_TestScaleCrd.md)
- [function TestScaleDeployment](symbols/function_TestScaleDeployment.md)
- [function TestScaleHPACrd](symbols/function_TestScaleHPACrd.md)
- [function TestScaleHpaDeployment](symbols/function_TestScaleHpaDeployment.md)
- [function TestScaleHpaStatefulSet](symbols/function_TestScaleHpaStatefulSet.md)
- [function TestScaleStatefulSet](symbols/function_TestScaleStatefulSet.md)
