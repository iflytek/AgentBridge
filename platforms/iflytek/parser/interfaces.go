package parser

import (
	"ai-agents-transformer/internal/models"
)

// NodeParser defines the interface for node parsing.
type NodeParser interface {
	// GetSupportedType returns the supported node type.
	GetSupportedType() string

	// ParseNode parses a node.
	ParseNode(iflytekNode IFlytekNode) (*models.Node, error)

	// ValidateNode validates node data.
	ValidateNode(iflytekNode IFlytekNode) error
}

// NodeParserFactory defines the interface for node parser factory.
type NodeParserFactory interface {
	// CreateParser creates a parser.
	CreateParser(nodeType string, variableRefSystem *models.VariableReferenceSystem) (NodeParser, error)

	// GetSupportedTypes returns supported node types.
	GetSupportedTypes() []string
}
