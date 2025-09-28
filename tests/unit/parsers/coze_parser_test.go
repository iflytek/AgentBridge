package parsers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/iflytek/agentbridge/platforms/coze/strategies"
	"github.com/stretchr/testify/require"
)

// TestCozeParser_BasicStartEnd validates Coze parser with basic start-end workflow
func TestCozeParser_BasicStartEnd(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewCozeStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "coze", "coze_basic_start_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse Coze DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_BasicStartEnd(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ Coze BasicStartEnd parser validation passed")
}

// TestCozeParser_ClassifierWorkflow validates Coze parser with classifier workflow
func TestCozeParser_ClassifierWorkflow(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewCozeStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "coze", "coze_start_classifier_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse Coze DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_ClassifierWorkflow(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ Coze ClassifierWorkflow parser validation passed")
}

// TestCozeParser_CodeWorkflow validates Coze parser with code execution workflow
func TestCozeParser_CodeWorkflow(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewCozeStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "coze", "coze_start_code_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse Coze DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_CodeWorkflow(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ Coze CodeWorkflow parser validation passed")
}

// TestCozeParser_ConditionWorkflow validates Coze parser with conditional branch workflow
func TestCozeParser_ConditionWorkflow(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewCozeStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "coze", "coze_start_condition_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse Coze DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_ConditionWorkflow(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ Coze ConditionWorkflow parser validation passed")
}

// TestCozeParser_IterationWorkflow validates Coze parser with iteration loop workflow
func TestCozeParser_IterationWorkflow(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewCozeStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "coze", "coze_start_iteration_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse Coze DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_IterationWorkflow(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ Coze IterationWorkflow parser validation passed")
}

// TestCozeParser_LLMWorkflow validates Coze parser with large language model workflow
func TestCozeParser_LLMWorkflow(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewCozeStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "coze", "coze_start_llm_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse Coze DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_LLMWorkflow(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ Coze LLMWorkflow parser validation passed")
}
