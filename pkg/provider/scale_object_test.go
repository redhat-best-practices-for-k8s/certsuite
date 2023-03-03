package provider

import (
	"testing"

	scalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCrScale_ToString(t *testing.T) {
	type fields struct {
		Scale *scalingv1.Scale
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test1",
			fields: fields{
				Scale: &scalingv1.Scale{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "testNS",
					},
				},
			},
			want: "cr: test1 ns: testNS",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crScale := CrScale{
				Scale: tt.fields.Scale,
			}
			if got := crScale.ToString(); got != tt.want {
				t.Errorf("CrScale.ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
