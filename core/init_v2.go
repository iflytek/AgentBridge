package core

import (
	"ai-agents-transformer/core/services"
	"ai-agents-transformer/internal/models"
	cozeStrategies "ai-agents-transformer/platforms/coze/strategies"
	"ai-agents-transformer/platforms/dify/strategies"
	iflytekStrategies "ai-agents-transformer/platforms/iflytek/strategies"
	"ai-agents-transformer/registry"
)

// InitializeArchitecture initializes the conversion architecture
func InitializeArchitecture() (*services.ConversionService, error) {
	// Create strategy registry
	strategyRegistry := registry.NewStrategyRegistry()

	// Register platform strategies
	cozeStrategy := cozeStrategies.NewCozeStrategy()
	difyStrategy := strategies.NewDifyStrategy()
	iflytekStrategy := iflytekStrategies.NewIFlytekStrategy()

	strategyRegistry.RegisterStrategy(models.PlatformCoze, cozeStrategy)
	strategyRegistry.RegisterStrategy(models.PlatformDify, difyStrategy)
	strategyRegistry.RegisterStrategy(models.PlatformIFlytek, iflytekStrategy)

	// Create conversion service
	conversionService := services.NewConversionService(strategyRegistry)

	return conversionService, nil
}
