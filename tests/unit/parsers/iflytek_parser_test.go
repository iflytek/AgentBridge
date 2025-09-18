package parsers

import (
	"os"
	"path/filepath"
	"testing"

	"ai-agents-transformer/platforms/iflytek/strategies"
	"github.com/stretchr/testify/require"
)

// TestIFlytekParser_BasicStartEnd validates iFlytek parser with basic start-end workflow
func TestIFlytekParser_BasicStartEnd(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewIFlytekStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "iflytek", "iflytek_basic_start_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse iFlytek DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_BasicStartEnd(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ iFlytek BasicStartEnd parser validation passed")
}

// TestIFlytekParser_ClassifierWorkflow validates iFlytek parser with classifier workflow
func TestIFlytekParser_ClassifierWorkflow(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewIFlytekStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "iflytek", "iflytek_start_classifier_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse iFlytek DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_ClassifierWorkflow(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ iFlytek ClassifierWorkflow parser validation passed")
}

// TestIFlytekParser_CodeWorkflow validates iFlytek parser with code execution workflow
func TestIFlytekParser_CodeWorkflow(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewIFlytekStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "iflytek", "iflytek_start_code_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse iFlytek DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_CodeWorkflow(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ iFlytek CodeWorkflow parser validation passed")
}

// TestIFlytekParser_ConditionWorkflow validates iFlytek parser with conditional branch workflow
func TestIFlytekParser_ConditionWorkflow(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewIFlytekStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "iflytek", "iflytek_start_condition_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse iFlytek DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_ConditionWorkflow(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ iFlytek ConditionWorkflow parser validation passed")
}

// TestIFlytekParser_IterationWorkflow validates iFlytek parser with iteration loop workflow
func TestIFlytekParser_IterationWorkflow(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewIFlytekStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "iflytek", "iflytek_start_iteration_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse iFlytek DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_IterationWorkflow(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ iFlytek IterationWorkflow parser validation passed")
}

// TestIFlytekParser_LLMWorkflow validates iFlytek parser with large language model workflow
func TestIFlytekParser_LLMWorkflow(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewIFlytekStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "iflytek", "iflytek_start_llm_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse iFlytek DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_LLMWorkflow(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ iFlytek LLMWorkflow parser validation passed")
}