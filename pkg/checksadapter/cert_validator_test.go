package checksadapter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	release "helm.sh/helm/v4/pkg/release/v1"
)

type mockCertificationStatusValidator struct {
	mock.Mock
}

func (m *mockCertificationStatusValidator) IsContainerCertified(registry, repository, tag, digest string) bool {
	args := m.Called(registry, repository, tag, digest)
	return args.Bool(0)
}

func (m *mockCertificationStatusValidator) IsOperatorCertified(csvName, ocpVersion string) bool {
	args := m.Called(csvName, ocpVersion)
	return args.Bool(0)
}

func (m *mockCertificationStatusValidator) IsHelmChartCertified(rel *release.Release, kubeVersion string) bool {
	args := m.Called(rel, kubeVersion)
	return args.Bool(0)
}

func TestOctValidatorAdapter_IsContainerCertified(t *testing.T) {
	inner := new(mockCertificationStatusValidator)
	inner.On("IsContainerCertified", "reg", "repo", "tag", "sha").Return(true)
	
	adapter := &octValidatorAdapter{inner: inner}
	assert.True(t, adapter.IsContainerCertified("reg", "repo", "tag", "sha"))
	inner.AssertExpectations(t)
}

func TestOctValidatorAdapter_IsOperatorCertified(t *testing.T) {
	inner := new(mockCertificationStatusValidator)
	inner.On("IsOperatorCertified", "csv", "4.15").Return(false)
	
	adapter := &octValidatorAdapter{inner: inner}
	assert.False(t, adapter.IsOperatorCertified("csv", "4.15"))
	inner.AssertExpectations(t)
}

func TestOctValidatorAdapter_IsHelmChartCertified(t *testing.T) {
	inner := new(mockCertificationStatusValidator)
	inner.On("IsHelmChartCertified", mock.Anything, "1.25").Return(true)
	
	adapter := &octValidatorAdapter{inner: inner}
	assert.True(t, adapter.IsHelmChartCertified("chart", "1.0.0", "1.25"))
	inner.AssertExpectations(t)
}

func TestNewCertValidator_Error(t *testing.T) {
	v := NewCertValidator("/nonexistent/path")
	_ = v
}
