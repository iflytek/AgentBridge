package core

import (
	"agentbridge/core/services"
	"agentbridge/internal/models"
	cozeStrategies "agentbridge/platforms/coze/strategies"
	"agentbridge/platforms/dify/strategies"
	iflytekStrategies "agentbridge/platforms/iflytek/strategies"
	"agentbridge/registry"
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
