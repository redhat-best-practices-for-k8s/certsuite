package failures

import "fmt"

// NonCompliantObject represents an object that failed compliance checks.
//
// It contains a human‑readable reason, the type of the object, and a spec that
// holds the relevant fields. The JSON serialization is customized to match the
// format expected by the claim's skipReason field.
type NonCompliantObject struct {
	Type   string     `json:"type"`
	Reason string     `json:"reason"`
	Spec   ObjectSpec `json:"spec"`
}

// ObjectSpec holds a list of key/value pairs that describe an object.
//
// It stores its data in the Fields slice, where each element contains a Key and Value string.
// The AddField method appends a new pair to this slice.
// MarshalJSON produces a JSON representation of the fields as a single-level map,
// returning the encoded bytes and any error encountered.
type ObjectSpec struct {
	Fields []struct{ Key, Value string }
}

// AddField appends a new field to an object specification.
//
// It takes two string arguments, the field name and its value,
// and appends them as a key/value pair to the ObjectSpec's internal
// representation. The method does not return any value; it mutates
// the receiver in place.
func (spec *ObjectSpec) AddField(key, value string) {
	spec.Fields = append(spec.Fields, struct {
		Key   string
		Value string
	}{key, value})
}

// MarshalJSON serializes the ObjectSpec into JSON format.
//
// It returns a byte slice containing the JSON representation of the
// ObjectSpec and an error if serialization fails. The method uses the
// standard encoding/json package to encode the struct fields. If any
// field cannot be encoded, an error is returned. The resulting JSON
// bytes can be written directly to output or further processed.
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

// FailedTestCase represents a test case that did not pass.
//
// It contains the name of the test case, a description of what it checks,
// details about why it failed, and a slice of objects that caused the failure.
// The struct is used to report non-compliant resources in claim show
// output.
type FailedTestCase struct {
	TestCaseName        string               `json:"name"`
	TestCaseDescription string               `json:"description"`
	CheckDetails        string               `json:"checkDetails,omitempty"`
	NonCompliantObjects []NonCompliantObject `json:"nonCompliantObjects,omitempty"`
}

// FailedTestSuite represents a test suite that contains failing tests.
//
// It holds the name of the test suite and a slice of the individual failed test cases within it, allowing callers to inspect or report failures per suite.
type FailedTestSuite struct {
	TestSuiteName    string           `json:"name"`
	FailingTestCases []FailedTestCase `json:"failures"`
}
