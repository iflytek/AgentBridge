package generator

import (
	"ai-agents-transformer/internal/models"
)

// NodeGenerator interface for iFlytek SparkAgent node generation
type NodeGenerator interface {
	// GetSupportedType returns the supported node type
	GetSupportedType() models.NodeType

	// GenerateNode generates a node
	GenerateNode(node models.Node) (IFlytekNode, error)

	// ValidateNode validates node data
	ValidateNode(node models.Node) error
}


// NodeGeneratorFactoryInterface interface for node generator factory
type NodeGeneratorFactoryInterface interface {
	// GetGenerator returns generator for specified node type
	GetGenerator(nodeType models.NodeType) (NodeGenerator, error)

	// GetSupportedTypes returns all supported node types
	GetSupportedTypes() []models.NodeType
}
