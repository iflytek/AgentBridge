package integration

import (
	"os"
	"path/filepath"
	"testing"

	"agentbridge/core"
	"agentbridge/internal/models"
)

// TestDifyToIFlytekToCoze tests multi-step conversion: Dify → iFlytek → Coze
func TestDifyToIFlytekToCoze(t *testing.T) {
	t.Log("Starting multi-step conversion test: Dify → iFlytek → Coze")

	// Input and intermediate/output file paths
	inputFile := "../fixtures/dify/dify_公众号文本生成.yml"
	intermediateFile := "test_output/dify_to_iflytek_intermediate.yml"
	outputFile := "test_output/dify_iflytek_coze_final.yml"

	// Initialize conversion architecture
	conversionService, err := core.InitializeArchitecture()
	if err != nil {
		t.Fatalf("Failed to initialize conversion architecture: %v", err)
	}

	// Step 1: Dify → iFlytek
	t.Log("Step 1: Converting Dify → iFlytek")
	t.Logf("Reading Dify file: %s", inputFile)
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to read input file: %v", err)
	}
	t.Logf("Input file size: %d bytes", len(inputData))

	// Convert Dify to iFlytek
	intermediateData, err := conversionService.Convert(inputData, models.PlatformDify, models.PlatformIFlytek)
	if err != nil {
		t.Fatalf("Step 1 conversion failed (Dify → iFlytek): %v", err)
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

	// Step 2: iFlytek → Coze
	t.Log("Step 2: Converting iFlytek → Coze")
	outputData, err := conversionService.Convert(intermediateData, models.PlatformIFlytek, models.PlatformCoze)
	if err != nil {
		t.Fatalf("Step 2 conversion failed (iFlytek → Coze): %v", err)
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
	t.Log("Multi-step conversion test completed successfully: Dify → iFlytek → Coze")
}

// TestDifyToIFlytekToCozeWithCustomFile tests multi-step conversion with custom file from environment variable
func TestDifyToIFlytekToCozeWithCustomFile(t *testing.T) {
	customFile := os.Getenv("CUSTOM_DIFY_FILE")
	if customFile == "" {
		t.Skip("CUSTOM_DIFY_FILE environment variable not set, skipping custom file test")
	}

	t.Logf("Starting multi-step conversion test with custom file: %s", customFile)
	t.Log("Conversion path: Dify → iFlytek → Coze")

	// Output file paths
	intermediateFile := "test_output/dify_to_iflytek_custom_intermediate.yml"
	outputFile := "test_output/dify_iflytek_coze_custom_final.yml"

	// Initialize conversion architecture
	conversionService, err := core.InitializeArchitecture()
	if err != nil {
		t.Fatalf("Failed to initialize conversion architecture: %v", err)
	}

	// Step 1: Dify → iFlytek
	t.Log("Step 1: Converting custom Dify file → iFlytek")
	t.Logf("Reading custom Dify file: %s", customFile)
	inputData, err := os.ReadFile(customFile)
	if err != nil {
		t.Fatalf("Failed to read custom input file: %v", err)
	}
	t.Logf("Custom input file size: %d bytes", len(inputData))

	// Convert Dify to iFlytek
	intermediateData, err := conversionService.Convert(inputData, models.PlatformDify, models.PlatformIFlytek)
	if err != nil {
		t.Fatalf("Step 1 conversion failed (Dify → iFlytek): %v", err)
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

	// Step 2: iFlytek → Coze
	t.Log("Step 2: Converting iFlytek → Coze")
	outputData, err := conversionService.Convert(intermediateData, models.PlatformIFlytek, models.PlatformCoze)
	if err != nil {
		t.Fatalf("Step 2 conversion failed (iFlytek → Coze): %v", err)
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
	t.Log("Custom file multi-step conversion test completed successfully: Dify → iFlytek → Coze")
}
