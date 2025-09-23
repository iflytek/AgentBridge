package generators

import (
	"testing"

	"agentbridge/platforms/iflytek/strategies"
	golden "agentbridge/tests/unit/golden/basic_start_end"
	codeGolden "agentbridge/tests/unit/golden/code_workflow"

	"github.com/stretchr/testify/require"
)

// TestIFlytekGenerator_BasicStartEnd_FromCoze tests iFlytek DSL generation from Coze parsed basic start-end workflow.
func TestIFlytekGenerator_BasicStartEnd_FromCoze(t *testing.T) {
	// Create generator instance
	strategy := strategies.NewIFlytekStrategy()
	generator, err := strategy.CreateGenerator()
	require.NoError(t, err, "generator creation failed")

	// Get golden data (active Go object from Coze parser)
	unifiedDSL := golden.GetCozeToUnified_Basic_start_end()
	require.NotNil(t, unifiedDSL, "unified DSL should not be nil")

	// Generate iFlytek DSL from unified DSL
	iflytekDSL, err := generator.Generate(unifiedDSL)
	require.NoError(t, err, "iFlytek DSL generation failed")
	require.NotEmpty(t, iflytekDSL, "generated iFlytek DSL should not be empty")

	// Run complete validation using helper functions
	allPassed := RunCompleteBasicStartEndValidation(t, iflytekDSL, "iflytek")
	require.True(t, allPassed, "Complete iFlytek DSL validation should pass")

	if allPassed {
		t.Logf("✅ Coze → iFlytek basic workflow conversion validation passed completely")
		t.Logf("Generated DSL length: %d bytes", len(iflytekDSL))
	}
}

// TestIFlytekGenerator_BasicStartEnd_FromDify tests iFlytek DSL generation from Dify parsed basic start-end workflow.
func TestIFlytekGenerator_BasicStartEnd_FromDify(t *testing.T) {
	// Create generator instance
	strategy := strategies.NewIFlytekStrategy()
	generator, err := strategy.CreateGenerator()
	require.NoError(t, err, "generator creation failed")

	// Get golden data (active Go object from Dify parser)
	unifiedDSL := golden.GetDifyToUnified_Basic_start_end()
	require.NotNil(t, unifiedDSL, "unified DSL should not be nil")

	// Generate iFlytek DSL from unified DSL
	iflytekDSL, err := generator.Generate(unifiedDSL)
	require.NoError(t, err, "iFlytek DSL generation failed")
	require.NotEmpty(t, iflytekDSL, "generated iFlytek DSL should not be empty")

	// Run complete validation using helper functions
	allPassed := RunCompleteBasicStartEndValidation(t, iflytekDSL, "iflytek")
	require.True(t, allPassed, "Complete iFlytek DSL validation should pass")

	if allPassed {
		t.Logf("✅ Dify → iFlytek basic workflow conversion validation passed completely")
		t.Logf("Generated DSL length: %d bytes", len(iflytekDSL))
	}
}

// TestIFlytekGenerator_CodeWorkflow_FromCoze tests iFlytek DSL generation from Coze parsed code workflow.
func TestIFlytekGenerator_CodeWorkflow_FromCoze(t *testing.T) {
	// Create generator instance
	strategy := strategies.NewIFlytekStrategy()
	generator, err := strategy.CreateGenerator()
	require.NoError(t, err, "generator creation failed")

	// Get golden data (active Go object from Coze parser)
	unifiedDSL := codeGolden.GetCozeToUnified_Code_workflow()
	require.NotNil(t, unifiedDSL, "unified DSL should not be nil")

	// Generate iFlytek DSL from unified DSL
	iflytekDSL, err := generator.Generate(unifiedDSL)
	require.NoError(t, err, "iFlytek DSL generation failed")
	require.NotEmpty(t, iflytekDSL, "generated iFlytek DSL should not be empty")

	// Validate generated iFlytek DSL using complete validation logic
	allPassed := RunCompleteCodeWorkflowValidation(t, iflytekDSL, "iflytek")
	require.True(t, allPassed, "Complete iFlytek code workflow validation should pass")

	if allPassed {
		t.Logf("✅ iFlytek generator code workflow conversion validation passed completely")
		t.Logf("Generated DSL length: %d bytes", len(iflytekDSL))
	}
}

// TestIFlytekGenerator_CodeWorkflow_FromDify tests iFlytek DSL generation from Dify parsed code workflow.
func TestIFlytekGenerator_CodeWorkflow_FromDify(t *testing.T) {
	// Create generator instance
	strategy := strategies.NewIFlytekStrategy()
	generator, err := strategy.CreateGenerator()
	require.NoError(t, err, "generator creation failed")

	// Get golden data (active Go object from Dify parser)
	unifiedDSL := codeGolden.GetDifyToUnified_Code_workflow()
	require.NotNil(t, unifiedDSL, "unified DSL should not be nil")

	// Generate iFlytek DSL from unified DSL
	iflytekDSL, err := generator.Generate(unifiedDSL)
	require.NoError(t, err, "iFlytek DSL generation failed")
	require.NotEmpty(t, iflytekDSL, "generated iFlytek DSL should not be empty")

	// Validate generated iFlytek DSL using complete validation logic
	allPassed := RunCompleteCodeWorkflowValidation(t, iflytekDSL, "iflytek")
	require.True(t, allPassed, "Complete iFlytek code workflow validation should pass")

	if allPassed {
		t.Logf("✅ iFlytek generator code workflow conversion validation passed completely")
		t.Logf("Generated DSL length: %d bytes", len(iflytekDSL))
	}
}
