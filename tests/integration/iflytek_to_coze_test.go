package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/iflytek/agentbridge/core"
	"github.com/iflytek/agentbridge/internal/models"
)

// TestIFlytekToCoze tests iFlytek platform to Coze conversion.
func TestIFlytekToCoze(t *testing.T) {
	t.Log("Starting iFlytek → Coze conversion test")

	// Input and output file paths
	inputFile := "../fixtures/iflytek/星辰agent_公众号文本生成.yml"
	outputFile := "test_output/iflytek_to_coze_converted.yml"

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
	t.Log("Starting conversion: iFlytek → Coze")
	outputData, err := conversionService.Convert(inputData, models.PlatformIFlytek, models.PlatformCoze)
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
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
		t.Fatalf("Failed to save output file: %v", err)
	}

	t.Logf("Conversion result saved to: %s", outputFile)
	t.Log("iFlytek → Coze conversion test completed")
}
