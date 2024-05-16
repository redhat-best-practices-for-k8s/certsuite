package failures

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddField(t *testing.T) {
	spec := ObjectSpec{}
	spec.AddField("key1", "value1")
	assert.Len(t, spec.Fields, 1)
}

func TestMarshalJSON(t *testing.T) {
	testCases := []struct {
		key          string
		value        string
		expectedJSON string
		clearFields  bool
	}{
		{
			key:          "key1",
			value:        "value1",
			expectedJSON: `{"key1":"value1"}`,
			clearFields:  false,
		},
		{
			key:          "key1",
			value:        "value1",
			expectedJSON: `{}`,
			clearFields:  true,
		},
	}

	for _, tc := range testCases {
		spec := ObjectSpec{}
		spec.AddField(tc.key, tc.value)

		if tc.clearFields {
			spec.Fields = nil
			result, err := spec.MarshalJSON()
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedJSON, string(result))
		} else {
			result, err := spec.MarshalJSON()
			assert.Nil(t, err)
			assert.Len(t, spec.Fields, 1)
			assert.Equal(t, tc.expectedJSON, string(result))
		}
	}
}
