package podhelper

import (
	"maps"
	"testing"

	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	appsv1 "k8s.io/api/apps/v1"

	k8sDynamicFake "k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

func Test_followOwnerReferences(t *testing.T) {
	type args struct {
		topOwners map[string]TopOwner
		namespace string
		ownerRefs []metav1.OwnerReference
	}

	csv1 := &olmv1Alpha.ClusterServiceVersion{
		TypeMeta: metav1.TypeMeta{Kind: "ClusterServiceVersion", APIVersion: "operators.coreos.com/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "csv1",
			Namespace:       "ns1",
			OwnerReferences: []metav1.OwnerReference{},
		},
	}
	dep1 := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "dep1",
			Namespace:       "ns1",
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "operators.coreos.com/v1alpha1", Kind: "ClusterServiceVersion", Name: "csv1"}},
		},
	}
	rep1 := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{Kind: "ReplicaSet", APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "rep1",
			Namespace:       "ns1",
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "Deployment", Name: "dep1"}},
		},
	}

	resourceList := []*metav1.APIResourceList{
		{GroupVersion: "operators.coreos.com/v1alpha1", APIResources: []metav1.APIResource{{Name: "clusterserviceversions", Kind: "ClusterServiceVersion"}}},
		{GroupVersion: "apps/v1", APIResources: []metav1.APIResource{{Name: "deployments", Kind: "Deployment"}}},
		{GroupVersion: "apps/v1", APIResources: []metav1.APIResource{{Name: "replicasets", Kind: "ReplicaSet"}}},
		{GroupVersion: "apps/v1", APIResources: []metav1.APIResource{{Name: "pods", Kind: "Pod"}}},
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{topOwners: map[string]TopOwner{"csv1": {Namespace: "ns1", Kind: "ClusterServiceVersion", Name: "csv1"}},
				namespace: "ns1",
				ownerRefs: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "ReplicaSet", Name: "rep1"}},
			},
		},
	}

	// Spoof the get and update functions
	client := k8sDynamicFake.NewSimpleDynamicClient(runtime.NewScheme(), rep1, dep1, csv1)
	client.Fake.AddReactor("get", "ClusterServiceVersion", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, csv1, nil
	})
	client.Fake.AddReactor("get", "Deployment", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, dep1, nil
	})
	client.Fake.AddReactor("get", "ReplicaSet", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, rep1, nil
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResults := map[string]TopOwner{}
			if err := followOwnerReferences(resourceList, client, gotResults, tt.args.namespace, tt.args.ownerRefs); (err != nil) != tt.wantErr {
				t.Errorf("followOwnerReferences() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !maps.Equal(gotResults, tt.args.topOwners) {
				t.Errorf("followOwnerReferences() = %v, want %v", gotResults, tt.args.topOwners)
			}
		})
	}
}
