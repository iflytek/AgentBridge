package generator

import (
	"github.com/iflytek/agentbridge/internal/models"
)

// NodeGenerator defines the node generator interface
type NodeGenerator interface {
	// GenerateNode generates a specific type of Dify node
	GenerateNode(node models.Node) (DifyNode, error)

	// GetSupportedType returns the supported node type
	GetSupportedType() models.NodeType
}

// BaseNodeGenerator provides base node generator functionality
type BaseNodeGenerator struct {
	nodeType models.NodeType
}

func NewBaseNodeGenerator(nodeType models.NodeType) *BaseNodeGenerator {
	return &BaseNodeGenerator{
		nodeType: nodeType,
	}
}

func (g *BaseNodeGenerator) GetSupportedType() models.NodeType {
	return g.nodeType
}

// generateBaseNode generates the base structure of a node
func (g *BaseNodeGenerator) generateBaseNode(node models.Node) DifyNode {
	difyNode := DifyNode{
		ID:   node.ID,
		Type: "custom",
		Data: DifyNodeData{
			Type:     mapNodeTypeToDify(node.Type),
			Title:    node.Title,
			Desc:     node.Description,
			Selected: false,
			Config:   make(map[string]interface{}), // Initialize empty configuration
		},
	}

	// Set position information
	if node.Position.X != 0 || node.Position.Y != 0 {
		difyNode.Position = DifyPosition{
			X: node.Position.X,
			Y: node.Position.Y,
		}
		difyNode.PositionAbsolute = DifyPosition{
			X: node.Position.X,
			Y: node.Position.Y,
		}
	}

	// Set size information
	if node.Size.Width != 0 || node.Size.Height != 0 {
		difyNode.Width = node.Size.Width
		difyNode.Height = node.Size.Height
	} else {
		// Default size
		difyNode.Width = 244
		difyNode.Height = 196
	}

	// Set connection point positions
	difyNode.SourcePosition = "right"
	difyNode.TargetPosition = "left"

	return difyNode
}

// mapNodeTypeToDify maps unified node types to Dify node types
func mapNodeTypeToDify(nodeType models.NodeType) string {
	switch nodeType {
	case models.NodeTypeStart:
		return "start"
	case models.NodeTypeEnd:
		return "end"
	case models.NodeTypeLLM:
		return "llm"
	case models.NodeTypeCode:
		return "code"
	case models.NodeTypeCondition:
		return "if-else"
	case models.NodeTypeClassifier:
		return "question-classifier"
	case models.NodeTypeIteration:
		return "iteration"
	default:
		return string(nodeType) // Fallback to original type
	}
}
