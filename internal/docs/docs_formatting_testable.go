package docs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
)

// textFormatFields defines the parameter-to-field mapping for text formatting.
// Each entry maps an argument name to its corresponding Docs API field mask name
// and an optional validation function.
var textFormatFields = []argFieldMapping{
	{arg: "bold", field: "bold", kind: kindBool},
	{arg: "italic", field: "italic", kind: kindBool},
	{arg: "underline", field: "underline", kind: kindBool},
	{arg: "strikethrough", field: "strikethrough", kind: kindBool},
	{arg: "small_caps", field: "smallCaps", kind: kindBool},
	{arg: "font_family", field: "weightedFontFamily", kind: kindNonEmptyString},
	{arg: "font_size", field: "fontSize", kind: kindPositiveFloat},
	{arg: "foreground_color", field: "foregroundColor", kind: kindNonEmptyString, validate: validateColor("foreground_color")},
	{arg: "background_color", field: "backgroundColor", kind: kindNonEmptyString, validate: validateColor("background_color")},
	{arg: "baseline_offset", field: "baselineOffset", kind: kindNonEmptyString, validate: validateEnum("baseline_offset", "NONE", "SUPERSCRIPT", "SUBSCRIPT")},
}

// paragraphStyleFields defines the parameter-to-field mapping for paragraph styles.
var paragraphStyleFields = []argFieldMapping{
	{arg: "alignment", field: "alignment", kind: kindNonEmptyString, validate: validateEnum("alignment", "START", "CENTER", "END", "JUSTIFIED")},
	{arg: "named_style_type", field: "namedStyleType", kind: kindNonEmptyString, validate: validateEnum("named_style_type", "NORMAL_TEXT", "TITLE", "SUBTITLE", "HEADING_1", "HEADING_2", "HEADING_3", "HEADING_4", "HEADING_5", "HEADING_6")},
	{arg: "line_spacing", field: "lineSpacing", kind: kindPositiveFloat},
	{arg: "indent_start", field: "indentStart", kind: kindFloat},
	{arg: "indent_end", field: "indentEnd", kind: kindFloat},
	{arg: "indent_first_line", field: "indentFirstLine", kind: kindFloat},
	{arg: "space_above", field: "spaceAbove", kind: kindFloat},
	{arg: "space_below", field: "spaceBelow", kind: kindFloat},
}

type fieldKind int

const (
	kindBool           fieldKind = iota // present as bool
	kindNonEmptyString                  // present as non-empty string
	kindFloat                           // present as float64
	kindPositiveFloat                   // present as float64 > 0
)

type argFieldMapping struct {
	arg      string
	field    string
	kind     fieldKind
	validate func(args map[string]any) *mcp.CallToolResult // optional validation returning error result
}

// collectFields iterates over field mappings, checks which arguments are present,
// runs optional validation, and returns the collected field names or an error result.
func collectFields(args map[string]any, mappings []argFieldMapping) ([]string, *mcp.CallToolResult) {
	var fields []string
	for _, m := range mappings {
		if !argPresent(args, m.arg, m.kind) {
			continue
		}
		if m.validate != nil {
			if errResult := m.validate(args); errResult != nil {
				return nil, errResult
			}
		}
		fields = append(fields, m.field)
	}
	return fields, nil
}

// argPresent checks whether the argument is present and satisfies the kind constraint.
func argPresent(args map[string]any, name string, kind fieldKind) bool {
	switch kind {
	case kindBool:
		_, ok := args[name].(bool)
		return ok
	case kindNonEmptyString:
		s, ok := args[name].(string)
		return ok && s != ""
	case kindFloat:
		_, ok := args[name].(float64)
		return ok
	case kindPositiveFloat:
		f, ok := args[name].(float64)
		return ok && f > 0
	}
	return false
}

// validateColor returns a validation function that parses a hex color argument.
func validateColor(argName string) func(args map[string]any) *mcp.CallToolResult {
	return func(args map[string]any) *mcp.CallToolResult {
		color := args[argName].(string)
		if _, _, _, err := parseColor(color); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid %s: %v", argName, err))
		}
		return nil
	}
}

// validateEnum returns a validation function that checks a string argument against allowed values.
func validateEnum(argName string, allowed ...string) func(args map[string]any) *mcp.CallToolResult {
	valid := make(map[string]bool, len(allowed))
	for _, v := range allowed {
		valid[v] = true
	}
	errMsg := fmt.Sprintf("%s must be one of: %s", argName, strings.Join(allowed, ", "))
	return func(args map[string]any) *mcp.CallToolResult {
		val := args[argName].(string)
		if !valid[val] {
			return mcp.NewToolResultError(errMsg)
		}
		return nil
	}
}

