package resultsspreadsheet

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/sheets/v4"
)

func TestPrepareRecordsForSpreadSheet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		records  [][]string
		validate func(t *testing.T, rows []*sheets.RowData)
	}{
		{
			name:    "empty input",
			records: [][]string{},
			validate: func(t *testing.T, rows []*sheets.RowData) {
				assert.Nil(t, rows)
			},
		},
		{
			name:    "normal records",
			records: [][]string{{"hello", "world"}, {"foo", "bar"}},
			validate: func(t *testing.T, rows []*sheets.RowData) {
				require.Len(t, rows, 2)
				require.Len(t, rows[0].Values, 2)
				assert.Equal(t, "hello", *rows[0].Values[0].UserEnteredValue.StringValue)
				assert.Equal(t, "world", *rows[0].Values[1].UserEnteredValue.StringValue)
			},
		},
		{
			name:    "cell content exceeding 50000 chars gets truncated",
			records: [][]string{{strings.Repeat("x", 60000)}},
			validate: func(t *testing.T, rows []*sheets.RowData) {
				require.Len(t, rows, 1)
				val := *rows[0].Values[0].UserEnteredValue.StringValue
				assert.LessOrEqual(t, len(val), cellContentLimit)
			},
		},
		{
			name:    "empty cells processed without panic",
			records: [][]string{{""}},
			validate: func(t *testing.T, rows []*sheets.RowData) {
				require.Len(t, rows, 1)
				assert.NotNil(t, rows[0].Values[0].UserEnteredValue.StringValue)
			},
		},
		{
			name:    "newlines replaced",
			records: [][]string{{"line1\nline2"}, {"line1\r\nline2"}},
			validate: func(t *testing.T, rows []*sheets.RowData) {
				require.Len(t, rows, 2)
				assert.NotContains(t, *rows[0].Values[0].UserEnteredValue.StringValue, "\n")
				assert.NotContains(t, *rows[1].Values[0].UserEnteredValue.StringValue, "\r\n")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rows := prepareRecordsForSpreadSheet(tc.records)
			tc.validate(t, rows)
		})
	}
}

func TestGetHeaderIndicesByColumnNames(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		headers        []string
		columnNames    []string
		expectedResult []int
		expectedErrMsg string
	}{
		{
			name:           "all columns found",
			headers:        []string{"Name", "Age", "City"},
			columnNames:    []string{"Age", "City"},
			expectedResult: []int{1, 2},
		},
		{
			name:           "missing column returns error",
			headers:        []string{"Name", "Age"},
			columnNames:    []string{"Missing"},
			expectedErrMsg: "column Missing doesn't exist in given headers list",
		},
		{
			name:           "empty headers",
			headers:        []string{},
			columnNames:    []string{"Name"},
			expectedErrMsg: "column Name doesn't exist in given headers list",
		},
		{
			name:           "single column found",
			headers:        []string{"A", "B", "C"},
			columnNames:    []string{"A"},
			expectedResult: []int{0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := GetHeaderIndicesByColumnNames(tc.headers, tc.columnNames)
			if tc.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErrMsg, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestExtractFolderIDFromURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		url            string
		expectedID     string
		expectedErrMsg string
	}{
		{
			name:       "valid Google Drive folder URL",
			url:        "https://drive.google.com/drive/folders/1AbCdEfGhIjKlMnOp",
			expectedID: "1AbCdEfGhIjKlMnOp",
		},
		{
			name:       "URL with trailing slash",
			url:        "https://drive.google.com/drive/folders/1AbCdEfGhIjKlMnOp/",
			expectedID: "",
		},
		{
			name:       "simple path",
			url:        "https://example.com/folderid123",
			expectedID: "folderid123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			id, err := extractFolderIDFromURL(tc.url)
			if tc.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedID, id)
			}
		})
	}
}
