package diff

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
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
			assert.Nil(t, err)

			fields := traverse(data, "", nil)
			assert.Equal(t, tc.expectedFields, fields)
		})
	}
}

func Test_traverseWithFilters(t *testing.T) {
	testCases := []struct {
		name           string
		JSONdata       string
		filters        []string
		expectedFields []field
	}{
		{
			name: "complex object with slices of objects, but show only field1 diffs",
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
			filters: []string{"field1"},
			expectedFields: []field{
				{Path: "/field1/0/internalField1", Value: "hello"},
				{Path: "/field1/1/internalField2", Value: "goodbye"},
			},
		},
		{
			name: "complex object with slices of objects, but show only internalField3 diffs",
			JSONdata: `
			{
			  "field1" : [
				{ "internalField1" : "hello" },
				{ "internalField2" : "goodbye"}
			  ],
	  		  "field2": "value2",
			  "field3": {
				"internalField3": {
					"interestingField" : ["hello3", "goodbye3"],
					"notInterestingField": 10
				},
				"internalField4": "field4Value"
			  }
			}`,
			filters: []string{"interestingField"},
			expectedFields: []field{
				{Path: "/field3/internalField3/interestingField/0", Value: "hello3"},
				{Path: "/field3/internalField3/interestingField/1", Value: "goodbye3"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var data interface{}
			err := json.Unmarshal([]byte(tc.JSONdata), &data)
			assert.Nil(t, err)

			fields := traverse(data, "", tc.filters)
			assert.Equal(t, tc.expectedFields, fields)
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
			assert.Nil(t, err)

			var data2 interface{}
			err = json.Unmarshal([]byte(tc.JSONData2), &data2)
			assert.Nil(t, err)

			differences := Compare(tc.objectName, data1, data2, nil)
			t.Logf("Expected: %+v", tc.expectedDiffs)
			t.Logf("Actual  : %+v", *differences)

			assert.Equal(t, tc.expectedDiffs, differences)
		})
	}
}

func TestCompareWithFilters(t *testing.T) {
	// Two json objects defined:
	// - They have same field "field1"
	// - "field7", inside "field5", has different values.
	// - JSONData1 has a <field8=value1> inside "field4" that is missing in JSONData1.
	JSONData1 := `
	{
		"field1" : {
			"field2": "value1",
			"field3": "value2"
		},
		"field4": {
			"field5": {
				"field6": 10,
				"field7": "hello"
			},
			"field8": "value1"
		}
	}`

	JSONData2 := `
	{
		"field1" : {
			"field2": "value1",
			"field3": "value2"
		},
		"field4": {
			"field5": {
				"field6": 10,
				"field7": "goodbye"
			}
		}
	}`

	testCases := []struct {
		name          string
		objectName    string
		filters       []string
		expectedDiffs *Diffs
	}{
		{
			name:       "Filter by field1, which is equal in both JSON objects",
			objectName: "Test",
			filters:    []string{"field1"},
			expectedDiffs: &Diffs{
				Name: "Test",
			},
		},
		{
			name:       "Filter by field5, which has a subfield field7 with different values.",
			objectName: "Test",
			filters:    []string{"field5"},
			expectedDiffs: &Diffs{
				Name: "Test",
				Fields: []FieldDiff{
					{
						FieldPath:   "/field4/field5/field7",
						Claim1Value: "hello",
						Claim2Value: "goodbye"},
				},
			},
		},
		{
			name:       "Filter by field4.",
			objectName: "Test",
			filters:    []string{"field4"},
			expectedDiffs: &Diffs{
				Name: "Test",
				Fields: []FieldDiff{
					{
						FieldPath:   "/field4/field5/field7",
						Claim1Value: "hello",
						Claim2Value: "goodbye",
					},
				},
				FieldsInClaim1Only: []string{"/field4/field8=value1"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var data1 interface{}
			err := json.Unmarshal([]byte(JSONData1), &data1)
			assert.Nil(t, err)

			var data2 interface{}
			err = json.Unmarshal([]byte(JSONData2), &data2)
			assert.Nil(t, err)

			differences := Compare(tc.objectName, data1, data2, tc.filters)
			t.Logf("Expected: %+v", tc.expectedDiffs)
			t.Logf("Actual  : %+v", *differences)

			assert.Equal(t, tc.expectedDiffs, differences)
		})
	}
}
