package diff

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
)

func Test_traverse(t *testing.T) {
	testCases := []struct {
		name           string
		JSONdata       string
		expectedFields []field
	}{
		{
			name:           "empty object",
			JSONdata:       "{}",
			expectedFields: []field{},
		},
		{
			name:     "object with one field only",
			JSONdata: `{"field1" : "value1"}`,
			expectedFields: []field{
				{Path: "/field1", Value: string("value1")},
			},
		},
		{
			name:     "object with two fields",
			JSONdata: `{"field1" : "value1", "field2": 5}`,
			expectedFields: []field{
				{Path: "/field1", Value: "value1"},
				{Path: "/field2", Value: float64(5)},
			},
		},
		{
			name:     "object with a field with another field inside",
			JSONdata: `{"field1" : { "internalField1" : "hello" } }`,
			expectedFields: []field{
				{Path: "/field1/internalField1", Value: "hello"},
			},
		},
		{
			name: "object with a field that is an array of objects with one field",
			JSONdata: `{"field1" : [
				{ "internalField1" : "hello" },
				{ "internalField2" : "goodbye"}
			  ]
			}`,
			expectedFields: []field{
				{Path: "/field1/0/internalField1", Value: "hello"},
				{Path: "/field1/1/internalField2", Value: "goodbye"},
			},
		},
		{
			name: "complex object with slices of objects",
			JSONdata: `
			{
			  "field1" : [
				{ "internalField1" : "hello" },
				{ "internalField2" : "goodbye"}
			  ],
	  		  "field2": "value2",
			  "field3": {
				"internalField3": ["hello3", "goodbye3"]
			  }
			}`,
			expectedFields: []field{
				{Path: "/field1/0/internalField1", Value: "hello"},
				{Path: "/field1/1/internalField2", Value: "goodbye"},
				{Path: "/field2", Value: "value2"},
				{Path: "/field3/internalField3/0", Value: "hello3"},
				{Path: "/field3/internalField3/1", Value: "goodbye3"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var data interface{}
			err := json.Unmarshal([]byte(tc.JSONdata), &data)
			assert.NilError(t, err)

			fields := traverse(data, "")
			assert.DeepEqual(t, tc.expectedFields, fields)
		})
	}
}

func TestCompare(t *testing.T) {
	testCases := []struct {
		name          string
		objectName    string
		JSONData1     string
		JSONData2     string
		expectedDiffs *Diffs
	}{
		{
			name:          "nil objects",
			objectName:    "nil",
			JSONData1:     "{}",
			JSONData2:     "{}",
			expectedDiffs: &Diffs{Name: "nil"},
		},
		{
			name:          "Equal objects with single field",
			objectName:    "Test",
			JSONData1:     `{ "field1" : "value1" }`,
			JSONData2:     `{ "field1" : "value1" }`,
			expectedDiffs: &Diffs{Name: "Test"},
		},
		{
			name:       "Equal complex objects with matching fields",
			objectName: "Test",
			JSONData1: `
			{
				"field1": [{
						"internalField1": "hello"
					}, {
						"internalField2": "goodbye"
					}
				],
				"field2": "value2",
				"field3": {
					"internalField3": ["hello3", "goodbye3"]
				}
			}`,
			JSONData2: `
			{
				"field1": [{
						"internalField1": "hello"
					}, {
						"internalField2": "goodbye"
					}
				],
				"field2": "value2",
				"field3": {
					"internalField3": ["hello3", "goodbye3"]
				}
			}`,
			expectedDiffs: &Diffs{Name: "Test"},
		},
		{
			name:       "Different complex objects 1: two non matching values",
			objectName: "Test",
			JSONData1: `
			{
				"field1": [{
						"internalField1": "hi"
					}, {
						"internalField2": "goodbye"
					}
				],
				"field2": "value2",
				"field3": {
					"internalField3": ["hi3", "goodbye3"]
				}
			}`,
			JSONData2: `
			{
				"field1": [{
						"internalField1": "hello"
					}, {
						"internalField2": "goodbye"
					}
				],
				"field2": "value2",
				"field3": {
					"internalField3": ["hello3", "goodbye3"]
				}
			}`,
			expectedDiffs: &Diffs{
				Name: "Test",
				Fields: []FieldDiff{
					{
						FieldPath:   "/field1/0/internalField1",
						Claim1Value: string("hi"),
						Claim2Value: string("hello"),
					},
					{
						FieldPath:   "/field3/internalField3/0",
						Claim1Value: string("hi3"),
						Claim2Value: string("hello3"),
					},
				},
			},
		},
		{
			name:       "Object1 has a field missing in object",
			objectName: "Test",
			JSONData1:  `{ "field1" : "value1", "field2": "value2" }`,
			JSONData2:  `{ "field1" : "value1" }`,
			expectedDiffs: &Diffs{
				Name:               "Test",
				FieldsInClaim1Only: []string{"/field2=value2"}},
		},
		{
			name:       "Object2 has a field missing in object1",
			objectName: "Test",
			JSONData1:  `{ "field1" : "value1" }`,
			JSONData2:  `{ "field1" : "value1", "field2": "value2" }`,
			expectedDiffs: &Diffs{
				Name:               "Test",
				FieldsInClaim2Only: []string{"/field2=value2"}},
		},
		{
			name:       "Different field1 value and missing fields in both objects",
			objectName: "Test",
			JSONData1:  `{ "field1" : "value1", "field2": "value3" }`,
			JSONData2:  `{ "field1" : "value2", "field3": "value4" }`,
			expectedDiffs: &Diffs{
				Name: "Test",
				Fields: []FieldDiff{
					{
						FieldPath:   "/field1",
						Claim1Value: string("value1"),
						Claim2Value: string("value2"),
					},
				},
				FieldsInClaim1Only: []string{"/field2=value3"},
				FieldsInClaim2Only: []string{"/field3=value4"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var data1 interface{}
			err := json.Unmarshal([]byte(tc.JSONData1), &data1)
			assert.NilError(t, err)

			var data2 interface{}
			err = json.Unmarshal([]byte(tc.JSONData2), &data2)
			assert.NilError(t, err)

			differences := Compare(tc.objectName, data1, data2)
			t.Logf("Expected: %+v", tc.expectedDiffs)
			t.Logf("Actual  : %+v", *differences)

			assert.DeepEqual(t, tc.expectedDiffs, differences)
		})
	}
}
