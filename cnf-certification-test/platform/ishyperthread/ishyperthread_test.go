package ishyperthread

import (
	"testing"
)

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
			name: "test2",
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

func Test_extractNumber(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test1",
			args: args{str: " number is 3"},
			want: 3,
		},
		{
			name: "test2",
			args: args{str: " number is 2 and thats ok"},
			want: 2,
		},
		{
			name: "test3",
			args: args{str: "there is no number"},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractNumber(tt.args.str); got != tt.want {
				t.Errorf("extractNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}
