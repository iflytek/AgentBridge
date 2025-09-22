package generator

import (
	"ai-agents-transformer/internal/models"
)

// CozeNodeGenerator defines the interface for Coze node generators
type CozeNodeGenerator interface {
	// GenerateNode generates a Coze workflow node from unified node
	GenerateNode(unifiedNode *models.Node) (*CozeNode, error)

	// GenerateSchemaNode generates a Coze schema node from unified node
	GenerateSchemaNode(unifiedNode *models.Node) (*CozeSchemaNode, error)

	// GetNodeType returns the node type this generator handles
	GetNodeType() models.NodeType

	// ValidateNode validates the unified node before generation
	ValidateNode(unifiedNode *models.Node) error

	// SetIDGenerator sets the shared ID generator
	SetIDGenerator(idGenerator *CozeIDGenerator)
}
