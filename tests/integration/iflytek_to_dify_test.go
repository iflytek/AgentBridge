package integration

import (
	"os"
	"path/filepath"
	"testing"

	"agentbridge/core"
	"agentbridge/internal/models"
)

// TestIFlytekToDify tests iFlytek platform to Dify conversion.
func TestIFlytekToDify(t *testing.T) {
	t.Log("Starting iFlytek → Dify conversion test")

	// Input and output file paths
	inputFile := "../fixtures/iflytek/星辰agent_公众号文本生成.yml"
	outputFile := "test_output/iflytek_to_dify_converted.yml"

	// Initialize conversion architecture
	conversionService, err := core.InitializeArchitecture()
	if err != nil {
		t.Fatalf("Failed to initialize conversion architecture: %v", err)
	}

	// Read input file
	t.Logf("Reading iFlytek file: %s", inputFile)
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	t.Logf("Input file size: %d bytes", len(inputData))

	// Execute conversion
	t.Log("Starting conversion: iFlytek → Dify")
	outputData, err := conversionService.Convert(inputData, models.PlatformIFlytek, models.PlatformDify)
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
