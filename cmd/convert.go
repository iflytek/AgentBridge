package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ai-agents-transformer/core"
	"ai-agents-transformer/internal/models"

	"github.com/spf13/cobra"
)

// NewConvertCmd creates the convert command
func NewConvertCmd() *cobra.Command {
	var convertCmd = &cobra.Command{
		Use:   "convert",
		Short: "Cross-platform DSL format conversion",
		Long: `Convert DSL formats between AI agent platforms with iFlytek Spark as the central hub.

🎯 Supported Conversion Paths (Star Architecture):
  • iFlytek Spark ↔ Dify Platform    ✅ Full Bidirectional
  • iFlytek Spark ↔ Coze Platform    ✅ Full Bidirectional
  • Support for Coze ZIP format      ✅ Auto-detection
  • Dify ↔ Coze                      ❌ Not Supported (use iFlytek as hub)

📋 Technical Features:
  • Unified DSL intermediate representation
  • Unsupported node placeholder conversion
  • Variable reference resolution
  • Platform-specific configuration preservation

🚀 Built on unified DSL standards, ensuring data integrity throughout the conversion process.`,
		Example: `  # Basic conversion (iFlytek to Dify)
  ai-agent-converter convert --from iflytek --to dify --input agent.yml --output dify.yml

  # Convert iFlytek to Coze (supports ZIP format)
  ai-agent-converter convert --from iflytek --to coze --input agent.yml --output coze.yml

  # Convert Coze ZIP to iFlytek
  ai-agent-converter convert --from coze --to iflytek --input workflow.zip --output agent.yml

  # Auto-detect source platform
  ai-agent-converter convert --to coze --input agent.yml --output coze.yml

  # Detailed conversion process
  ai-agent-converter convert --from iflytek --to coze --input agent.yml --output coze.yml --verbose`,
		RunE: runConvert,
	}

	// Configure convert command flags
	convertCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input DSL file path (required)")
	convertCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output DSL file path (required)")
	convertCmd.Flags().StringVar(&sourceType, "from", "", "Source platform (iflytek|dify|coze, auto-detect if not specified)")
	convertCmd.Flags().StringVar(&targetType, "to", "", "Target platform (iflytek|dify|coze) (required)")
	
	// Mark required flags
	convertCmd.MarkFlagRequired("input")
	convertCmd.MarkFlagRequired("output")
	convertCmd.MarkFlagRequired("to")

	return convertCmd
}

// runConvert executes the conversion command
func runConvert(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	// Set verbose environment variable for parsers
	if verbose {
		os.Setenv("AI_AGENT_VERBOSE", "true")
	} else {
		os.Setenv("AI_AGENT_VERBOSE", "false")
	}

	if err := executeConversionPipeline(startTime); err != nil {
		return err
	}

	return nil
}

// executeConversionPipeline handles the complete conversion pipeline
func executeConversionPipeline(startTime time.Time) error {
	// Step 1: Initialize and validate input
	inputData, err := initializeAndValidateInput()
	if err != nil {
		return err
	}

	// Step 2: Detect and validate source format
	if err := detectAndValidateSourceFormat(inputData); err != nil {
		return err
	}

	// Step 3: Execute the conversion
	outputData, err := executeConversion(inputData)
	if err != nil {
		return err
	}

	// Step 4: Write output and report results
	return writeOutputAndReport(inputData, outputData, startTime)
}

// initializeAndValidateInput initializes UI and validates input file
func initializeAndValidateInput() ([]byte, error) {
	if !quiet {
		printHeader("Cross-Platform DSL Conversion")
	}

	// Validate input file
	if err := validateInputFile(inputFile); err != nil {
		return nil, fmt.Errorf("input file validation failed: %w", err)
	}

	// Read input file
	if verbose {
		fmt.Printf("📖 Reading input file: %s\n", inputFile)
	}

	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	if verbose {
		fmt.Printf("   File size: %d bytes\n", len(inputData))
	}

	return inputData, nil
}

// detectAndValidateSourceFormat detects source format and validates conversion path
func detectAndValidateSourceFormat(inputData []byte) error {
	// Auto-detect source format (if not specified)
	if sourceType == "" {
		sourceType = detectSourceType(inputData)
		if verbose {
			fmt.Printf("🔍 Auto-detected source platform: %s\n", sourceType)
		}
	}

	// Validate format types
	return validateFormatTypes(sourceType, targetType)
}

// executeConversion performs the actual DSL conversion
func executeConversion(inputData []byte) ([]byte, error) {
	if verbose {
		fmt.Printf("🔄 Starting conversion: %s → %s\n", sourceType, targetType)
	}

	var outputData []byte
	var err error
	
	switch {
	case sourceType == "iflytek" && targetType == "dify":
		outputData, err = convertBetweenPlatforms(inputData, models.PlatformIFlytek, models.PlatformDify)
	case sourceType == "dify" && targetType == "iflytek":
		outputData, err = convertBetweenPlatforms(inputData, models.PlatformDify, models.PlatformIFlytek)
	case sourceType == "iflytek" && targetType == "coze":
		outputData, err = convertBetweenPlatforms(inputData, models.PlatformIFlytek, models.PlatformCoze)
	case sourceType == "coze" && targetType == "iflytek":
		outputData, err = convertBetweenPlatforms(inputData, models.PlatformCoze, models.PlatformIFlytek)
	default:
		return nil, fmt.Errorf("unsupported conversion path: %s → %s", sourceType, targetType)
	}

	if err != nil {
		return nil, fmt.Errorf("conversion failed: %w", err)
	}

	return outputData, nil
}

// writeOutputAndReport writes output file and reports conversion results
func writeOutputAndReport(inputData, outputData []byte, startTime time.Time) error {
	// Create output directory
	if err := createOutputDirectory(); err != nil {
		return err
	}

	// Write output file
	if err := writeOutputFile(outputData); err != nil {
		return err
	}

	// Report results
	reportConversionResults(inputData, outputData, startTime)
	return nil
}

// createOutputDirectory creates the output directory if needed
func createOutputDirectory() error {
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	return nil
}

// writeOutputFile writes the converted data to output file
func writeOutputFile(outputData []byte) error {
	if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	return nil
}

// reportConversionResults reports the conversion results to user
func reportConversionResults(inputData, outputData []byte, startTime time.Time) {
	if quiet {
		return
	}

	elapsed := time.Since(startTime)
	
	fmt.Printf("✅ Conversion completed successfully!\n")
	fmt.Printf("   Input file: %s (%d bytes)\n", inputFile, len(inputData))
	fmt.Printf("   Output file: %s (%d bytes)\n", outputFile, len(outputData))
	fmt.Printf("   Conversion path: %s → %s\n", sourceType, targetType)
	fmt.Printf("   Duration: %v\n", elapsed)
	fmt.Printf("   Throughput: %.2f KB/s\n", float64(len(inputData))/1024/elapsed.Seconds())
}

// convertBetweenPlatforms performs conversion between platforms
func convertBetweenPlatforms(inputData []byte, fromPlatform, toPlatform models.PlatformType) ([]byte, error) {
	// Initialize conversion service
	conversionService, err := core.InitializeArchitecture()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize architecture: %w", err)
	}

	// Execute conversion
	outputData, err := conversionService.Convert(inputData, fromPlatform, toPlatform)
	if err != nil {
		return nil, fmt.Errorf("conversion failed: %w", err)
	}

	if verbose {
		fmt.Printf("   Conversion completed\n")
	}

	return outputData, nil
}