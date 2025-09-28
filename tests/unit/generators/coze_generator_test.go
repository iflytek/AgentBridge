package generators

import (
	"testing"

	"github.com/iflytek/agentbridge/platforms/coze/strategies"
	golden "github.com/iflytek/agentbridge/tests/unit/golden/basic_start_end"
	codeGolden "github.com/iflytek/agentbridge/tests/unit/golden/code_workflow"

	"github.com/stretchr/testify/require"
)

// TestCozeGenerator_BasicStartEnd_FromIFlytek tests Coze DSL generation from iFlytek parsed basic start-end workflow.
func TestCozeGenerator_BasicStartEnd_FromIFlytek(t *testing.T) {
	// Create generator instance
	strategy := strategies.NewCozeStrategy()
	generator, err := strategy.CreateGenerator()
	require.NoError(t, err, "generator creation failed")

	// Get golden data (active Go object from iFlytek parser)
	unifiedDSL := golden.GetIFlytekToUnified_BasicStartEnd()
	require.NotNil(t, unifiedDSL, "unified DSL should not be nil")

	// Generate Coze DSL from unified DSL
	cozeDSL, err := generator.Generate(unifiedDSL)
	require.NoError(t, err, "Coze DSL generation failed")
	require.NotEmpty(t, cozeDSL, "generated Coze DSL should not be empty")

	// Validate generated Coze DSL using complete validation logic
	allPassed := RunCompleteBasicStartEndValidation(t, cozeDSL, "coze")
	require.True(t, allPassed, "Complete Coze DSL validation should pass")

	if allPassed {
		t.Logf("✅ Coze → iFlytek basic workflow conversion validation passed completely")
		t.Logf("Generated DSL length: %d bytes", len(cozeDSL))
	}
}

// TestCozeGenerator_CodeWorkflow_FromIFlytek tests Coze DSL generation from iFlytek parsed code workflow.
func TestCozeGenerator_CodeWorkflow_FromIFlytek(t *testing.T) {
	// Create generator instance
	strategy := strategies.NewCozeStrategy()
	generator, err := strategy.CreateGenerator()
	require.NoError(t, err, "generator creation failed")

	// Get golden data (active Go object from iFlytek parser)
	unifiedDSL := codeGolden.GetIflytekToUnified_Code_workflow()
	require.NotNil(t, unifiedDSL, "unified DSL should not be nil")

	// Generate Coze DSL from unified DSL
	cozeDSL, err := generator.Generate(unifiedDSL)
	require.NoError(t, err, "Coze DSL generation failed")
	require.NotEmpty(t, cozeDSL, "generated Coze DSL should not be empty")

	// Validate generated Coze DSL using complete validation logic
	allPassed := RunCompleteCodeWorkflowValidation(t, cozeDSL, "coze")
	require.True(t, allPassed, "Complete Coze code workflow validation should pass")

	if allPassed {
		t.Logf("✅ Coze generator code workflow conversion validation passed completely")
		t.Logf("Generated DSL length: %d bytes", len(cozeDSL))
	}
}
