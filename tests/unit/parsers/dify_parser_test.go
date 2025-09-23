package parsers

import (
	"os"
	"path/filepath"
	"testing"

	"agentbridge/platforms/dify/strategies"
	"github.com/stretchr/testify/require"
)

// TestDifyParser_BasicStartEnd validates Dify parser with basic start-end workflow
func TestDifyParser_BasicStartEnd(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewDifyStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "dify", "dify_basic_start_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse Dify DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_BasicStartEnd(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ Dify BasicStartEnd parser validation passed")
}

// TestDifyParser_ClassifierWorkflow validates Dify parser with classifier workflow
func TestDifyParser_ClassifierWorkflow(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewDifyStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "dify", "dify_start_classifier_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse Dify DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_ClassifierWorkflow(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ Dify ClassifierWorkflow parser validation passed")
}

// TestDifyParser_CodeWorkflow validates Dify parser with code execution workflow
func TestDifyParser_CodeWorkflow(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewDifyStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "dify", "dify_start_code_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse Dify DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_CodeWorkflow(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ Dify CodeWorkflow parser validation passed")
}

// TestDifyParser_ConditionWorkflow validates Dify parser with conditional branch workflow
func TestDifyParser_ConditionWorkflow(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewDifyStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "dify", "dify_start_condition_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse Dify DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_ConditionWorkflow(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ Dify ConditionWorkflow parser validation passed")
}

// TestDifyParser_IterationWorkflow validates Dify parser with iteration loop workflow
func TestDifyParser_IterationWorkflow(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewDifyStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "dify", "dify_start_iteration_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse Dify DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_IterationWorkflow(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ Dify IterationWorkflow parser validation passed")
}

// TestDifyParser_LLMWorkflow validates Dify parser with large language model workflow
func TestDifyParser_LLMWorkflow(t *testing.T) {
	// Create parser instance
	strategy := strategies.NewDifyStrategy()
	parser, err := strategy.CreateParser()
	require.NoError(t, err, "parser creation failed")

	// Load test fixture data
	inputFile := filepath.Join("..", "..", "fixtures", "dify", "dify_start_llm_end.yml")
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "file read failed")

	// Parse Dify DSL to unified format
	unifiedDSL, err := parser.Parse(inputData)
	require.NoError(t, err, "DSL parsing failed")
	require.NotNil(t, unifiedDSL, "parsed DSL is nil")

	// Validate parsed result structure
	err = ValidateParserResult_LLMWorkflow(unifiedDSL)
	require.NoError(t, err, "result validation failed")

	t.Logf("✅ Dify LLMWorkflow parser validation passed")
}
