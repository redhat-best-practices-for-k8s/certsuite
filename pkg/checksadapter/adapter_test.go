package checksadapter

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/checks"
	"github.com/stretchr/testify/assert"
)

func TestNewAdapter(t *testing.T) {
	checkFunc := func(resources *checks.DiscoveredResources) checks.CheckResult {
		return checks.CheckResult{}
	}
	adapter := NewAdapter(checkFunc)
	assert.NotNil(t, adapter)
	assert.NotNil(t, adapter.checkFunc)
}

func TestAdapter_Execute(t *testing.T) {
	// Initialize fake clientsholder to avoid log.Fatal
	clientsholder.GetTestClientsHolder(nil)

	env := &provider.TestEnvironment{}
	check := &checksdb.Check{}

	called := false
	checkFunc := func(resources *checks.DiscoveredResources) checks.CheckResult {
		called = true
		assert.NotNil(t, resources)
		return checks.CheckResult{
			ComplianceStatus: checks.StatusCompliant,
		}
	}

	adapter := NewAdapter(checkFunc)
	err := adapter.Execute(check, env)

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestAdapter_ExecuteIntrusive(t *testing.T) {
	clientsholder.GetTestClientsHolder(nil)
	env := &provider.TestEnvironment{}
	check := &checksdb.Check{}

	checkFunc := func(resources *checks.DiscoveredResources) checks.CheckResult {
		return checks.CheckResult{
			ComplianceStatus: checks.StatusCompliant,
		}
	}

	adapter := NewAdapter(checkFunc)
	err := adapter.ExecuteIntrusive(check, env)

	assert.NoError(t, err)
}

func TestAdapter_MakeCheckFn(t *testing.T) {
	clientsholder.GetTestClientsHolder(nil)
	env := &provider.TestEnvironment{}
	check := &checksdb.Check{}

	called := false
	checkFunc := func(resources *checks.DiscoveredResources) checks.CheckResult {
		called = true
		return checks.CheckResult{
			ComplianceStatus: checks.StatusCompliant,
		}
	}

	adapter := NewAdapter(checkFunc)
	fn := adapter.MakeCheckFn(env)
	err := fn(check)

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestAdapter_MakeIntrusiveCheckFn(t *testing.T) {
	clientsholder.GetTestClientsHolder(nil)
	env := &provider.TestEnvironment{}
	check := &checksdb.Check{}

	checkFunc := func(resources *checks.DiscoveredResources) checks.CheckResult {
		return checks.CheckResult{
			ComplianceStatus: checks.StatusCompliant,
		}
	}

	adapter := NewAdapter(checkFunc)
	fn := adapter.MakeIntrusiveCheckFn(env)
	err := fn(check)

	assert.NoError(t, err)
}
