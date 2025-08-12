package openapi

import (
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// IsCRDDefinedWithOpenAPI3Schema reports whether a CustomResourceDefinition includes an OpenAPI v3 schema.
//
// It examines the Spec.Versions slice of the given CRD, looking for a version that has
// a non-nil Schema field with an OpenAPIV3Schema defined. If such a schema exists,
// it returns true; otherwise it returns false. The function does not modify the CRD.
func IsCRDDefinedWithOpenAPI3Schema(crd *apiextv1.CustomResourceDefinition) bool {
	for _, version := range crd.Spec.Versions {
		crdSchema := version.Schema.String()

		containsOpenAPIV3SchemaSubstr := strings.Contains(strings.ToLower(crdSchema),
			strings.ToLower(testhelper.OpenAPIV3Schema))

		if containsOpenAPIV3SchemaSubstr {
			return true
		}
	}

	return false
}
