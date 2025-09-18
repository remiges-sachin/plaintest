package csv

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Processor struct{}

func NewProcessor() *Processor {
	return &Processor{}
}

func (p *Processor) ProcessRows(csvFile string, rowSelection string) (string, error) {
	if rowSelection == "" {
		return "", errors.New("row selection is required")
	}

	// Parse which rows to include
	rows, err := p.ParseRowSelection(rowSelection)
	if err != nil {
		return "", fmt.Errorf("invalid row selection: %w", err)
	}

	// Create temporary output file
	tmpDir := os.TempDir()
	outputFile := filepath.Join(tmpDir, fmt.Sprintf("plaintest_rows_%s.csv", strings.ReplaceAll(rowSelection, ",", "_")))

	// Read input CSV
	input, err := os.Open(csvFile)
	if err != nil {
		return "", fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer input.Close()

	// Create output CSV
	output, err := os.Create(outputFile)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	scanner := bufio.NewScanner(input)
	writer := bufio.NewWriter(output)
	defer writer.Flush()

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Always include header (line 1)
		if lineNum == 1 {
			_, _ = writer.WriteString(line + "\n")
			continue
		}

		// Include line if it's in our selection (convert to 0-based index)
		dataRowNum := lineNum - 1
		if p.containsRow(rows, dataRowNum) {
			_, _ = writer.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading CSV: %w", err)
	}

	return outputFile, nil
}

func (p *Processor) ParseRowSelection(rowSelection string) ([]int, error) {
	if rowSelection == "" {
		return nil, errors.New("empty row selection")
	}

	// Single number: "2"
	if matched, _ := regexp.MatchString(`^[0-9]+$`, rowSelection); matched {
		num, err := strconv.Atoi(rowSelection)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", rowSelection)
		}
		return []int{num}, nil
	}

	// Range: "2-5"
	if matched, _ := regexp.MatchString(`^[0-9]+-[0-9]+$`, rowSelection); matched {
		parts := strings.Split(rowSelection, "-")
		start, err1 := strconv.Atoi(parts[0])
		end, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("invalid range: %s", rowSelection)
		}
		if start > end {
			return nil, fmt.Errorf("invalid range: start > end")
		}

		var rows []int
		for i := start; i <= end; i++ {
			rows = append(rows, i)
		}
		return rows, nil
	}

	// Comma-separated: "1,3,5"
	if matched, _ := regexp.MatchString(`^[0-9,]+$`, rowSelection); matched {
		parts := strings.Split(rowSelection, ",")
		var rows []int
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			num, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid number in list: %s", part)
			}
			rows = append(rows, num)
		}
		return rows, nil
	}

	return nil, fmt.Errorf("invalid row selection format: %s", rowSelection)
}

func (p *Processor) containsRow(rows []int, target int) bool {
	for _, row := range rows {
		if row == target {
			return true
		}
	}
	return false
}
