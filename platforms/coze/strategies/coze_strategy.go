package strategies

import (
	"ai-agents-transformer/core/interfaces"
	"ai-agents-transformer/core/services"
	"ai-agents-transformer/internal/models"
	"ai-agents-transformer/platforms/common"
	cozeGenerator "ai-agents-transformer/platforms/coze/generator"
	cozeParser "ai-agents-transformer/platforms/coze/parser"
	"os"
)

// CozeStrategy implements platform-specific strategy for ByteDance Coze workflow platform.
// Supports bidirectional conversion: iFlytek ↔ Coze
type CozeStrategy struct {
	validator *common.UnifiedDSLValidator
}

// NewCozeStrategy creates a Coze platform strategy with initialized components.
func NewCozeStrategy() services.PlatformStrategy {
	strategy := &CozeStrategy{
		validator: common.NewUnifiedDSLValidator(),
	}

	return strategy
}

// GetPlatformType returns the platform identifier for Coze.
func (s *CozeStrategy) GetPlatformType() models.PlatformType {
	return models.PlatformCoze
}

// CreateParser creates a Coze DSL parser instance.
// Supports Coze → iFlytek conversion (implemented by colleague)
func (s *CozeStrategy) CreateParser() (interfaces.DSLParser, error) {
	parser := cozeParser.NewCozeParser()

	// Check for verbose flag from environment or command line
	// This is a temporary solution to pass verbose flag to parser
	if os.Getenv("AI_AGENT_VERBOSE") == "true" {
		parser.SetVerbose(true)
	}

	return parser, nil
}

// CreateGenerator creates a Coze DSL generator instance.
// Supports iFlytek → Coze conversion
func (s *CozeStrategy) CreateGenerator() (interfaces.DSLGenerator, error) {
	return cozeGenerator.NewCozeGenerator(), nil
}

// GetValidator returns the DSL validator instance.
func (s *CozeStrategy) GetValidator() *common.UnifiedDSLValidator {
	return s.validator
}
