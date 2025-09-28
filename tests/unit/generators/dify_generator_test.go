package generators

import (
	"testing"

	"github.com/iflytek/agentbridge/platforms/dify/strategies"
	golden "github.com/iflytek/agentbridge/tests/unit/golden/basic_start_end"
	codeGolden "github.com/iflytek/agentbridge/tests/unit/golden/code_workflow"
	"github.com/stretchr/testify/require"
)

// TestDifyGenerator_BasicStartEnd_FromIFlytek tests Dify DSL generation from iFlytek parsed basic start-end workflow.
func TestDifyGenerator_BasicStartEnd_FromIFlytek(t *testing.T) {
	// Create generator instance
	strategy := strategies.NewDifyStrategy()
	generator, err := strategy.CreateGenerator()
	require.NoError(t, err, "generator creation failed")

	// Get golden data (active Go object from iFlytek parser)
	unifiedDSL := golden.GetIFlytekToUnified_BasicStartEnd()
	require.NotNil(t, unifiedDSL, "unified DSL should not be nil")

	// Generate Dify DSL from unified DSL
	difyDSL, err := generator.Generate(unifiedDSL)
	require.NoError(t, err, "Dify DSL generation failed")
	require.NotEmpty(t, difyDSL, "generated Dify DSL should not be empty")

	// Run complete validation using helper functions
	allPassed := RunCompleteBasicStartEndValidation(t, difyDSL, "dify")
	require.True(t, allPassed, "Complete Dify DSL validation should pass")

	if allPassed {
		t.Logf("✅ iFlytek → Dify basic workflow conversion validation passed completely")
		t.Logf("Generated DSL length: %d bytes", len(difyDSL))
	}
}

// TestDifyGenerator_CodeWorkflow_FromIFlytek tests Dify DSL generation from iFlytek parsed code workflow.
func TestDifyGenerator_CodeWorkflow_FromIFlytek(t *testing.T) {
	// Create generator instance
	strategy := strategies.NewDifyStrategy()
	generator, err := strategy.CreateGenerator()
	require.NoError(t, err, "generator creation failed")

	// Get golden data (active Go object from iFlytek parser)
	unifiedDSL := codeGolden.GetIflytekToUnified_Code_workflow()
	require.NotNil(t, unifiedDSL, "unified DSL should not be nil")

	// Generate Dify DSL from unified DSL
	difyDSL, err := generator.Generate(unifiedDSL)
	require.NoError(t, err, "Dify DSL generation failed")
	require.NotEmpty(t, difyDSL, "generated Dify DSL should not be empty")

	// Validate generated Dify DSL using complete validation logic
	allPassed := RunCompleteCodeWorkflowValidation(t, difyDSL, "dify")
	require.True(t, allPassed, "Complete Dify code workflow validation should pass")

	if allPassed {
		t.Logf("✅ Dify generator code workflow conversion validation passed completely")
		t.Logf("Generated DSL length: %d bytes", len(difyDSL))
	}
}
