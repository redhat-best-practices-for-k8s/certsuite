package failures

import "fmt"

// Custom object type needed to provide a different JSON serialization than
// the one in claim's test cases' skipReason field.
type NonCompliantObject struct {
	Type   string     `json:"type"`
	Reason string     `json:"reason"`
	Spec   ObjectSpec `json:"spec"`
}

type ObjectSpec struct {
	Fields []struct{ Key, Value string }
}

func (spec *ObjectSpec) AddField(key, value string) {
	spec.Fields = append(spec.Fields, struct {
		Key   string
		Value string
	}{key, value})
}

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

type FailedTestCase struct {
	TestCaseName        string               `json:"name"`
	TestCaseDescription string               `json:"description"`
	SkipReason          string               `json:"skipReason,omitempty"`
	NonCompliantObjects []NonCompliantObject `json:"nonCompliantObjects,omitempty"`
}

type FailedTestSuite struct {
	TestSuiteName    string           `json:"name"`
	FailingTestCases []FailedTestCase `json:"failures"`
}
