package csv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCSVProcessor_ProcessRows(t *testing.T) {
	// Create a temporary CSV file for testing
	testCSV := `test_name,input_data,expected_result
test1,data1,result1
test2,data2,result2
test3,data3,result3
test4,data4,result4
test5,data5,result5`

	tmpFile := createTempCSV(t, testCSV)
	defer os.Remove(tmpFile)

	processor := NewProcessor()

	tests := []struct {
		name         string
		rowSelection string
		wantRows     int // including header
		wantErr      bool
	}{
		{
			name:         "single row",
			rowSelection: "2",
			wantRows:     2, // header + row 2
			wantErr:      false,
		},
		{
			name:         "range of rows",
			rowSelection: "2-4",
			wantRows:     4, // header + rows 2,3,4
			wantErr:      false,
		},
		{
			name:         "comma-separated rows",
			rowSelection: "1,3,5",
			wantRows:     4, // header + rows 1,3,5
			wantErr:      false,
		},
		{
			name:         "invalid format",
			rowSelection: "abc",
			wantRows:     0,
			wantErr:      true,
		},
		{
			name:         "empty selection",
			rowSelection: "",
			wantRows:     0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputFile, err := processor.ProcessRows(tmpFile, tt.rowSelection)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessRows() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Check that output file exists and has correct number of rows
			if _, err := os.Stat(outputFile); os.IsNotExist(err) {
				t.Errorf("Output file does not exist: %s", outputFile)
				return
			}

			// Count rows in output file
			content, err := os.ReadFile(outputFile)
			if err != nil {
				t.Errorf("Failed to read output file: %v", err)
				return
			}

			rows := strings.Split(strings.TrimSpace(string(content)), "\n")
			if len(rows) != tt.wantRows {
				t.Errorf("ProcessRows() got %d rows, want %d", len(rows), tt.wantRows)
			}

			// Clean up temp file
			os.Remove(outputFile)
		})
	}
}

func TestCSVProcessor_ParseRowSelection(t *testing.T) {
	processor := NewProcessor()

	tests := []struct {
		name         string
		rowSelection string
		wantRows     []int
		wantErr      bool
	}{
		{
			name:         "single row",
			rowSelection: "2",
			wantRows:     []int{2},
			wantErr:      false,
		},
		{
			name:         "range",
			rowSelection: "2-5",
			wantRows:     []int{2, 3, 4, 5},
			wantErr:      false,
		},
		{
			name:         "comma-separated",
			rowSelection: "1,3,5",
			wantRows:     []int{1, 3, 5},
			wantErr:      false,
		},
		{
			name:         "invalid format",
			rowSelection: "abc",
			wantRows:     nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := processor.ParseRowSelection(tt.rowSelection)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRowSelection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !equalSlices(rows, tt.wantRows) {
				t.Errorf("ParseRowSelection() = %v, want %v", rows, tt.wantRows)
			}
		})
	}
}

func createTempCSV(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.csv")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp CSV: %v", err)
	}

	return tmpFile
}

func equalSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
