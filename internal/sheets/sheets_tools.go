package sheets

import (
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
)

// === Handle functions - generated via WrapHandler ===

var (
	HandleSheetsGet        = common.WrapHandler[SheetsService](TestableSheetsGet)
	HandleSheetsRead       = common.WrapHandler[SheetsService](TestableSheetsRead)
	HandleSheetsWrite      = common.WrapHandler[SheetsService](TestableSheetsWrite)
	HandleSheetsAppend     = common.WrapHandler[SheetsService](TestableSheetsAppend)
	HandleSheetsCreate     = common.WrapHandler[SheetsService](TestableSheetsCreate)
	HandleSheetsBatchRead  = common.WrapHandler[SheetsService](TestableSheetsBatchRead)
	HandleSheetsBatchWrite = common.WrapHandler[SheetsService](TestableSheetsBatchWrite)
	HandleSheetsClear      = common.WrapHandler[SheetsService](TestableSheetsClear)
)

// parseValues converts various input formats to [][]any for Sheets API.
func parseValues(input any) ([][]any, error) {
	if input == nil {
		return nil, fmt.Errorf("values cannot be nil")
	}

	// Handle [][]any directly
	if rows, ok := input.([][]any); ok {
		return rows, nil
	}

	// Handle []any (array of rows)
	rows, ok := input.([]any)
	if !ok {
		return nil, fmt.Errorf("values must be a 2D array (array of rows)")
	}

	result := make([][]any, 0, len(rows))
	for i, row := range rows {
		rowSlice, ok := row.([]any)
		if !ok {
			return nil, fmt.Errorf("row %d must be an array", i)
		}
		result = append(result, rowSlice)
	}

	return result, nil
}
