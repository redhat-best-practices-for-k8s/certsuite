package failures

import "fmt"

// Custom object type needed to provide a different JSON serialization than
// the one in claim's test cases' failureReason field.
type NonCompliantObject struct {
	Type   string     `json:"type"`
	Reason string     `json:"reason"`
	Spec   ObjectSpec `json:"spec"`
}

type ObjectSpec struct {
	fields []struct{ key, value string }
}

func (spec *ObjectSpec) AddField(key, value string) {
	spec.fields = append(spec.fields, struct {
		key   string
		value string
	}{key, value})
}

func (spec *ObjectSpec) MarshalJSON() ([]byte, error) {
	if len(spec.fields) == 0 {
		return []byte("{}"), nil
	}

	specStr := "{"
	for i := range spec.fields {
		if i != 0 {
			specStr += ", "
		}
		specStr += fmt.Sprintf("%q:%q", spec.fields[i].key, spec.fields[i].value)
	}

	specStr += "}"

	return []byte(specStr), nil
}

type FailedTestCase struct {
	TestCaseName        string               `json:"name"`
	TestCaseDescription string               `json:"description"`
	NonCompliantObjects []NonCompliantObject `json:"nonCompliantObjects"`
}

type FailedTestSuite struct {
	TestSuiteName    string           `json:"name"`
	FailingTestCases []FailedTestCase `json:"failures"`
}
