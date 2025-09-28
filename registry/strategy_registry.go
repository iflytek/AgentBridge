package registry

import (
	"github.com/iflytek/agentbridge/core/services"
	"github.com/iflytek/agentbridge/internal/models"
	"fmt"
	"sync"
)

// StrategyRegistry manages platform-specific conversion strategies
type StrategyRegistry struct {
	mutex      sync.RWMutex
	strategies map[models.PlatformType]services.PlatformStrategy
}

func NewStrategyRegistry() services.StrategyRegistry {
	return &StrategyRegistry{
		strategies: make(map[models.PlatformType]services.PlatformStrategy),
	}
}

// RegisterStrategy registers a strategy for a platform
func (r *StrategyRegistry) RegisterStrategy(platform models.PlatformType, strategy services.PlatformStrategy) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.strategies[platform] = strategy
}

// GetStrategy retrieves the strategy for the specified platform
func (r *StrategyRegistry) GetStrategy(platform models.PlatformType) (services.PlatformStrategy, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	strategy, exists := r.strategies[platform]
	if !exists {
		return nil, fmt.Errorf("strategy for platform %s not found", platform)
	}

	return strategy, nil
}

// GetSupportedPlatforms returns list of supported platforms
func (r *StrategyRegistry) GetSupportedPlatforms() []models.PlatformType {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	platforms := make([]models.PlatformType, 0, len(r.strategies))
	for platform := range r.strategies {
		platforms = append(platforms, platform)
	}

	return platforms
}

// HasStrategy checks if a platform is supported
func (r *StrategyRegistry) HasStrategy(platform models.PlatformType) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.strategies[platform]
	return exists
}

// UnregisterStrategy removes a registered platform strategy
func (r *StrategyRegistry) UnregisterStrategy(platform models.PlatformType) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.strategies, platform)
}

// Clear removes all registered strategies
func (r *StrategyRegistry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.strategies = make(map[models.PlatformType]services.PlatformStrategy)
}
