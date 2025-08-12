package diff

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Diffs holds the differences between two interface{} objects that have
// been obtained by unmarshalling JSON strings.
type Diffs struct {
	// Name of the json object whose diffs are stored here.
	// It will be used when serializing the data in table format.
	Name string
	// CNI Fields that appear in both claim Fields but their values are different.
	Fields []FieldDiff

	FieldsInClaim1Only []string
	FieldsInClaim2Only []string
}

// FieldDIff holds the field path and the values from both claim files
// that have been found to be different.
type FieldDiff struct {
	FieldPath   string      `json:"field"`
	Claim1Value interface{} `json:"claim1Value"`
	Claim2Value interface{} `json:"claim2Value"`
}

// Stringer method. The output string is a table like this:
// <name>: Differences
// FIELD                           CLAIM 1     CLAIM 2
// /jsonpath/to/field1             value1      value2
// /jsonpath/to/another/field2     value3      value4
// ...
//
// <name>: Only in CLAIM 1
// /jsonpath/to/field/in/claim1/only
// ...
//
// <name>: Only in CLAIM 2
// /jsonpath/to/field/in/claim2/only
// ...
//
// Where <name> is replaced by the value of d.Name.
// The columns "FIELD" and "CLAIM 1" have a dynamic width that depends
// on the longest field path and longest value.
func (d *Diffs) String() string {
	const (
		noDiffs        = "<none>"
		columnsGapSize = 5
	)

	// Get the length of the longest field path so we can use it as the column size.
	maxFieldPathLength := len("FIELD")
	// Same for the column for the values from the claim1 file.
	maxClaim1FieldValueLength := len("CLAIM 1")
	for _, diff := range d.Fields {
		fieldPathLength := len(diff.FieldPath)
		if fieldPathLength > maxFieldPathLength {
			maxFieldPathLength = len(diff.FieldPath)
		}

		claim1ValueLength := len(fmt.Sprint(diff.Claim1Value))
		if claim1ValueLength > maxClaim1FieldValueLength {
			maxClaim1FieldValueLength = claim1ValueLength
		}
	}

	// Add an extra gap to avoid columns to appear too close.
	fieldRowLen := maxFieldPathLength + columnsGapSize
	claim1FieldValueRowLen := maxClaim1FieldValueLength + columnsGapSize

	// Create the format string using those dynamic widths.
	cniDiffRowFmt := "%-" + fmt.Sprint(fieldRowLen) + "s%-" + fmt.Sprint(claim1FieldValueRowLen) + "v%-v\n"

	// Generate a line per different field with their values in both claim files.
	str := d.Name + ": Differences\n"
	str += fmt.Sprintf(cniDiffRowFmt, "FIELD", "CLAIM 1", "CLAIM 2")
	if len(d.Fields) != 0 {
		for _, diff := range d.Fields {
			str += fmt.Sprintf(cniDiffRowFmt, diff.FieldPath, diff.Claim1Value, diff.Claim2Value)
		}
	} else {
		str += noDiffs + "\n"
	}

	// Generate a line per field that was found in claim1 only.
	str += "\n" + d.Name + ": Only in CLAIM 1\n"
	if len(d.FieldsInClaim1Only) > 0 {
		for _, field := range d.FieldsInClaim1Only {
			str += field + "\n"
		}
	} else {
		str += noDiffs + "\n"
	}

	// Generate a line per field that was found in claim2 only.
	str += "\n" + d.Name + ": Only in CLAIM 2\n"
	if len(d.FieldsInClaim2Only) > 0 {
		for _, field := range d.FieldsInClaim2Only {
			str += field + "\n"
		}
	} else {
		str += noDiffs + "\n"
	}

	return str
}

// Compares to interface{} objects obtained through json.Unmarshal() and returns
// a pointer to a Diffs object.
// A simple filtering of json subtrees can be achieved using the filters slice parameter.
// This might be helpful with big json trees that could have too many potential differences,
// but we want to get just the differences for some custom nodes/subtrees.
// E.g.: filters = []string{"labels"} : only the nodes/subtrees under all the
// labels nodes/branches will be traversed and compared.
func Compare(objectName string, claim1Object, claim2Object interface{}, filters []string) *Diffs {
	objectsDiffs := Diffs{Name: objectName}

	claim1Fields := traverse(claim1Object, "", filters)
	claim2Fields := traverse(claim2Object, "", filters)

	// Build helper maps, to make it easier to find fields.
	claim1FieldsMap := map[string]interface{}{}
	for _, field := range claim1Fields {
		claim1FieldsMap[field.Path] = field.Value
	}

	claim2FieldsMap := map[string]interface{}{}
	for _, field := range claim2Fields {
		claim2FieldsMap[field.Path] = field.Value
	}

	// Start comparing, keeping the original order.
	for _, claim1Field := range claim1Fields {
		// Does the field (path) in claim1 exist in claim2?
		if claim2Value, exist := claim2FieldsMap[claim1Field.Path]; exist {
			// Do they have the same value?
			if !reflect.DeepEqual(claim1Field.Value, claim2Value) {
				objectsDiffs.Fields = append(objectsDiffs.Fields, FieldDiff{
					FieldPath:   claim1Field.Path,
					Claim1Value: claim1Field.Value,
					Claim2Value: claim2Value})
			}
		} else {
			fieldAndValue := fmt.Sprintf("%s=%v", claim1Field.Path, claim1Field.Value)
			objectsDiffs.FieldsInClaim1Only = append(objectsDiffs.FieldsInClaim1Only, fieldAndValue)
		}
	}

	// Fields that appear in both claim files have been already checked,
	// so we only need to search fields in claim2 that will not exist in claim 1.
	for _, claim2Field := range claim2Fields {
		if _, exist := claim1FieldsMap[claim2Field.Path]; !exist {
			fieldAndValue := fmt.Sprintf("%s=%v", claim2Field.Path, claim2Field.Value)
			objectsDiffs.FieldsInClaim2Only = append(objectsDiffs.FieldsInClaim2Only, fieldAndValue)
		}
	}

	return &objectsDiffs
}

type field struct {
	Path  string
	Value interface{}
}

// Helper function that traverses recursively a node to return a list
// of each field (leaf) path and its value.
func traverse(node interface{}, path string, filters []string) []field {
	if node == nil {
		return nil
	}

	leavePathDelimiter := `/`
	fields := []field{}

	switch value := node.(type) {
	// map object
	case map[string]interface{}:
		// Get all keys for sorting
		keys := make([]string, 0)
		for k := range value {
			keys = append(keys, k)
		}

		// Sort keys
		sort.Strings(keys)
		for _, key := range keys {
			fields = append(fields, traverse(value[key], path+leavePathDelimiter+key, filters)...)
		}
	// list object
	case []interface{}:
		for i, v := range value {
			fields = append(fields, traverse(v, path+leavePathDelimiter+strconv.Itoa(i), filters)...)
		}
	// simple value (int, string...)
	default:
		// No filters: append every field's path=value
		if len(filters) == 0 {
			fields = append(fields, field{
				Path:  path,
				Value: value,
			})
		}

		// Append field's whose path matches some filter.
		for _, filter := range filters {
			if strings.Contains(path, "/"+filter+"/") {
				fields = append(fields, field{
					Path:  path,
					Value: value,
				})
			}
		}
	}

	return fields
}
