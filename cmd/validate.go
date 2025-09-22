package main

import (
	"fmt"
	"os"

	"ai-agents-transformer/core"
	"ai-agents-transformer/internal/models"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewValidateCmd creates the validate command
func NewValidateCmd() *cobra.Command {
	var validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Validate DSL file format",
		Long: `Validate DSL file format correctness, checking for syntax errors and structural integrity.

Supports automatic platform detection and comprehensive validation rules.`,
		Example: `  # Validate iFlytek Spark Agent DSL file
  ai-agent-converter validate --input agent.yml --from iflytek

  # Validate Dify DSL file
  ai-agent-converter validate --input dify.yml --from dify

  # Auto-detect format and validate
  ai-agent-converter validate --input workflow.yml`,
		RunE: runValidate,
	}

	// Configure validate command flags
	validateCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input DSL file path (required)")
	validateCmd.Flags().StringVar(&sourceType, "from", "", "Source platform (iflytek|dify|coze, auto-detect if not specified)")

	// Mark required flags
	validateCmd.MarkFlagRequired("input")

	return validateCmd
}

// runValidate executes the validation command
func runValidate(cmd *cobra.Command, args []string) error {
	restore := redirectStdoutIfQuiet()
	defer restore()
	if quiet {
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
	}
	if !quiet {
		printHeader("DSL File Validation")
	}

	// Prepare validation context
	ctx, err := prepareValidationContext()
	if err != nil {
		return err
	}

	// Execute validation
	validationErrors, err := executeValidation(ctx)
	if err != nil {
		return err
	}

	// Output results
	return outputValidationResults(ctx, validationErrors)
}

// validationContext holds validation context data
type validationContext struct {
	inputFile  string
	inputData  []byte
	sourceType string
}

// prepareValidationContext prepares validation context
func prepareValidationContext() (*validationContext, error) {
	// Validate input file
	if err := validateInputFile(inputFile); err != nil {
		return nil, fmt.Errorf("input file validation failed: %w", err)
	}

	// Read input file
	if verbose {
		fmt.Printf("📖 Reading file: %s\n", inputFile)
	}

	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Auto-detect format if not specified
	detectedType := sourceType
	if detectedType == "" {
		detectedType = detectSourceType(inputData)
		if verbose {
			fmt.Printf("🔍 Detected platform: %s\n", detectedType)
		}
	}

	return &validationContext{
		inputFile:  inputFile,
		inputData:  inputData,
		sourceType: detectedType,
	}, nil
}

// executeValidation executes DSL validation
func executeValidation(ctx *validationContext) ([]string, error) {
	if verbose {
		fmt.Printf("🔍 Validating DSL format: %s\n", ctx.sourceType)
	}

	switch ctx.sourceType {
	case "iflytek":
		return validateIflytekDSL(ctx.inputData), nil
	case "dify":
		return validateDifyDSL(ctx.inputData), nil
	case "coze":
		return validateCozeDSL(ctx.inputData), nil
	default:
		return nil, fmt.Errorf("unsupported platform type: %s", ctx.sourceType)
	}
}

// outputValidationResults outputs validation results
func outputValidationResults(ctx *validationContext, validationErrors []string) error {
	if len(validationErrors) == 0 {
		return outputValidationSuccess(ctx)
	}
	return outputValidationFailure(validationErrors)
}

// outputValidationSuccess outputs success results
func outputValidationSuccess(ctx *validationContext) error {
	if !quiet {
		fmt.Printf("✅ DSL file validation passed\n")
		fmt.Printf("   File: %s\n", ctx.inputFile)
		fmt.Printf("   Platform: %s\n", ctx.sourceType)
		fmt.Printf("   Size: %d bytes\n", len(ctx.inputData))
	}
	return nil
}

// outputValidationFailure outputs failure results
func outputValidationFailure(validationErrors []string) error {
	fmt.Printf("❌ DSL file validation failed, found %d issues:\n", len(validationErrors))
	for i, err := range validationErrors {
		fmt.Printf("   %d. %s\n", i+1, err)
	}
	return fmt.Errorf("validation failed")
}

// validateIflytekDSL validates iFlytek DSL format
func validateIflytekDSL(data []byte) []string {
	var errors []string

	// Basic YAML format validation
	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		errors = append(errors, fmt.Sprintf("YAML format error: %v", err))
		return errors
	}

	// Check required fields
	if _, exists := yamlData["flowMeta"]; !exists {
		errors = append(errors, "missing required field: flowMeta")
	}

	if _, exists := yamlData["flowData"]; !exists {
		errors = append(errors, "missing required field: flowData")
	}

	// Use conversion service for detailed validation
	conversionService, err := core.InitializeArchitecture()
	if err != nil {
		errors = append(errors, fmt.Sprintf("failed to initialize architecture: %v", err))
		return errors
	}

	if err := conversionService.ValidateSourceData(data, models.PlatformIFlytek); err != nil {
		errors = append(errors, fmt.Sprintf("structural validation failed: %v", err))
	}

	return errors
}

// validateDifyDSL validates Dify DSL format
func validateDifyDSL(data []byte) []string {
	var errors []string

	// Basic YAML format validation
	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		errors = append(errors, fmt.Sprintf("YAML format error: %v", err))
		return errors
	}

	// Check required fields
	requiredFields := []string{"app", "workflow", "kind", "version"}
	for _, field := range requiredFields {
		if _, exists := yamlData[field]; !exists {
			errors = append(errors, fmt.Sprintf("missing required field: %s", field))
		}
	}

	return errors
}

// validateCozeDSL validates Coze DSL format
func validateCozeDSL(data []byte) []string {
	var errors []string

	// Basic YAML format validation
	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		errors = append(errors, fmt.Sprintf("YAML format error: %v", err))
		return errors
	}

	// Check for Coze DSL structure
	if _, exists := yamlData["workflow_id"]; !exists {
		if _, exists := yamlData["name"]; !exists {
			errors = append(errors, "missing required Coze DSL fields: workflow_id or name")
		}
	}

	return errors
}
