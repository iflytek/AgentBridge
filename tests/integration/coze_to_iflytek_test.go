package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/iflytek/agentbridge/core"
	"github.com/iflytek/agentbridge/internal/models"
)

// TestCozeToIFlytek tests Coze DSL to iFlytek DSL conversion.
func TestCozeToIFlytek(t *testing.T) {
	t.Log("Starting Coze → iFlytek conversion test")

	// Input and output file paths
	inputFile := "../fixtures/coze/coze_公众号文本生成.yml"
	outputFile := "test_output/coze_to_iflytek_converted.yml"

	// Initialize conversion architecture
	conversionService, err := core.InitializeArchitecture()
	if err != nil {
		t.Fatalf("Failed to initialize conversion architecture: %v", err)
	}

	// Read input file
	t.Logf("Reading Coze file: %s", inputFile)
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	t.Logf("Input file size: %d bytes", len(inputData))

	// Execute conversion
	t.Log("Starting conversion: Coze → iFlytek")
	outputData, err := conversionService.Convert(inputData, models.PlatformCoze, models.PlatformIFlytek)
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
