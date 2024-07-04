package configurations

import (
	"fmt"

	"github.com/test-network-function/cnf-certification-test/cmd/certsuite/claim/compare/diff"
	"github.com/test-network-function/cnf-certification-test/cmd/certsuite/pkg/claim"
)

type AbnormalEventsCount struct {
	Claim1 int `json:"claim1"`
	Claim2 int `json:"claim2"`
}

func (c *AbnormalEventsCount) String() string {
	const (
		rowHeaderFmt = "%-12s%-s\n"
		rowDataFmt   = "%-12d%-d\n"
	)

	str := "Cluster abnormal events count\n"
	str += fmt.Sprintf(rowHeaderFmt, "CLAIM 1", "CLAIM 2")
	str += fmt.Sprintf(rowDataFmt, c.Claim1, c.Claim2)

	return str
}

type DiffReport struct {
	Config         *diff.Diffs         `json:"CNFCertSuiteConfig"`
	AbnormalEvents AbnormalEventsCount `json:"abnormalEventsCount"`
}

func (d *DiffReport) String() string {
	str := "CONFIGURATIONS\n"
	str += "--------------\n\n"

	str += d.Config.String()

	str += "\n"
	str += d.AbnormalEvents.String()

	return str
}

func GetDiffReport(claim1Configurations, claim2Configurations *claim.Configurations) *DiffReport {
	return &DiffReport{
		Config: diff.Compare("CNF Cert Suite Configuration", claim1Configurations.Config, claim2Configurations.Config, nil),
		AbnormalEvents: AbnormalEventsCount{
			Claim1: len(claim1Configurations.AbnormalEvents),
			Claim2: len(claim2Configurations.AbnormalEvents),
		},
	}
}
