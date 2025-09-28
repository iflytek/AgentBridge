package parser

import (
	"github.com/iflytek/agentbridge/internal/models"
	"fmt"
)

// StartNodeParser parses start nodes.
type StartNodeParser struct {
	*BaseNodeParser
}

func NewStartNodeParser(variableRefSystem *models.VariableReferenceSystem) *StartNodeParser {
	return &StartNodeParser{
		BaseNodeParser: NewBaseNodeParser("1", variableRefSystem),
	}
}

// ParseNode parses start node.
func (p *StartNodeParser) ParseNode(cozeNode CozeNode) (*models.Node, error) {
	// Validate node
	if err := p.ValidateNode(cozeNode); err != nil {
		return nil, err
	}

	// Create basic node structure
	node := p.createBasicStartNode(cozeNode)

	// Parse configuration and outputs
	config := p.parseStartConfiguration(cozeNode)
	node.Config = config

	// Generate outputs from variables
	node.Outputs = p.generateOutputsFromCozeOutputs(cozeNode.Data.Outputs)

	return node, nil
}

// createBasicStartNode creates the basic start node structure
func (p *StartNodeParser) createBasicStartNode(cozeNode CozeNode) *models.Node {
	node := p.parseBasicNodeInfo(cozeNode)
	node.Type = models.NodeTypeStart
	return node
}

// parseStartConfiguration parses start node configuration from Coze outputs
func (p *StartNodeParser) parseStartConfiguration(cozeNode CozeNode) models.StartConfig {
	config := models.StartConfig{
		Variables: make([]models.Variable, 0, len(cozeNode.Data.Outputs)),
	}

	for _, cozeOutput := range cozeNode.Data.Outputs {
		startVar := p.convertCozeOutputToStartVariable(cozeOutput)
		config.Variables = append(config.Variables, startVar)
	}

	return config
}

// convertCozeOutputToStartVariable converts a Coze output to start variable with validation
func (p *StartNodeParser) convertCozeOutputToStartVariable(cozeOutput CozeOutput) models.Variable {
	startVar := models.Variable{
		Name:     cozeOutput.Name,
		Label:    cozeOutput.Name,
		Type:     string(p.convertDataType(cozeOutput.Type)),
		Required: cozeOutput.Required,
	}

	// Set reasonable default values for non-required fields
	if !cozeOutput.Required {
		// Set reasonable default values for non-required fields
		switch p.convertDataType(cozeOutput.Type) {
		case models.DataTypeString:
			startVar.Default = ""
		case models.DataTypeInteger:
			startVar.Default = 0
		case models.DataTypeFloat:
			startVar.Default = 0.0
		case models.DataTypeBoolean:
			startVar.Default = false
		}
	}

	return startVar
}

// generateOutputsFromCozeOutputs generates outputs from Coze output configuration with mapping
func (p *StartNodeParser) generateOutputsFromCozeOutputs(cozeOutputs []CozeOutput) []models.Output {
	outputs := make([]models.Output, 0, len(cozeOutputs))

	for _, cozeOutput := range cozeOutputs {
		output := models.Output{
			Name:        cozeOutput.Name,
			Label:       cozeOutput.Name,
			Type:        p.convertDataType(cozeOutput.Type),
			Required:    cozeOutput.Required,
			Description: p.generateOutputDescription(cozeOutput),
		}
		outputs = append(outputs, output)
	}

	return outputs
}

// generateOutputDescription generates description for output based on type and name
func (p *StartNodeParser) generateOutputDescription(cozeOutput CozeOutput) string {
	if cozeOutput.Name == "AGENT_USER_INPUT" {
		return "User input content"
	}

	// Generate appropriate description based on type
	switch cozeOutput.Type {
	case "string":
		return fmt.Sprintf("String type variable: %s", cozeOutput.Name)
	case "integer":
		return fmt.Sprintf("Integer type variable: %s", cozeOutput.Name)
	case "float":
		return fmt.Sprintf("Float type variable: %s", cozeOutput.Name)
	case "boolean":
		return fmt.Sprintf("Boolean type variable: %s", cozeOutput.Name)
	default:
		return fmt.Sprintf("Workflow variable: %s", cozeOutput.Name)
	}
}
