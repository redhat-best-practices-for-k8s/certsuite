package configurations

import (
	"fmt"

	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/diff"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/pkg/claim"
)

// AbnormalEventsCount Displays counts of abnormal events for two claims
//
// This struct holds integer counts of abnormal events for two distinct claims,
// named Claim1 and Claim2. The String method formats these values into a
// readable table with headers, producing a string that summarizes the event
// counts for comparison purposes.
type AbnormalEventsCount struct {
	Claim1 int `json:"claim1"`
	Claim2 int `json:"claim2"`
}

// AbnormalEventsCount.String Formats abnormal event counts for two claims
//
// This method builds a multi-line string that displays the number of abnormal
// events detected in two separate claims. It starts with a header line, then
// adds a formatted table row showing the claim identifiers and their
// corresponding counts using printf-style formatting. The resulting string is
// returned for display or logging.
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

// DiffReport captures configuration differences and abnormal event counts
//
// This structure contains a diff of Cert Suite configuration objects and a
// count of abnormal events for two claims. The Config field holds the result
// from a diff comparison, while AbnormalEvents stores how many abnormal events
// each claim reported. It is used to report and display discrepancies between
// claims.
type DiffReport struct {
	Config         *diff.Diffs         `json:"CertSuiteConfig"`
	AbnormalEvents AbnormalEventsCount `json:"abnormalEventsCount"`
}

// DiffReport.String Formats the diff report into a readable string
//
// This method builds a formatted representation of a configuration comparison,
// beginning with header lines and then appending the configuration details
// followed by any abnormal events. It concatenates strings from the embedded
// Config and AbnormalEvents fields and returns the final result as a single
// string.
func (d *DiffReport) String() string {
	str := "CONFIGURATIONS\n"
	str += "--------------\n\n"

	str += d.Config.String()

	str += "\n"
	str += d.AbnormalEvents.String()

	return str
}

// GetDiffReport Creates a report of configuration differences
//
// The function compares two configuration objects from claim files, generating
// a DiffReport that includes field-by-field differences in the main
// configuration map and counts of abnormal events present in each file. It uses
// an external diff utility to compute the detailed comparison and returns the
// assembled report for further processing or display.
func GetDiffReport(claim1Configurations, claim2Configurations *claim.Configurations) *DiffReport {
	return &DiffReport{
		Config: diff.Compare("Cert Suite Configuration", claim1Configurations.Config, claim2Configurations.Config, nil),
		AbnormalEvents: AbnormalEventsCount{
			Claim1: len(claim1Configurations.AbnormalEvents),
			Claim2: len(claim2Configurations.AbnormalEvents),
		},
	}
}
