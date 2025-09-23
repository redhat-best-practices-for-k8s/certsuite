package diff

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Diffs Captures differences between two JSON objects
//
// This structure records fields that differ, as well as those present only in
// one of the compared claims. It stores the object name for contextual output
// and provides a method to format the differences into a readable table. The
// fields are populated by comparing flattened representations of each claim.
type Diffs struct {
	// Name of the json object whose diffs are stored here.
	// It will be used when serializing the data in table format.
	Name string
	// CNI Fields that appear in both claim Fields but their values are different.
	Fields []FieldDiff

	FieldsInClaim1Only []string
	FieldsInClaim2Only []string
}

// FieldDiff Represents a mismatch between two claim files
//
// This structure records the location of a differing field along with its value
// from each claim file. It is used during comparison to track which fields
// differ, enabling further processing or reporting. The field path indicates
// where in the document the discrepancy occurs.
type FieldDiff struct {
	FieldPath   string      `json:"field"`
	Claim1Value interface{} `json:"claim1Value"`
	Claim2Value interface{} `json:"claim2Value"`
}

// Diffs.String Formats a readable report of claim differences
//
// The method builds a string that lists fields with differing values between
// two claims, as well as fields present only in one claim or the other. It
// calculates column widths based on longest field paths and values to align the
// table neatly. If no differences exist it displays a placeholder indicating
// none were found.
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

// Compare Compares two JSON structures for differences
//
// This function takes two interface values that were previously unmarshaled
// from JSON, walks each tree to collect paths and values, then compares
// corresponding entries. It records mismatched values, fields present only in
// the first object, and fields present only in the second. Optional filters
// allow limiting comparison to specified subtrees.
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

// field represents a node in the traversal result
//
// This structure holds the full path to a value and the value itself as
// encountered during tree walking. The Path string records the hierarchical
// location using delimiters, while Value captures any type of data found at
// that point. It is used by the traversal routine to aggregate matching fields
// for comparison.
type field struct {
	Path  string
	Value interface{}
}

// traverse recursively collects leaf paths and values from a nested data structure
//
// The function walks through maps, slices, or simple values, building a path
// string for each leaf node separated by slashes. It optionally filters the
// collected fields based on provided substrings in the path. The result is a
// slice of field structs containing the full path and the corresponding value.
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
