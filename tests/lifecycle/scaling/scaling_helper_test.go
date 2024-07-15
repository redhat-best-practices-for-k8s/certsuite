package scaling

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	scalingv1 "k8s.io/api/autoscaling/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsManaged(t *testing.T) {
	type args struct {
		podSetName    string
		managedPodSet []configuration.ManagedDeploymentsStatefulsets
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				podSetName: "test1",
				managedPodSet: []configuration.ManagedDeploymentsStatefulsets{
					{
						Name: "test1",
					},
				},
			},
			want: true,
		},
		{
			name: "test2",
			args: args{
				podSetName: "test1",
				managedPodSet: []configuration.ManagedDeploymentsStatefulsets{
					{
						Name: "test2",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsManaged(tt.args.podSetName, tt.args.managedPodSet); got != tt.want {
				t.Errorf("IsManaged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetResourceHPA(t *testing.T) {
	generateHPAList := func() []*scalingv1.HorizontalPodAutoscaler {
		return []*scalingv1.HorizontalPodAutoscaler{
			{
				Spec: scalingv1.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: scalingv1.CrossVersionObjectReference{
						Kind: "testKind",
						Name: "testName",
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "testNamespace",
				},
			},
		}
	}

	testCases := []struct {
		existsInList bool
	}{
		{
			existsInList: true,
		},
		{
			existsInList: false,
		},
	}

	for _, testCase := range testCases {
		if testCase.existsInList {
			assert.NotNil(t, GetResourceHPA(generateHPAList(), "testName", "testNamespace", "testKind"))
		} else {
			assert.Nil(t, GetResourceHPA([]*scalingv1.HorizontalPodAutoscaler{}, "testName", "testNamespace", "testKind"))
		}
	}
}

func TestCheckOwnerReference(t *testing.T) {
	generateOwnerReferences := func() []metav1.OwnerReference {
		return []metav1.OwnerReference{
			{
				Kind: "testKind",
			},
		}
	}

	generateCrdFilter := func(scalable bool) []configuration.CrdFilter {
		return []configuration.CrdFilter{
			{
				NameSuffix: "testSuffix",
				Scalable:   scalable,
			},
		}
	}

	generateCrds := func() []*apiextv1.CustomResourceDefinition {
		return []*apiextv1.CustomResourceDefinition{
			{
				Spec: apiextv1.CustomResourceDefinitionSpec{
					Names: apiextv1.CustomResourceDefinitionNames{
						Kind: "testKind",
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "testSuffix",
				},
			},
		}
	}

	testCases := []struct {
		scalable bool
	}{
		{
			scalable: true,
		},
		{
			scalable: false,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.scalable, CheckOwnerReference(generateOwnerReferences(), generateCrdFilter(testCase.scalable), generateCrds()))
	}

	// Test case when owner reference is not found
	assert.False(t, CheckOwnerReference([]metav1.OwnerReference{}, generateCrdFilter(true), generateCrds()))
}
