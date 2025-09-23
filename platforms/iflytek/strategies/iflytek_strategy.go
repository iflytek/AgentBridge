package strategies

import (
	"agentbridge/core/interfaces"
	"agentbridge/core/services"
	"agentbridge/internal/models"
	"agentbridge/platforms/common"
	iflytekGenerator "agentbridge/platforms/iflytek/generator"
	iflytekParser "agentbridge/platforms/iflytek/parser"
)

// IFlytekStrategy implements platform-specific strategy for iFlytek SparkAgent platform.
type IFlytekStrategy struct {
	validator *common.UnifiedDSLValidator
}

// NewIFlytekStrategy creates an iFlytek platform strategy with initialized components.
func NewIFlytekStrategy() services.PlatformStrategy {
	strategy := &IFlytekStrategy{
		validator: common.NewUnifiedDSLValidator(),
	}

	return strategy
}

// GetPlatformType returns the platform identifier for iFlytek.
func (s *IFlytekStrategy) GetPlatformType() models.PlatformType {
	return models.PlatformIFlytek
}

// CreateParser creates an iFlytek DSL parser instance.
func (s *IFlytekStrategy) CreateParser() (interfaces.DSLParser, error) {
	return iflytekParser.NewIFlytekParser(), nil
}

// CreateGenerator creates an iFlytek DSL generator instance.
func (s *IFlytekStrategy) CreateGenerator() (interfaces.DSLGenerator, error) {
	return iflytekGenerator.NewIFlytekGenerator(), nil
}

// GetValidator returns the DSL validator instance.
func (s *IFlytekStrategy) GetValidator() *common.UnifiedDSLValidator {
	return s.validator
}
