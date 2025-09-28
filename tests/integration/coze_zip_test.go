package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iflytek/agentbridge/core"
	"github.com/iflytek/agentbridge/internal/models"
)

// TestCozeZipToIFlytek tests Coze ZIP format to iFlytek conversion.
func TestCozeZipToIFlytek(t *testing.T) {
	t.Log("Starting Coze ZIP to iFlytek conversion test")

	// Read actual ZIP file
	zipFile := "../fixtures/coze/Workflow-X74_Wcaisehuochairen_video_1-draft-2293.zip"
	zipData, err := os.ReadFile(zipFile)
	if err != nil {
		t.Fatalf("Failed to read ZIP file: %v", err)
	}

	// Output file path
	outputFile := "test_output/coze_zip_to_iflytek_converted.yml"

	// Initialize conversion architecture
	conversionService, err := core.InitializeArchitecture()
	if err != nil {
		t.Fatalf("Failed to initialize conversion architecture: %v", err)
	}

	// Execute conversion - from Coze ZIP to iFlytek
	t.Log("Starting conversion: Coze ZIP → iFlytek")
	outputData, err := conversionService.Convert(zipData, models.PlatformCoze, models.PlatformIFlytek)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Validate conversion result
	if len(outputData) == 0 {
		t.Fatal("Conversion result is empty")
	}

	// Log conversion result
	t.Logf("Conversion successful!")
	t.Logf("Input ZIP file: %s", zipFile)
	t.Logf("Output file size: %d bytes", len(outputData))

	// Validate output contains basic workflow elements
	outputString := string(outputData)
	if !strings.Contains(outputString, "nodes:") {
		t.Error("Output missing node definitions")
	}
	if !strings.Contains(outputString, "edges:") {
		t.Error("Output missing edge definitions")
	}

	// Create output directory and save file
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
		t.Fatalf("Failed to save output file: %v", err)
	}

	t.Logf("Conversion result saved to: %s", outputFile)
	t.Log("Coze ZIP to iFlytek conversion test completed")
}
