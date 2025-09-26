package main

import (
	"agentbridge/internal/models"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// Common variables used across commands
var (
	inputFile    string
	outputFile   string
	inputDir     string
	outputDir    string
	sourceType   string
	targetType   string
	pattern      string
	showNodes    bool
	showTypes    bool
	showAll      bool
	showDetailed bool
)

// printHeader prints a formatted header
func printHeader(title string) {
	fmt.Printf("🚀 %s %s\n", appName, version)
	fmt.Printf("📋 %s\n", title)
	fmt.Println(strings.Repeat("=", 60))
}

// validateInputFile validates that the input file exists and has correct format
func validateInputFile(filename string) error {
	if filename == "" {
		return fmt.Errorf("input file path cannot be empty")
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", filename)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".yml" && ext != ".yaml" && ext != ".zip" {
		return fmt.Errorf("input file must be in YAML format (.yml or .yaml) or ZIP format (.zip)")
	}

	return nil
}

// detectSourceType auto-detects the source platform type from file content
func detectSourceType(data []byte) string {
	// Detect ZIP signature first; treat any ZIP as Coze export package
	if isZipData(data) {
		return "coze"
	}

	// Use case-insensitive heuristic checks
	content := string(data)
	lower := strings.ToLower(content)

	// iFlytek Spark Agent characteristics
	if strings.Contains(content, "flowMeta") && strings.Contains(content, "flowData") {
		return "iflytek"
	}

	// Dify characteristics
	if strings.Contains(content, "workflow") && strings.Contains(content, "app") && strings.Contains(content, "kind") {
		return "dify"
	}

	// Coze characteristics (be tolerant to different field spellings)
	// - workflow_id (official) or workflowid (observed in fixtures)
	// - export_format hint
	// - schema section with nodes/edges
	// - trigger_parameters key appears in start/end
	if strings.Contains(lower, "workflow_id") ||
		strings.Contains(lower, "workflowid") ||
		strings.Contains(lower, "export_format") ||
		(strings.Contains(lower, "schema:") && (strings.Contains(lower, "nodes:") || strings.Contains(lower, "edges:"))) ||
		strings.Contains(lower, "trigger_parameters") {
		return "coze"
	}

	// Default to iFlytek format
	return "iflytek"
}

// isZipData returns true if data starts with a ZIP file signature ("PK")
func isZipData(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	return data[0] == 'P' && data[1] == 'K'
}

// validateFormatTypes validates source and target platform types
func validateFormatTypes(source, target string) error {
	validTypes := []string{"iflytek", "dify", "coze"}

	isValidSource := false
	isValidTarget := false

	for _, t := range validTypes {
		if source == t {
			isValidSource = true
		}
		if target == t {
			isValidTarget = true
		}
	}

	if !isValidSource {
		return fmt.Errorf("unsupported source platform: %s, supported platforms: %v", source, validTypes)
	}

	if !isValidTarget {
		return fmt.Errorf("unsupported target platform: %s, supported platforms: %v", target, validTypes)
	}

	if source == target {
		return fmt.Errorf("source and target platforms cannot be the same")
	}

	// Validate supported conversion paths (star architecture with iFlytek as hub)
	if (source == "dify" && target == "coze") || (source == "coze" && target == "dify") {
		return fmt.Errorf("direct conversion between %s and %s is not supported. Please use iFlytek as intermediate hub:\n  1. Convert %s → iflytek\n  2. Convert iflytek → %s", source, target, source, target)
	}

	return nil
}

// ErrorCodeMapping represents a mapping from error pattern to user-friendly message
type ErrorCodeMapping struct {
	Pattern     string
	Message     string
	Suggestions []string
	Severity    models.ErrorSeverity
}

// errorCodeMappings defines the mapping table for error codes to user-friendly messages
var errorCodeMappings = []ErrorCodeMapping{
	{
		Pattern: "CONV_000001",
		Message: "Invalid or empty file format",
		Suggestions: []string{
			"Check if the input file is a valid workflow configuration file",
			"Verify the file is not corrupted",
		},
		Severity: models.SeverityError,
	},
	{
		Pattern: "INTERNAL_2",
		Message: "Internal format validation failed",
		Suggestions: []string{
			"Check if the input file follows the correct schema",
			"Try validating the file with the validate command first",
		},
		Severity: models.SeverityError,
	},
	{
		Pattern: "PARSE_FAILED",
		Message: "Failed to parse source file",
		Suggestions: []string{
			"Verify the file format is correct",
			"Check for syntax errors in YAML/JSON",
		},
		Severity: models.SeverityError,
	},
	{
		Pattern: "GENERATION_FAILED",
		Message: "Failed to generate target format",
		Suggestions: []string{
			"Check the conversion configuration",
			"Ensure the source platform is supported",
		},
		Severity: models.SeverityError,
	},
	{
		Pattern: "failed to read",
		Message: "File read operation failed",
		Suggestions: []string{
			"Check the file path and permissions",
			"Ensure the file exists and is readable",
		},
		Severity: models.SeverityError,
	},
	{
		Pattern: "file does not exist",
		Message: "Input file not found",
		Suggestions: []string{
			"Check the file path is correct",
			"Ensure the file exists in the specified location",
		},
		Severity: models.SeverityError,
	},
	{
		Pattern: "unsupported conversion path",
		Message: "Direct conversion not supported",
		Suggestions: []string{
			"Use iFlytek as intermediate hub for Dify ↔ Coze conversion",
			"Convert source → iflytek, then iflytek → target",
		},
		Severity: models.SeverityWarning,
	},
	{
		Pattern: "YAML format error",
		Message: "YAML syntax error detected",
		Suggestions: []string{
			"Check YAML file syntax is correct",
			"Verify proper indentation and structure",
		},
		Severity: models.SeverityError,
	},
	{
		Pattern: "failed to initialize",
		Message: "System initialization failed",
		Suggestions: []string{
			"Retry the operation",
			"Contact support if the issue persists",
		},
		Severity: models.SeverityCritical,
	},
}

// wrapUserFriendlyError converts internal errors to user-friendly messages using error code mapping
func wrapUserFriendlyError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Try to match error patterns using the mapping table
	for _, mapping := range errorCodeMappings {
		if strings.Contains(errStr, mapping.Pattern) {
			// Create a new ConversionError with enhanced information
			convErr := &models.ConversionError{
				Code:        mapping.Pattern,
				Message:     mapping.Message,
				Suggestions: mapping.Suggestions,
				Severity:    mapping.Severity,
				// Preserve original error details for debugging
				Details: errStr,
			}
			return convErr
		}
	}

	// Handle special case for conversion failed errors
	if strings.Contains(errStr, "conversion failed") {
		return &models.ConversionError{
			Code:    "CONV_GENERAL",
			Message: "Conversion operation failed",
			Suggestions: []string{
				"Check source file format",
				"Verify conversion path is supported",
				"Use --verbose for detailed information",
			},
			Severity: models.SeverityError,
			Details:  errStr,
		}
	}

	// For unknown errors, return original error
	return err
}

// markRequiredFlags marks multiple flags as required for a command
func markRequiredFlags(cmd *cobra.Command, flagNames []string) {
	for _, flagName := range flagNames {
		if err := cmd.MarkFlagRequired(flagName); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to mark %s flag as required: %v\n", flagName, err)
		}
	}
}

// redirectStdoutIfQuiet redirects stdout to the OS null device when quiet mode is enabled.
// It returns a restore function to recover the original stdout.
func redirectStdoutIfQuiet() func() {
	if !quiet {
		return func() {}
	}
	old := os.Stdout
	nullDevice := "/dev/null"
	if runtime.GOOS == "windows" {
		nullDevice = "NUL"
	}
	f, err := os.OpenFile(nullDevice, os.O_WRONLY, 0)
	if err != nil {
		return func() {}
	}
	os.Stdout = f
	return func() {
		_ = f.Close()
		os.Stdout = old
	}
}
