package configuration_test

import (
	"testing"

	configuration "github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	"github.com/stretchr/testify/assert"
)

const (
	filePath   = "testdata/tnf_test_config.yml"
	nsLength   = 2
	ns1        = "tnf"
	ns2        = "test2"
	crds       = 2
	crdSuffix1 = "group1.test.com"
	crdSuffix2 = "group2.test.com"
)

func TestLoadConfiguration(t *testing.T) {
	env, err := configuration.LoadConfiguration(filePath)
	assert.Nil(t, err)
	// check if targetNameSpaces section is parsed properly
	assert.Equal(t, nsLength, len(env.TargetNameSpaces))
	ns := configuration.Namespace{Name: ns1}
	assert.Contains(t, env.TargetNameSpaces, ns)
	ns.Name = ns2
	assert.Contains(t, env.TargetNameSpaces, ns)
	// check if targetCrdFilters section is parsed properly
	assert.Equal(t, crds, len(env.CrdFilters))
	crd1 := configuration.CrdFilter{NameSuffix: crdSuffix1}
	assert.Contains(t, env.CrdFilters, crd1)
	crd2 := configuration.CrdFilter{NameSuffix: crdSuffix2}
	assert.Contains(t, env.CrdFilters, crd2)
}
