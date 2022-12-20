package claim

import (
	"reflect"
	"testing"
)

//nolint:funlen
func Test_compare2TestCaseResults(t *testing.T) {
	type args struct {
		testcaseResult1 []testCase
		testcaseResult2 []testCase
	}
	tests := []struct {
		name              string
		args              args
		wantDiffresult    []testCase
		wantNotFoundtest  []string
		wantNotFoundtest2 []string
	}{
		{
			name: "test1",
			args: args{
				testcaseResult1: []testCase{
					{
						Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
						Status: "skipped",
					},
				},
				testcaseResult2: []testCase{
					{
						Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
						Status: "failed",
					},
				},
			},
			wantDiffresult: []testCase{
				{
					Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
					Status: "skipped",
				},
			},
			wantNotFoundtest:  []string{},
			wantNotFoundtest2: []string{},
		},
		{
			name: "test2",
			args: args{
				testcaseResult1: []testCase{
					{
						Name:   "[It] observability observability-crd-status [common, observability, observability-crd-status]",
						Status: "skipped",
					},
					{
						Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
						Status: "skipped",
					},
				},
				testcaseResult2: []testCase{
					{
						Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
						Status: "failed",
					},
				},
			},
			wantDiffresult: []testCase{
				{
					Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
					Status: "skipped",
				},
			},
			wantNotFoundtest:  []string{},
			wantNotFoundtest2: []string{"[It] observability observability-crd-status [common, observability, observability-crd-status]"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDiffresult, gotNotFoundtest, gotNotFoundtest2 := compare2TestCaseResults(tt.args.testcaseResult1, tt.args.testcaseResult2)
			if !reflect.DeepEqual(gotDiffresult, tt.wantDiffresult) {
				t.Errorf("compare2TestCaseResults() gotDiffresult = %v, want %v", gotDiffresult, tt.wantDiffresult)
			}
			if !reflect.DeepEqual(gotNotFoundtest, tt.wantNotFoundtest) {
				t.Errorf("compare2TestCaseResults() gotNotFoundtest = %v, want %v", gotNotFoundtest, tt.wantNotFoundtest)
			}
			if !reflect.DeepEqual(gotNotFoundtest2, tt.wantNotFoundtest2) {
				t.Errorf("compare2TestCaseResults() gotNotFoundtest2 = %v, want %v", gotNotFoundtest2, tt.wantNotFoundtest2)
			}
		})
	}
}

//nolint:funlen
func Test_compare2cnis(t *testing.T) {
	type args struct {
		cniList1 []Cni
		cniList2 []Cni
		nodeName string
	}
	tests := []struct {
		name               string
		args               args
		wantDiffplugins    []Cni
		wantNotFoundNames  []string
		wantNotFoundNames2 []string
	}{
		{
			name: "test1",
			args: args{
				cniList1: []Cni{
					{
						Name:    "podman",
						Plugins: nil,
					},
					{
						Name:    "crio",
						Plugins: nil,
					},
				},
				cniList2: []Cni{
					{
						Name:    "podman",
						Plugins: nil,
					},
				},
				nodeName: "master1",
			},
			wantDiffplugins:    nil,
			wantNotFoundNames:  []string{"crio"},
			wantNotFoundNames2: []string{},
		},
		{
			name: "test2",
			args: args{
				cniList1: []Cni{
					{
						Name:    "podman",
						Plugins: nil,
					},
				},
				cniList2: []Cni{
					{
						Name:    "podman",
						Plugins: nil,
					},
					{
						Name:    "crio",
						Plugins: nil,
					},
				},
				nodeName: "master1",
			},
			wantDiffplugins:    nil,
			wantNotFoundNames:  []string{},
			wantNotFoundNames2: []string{"crio"},
		},
		{
			name: "test3",
			args: args{
				cniList1: nil,
				cniList2: []Cni{
					{
						Name:    "podman",
						Plugins: nil,
					},
					{
						Name:    "crio",
						Plugins: nil,
					},
				},
				nodeName: "master1",
			},
			wantDiffplugins:    nil,
			wantNotFoundNames:  nil,
			wantNotFoundNames2: nil,
		},
		{
			name: "test4",
			args: args{
				cniList1: nil,
				cniList2: nil,
				nodeName: "master1",
			},
			wantDiffplugins:    nil,
			wantNotFoundNames:  nil,
			wantNotFoundNames2: nil,
		},
		{
			name: "test5",
			args: args{
				cniList1: []Cni{},
				cniList2: []Cni{},
				nodeName: "master1",
			},
			wantDiffplugins:    nil,
			wantNotFoundNames:  nil,
			wantNotFoundNames2: nil,
		},
		{
			name: "test6",
			args: args{
				cniList1: []Cni{
					{
						Name:    "podman",
						Plugins: nil,
					},
				},
				cniList2: []Cni{},
				nodeName: "master1",
			},
			wantDiffplugins:    nil,
			wantNotFoundNames:  nil,
			wantNotFoundNames2: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDiffplugins, gotNotFoundNames, gotNotFoundNames2 := compare2cniHelper(tt.args.cniList1, tt.args.cniList2, tt.args.nodeName)
			if !reflect.DeepEqual(gotDiffplugins, tt.wantDiffplugins) {
				t.Errorf("compare2cnis() gotDiffplugins = %v, want %v", gotDiffplugins, tt.wantDiffplugins)
			}
			if !reflect.DeepEqual(gotNotFoundNames, tt.wantNotFoundNames) {
				t.Errorf("compare2cnis() gotNotFoundNames = %v, want %v", gotNotFoundNames, tt.wantNotFoundNames)
			}
			if !reflect.DeepEqual(gotNotFoundNames2, tt.wantNotFoundNames2) {
				t.Errorf("compare2cnis() gotNotFoundNames2 = %v, want %v", gotNotFoundNames2, tt.wantNotFoundNames2)
			}
		})
	}
}

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
