package strategies

import (
	"github.com/iflytek/agentbridge/core/interfaces"
	"github.com/iflytek/agentbridge/core/services"
	"github.com/iflytek/agentbridge/internal/models"
	"github.com/iflytek/agentbridge/platforms/common"
	difyGenerator "github.com/iflytek/agentbridge/platforms/dify/generator"
	difyParser "github.com/iflytek/agentbridge/platforms/dify/parser"
)

// DifyStrategy implements platform-specific strategy for Dify workflow platform.
type DifyStrategy struct {
	validator *common.UnifiedDSLValidator
}

// NewDifyStrategy creates a Dify platform strategy with initialized components.
func NewDifyStrategy() services.PlatformStrategy {
	strategy := &DifyStrategy{
		validator: common.NewUnifiedDSLValidator(),
	}

	return strategy
}

// GetPlatformType returns the platform identifier for Dify.
func (s *DifyStrategy) GetPlatformType() models.PlatformType {
	return models.PlatformDify
}

// CreateParser creates a Dify DSL parser instance.
func (s *DifyStrategy) CreateParser() (interfaces.DSLParser, error) {
	return difyParser.NewDifyParser(), nil
}

// CreateGenerator creates a Dify DSL generator instance.
func (s *DifyStrategy) CreateGenerator() (interfaces.DSLGenerator, error) {
	return difyGenerator.NewDifyGenerator(), nil
}

// GetValidator returns the DSL validator instance.
func (s *DifyStrategy) GetValidator() *common.UnifiedDSLValidator {
	return s.validator
}
