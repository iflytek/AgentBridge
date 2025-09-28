package parser

import (
	"github.com/iflytek/agentbridge/internal/models"
	"fmt"
)

// BaseNodeParser provides the base implementation for Dify node parsing.
type BaseNodeParser struct {
	nodeType          string
	variableRefSystem *models.VariableReferenceSystem
}

func NewBaseNodeParser(nodeType string, variableRefSystem *models.VariableReferenceSystem) *BaseNodeParser {
	return &BaseNodeParser{
		nodeType:          nodeType,
		variableRefSystem: variableRefSystem,
	}
}

// GetSupportedType returns the supported node type.
func (p *BaseNodeParser) GetSupportedType() string {
	return p.nodeType
}

// ParseNode performs basic node parsing with default implementation.
func (p *BaseNodeParser) ParseNode(difyNode DifyNode) (*models.Node, error) {
	// Validate node
	if err := p.ValidateNode(difyNode); err != nil {
		return nil, err
	}

	// Parse basic node information
	node := p.parseBasicNodeInfo(difyNode)

	// Set corresponding unified DSL type based on node type
	switch difyNode.Data.Type {
	case "start":
		node.Type = models.NodeTypeStart
	case "end":
		node.Type = models.NodeTypeEnd
	case "llm":
		node.Type = models.NodeTypeLLM
	case "code":
		node.Type = models.NodeTypeCode
	case "if-else":
		node.Type = models.NodeTypeCondition
	case "question-classifier":
		node.Type = models.NodeTypeClassifier
	case "iteration":
		node.Type = models.NodeTypeIteration
	default:
		return nil, fmt.Errorf("unsupported node type: %s", difyNode.Data.Type)
	}

	// Parse inputs and outputs
	node.Inputs = p.parseInputs(difyNode.Data.Variables)
	node.Outputs = p.parseOutputsInterface(difyNode.Data.Outputs)

	// Node configuration is implemented by specific node parsers
	// Base parser provides default configuration
	node.Config = models.StartConfig{Variables: []models.Variable{}}

	return node, nil
}

// ValidateNode performs basic node validation.
func (p *BaseNodeParser) ValidateNode(difyNode DifyNode) error {
	if difyNode.ID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	if difyNode.Data.Type == "" {
		return fmt.Errorf("node type cannot be empty")
	}

	if difyNode.Data.Title == "" {
		return fmt.Errorf("node title cannot be empty")
	}

	return nil
}

// parseBasicNodeInfo parses basic node information.
func (p *BaseNodeParser) parseBasicNodeInfo(difyNode DifyNode) *models.Node {
	node := &models.Node{
		ID:          difyNode.ID,
		Title:       difyNode.Data.Title,
		Description: difyNode.Data.Desc,
		Position: models.Position{
			X: difyNode.Position.X,
			Y: difyNode.Position.Y,
		},
		Size: models.Size{
			Width:  difyNode.Width,
			Height: difyNode.Height,
		},
		Inputs:  []models.Input{},
		Outputs: []models.Output{},
		PlatformConfig: models.PlatformConfig{
			Dify: map[string]interface{}{
				"selected":       difyNode.Selected,
				"draggable":      difyNode.Draggable,
				"selectable":     difyNode.Selectable,
				"sourcePosition": difyNode.SourcePosition,
				"targetPosition": difyNode.TargetPosition,
				"zIndex":         difyNode.ZIndex,
				"isInIteration":  difyNode.Data.IsInIteration,
				"parentId":       difyNode.ParentID,
				"extent":         difyNode.Extent,
			},
		},
	}

	return node
}

// parseInputs parses node inputs.
func (p *BaseNodeParser) parseInputs(variables []DifyVariable) []models.Input {
	inputs := make([]models.Input, 0, len(variables))

	for _, variable := range variables {
		input := models.Input{
			Name:        variable.Variable,
			Label:       variable.Label,
			Type:        p.convertDataType(variable.Type),
			Required:    variable.Required,
			Description: "",
		}

		// Parse constraint conditions
		if variable.MaxLength > 0 {
			input.Constraints = &models.Constraints{
				MaxLength: variable.MaxLength,
			}
		}

		inputs = append(inputs, input)
	}

	return inputs
}

// parseOutputs parses node outputs for backward compatibility.

// parseOutputsInterface parses node outputs (supports interface{} type).
func (p *BaseNodeParser) parseOutputsInterface(outputs interface{}) []models.Output {
	if outputs == nil {
		return []models.Output{}
	}

	// Try array format first (end node)
	if result := p.parseArrayOutputs(outputs); result != nil {
		return result
	}

	// Try object format (code node)
	if result := p.parseObjectOutputs(outputs); result != nil {
		return result
	}

	return []models.Output{}
}

// parseArrayOutputs parses array format outputs
func (p *BaseNodeParser) parseArrayOutputs(outputs interface{}) []models.Output {
	outputArray, ok := outputs.([]interface{})
	if !ok {
		return nil
	}

	var result []models.Output
	for _, item := range outputArray {
		if output := p.parseArrayOutputItem(item); output != nil {
			result = append(result, *output)
		}
	}
	return result
}

// parseArrayOutputItem parses single array output item
func (p *BaseNodeParser) parseArrayOutputItem(item interface{}) *models.Output {
	outputMap, ok := item.(map[interface{}]interface{})
	if !ok {
		return nil
	}

	output := models.Output{Description: ""}

	if variable, ok := outputMap["variable"].(string); ok {
		output.Name = variable
	}
	if valueType, ok := outputMap["value_type"].(string); ok {
		output.Type = p.convertDataType(valueType)
	}

	return &output
}

// parseObjectOutputs parses object format outputs
func (p *BaseNodeParser) parseObjectOutputs(outputs interface{}) []models.Output {
	outputMap, ok := outputs.(map[interface{}]interface{})
	if !ok {
		return nil
	}

	var result []models.Output
	for key, value := range outputMap {
		if output := p.parseObjectOutputItem(key, value); output != nil {
			result = append(result, *output)
		}
	}
	return result
}

// parseObjectOutputItem parses single object output item
func (p *BaseNodeParser) parseObjectOutputItem(key, value interface{}) *models.Output {
	outputName, ok := key.(string)
	if !ok {
		return nil
	}

	outputInfo, ok := value.(map[interface{}]interface{})
	if !ok {
		return nil
	}

	return &models.Output{
		Name:        outputName,
		Type:        p.convertOutputType(outputInfo),
		Description: "Output result",
	}
}

// convertOutputType converts output type (from object format).
func (p *BaseNodeParser) convertOutputType(outputInfo map[interface{}]interface{}) models.UnifiedDataType {
	if typeStr, ok := outputInfo["type"].(string); ok {
		return p.convertDataType(typeStr)
	}
	return models.DataTypeString
}

// convertDataType converts data types.
func (p *BaseNodeParser) convertDataType(difyType string) models.UnifiedDataType {
	switch difyType {
	case "text-input", "string":
		return models.DataTypeString
	case "number":
		return models.DataTypeNumber
	case "boolean":
		return models.DataTypeBoolean
	case "array[string]":
		return models.DataTypeArrayString
	case "array[object]":
		return models.DataTypeArrayObject
	case "object":
		return models.DataTypeObject
	default:
		return models.DataTypeString // Default to string type
	}
}
