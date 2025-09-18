package services

import (
	"ai-agents-transformer/core/interfaces"
	"ai-agents-transformer/internal/models"
	"context"
	"fmt"
)

// ConversionService orchestrates DSL conversion between platforms.
type ConversionService struct {
	strategyRegistry StrategyRegistry
}

// NewConversionService creates a conversion service with the provided strategy registry.
func NewConversionService(registry StrategyRegistry) *ConversionService {
	return &ConversionService{
		strategyRegistry: registry,
	}
}

// Convert performs DSL conversion from source to target format.
func (s *ConversionService) Convert(
	sourceData []byte,
	sourcePlatform, targetPlatform models.PlatformType,
) ([]byte, error) {
	return s.ConvertWithContext(context.Background(), sourceData, sourcePlatform, targetPlatform)
}

// ConvertWithContext performs DSL conversion with request context support.
func (s *ConversionService) ConvertWithContext(
	ctx context.Context,
	sourceData []byte,
	sourcePlatform, targetPlatform models.PlatformType,
) ([]byte, error) {
	// Check platform support
	if err := s.validatePlatformSupport(sourcePlatform, targetPlatform); err != nil {
		return nil, fmt.Errorf("platform validation failed: %w", err)
	}

	// Get source platform parser
	parser, err := s.getParser(sourcePlatform)
	if err != nil {
		return nil, fmt.Errorf("failed to get parser for %s: %w", sourcePlatform, err)
	}

	// Parse source DSL to unified format
	unifiedDSL, err := parser.Parse(sourceData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source DSL: %w", err)
	}

	// Basic validation (simplified)
	if err := s.performBasicValidation(unifiedDSL); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get target platform generator
	generator, err := s.getGenerator(targetPlatform)
	if err != nil {
		return nil, fmt.Errorf("failed to get generator for %s: %w", targetPlatform, err)
	}

	// Generate target platform DSL
	targetData, err := generator.Generate(unifiedDSL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate target DSL: %w", err)
	}

	return targetData, nil
}

// validatePlatformSupport checks if the source and target platforms are supported.
func (s *ConversionService) validatePlatformSupport(sourcePlatform, targetPlatform models.PlatformType) error {
	supportedPlatforms := s.strategyRegistry.GetSupportedPlatforms()

	// Check source platform support
	sourceSupported := false
	targetSupported := false

	for _, platform := range supportedPlatforms {
		if platform == sourcePlatform {
			sourceSupported = true
		}
		if platform == targetPlatform {
			targetSupported = true
		}
	}

	if !sourceSupported {
		return fmt.Errorf("source platform %s is not supported", sourcePlatform)
	}

	if !targetSupported {
		return fmt.Errorf("target platform %s is not supported", targetPlatform)
	}

	return nil
}

// performBasicValidation performs simplified validation on the unified DSL.
func (s *ConversionService) performBasicValidation(unifiedDSL *models.UnifiedDSL) error {
	// Basic null checks
	if unifiedDSL == nil {
		return fmt.Errorf("unified DSL is nil")
	}

	// Check for minimum required nodes
	if len(unifiedDSL.Workflow.Nodes) == 0 {
		return fmt.Errorf("workflow must contain at least one node")
	}

	// Basic node validation
	for i, node := range unifiedDSL.Workflow.Nodes {
		if node.ID == "" {
			return fmt.Errorf("node at index %d has empty ID", i)
		}
		if node.Type == "" {
			return fmt.Errorf("node %s has empty type", node.ID)
		}
	}

	return nil
}

// ValidateSourceData validates input data against source platform requirements.
func (s *ConversionService) ValidateSourceData(
	data []byte,
	platform models.PlatformType,
) error {
	parser, err := s.getParser(platform)
	if err != nil {
		return fmt.Errorf("failed to get parser for validation: %w", err)
	}

	return parser.Validate(data)
}

// ValidateTargetCompatibility verifies unified DSL compatibility with target platform.
func (s *ConversionService) ValidateTargetCompatibility(
	unifiedDSL *models.UnifiedDSL,
	platform models.PlatformType,
) error {
	generator, err := s.getGenerator(platform)
	if err != nil {
		return fmt.Errorf("failed to get generator for validation: %w", err)
	}

	return generator.Validate(unifiedDSL)
}

func (s *ConversionService) getParser(platform models.PlatformType) (interfaces.DSLParser, error) {
	strategy, err := s.strategyRegistry.GetStrategy(platform)
	if err != nil {
		return nil, err
	}

	return strategy.CreateParser()
}

func (s *ConversionService) getGenerator(platform models.PlatformType) (interfaces.DSLGenerator, error) {
	strategy, err := s.strategyRegistry.GetStrategy(platform)
	if err != nil {
		return nil, err
	}

	return strategy.CreateGenerator()
}

// StrategyRegistry manages platform-specific conversion strategies.
type StrategyRegistry interface {
	// GetStrategy retrieves the strategy for the specified platform
	GetStrategy(platform models.PlatformType) (PlatformStrategy, error)

	// RegisterStrategy registers a strategy for a platform
	RegisterStrategy(platform models.PlatformType, strategy PlatformStrategy)

	// GetSupportedPlatforms returns list of supported platforms
	GetSupportedPlatforms() []models.PlatformType
}

// PlatformStrategy defines the interface for platform-specific conversion strategies.
// Simplified to include only the methods that are actually used.
type PlatformStrategy interface {
	// GetPlatformType returns the platform identifier
	GetPlatformType() models.PlatformType

	// CreateParser creates a DSL parser for this platform
	CreateParser() (interfaces.DSLParser, error)

	// CreateGenerator creates a DSL generator for this platform
	CreateGenerator() (interfaces.DSLGenerator, error)
}