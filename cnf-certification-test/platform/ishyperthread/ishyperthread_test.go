package ishyperthread

import "testing"

func TestIsBareMetal(t *testing.T) {
	type args struct {
		providerID string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{providerID: "baremetalhost://test"},
			want: true,
		}, {
			name: "test1",
			args: args{providerID: "VM://test"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBareMetal(tt.args.providerID); got != tt.want {
				t.Errorf("IsBareMetal() = %v, want %v", got, tt.want)
			}
		})
	}
}
