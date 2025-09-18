package parser

import (
	"ai-agents-transformer/internal/models"
)

// StartNodeParser parses start nodes.
type StartNodeParser struct {
	*BaseNodeParser
}

func NewStartNodeParser(variableRefSystem *models.VariableReferenceSystem) *StartNodeParser {
	return &StartNodeParser{
		BaseNodeParser: NewBaseNodeParser("start", variableRefSystem),
	}
}

// ParseNode parses start node.
func (p *StartNodeParser) ParseNode(difyNode DifyNode) (*models.Node, error) {
	// Validate node
	if err := p.ValidateNode(difyNode); err != nil {
		return nil, err
	}

	// Create basic node structure
	node := p.createBasicStartNode(difyNode)

	// Parse configuration and outputs
	config := p.parseStartConfiguration(difyNode.Data.Variables)
	node.Config = config

	// Generate outputs from variables
	node.Outputs = p.generateOutputsFromVariables(config.Variables)

	return node, nil
}

// createBasicStartNode creates the basic start node structure
func (p *StartNodeParser) createBasicStartNode(difyNode DifyNode) *models.Node {
	node := p.parseBasicNodeInfo(difyNode)
	node.Type = models.NodeTypeStart
	return node
}

// parseStartConfiguration parses start node configuration from variables
func (p *StartNodeParser) parseStartConfiguration(difyVariables []DifyVariable) models.StartConfig {
	config := models.StartConfig{
		Variables: make([]models.Variable, 0, len(difyVariables)),
	}

	for _, difyVar := range difyVariables {
		startVar := p.convertDifyVariableToStartVariable(difyVar)
		config.Variables = append(config.Variables, startVar)
	}

	return config
}

// convertDifyVariableToStartVariable converts a Dify variable to start variable
func (p *StartNodeParser) convertDifyVariableToStartVariable(difyVar DifyVariable) models.Variable {
	startVar := models.Variable{
		Name:     difyVar.Variable,
		Label:    difyVar.Label,
		Type:     string(p.convertDataType(difyVar.Type)),
		Required: difyVar.Required,
	}

	// Add constraints if needed
	p.addVariableConstraints(&startVar, difyVar)

	return startVar
}

// addVariableConstraints adds constraints to a variable
func (p *StartNodeParser) addVariableConstraints(startVar *models.Variable, difyVar DifyVariable) {
	// Parse constraint conditions
	if difyVar.MaxLength > 0 {
		startVar.Constraints = &models.Constraints{
			MaxLength: difyVar.MaxLength,
		}
	}

	// Parse options
	if len(difyVar.Options) > 0 {
		p.addConstraintOptions(startVar, difyVar.Options)
	}
}

// addConstraintOptions adds options to variable constraints
func (p *StartNodeParser) addConstraintOptions(startVar *models.Variable, options []string) {
	if startVar.Constraints == nil {
		startVar.Constraints = &models.Constraints{}
	}

	// Convert string array to interface{} array
	interfaceOptions := make([]interface{}, len(options))
	for i, opt := range options {
		interfaceOptions[i] = opt
	}
	startVar.Constraints.Options = interfaceOptions
}

// generateOutputsFromVariables generates outputs from variable configuration
func (p *StartNodeParser) generateOutputsFromVariables(variables []models.Variable) []models.Output {
	outputs := make([]models.Output, 0, len(variables))

	for _, variable := range variables {
		output := models.Output{
			Name:        variable.Name,
			Label:       variable.Label,
			Type:        models.UnifiedDataType(variable.Type),
			Description: variable.Description,
		}
		outputs = append(outputs, output)
	}

	return outputs
}
