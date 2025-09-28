package core

import (
	"github.com/iflytek/agentbridge/core/services"
	"github.com/iflytek/agentbridge/internal/models"
	cozeStrategies "github.com/iflytek/agentbridge/platforms/coze/strategies"
	"github.com/iflytek/agentbridge/platforms/dify/strategies"
	iflytekStrategies "github.com/iflytek/agentbridge/platforms/iflytek/strategies"
	"github.com/iflytek/agentbridge/registry"
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
