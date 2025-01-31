package openapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestIsCRDDefinedWithOpenAPI3Schema(t *testing.T) {
	testCases := []struct {
		testCRD        *apiextv1.CustomResourceDefinition
		expectedOutput bool
	}{
		{
			testCRD: &apiextv1.CustomResourceDefinition{
				Spec: apiextv1.CustomResourceDefinitionSpec{
					Versions: []apiextv1.CustomResourceDefinitionVersion{
						{
							Schema: &apiextv1.CustomResourceValidation{
								OpenAPIV3Schema: &apiextv1.JSONSchemaProps{},
							},
						},
					},
				},
			},
			expectedOutput: true,
		},
		{
			testCRD: &apiextv1.CustomResourceDefinition{
				Spec: apiextv1.CustomResourceDefinitionSpec{
					Versions: []apiextv1.CustomResourceDefinitionVersion{
						{
							Schema: &apiextv1.CustomResourceValidation{},
						},
					},
				},
			},
			expectedOutput: true,
		},
		{
			testCRD: &apiextv1.CustomResourceDefinition{
				Spec: apiextv1.CustomResourceDefinitionSpec{
					Versions: []apiextv1.CustomResourceDefinitionVersion{
						{
							Schema: &apiextv1.CustomResourceValidation{
								OpenAPIV3Schema: &apiextv1.JSONSchemaProps{},
							},
						},
					},
				},
			},
			expectedOutput: true,
		},
		{
			testCRD: &apiextv1.CustomResourceDefinition{
				Spec: apiextv1.CustomResourceDefinitionSpec{
					Versions: nil,
				},
			},
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, IsCRDDefinedWithOpenAPI3Schema(tc.testCRD))
	}
}
