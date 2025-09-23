package interfaces

import (
	"agentbridge/internal/models"
)

// DSLGenerator defines the unified DSL generator interface
type DSLGenerator interface {
	// Generate creates target platform DSL from unified DSL
	Generate(unifiedDSL *models.UnifiedDSL) ([]byte, error)

	// Validate checks if unified DSL meets target platform requirements
	Validate(unifiedDSL *models.UnifiedDSL) error

	// GetPlatformType returns the platform type supported by the generator
	GetPlatformType() models.PlatformType
}

// DSLParser defines the unified DSL parser interface
type DSLParser interface {
	// Parse converts DSL file to unified format
	Parse(data []byte) (*models.UnifiedDSL, error)

	// Validate checks if input data format is correct
	Validate(data []byte) error

	// GetPlatformType returns the platform type supported by the parser
	GetPlatformType() models.PlatformType
}

// IDMapper defines the ID mapper interface
type IDMapper interface {
	// MapNodeID maps node ID
	MapNodeID(originalID string, nodeType models.NodeType) string

	// GetMapping returns the complete mapping table
	GetMapping() map[string]string

	// SetMapping sets the mapping table
	SetMapping(mapping map[string]string)
}