// TestableDocsFormatText is the testable version of HandleDocsFormatText.
func TestableDocsFormatText(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, _ := request.Params.Arguments["document_id"].(string)
	if docID == "" {
		return mcp.NewToolResultError("document_id parameter is required"), nil
	}

	startIndex, endIndex, errResult := extractIndexRange(request)
	if errResult != nil {
		return errResult, nil
	}

	docID = common.ExtractGoogleResourceID(docID)

	fields, validationErr := collectFields(request.Params.Arguments, textFormatFields)
	if validationErr != nil {
		return validationErr, nil
	}
	if len(fields) == 0 {
		return mcp.NewToolResultError("at least one formatting option must be specified"), nil
	}

	_, err := srv.BatchUpdate(ctx, docID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	result := map[string]any{
		"success":        true,
		"document_id":    docID,
		"fields_updated": fields,
		"message":        fmt.Sprintf("Applied formatting to text from index %d to %d", startIndex, endIndex),
		"url":            fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}

// TestableDocsClearFormatting is the testable version of HandleDocsClearFormatting.
func TestableDocsClearFormatting(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, _ := request.Params.Arguments["document_id"].(string)
	if docID == "" {
		return mcp.NewToolResultError("document_id parameter is required"), nil
	}

	startIndex, endIndex, errResult := extractIndexRange(request)
	if errResult != nil {
		return errResult, nil
	}

	docID = common.ExtractGoogleResourceID(docID)

	_, err := srv.BatchUpdate(ctx, docID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	result := map[string]any{
		"success":     true,
		"document_id": docID,
		"message":     fmt.Sprintf("Cleared formatting from text at index %d to %d", startIndex, endIndex),
		"url":         fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}

// TestableDocsSetParagraphStyle is the testable version of HandleDocsSetParagraphStyle.
func TestableDocsSetParagraphStyle(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, _ := request.Params.Arguments["document_id"].(string)
	if docID == "" {
		return mcp.NewToolResultError("document_id parameter is required"), nil
	}

	startIndex, endIndex, errResult := extractIndexRange(request)
	if errResult != nil {
		return errResult, nil
	}

	docID = common.ExtractGoogleResourceID(docID)

	fields, validationErr := collectFields(request.Params.Arguments, paragraphStyleFields)
	if validationErr != nil {
		return validationErr, nil
	}
	if len(fields) == 0 {
		return mcp.NewToolResultError("at least one paragraph style option must be specified"), nil
	}

	_, err := srv.BatchUpdate(ctx, docID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	result := map[string]any{
		"success":        true,
		"document_id":    docID,
		"fields_updated": fields,
		"message":        fmt.Sprintf("Applied paragraph style to range %d to %d", startIndex, endIndex),
		"url":            fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}

// TestableDocsCreateList is the testable version of HandleDocsCreateList.
func TestableDocsCreateList(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, _ := request.Params.Arguments["document_id"].(string)
	if docID == "" {
		return mcp.NewToolResultError("document_id parameter is required"), nil
	}

	startIndex, endIndex, errResult := extractIndexRange(request)
	if errResult != nil {
		return errResult, nil
	}

	bulletPreset := "BULLET_DISC_CIRCLE_SQUARE"
	if preset, ok := request.Params.Arguments["bullet_preset"].(string); ok && preset != "" {
		validPresets := map[string]bool{
			"BULLET_DISC_CIRCLE_SQUARE":              true,
			"BULLET_DIAMONDX_ARROW3D_SQUARE":         true,
			"BULLET_CHECKBOX":                        true,
			"BULLET_ARROW_DIAMOND_DISC":              true,
			"BULLET_STAR_CIRCLE_SQUARE":              true,
			"BULLET_ARROW3D_CIRCLE_SQUARE":           true,
			"BULLET_LEFTTRIANGLE_DIAMOND_DISC":       true,
			"BULLET_DIAMONDX_HOLLOWDIAMOND_SQUARE":   true,
			"BULLET_DIAMOND_CIRCLE_SQUARE":           true,
			"NUMBERED_DECIMAL_ALPHA_ROMAN":           true,
			"NUMBERED_DECIMAL_ALPHA_ROMAN_PARENS":    true,
			"NUMBERED_DECIMAL_NESTED":                true,
			"NUMBERED_UPPERALPHA_ALPHA_ROMAN":        true,
			"NUMBERED_UPPERROMAN_UPPERALPHA_DECIMAL": true,
			"NUMBERED_ZERODECIMAL_ALPHA_ROMAN":       true,
		}
		if !validPresets[preset] {
			return mcp.NewToolResultError("invalid bullet_preset - see Google Docs API documentation for valid values"), nil
		}
		bulletPreset = preset
	}

	docID = common.ExtractGoogleResourceID(docID)

	_, err := srv.BatchUpdate(ctx, docID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	listType := "bulleted"
	if strings.HasPrefix(bulletPreset, "NUMBERED_") {
		listType = "numbered"
	}

	result := map[string]any{
		"success":       true,
		"document_id":   docID,
		"list_type":     listType,
		"bullet_preset": bulletPreset,
		"message":       fmt.Sprintf("Created %s list from index %d to %d", listType, startIndex, endIndex),
		"url":           fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}

// TestableDocsRemoveList is the testable version of HandleDocsRemoveList.
func TestableDocsRemoveList(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, _ := request.Params.Arguments["document_id"].(string)
	if docID == "" {
		return mcp.NewToolResultError("document_id parameter is required"), nil
	}

	startIndex, endIndex, errResult := extractIndexRange(request)
	if errResult != nil {
		return errResult, nil
	}

	docID = common.ExtractGoogleResourceID(docID)

	_, err := srv.BatchUpdate(ctx, docID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	result := map[string]any{
		"success":     true,
		"document_id": docID,
		"message":     fmt.Sprintf("Removed list formatting from index %d to %d", startIndex, endIndex),
		"url":         fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}
