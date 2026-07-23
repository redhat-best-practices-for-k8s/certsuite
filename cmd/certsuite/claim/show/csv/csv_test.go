package csv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildCatalogByID(t *testing.T) {
	catalogMap := buildCatalogByID()
	assert.NotNil(t, catalogMap)

	t.Run("catalog map is populated from identifiers.Catalog", func(t *testing.T) {
		assert.GreaterOrEqual(t, len(catalogMap), 1)
	})

	t.Run("entries are keyed by ID string", func(t *testing.T) {
		for id, desc := range catalogMap {
			assert.NotEmpty(t, id)
			assert.NotEmpty(t, desc.Identifier.Id)
			assert.Equal(t, id, desc.Identifier.Id)
		}
	})
}
