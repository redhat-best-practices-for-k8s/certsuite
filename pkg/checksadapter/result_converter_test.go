package checksadapter

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/checks"
	"github.com/stretchr/testify/assert"
)

func TestConvertAndSetResult_CompliantDetails(t *testing.T) {
	check := checksdb.NewCheck("test-compliant", []string{"test"})
	result := checks.CheckResult{
		ComplianceStatus: checks.StatusCompliant,
		Details: []checks.ResourceDetail{
			{Kind: "Pod", Name: "pod1", Namespace: "ns1", Compliant: true, Message: "ok"},
		},
	}
	ConvertAndSetResult(check, result)
	assert.Equal(t, "passed", check.Result.String())
}

func TestConvertAndSetResult_NonCompliantDetails(t *testing.T) {
	check := checksdb.NewCheck("test-noncompliant", []string{"test"})
	result := checks.CheckResult{
		ComplianceStatus: checks.StatusNonCompliant,
		Details: []checks.ResourceDetail{
			{Kind: "Pod", Name: "pod1", Namespace: "ns1", Compliant: false, Message: "bad config"},
		},
	}
	ConvertAndSetResult(check, result)
	assert.Equal(t, "failed", check.Result.String())
}

func TestConvertAndSetResult_EmptyDetailsProducesSkip(t *testing.T) {
	check := checksdb.NewCheck("test-skip", []string{"test"})
	result := checks.CheckResult{
		ComplianceStatus: checks.StatusCompliant,
		Details:          []checks.ResourceDetail{},
	}
	ConvertAndSetResult(check, result)
	assert.Equal(t, "skipped", check.Result.String())
}

func TestConvertAndSetResult_NotFoundPodFilteredWithCompliant(t *testing.T) {
	check := checksdb.NewCheck("test-notfound-pod", []string{"test"})
	result := checks.CheckResult{
		ComplianceStatus: checks.StatusCompliant,
		Details: []checks.ResourceDetail{
			{Kind: "Pod", Name: "gone-pod", Namespace: "ns1", Compliant: false, Message: "pod not found"},
			{Kind: "Pod", Name: "good-pod", Namespace: "ns1", Compliant: true, Message: "ok"},
		},
	}
	ConvertAndSetResult(check, result)
	assert.Equal(t, "passed", check.Result.String())
}

func TestConvertAndSetResult_AllNotFoundPodsFilteredSynthesizesPass(t *testing.T) {
	check := checksdb.NewCheck("test-all-notfound", []string{"test"})
	result := checks.CheckResult{
		ComplianceStatus: checks.StatusCompliant,
		Details: []checks.ResourceDetail{
			{Kind: "Pod", Name: "gone-pod", Namespace: "ns1", Compliant: false, Message: "pod not found"},
		},
	}
	ConvertAndSetResult(check, result)
	assert.Equal(t, "passed", check.Result.String())
}

func TestConvertAndSetResult_NotFoundContainerFilteredWithCompliant(t *testing.T) {
	check := checksdb.NewCheck("test-notfound-container", []string{"test"})
	result := checks.CheckResult{
		ComplianceStatus: checks.StatusCompliant,
		Details: []checks.ResourceDetail{
			{Kind: "Container", Name: "gone-ctr", Namespace: "ns1", Compliant: false, Message: "container not found"},
			{Kind: "Container", Name: "good-ctr", Namespace: "ns1", Compliant: true, Message: "ok"},
		},
	}
	ConvertAndSetResult(check, result)
	assert.Equal(t, "passed", check.Result.String())
}

func TestConvertAndSetResult_NotFoundDeploymentNotFiltered(t *testing.T) {
	check := checksdb.NewCheck("test-notfound-deploy", []string{"test"})
	result := checks.CheckResult{
		ComplianceStatus: checks.StatusNonCompliant,
		Details: []checks.ResourceDetail{
			{Kind: "Deployment", Name: "gone-deploy", Namespace: "ns1", Compliant: false, Message: "not found in cluster"},
		},
	}
	ConvertAndSetResult(check, result)
	assert.Equal(t, "failed", check.Result.String())
}

