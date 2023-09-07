package configuration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	configuration "github.com/test-network-function/cnf-certification-test/pkg/configuration"
)

const (
	filePath          = "testdata/tnf_test_config.yml"
	nsLength          = 2
	ns1               = "tnf"
	ns2               = "test2"
	crds              = 2
	crdSuffix1        = "group1.test.com"
	crdSuffix2        = "group2.test.com"
	containers        = 1
	containerInfoName = "nginx-116"
	containerRepo     = "rhel8"
	operators         = 1
	operatorName      = "etcd"
	operatorOrg       = "community-operators"
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
	// PodsUnderTestLabelsObjects
	assert.Equal(t, 4, len(env.PodsUnderTestLabelsObjects))
	assert.Equal(t, env.PodsUnderTestLabelsObjects[0].LabelKey, "test", "pod")
	assert.Equal(t, env.PodsUnderTestLabelsObjects[1].LabelKey, "cnf", "pod")
	assert.Equal(t, env.PodsUnderTestLabelsObjects[2].LabelKey, "cnf/test", "pod1")
	assert.Equal(t, env.PodsUnderTestLabelsObjects[3].LabelKey, "cnf/testEmpty", "")
	assert.Equal(t, env.PodsUnderTestLabelsObjects[0].LabelValue, "pod")
	assert.Equal(t, env.PodsUnderTestLabelsObjects[1].LabelValue, "pod")
	assert.Equal(t, env.PodsUnderTestLabelsObjects[2].LabelValue, "pod1")
	assert.Equal(t, env.PodsUnderTestLabelsObjects[3].LabelValue, "")
	// OperatorsUnderTestLabelsObjects
	assert.Equal(t, 4, len(env.OperatorsUnderTestLabelsObjects))
	assert.Equal(t, env.OperatorsUnderTestLabelsObjects[0].LabelKey, "test")
	assert.Equal(t, env.OperatorsUnderTestLabelsObjects[1].LabelKey, "cnf/test")
	assert.Equal(t, env.OperatorsUnderTestLabelsObjects[2].LabelKey, "cnf")
	assert.Equal(t, env.PodsUnderTestLabelsObjects[3].LabelKey, "cnf/testEmpty")
	assert.Equal(t, env.OperatorsUnderTestLabelsObjects[0].LabelValue, "operator")
	assert.Equal(t, env.OperatorsUnderTestLabelsObjects[1].LabelValue, "operator1")
	assert.Equal(t, env.OperatorsUnderTestLabelsObjects[2].LabelValue, "operator")
	assert.Equal(t, env.PodsUnderTestLabelsObjects[3].LabelValue, "")
	// check if targetCrdFilters section is parsed properly
	assert.Equal(t, crds, len(env.CrdFilters))
	crd1 := configuration.CrdFilter{NameSuffix: crdSuffix1}
	assert.Contains(t, env.CrdFilters, crd1)
	crd2 := configuration.CrdFilter{NameSuffix: crdSuffix2}
	assert.Contains(t, env.CrdFilters, crd2)
	// check if certifiedcontainerinfo section is parsed properly
	assert.Equal(t, containers, len(env.CertifiedContainerInfo))
	containerInfo := configuration.ContainerImageIdentifier{Repository: containerInfoName, Registry: containerRepo, Tag: "", Digest: ""}
	assert.Contains(t, env.CertifiedContainerInfo, containerInfo)
	// check if certifiedoperatorinfo section is parsed properly
	assert.Equal(t, operators, len(env.CertifiedOperatorInfo))
	operator := configuration.CertifiedOperatorRequestInfo{Name: operatorName, Organization: operatorOrg}
	assert.Contains(t, env.CertifiedOperatorInfo, operator)
	e, er := configuration.LoadConfiguration(filePath)
	assert.Equal(t, &env, &e)
	assert.Nil(t, er)
}
