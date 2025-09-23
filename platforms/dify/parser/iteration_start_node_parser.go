package parser

import (
	"agentbridge/internal/models"
)

// IterationStartNodeParser parses iteration start nodes.
type IterationStartNodeParser struct {
	*BaseNodeParser
}

func NewIterationStartNodeParser(variableRefSystem *models.VariableReferenceSystem) NodeParser {
	return &IterationStartNodeParser{
		BaseNodeParser: NewBaseNodeParser("iteration-start", variableRefSystem),
	}
}

// ParseNode parses iteration start node.
func (p *IterationStartNodeParser) ParseNode(difyNode DifyNode) (*models.Node, error) {
	// Extract basic information
	id := difyNode.ID
	data := difyNode.Data

	title := data.Title
	if title == "" {
		title = "Iteration Start"
	}

	description := data.Desc

	// Check if inside iteration
	isInIteration := data.IsInIteration

	// Get parent iteration node ID
	parentIterationID := difyNode.ParentID

	// Special configuration for iteration start node
	config := models.StartConfig{
		Variables:     []models.Variable{}, // Iteration start nodes usually have no external variables
		IsInIteration: isInIteration,
		ParentID:      parentIterationID,
	}

	// Output: iteration item
	outputs := []models.Output{
		{
			Name: "item",
			Type: models.DataTypeString, // Iteration item type, usually string or object
		},
	}

	// Create unified node
	node := &models.Node{
		ID:          id,
		Type:        models.NodeTypeStart, // Iteration start node is essentially a start node
		Title:       title,
		Description: description,
		Position:    models.Position{X: difyNode.Position.X, Y: difyNode.Position.Y},
		Size:        models.Size{Width: difyNode.Width, Height: difyNode.Height},
		Config:      config,
		Inputs:      []models.Input{}, // Start nodes usually have no inputs
		Outputs:     outputs,
	}

	// Save iteration-related platform-specific information
	if node.PlatformConfig.Dify == nil {
		node.PlatformConfig.Dify = make(map[string]interface{})
	}
	node.PlatformConfig.Dify["isInIteration"] = isInIteration
	node.PlatformConfig.Dify["parentIterationID"] = parentIterationID

	return node, nil
}
