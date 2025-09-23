package common

import (
	"agentbridge/internal/models"
)

// BaseGenerator provides base implementation for generators
type BaseGenerator struct {
	platformType models.PlatformType
}

func NewBaseGenerator(platformType models.PlatformType) *BaseGenerator {
	return &BaseGenerator{
		platformType: platformType,
	}
}

// GetPlatformType returns the platform type
func (g *BaseGenerator) GetPlatformType() models.PlatformType {
	return g.platformType
}

// BaseParser provides base implementation for parsers
type BaseParser struct {
	platformType models.PlatformType
}

func NewBaseParser(platformType models.PlatformType) *BaseParser {
	return &BaseParser{
		platformType: platformType,
	}
}

// GetPlatformType returns the platform type
func (p *BaseParser) GetPlatformType() models.PlatformType {
	return p.platformType
}
