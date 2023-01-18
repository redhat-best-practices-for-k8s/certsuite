package scaling

import (
	"testing"

	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
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
