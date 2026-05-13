package preflight

import (
	"context"

	"github.com/redhat-openshift-ecosystem/openshift-preflight/artifacts"
	plibContainer "github.com/redhat-openshift-ecosystem/openshift-preflight/container"
	plibOperator "github.com/redhat-openshift-ecosystem/openshift-preflight/operator"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
)

var catalogLoaded bool

func LoadCatalogChecks() {
	if catalogLoaded {
		return
	}
	catalogLoaded = true
	const dummy = "dummy"

	artifactsWriter, err := artifacts.NewMapWriter()
	if err != nil {
		log.Error("Failed to create artifact writer for preflight catalog: %v", err)
		return
	}
	ctx := artifacts.ContextWithWriter(context.TODO(), artifactsWriter)

	checkContainer := plibContainer.NewCheck(dummy)
	checkOperator := plibOperator.NewCheck(dummy, dummy, []byte(""))

	_, checksContainer, err := checkContainer.List(ctx)
	if err != nil {
		log.Error("Error listing preflight container checks: %v", err)
	}

	_, checksOperator, err := checkOperator.List(ctx)
	if err != nil {
		log.Error("Error listing preflight operator checks: %v", err)
	}

	allChecks := checksContainer
	allChecks = append(allChecks, checksOperator...)

	checksGroup := checksdb.NewChecksGroup(common.PreflightTestKey)

	for _, c := range allChecks {
		remediation := c.Help().Suggestion
		if c.Name() == "FollowsRestrictedNetworkEnablementGuidelines" {
			remediation = "If consumers of your operator may need to do so on a restricted network, implement the guidelines outlined in OCP documentation: https://docs.redhat.com/en/documentation/openshift_container_platform/latest/html/disconnected_environments/olm-restricted-networks"
		}

		aID := identifiers.AddCatalogEntry(
			c.Name(),
			common.PreflightTestKey,
			c.Metadata().Description,
			remediation,
			identifiers.NoDocumentedProcess,
			identifiers.NoDocLink,
			true,
			map[string]string{
				identifiers.FarEdge:  identifiers.Optional,
				identifiers.Telco:    identifiers.Optional,
				identifiers.NonTelco: identifiers.Optional,
				identifiers.Extended: identifiers.Optional,
			},
			identifiers.TagPreflight)

		checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(aID)))
	}
}
