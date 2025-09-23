package parser

import (
	"agentbridge/internal/models"
	"fmt"
)

// StartNodeParser parses start nodes.
type StartNodeParser struct {
	*BaseNodeParser
}

func NewStartNodeParser(variableRefSystem *models.VariableReferenceSystem) *StartNodeParser {
	return &StartNodeParser{
		BaseNodeParser: NewBaseNodeParser(variableRefSystem),
	}
}

// GetSupportedType returns the supported node type.
func (p *StartNodeParser) GetSupportedType() string {
	return "开始节点"
}

// ValidateNode validates node data.
func (p *StartNodeParser) ValidateNode(iflytekNode IFlytekNode) error {
	if iflytekNode.ID == "" {
		return fmt.Errorf("node ID is empty")
	}

	if iflytekNode.Type != p.GetSupportedType() {
		return fmt.Errorf("invalid node type: expected %s, got %s", p.GetSupportedType(), iflytekNode.Type)
	}

	return nil
}

// ParseNode parses a node.
func (p *StartNodeParser) ParseNode(iflytekNode IFlytekNode) (*models.Node, error) {
	if err := p.ValidateNode(iflytekNode); err != nil {
		return nil, err
	}

	// Parse basic information
	node := p.ParseBasicNodeInfo(iflytekNode, models.NodeTypeStart)

	// Parse outputs (start node outputs are workflow input variables)
	var config models.StartConfig
	if outputs, ok := iflytekNode.Data["outputs"].([]interface{}); ok {
		nodeOutputs, err := p.ParseNodeOutputs(outputs)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start node outputs: %w", err)
		}
		node.Outputs = nodeOutputs

		// Generate variable configuration from outputs to avoid duplicate parsing
		variables, err := p.parseVariablesFromOutputs(outputs)
		if err != nil {
			return nil, fmt.Errorf("failed to parse variables from outputs: %w", err)
		}
		config = models.StartConfig{Variables: variables}
	} else {
		config = models.StartConfig{Variables: make([]models.Variable, 0)}
	}
	node.Config = config

	// Save platform-specific configuration
	p.SavePlatformConfig(node, iflytekNode)

	return node, nil
}

// parseVariablesFromOutputs parses variable definitions from outputs.
func (p *StartNodeParser) parseVariablesFromOutputs(outputs []interface{}) ([]models.Variable, error) {
	variables := make([]models.Variable, 0, len(outputs))

	for _, output := range outputs {
		if outputMap, ok := output.(map[string]interface{}); ok {
			variable, err := p.parseVariable(outputMap)
			if err != nil {
				return nil, fmt.Errorf("failed to parse variable: %w", err)
			}
			if variable != nil {
				variables = append(variables, *variable)
			}
		}
	}

	return variables, nil
}

// parseVariable parses variable definition.
func (p *StartNodeParser) parseVariable(outputData map[string]interface{}) (*models.Variable, error) {
	name, _ := outputData["name"].(string)
	if name == "" {
		return nil, fmt.Errorf("variable name is empty")
	}

	variable := &models.Variable{
		Name:     name,
		Label:    name,
		Required: false,
	}

	p.parseVariableID(variable, outputData)
	p.parseVariableSchema(variable, outputData)
	p.parseVariableCustomType(variable, outputData)
	p.parseVariableConstraints(variable, outputData)
	p.parseVariableErrorMessage(variable, outputData)

	return variable, nil
}

// parseVariableID parses variable ID information
func (p *StartNodeParser) parseVariableID(variable *models.Variable, outputData map[string]interface{}) {
	if id, ok := outputData["id"].(string); ok {
		variable.ID = id
	}
}

// parseVariableSchema parses schema information including type, default value and properties
func (p *StartNodeParser) parseVariableSchema(variable *models.Variable, outputData map[string]interface{}) {
	schema, ok := outputData["schema"].(map[string]interface{})
	if !ok {
		return
	}

	p.parseVariableType(variable, schema)
	p.parseVariableDefault(variable, schema)
	p.parseVariableProperties(variable, schema)
}

// parseVariableType parses data type from schema
func (p *StartNodeParser) parseVariableType(variable *models.Variable, schema map[string]interface{}) {
	schemaType, ok := schema["type"].(string)
	if !ok {
		return
	}

	mapping := models.GetDefaultDataTypeMapping()
	unifiedType := mapping.FromIFlytekType(schemaType)
	variable.Type = string(unifiedType)
}

// parseVariableDefault parses default value and updates label if meaningful
func (p *StartNodeParser) parseVariableDefault(variable *models.Variable, schema map[string]interface{}) {
	defaultValue, ok := schema["default"]
	if !ok {
		return
	}

	variable.Default = defaultValue

	// If default is a non-empty string, use it as a more meaningful label
	if defaultStr, ok := defaultValue.(string); ok && defaultStr != "" && defaultStr != variable.Name {
		variable.Label = defaultStr
	}
}

// parseVariableProperties parses property constraints
func (p *StartNodeParser) parseVariableProperties(variable *models.Variable, schema map[string]interface{}) {
	if properties, ok := schema["properties"].([]interface{}); ok {
		variable.Properties = properties
	}
}

// parseVariableCustomType parses custom parameter type and adjusts data type accordingly
func (p *StartNodeParser) parseVariableCustomType(variable *models.Variable, outputData map[string]interface{}) {
	customType, ok := outputData["customParameterType"].(string)
	if !ok {
		return
	}

	variable.CustomParameterType = customType

	// Adjust data type based on custom type
	if customType == "xfyun-file" {
		variable.Type = string(models.DataTypeString)
	}
}

// parseVariableConstraints parses constraint conditions
func (p *StartNodeParser) parseVariableConstraints(variable *models.Variable, outputData map[string]interface{}) {
	if required, ok := outputData["required"].(bool); ok {
		variable.Required = required
	}

	if deleteDisabled, ok := outputData["deleteDisabled"].(bool); ok {
		variable.DeleteDisabled = deleteDisabled
	}
}

// parseVariableErrorMessage parses name error message
func (p *StartNodeParser) parseVariableErrorMessage(variable *models.Variable, outputData map[string]interface{}) {
	if nameErrMsg, ok := outputData["nameErrMsg"].(string); ok {
		variable.NameErrMsg = nameErrMsg
	}
}
