package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/iflytek/agentbridge/core"
	"github.com/iflytek/agentbridge/internal/models"
)

// TestCozeToIFlytekToDify tests multi-step conversion: Coze → iFlytek → Dify
func TestCozeToIFlytekToDify(t *testing.T) {
	t.Log("Starting multi-step conversion test: Coze → iFlytek → Dify")

	// Input and intermediate/output file paths
	inputFile := "../fixtures/coze/coze1111t.yml"
	intermediateFile := "test_output/coze_公众号文本生成.yml"
	outputFile := "test_output/coze_iflytek_dify_final.yml"

	// Initialize conversion architecture
	conversionService, err := core.InitializeArchitecture()
	if err != nil {
		t.Fatalf("Failed to initialize conversion architecture: %v", err)
	}

	// Step 1: Coze → iFlytek
	t.Log("Step 1: Converting Coze → iFlytek")
	t.Logf("Reading Coze file: %s", inputFile)
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to read input file: %v", err)
	}
	t.Logf("Input file size: %d bytes", len(inputData))

	// Convert Coze to iFlytek
	intermediateData, err := conversionService.Convert(inputData, models.PlatformCoze, models.PlatformIFlytek)
	if err != nil {
		t.Fatalf("Step 1 conversion failed (Coze → iFlytek): %v", err)
	}

	// Validate intermediate result
	if len(intermediateData) == 0 {
		t.Fatal("Step 1 conversion result is empty")
	}
	t.Logf("Step 1 successful! Intermediate file size: %d bytes", len(intermediateData))

	// Save intermediate file for debugging
	err = os.MkdirAll(filepath.Dir(intermediateFile), 0o755)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	err = os.WriteFile(intermediateFile, intermediateData, 0o644)
	if err != nil {
		t.Fatalf("Failed to save intermediate file: %v", err)
	}
	t.Logf("Intermediate result saved: %s", intermediateFile)

	// Step 2: iFlytek → Dify
	t.Log("Step 2: Converting iFlytek → Dify")
	outputData, err := conversionService.Convert(intermediateData, models.PlatformIFlytek, models.PlatformDify)
	if err != nil {
		t.Fatalf("Step 2 conversion failed (iFlytek → Dify): %v", err)
	}

	// Validate final result
	if len(outputData) == 0 {
		t.Fatal("Step 2 conversion result is empty")
	}
	t.Logf("Step 2 successful! Final file size: %d bytes", len(outputData))

	// Save final output file
	err = os.WriteFile(outputFile, outputData, 0o644)
	if err != nil {
		t.Fatalf("Failed to save final output file: %v", err)
	}

	t.Logf("Final conversion result saved: %s", outputFile)
	t.Log("Multi-step conversion test completed successfully: Coze → iFlytek → Dify")
}

// TestCozeToIFlytekToDifyWithCustomFile tests multi-step conversion with custom file from environment variable
func TestCozeToIFlytekToDifyWithCustomFile(t *testing.T) {
	customFile := os.Getenv("CUSTOM_COZE_FILE")
	if customFile == "" {
		t.Skip("CUSTOM_COZE_FILE environment variable not set, skipping custom file test")
	}

	t.Logf("Starting multi-step conversion test with custom file: %s", customFile)
	t.Log("Conversion path: Coze → iFlytek → Dify")

	// Output file paths
	intermediateFile := "test_output/coze_to_iflytek_custom_intermediate.yml"
	outputFile := "test_output/coze_iflytek_dify_custom_final.yml"

	// Initialize conversion architecture
	conversionService, err := core.InitializeArchitecture()
	if err != nil {
		t.Fatalf("Failed to initialize conversion architecture: %v", err)
	}

	// Step 1: Coze → iFlytek
	t.Log("Step 1: Converting custom Coze file → iFlytek")
	t.Logf("Reading custom Coze file: %s", customFile)
	inputData, err := os.ReadFile(customFile)
	if err != nil {
		t.Fatalf("Failed to read custom input file: %v", err)
	}
	t.Logf("Custom input file size: %d bytes", len(inputData))

	// Convert Coze to iFlytek
	intermediateData, err := conversionService.Convert(inputData, models.PlatformCoze, models.PlatformIFlytek)
	if err != nil {
		t.Fatalf("Step 1 conversion failed (Coze → iFlytek): %v", err)
	}

	// Validate intermediate result
	if len(intermediateData) == 0 {
		t.Fatal("Step 1 conversion result is empty")
	}
	t.Logf("Step 1 successful! Intermediate file size: %d bytes", len(intermediateData))

	// Save intermediate file for debugging
	err = os.MkdirAll(filepath.Dir(intermediateFile), 0o755)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	err = os.WriteFile(intermediateFile, intermediateData, 0o644)
	if err != nil {
		t.Fatalf("Failed to save intermediate file: %v", err)
	}
	t.Logf("Intermediate result saved: %s", intermediateFile)

	// Step 2: iFlytek → Dify
	t.Log("Step 2: Converting iFlytek → Dify")
	outputData, err := conversionService.Convert(intermediateData, models.PlatformIFlytek, models.PlatformDify)
	if err != nil {
		t.Fatalf("Step 2 conversion failed (iFlytek → Dify): %v", err)
	}

	// Validate final result
	if len(outputData) == 0 {
		t.Fatal("Step 2 conversion result is empty")
	}
	t.Logf("Step 2 successful! Final file size: %d bytes", len(outputData))

	// Save final output file
	err = os.WriteFile(outputFile, outputData, 0o644)
	if err != nil {
		t.Fatalf("Failed to save final output file: %v", err)
	}

	t.Logf("Final conversion result saved: %s", outputFile)
	t.Log("Custom file multi-step conversion test completed successfully: Coze → iFlytek → Dify")
}
