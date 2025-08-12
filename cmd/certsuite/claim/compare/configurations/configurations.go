package configurations

import (
	"fmt"

	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/diff"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/pkg/claim"
)

// AbnormalEventsCount holds the number of abnormal events for two claims.
//
// It contains integer fields Claim1 and Claim2 that represent the count of
// abnormal events detected for each respective claim. The String method
// returns a formatted string summarizing these counts.
type AbnormalEventsCount struct {
	Claim1 int `json:"claim1"`
	Claim2 int `json:"claim2"`
}

// String returns a human‑readable representation of the abnormal events count.
//
// It formats the internal counters into a single string, allowing easy
// printing or logging of the number of abnormal events recorded.
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

// DiffReport represents the result of comparing two configurations.
//
// It contains the number of abnormal events detected during the comparison
// in the AbnormalEvents field, and a pointer to a diff.Diffs value that
// describes the differences between the two configuration objects.
// The String method returns a human‑readable summary of the report.
type DiffReport struct {
	Config         *diff.Diffs         `json:"CertSuiteConfig"`
	AbnormalEvents AbnormalEventsCount `json:"abnormalEventsCount"`
}

// String returns a human‑readable representation of the DiffReport.
//
// It formats the differences between two configuration sets into a single
// string, suitable for printing or logging. The returned value is a plain
// text description that lists added, removed, and changed items in a
// readable layout.
func (d *DiffReport) String() string {
	str := "CONFIGURATIONS\n"
	str += "--------------\n\n"

	str += d.Config.String()

	str += "\n"
	str += d.AbnormalEvents.String()

	return str
}

// GetDiffReport produces a diff report between two configuration sets.
//
// It compares the first configuration set with the second and returns a DiffReport
// that summarizes differences, additions, and removals. The returned value contains
// details of what changed and is nil if an error occurs during comparison.
func GetDiffReport(claim1Configurations, claim2Configurations *claim.Configurations) *DiffReport {
	return &DiffReport{
		Config: diff.Compare("Cert Suite Configuration", claim1Configurations.Config, claim2Configurations.Config, nil),
		AbnormalEvents: AbnormalEventsCount{
			Claim1: len(claim1Configurations.AbnormalEvents),
			Claim2: len(claim2Configurations.AbnormalEvents),
		},
	}
}
