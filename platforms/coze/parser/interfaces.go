package parser

import (
	"github.com/iflytek/agentbridge/internal/models"
)

// NodeParser defines the interface for Coze node parsing.
type NodeParser interface {
	// GetSupportedType returns the supported node type.
	GetSupportedType() string

	// ParseNode parses a node.
	ParseNode(cozeNode CozeNode) (*models.Node, error)

	// ValidateNode validates node data.
	ValidateNode(cozeNode CozeNode) error
}

// ConfigParser defines the interface for configuration parsing.
type ConfigParser interface {
	// ParseConfig parses node-specific configuration.
	ParseConfig(nodeData map[string]interface{}) (models.NodeConfig, error)
}

// NodeParserFactory defines the interface for node parser factory.
type NodeParserFactory interface {
	// CreateParser creates a parser.
	CreateParser(nodeType string, variableRefSystem *models.VariableReferenceSystem) (NodeParser, error)

	// GetSupportedTypes returns supported node types.
	GetSupportedTypes() []string
}
