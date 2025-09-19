package openapi

import (
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// IsCRDDefinedWithOpenAPI3Schema Checks if a CRD uses an OpenAPI v3 schema
//
// The function inspects each version of the provided CustomResourceDefinition,
// converting its schema definition to a string. It searches for the substring
// that identifies an OpenAPI v3 schema, ignoring case. If any version contains
// this substring, it returns true; otherwise, it returns false.
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
