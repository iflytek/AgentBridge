package integration

import (
	"os"
	"path/filepath"
	"testing"

	"ai-agents-transformer/core"
	"ai-agents-transformer/internal/models"
)

// TestDifyToIFlytek tests Dify to iFlytek platform conversion.
func TestDifyToIFlytek(t *testing.T) {
	t.Log("Starting Dify → iFlytek conversion test")

	// Input and output file paths
	inputFile := "../fixtures/dify/dify_basic_start_end.yml"
	outputFile := "test_output/dify_to_iflytek_converted.yml"

	// Initialize conversion architecture
	conversionService, err := core.InitializeArchitecture()
	if err != nil {
		t.Fatalf("Failed to initialize conversion architecture: %v", err)
	}

	// Read input file
	t.Logf("Reading Dify file: %s", inputFile)
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	t.Logf("Input file size: %d bytes", len(inputData))

	// Execute conversion
	t.Log("Starting conversion: Dify → iFlytek")
	outputData, err := conversionService.Convert(inputData, models.PlatformDify, models.PlatformIFlytek)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Validate conversion result
	if len(outputData) == 0 {
		t.Fatal("Conversion result is empty")
	}

	// Log conversion result
	t.Logf("Conversion successful!")
	t.Logf("Output file size: %d bytes", len(outputData))

	// Basic conversion validation completed

	// Create output directory and save file
	err = os.MkdirAll(filepath.Dir(outputFile), 0o755)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	err = os.WriteFile(outputFile, outputData, 0o644)
	if err != nil {
		t.Fatalf("Failed to save output file: %v", err)
	}

	t.Logf("Conversion result saved: %s", outputFile)
	t.Log("Test completed!")
}
