package compare

import (
	"testing"
)

func Test_claimCompareFilesfunc(t *testing.T) {
	type args struct {
		claim1 string
		claim2 string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{{

		name: "test1",
		args: args{
			claim1: "claim_test1.json",
			claim2: "claim_test2.json",
		},
		wantErr: false,
	},
		{
			name: "test2",
			args: args{
				claim1: "claim_test1.json",
				claim2: "claim_test4.json",
			},
			wantErr: true,
		}, {
			name: "test3",
			args: args{
				claim1: "claim_test1.json",
				claim2: "claim_test3.json",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := claimCompareFilesfunc(tt.args.claim1, tt.args.claim2); (err != nil) != tt.wantErr {
				t.Errorf("claimCompareFilesfunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