func TestConvertAndSetResult_NonCompliantStatusNoDetails(t *testing.T) {
	check := checksdb.NewCheck("test-noncompliant-nodetails", []string{"test"})
	result := checks.CheckResult{
		ComplianceStatus: checks.StatusNonCompliant,
		Reason:           "general failure",
	}
	ConvertAndSetResult(check, result)
	assert.Equal(t, "failed", check.Result.String())
}

func TestConvertAndSetResult_ErrorStatusNoDetails(t *testing.T) {
	check := checksdb.NewCheck("test-error-nodetails", []string{"test"})
	result := checks.CheckResult{
		ComplianceStatus: checks.StatusError,
		Reason:           "internal error",
	}
	ConvertAndSetResult(check, result)
	assert.Equal(t, "failed", check.Result.String())
}

func TestConvertAndSetResult_MixedCompliance(t *testing.T) {
	check := checksdb.NewCheck("test-mixed", []string{"test"})
	result := checks.CheckResult{
		ComplianceStatus: checks.StatusNonCompliant,
		Details: []checks.ResourceDetail{
			{Kind: "Pod", Name: "good-pod", Namespace: "ns1", Compliant: true, Message: "ok"},
			{Kind: "Pod", Name: "bad-pod", Namespace: "ns1", Compliant: false, Message: "failed check"},
		},
	}
	ConvertAndSetResult(check, result)
	assert.Equal(t, "failed", check.Result.String())
}

func TestConvertDetailToReportObject_AllKinds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kind         string
		expectedType string
	}{
		{"Pod", testhelper.PodType},
		{"Container", testhelper.ContainerType},
		{"Deployment", testhelper.DeploymentType},
		{"StatefulSet", testhelper.StatefulSetType},
		{"Service", testhelper.ServiceType},
		{"RoleBinding", testhelper.RoleType},
		{"ClusterRoleBinding", testhelper.RoleType},
		{"CustomResourceDefinition", testhelper.CustomResourceDefinitionType},
		{"ClusterServiceVersion", testhelper.OperatorType},
		{"Namespace", testhelper.Namespace},
		{"Node", testhelper.NodeType},
		{"CatalogSource", testhelper.CatalogSourceType},
		{"HelmRelease", testhelper.HelmVersionType},
		{"UnknownKind", testhelper.UndefinedType},
	}

	for _, tc := range tests {
		t.Run(tc.kind, func(t *testing.T) {
			t.Parallel()
			detail := checks.ResourceDetail{
				Kind:      tc.kind,
				Name:      "test-resource",
				Namespace: "test-ns",
				Compliant: true,
				Message:   "ok",
			}
			ro := convertDetailToReportObject(detail)
			assert.Equal(t, tc.expectedType, ro.ObjectType)
		})
	}
}

func TestConvertDetailToReportObject_EmptyNamespace(t *testing.T) {
	t.Parallel()

	detail := checks.ResourceDetail{
		Kind:      "Node",
		Name:      "worker-1",
		Namespace: "",
		Compliant: true,
		Message:   "ok",
	}
	ro := convertDetailToReportObject(detail)
	assert.Equal(t, testhelper.NodeType, ro.ObjectType)
	assert.NotContains(t, ro.ObjectFieldsKeys, testhelper.Namespace)
}

func TestConvertDetailToReportObject_EmptyName(t *testing.T) {
	t.Parallel()

	detail := checks.ResourceDetail{
		Kind:      "Namespace",
		Name:      "",
		Namespace: "test-ns",
		Compliant: true,
		Message:   "ok",
	}
	ro := convertDetailToReportObject(detail)
	assert.NotContains(t, ro.ObjectFieldsKeys, testhelper.Name)
}
