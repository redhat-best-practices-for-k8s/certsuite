package failures

import "fmt"

// NonCompliantObject represents a non‑compliant object extracted from failure data
//
// This type holds information about objects that failed compliance checks,
// including the object's kind, the reason for failure, and its specification
// details. The Spec field aggregates key/value pairs representing the object's
// configuration at the time of the check. Instances are created by parsing JSON
// output from a compliance test and converting it into a more convenient
// structure for reporting.
type NonCompliantObject struct {
	Type   string     `json:"type"`
	Reason string     `json:"reason"`
	Spec   ObjectSpec `json:"spec"`
}

// ObjectSpec Represents a collection of key/value pairs for JSON output
//
// This structure holds an ordered list of fields where each field has a string
// key and value. It provides methods to add new fields and to marshal the
// collection into a valid JSON object. If no fields are present, marshaling
// returns an empty JSON object.
type ObjectSpec struct {
	Fields []struct{ Key, Value string }
}

// ObjectSpec.AddField Adds a key/value pair to the object's specification
//
// This method appends a new field containing the provided key and value strings
// to the spec's internal slice of fields. It does not return any value or
// perform validation, simply extending the slice. The updated spec can then be
// used elsewhere to represent object metadata.
func (spec *ObjectSpec) AddField(key, value string) {
	spec.Fields = append(spec.Fields, struct {
		Key   string
		Value string
	}{key, value})
}

// ObjectSpec.MarshalJSON Converts the ObjectSpec into JSON bytes
//
// The method checks if there are any fields; if none, it returns an empty JSON
// object. Otherwise, it builds a JSON string by iterating over each field and
// formatting key/value pairs as quoted strings separated by commas. The
// resulting byte slice is returned with no error.
func (spec *ObjectSpec) MarshalJSON() ([]byte, error) {
	if len(spec.Fields) == 0 {
		return []byte("{}"), nil
	}

	specStr := "{"
	for i := range spec.Fields {
		if i != 0 {
			specStr += ", "
		}
		specStr += fmt.Sprintf("%q:%q", spec.Fields[i].Key, spec.Fields[i].Value)
	}

	specStr += "}"

	return []byte(specStr), nil
}

// FailedTestCase Represents a test case that did not pass
//
// It holds the name and description of the test case, optional details about
// the check, and any objects that failed to meet compliance criteria. The
// structure is used to aggregate failure information for reporting or logging
// purposes.
type FailedTestCase struct {
	TestCaseName        string               `json:"name"`
	TestCaseDescription string               `json:"description"`
	CheckDetails        string               `json:"checkDetails,omitempty"`
	NonCompliantObjects []NonCompliantObject `json:"nonCompliantObjects,omitempty"`
}

// FailedTestSuite represents a test suite with failures
//
// This struct holds the name of a test suite and a list of its failing test
// cases. It is used when reporting or displaying results, allowing consumers to
// see which specific tests failed within each suite.
type FailedTestSuite struct {
	TestSuiteName    string           `json:"name"`
	FailingTestCases []FailedTestCase `json:"failures"`
}
