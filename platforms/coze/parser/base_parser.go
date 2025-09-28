package parser

import (
	"fmt"
	"github.com/iflytek/agentbridge/internal/models"
)

// BaseNodeParser provides the base implementation for Coze node parsing.
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
func (p *BaseNodeParser) ParseNode(cozeNode CozeNode) (*models.Node, error) {
	// Validate node
	if err := p.ValidateNode(cozeNode); err != nil {
		return nil, err
	}

	// Parse basic node information
	node := p.parseBasicNodeInfo(cozeNode)

	// Set corresponding unified DSL type based on node type
	switch cozeNode.Type {
	case "1":
		node.Type = models.NodeTypeStart
	case "2":
		node.Type = models.NodeTypeEnd
	case "3":
		node.Type = models.NodeTypeLLM
	case "4":
		node.Type = models.NodeTypeCondition
	case "5":
		node.Type = models.NodeTypeCode
	case "7":
		node.Type = models.NodeTypeIteration
	case "8":
		node.Type = models.NodeTypeCondition // Selector nodes map to condition type
	case "21":
		node.Type = models.NodeTypeIteration // Loop nodes map to iteration type
	case "22":
		node.Type = models.NodeTypeClassifier // Intent detection nodes map to classifier type
	default:
		return nil, fmt.Errorf("unsupported node type: %s", cozeNode.Type)
	}

	// Parse inputs and outputs
	node.Inputs = p.parseInputs(cozeNode)
	node.Outputs = p.parseOutputs(cozeNode)

	// Node configuration is implemented by specific node parsers
	// Base parser provides default configuration
	node.Config = models.StartConfig{Variables: []models.Variable{}}

	return node, nil
}

// ValidateNode performs basic node validation.
func (p *BaseNodeParser) ValidateNode(cozeNode CozeNode) error {
	if cozeNode.ID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	if cozeNode.Type == "" {
		return fmt.Errorf("node type cannot be empty")
	}

	if cozeNode.Data.Meta.Title == "" {
		return fmt.Errorf("node title cannot be empty")
	}

	return nil
}

// parseBasicNodeInfo parses basic node information.
func (p *BaseNodeParser) parseBasicNodeInfo(cozeNode CozeNode) *models.Node {
	node := &models.Node{
		ID:          cozeNode.ID,
		Title:       cozeNode.Data.Meta.Title,
		Description: cozeNode.Data.Meta.Description,
		Position: models.Position{
			X: cozeNode.Meta.Position.X,
			Y: cozeNode.Meta.Position.Y,
		},
		Size: models.Size{
			Width:  244, // Default width
			Height: 118, // Default height
		},
		Inputs:  []models.Input{},
		Outputs: []models.Output{},
		PlatformConfig: models.PlatformConfig{
			IFlytek: make(map[string]interface{}),
			Dify:    make(map[string]interface{}),
		},
	}

	return node
}

// parseInputs parses node inputs from Coze node structure.
func (p *BaseNodeParser) parseInputs(cozeNode CozeNode) []models.Input {
	inputs := make([]models.Input, 0)

	// Parse inputs from node inputs structure - check both formats
	var inputParams []CozeNodeInputParam
	if cozeNode.Data.Inputs != nil {
		if cozeNode.Data.Inputs.InputParameters != nil {
			inputParams = cozeNode.Data.Inputs.InputParameters
		} else if cozeNode.Data.Inputs.InputParametersAlt != nil {
			inputParams = cozeNode.Data.Inputs.InputParametersAlt
		}
	}

	if len(inputParams) > 0 {
		for _, param := range inputParams {
			input := models.Input{
				Name:        param.Name,
				Label:       param.Name,
				Type:        p.convertDataType(param.Input.Type),
				Required:    true, // Default to required
				Description: "",
			}

			// Parse reference if it exists
			if param.Input.Value.Type == "ref" {
				blockID := param.Input.Value.Content.BlockID
				outputName := param.Input.Value.Content.Name

				// Apply output name mapping if available
				if p.variableRefSystem != nil {
					outputName = p.variableRefSystem.ResolveOutputName(blockID, outputName)
				}

				input.Reference = &models.VariableReference{
					Type:       models.ReferenceTypeNodeOutput,
					NodeID:     blockID,
					OutputName: outputName,
					DataType:   p.convertDataType(param.Input.Type),
				}
			}

			inputs = append(inputs, input)
		}
	}

	return inputs
}

// parseOutputs parses node outputs from Coze node structure.
func (p *BaseNodeParser) parseOutputs(cozeNode CozeNode) []models.Output {
	outputs := make([]models.Output, 0)

	// Parse outputs from node data structure
	for _, output := range cozeNode.Data.Outputs {
		// Determine the correct data type based on Coze output structure
		var outputType models.UnifiedDataType

		// Check if output has schema information for list types
		if output.Type == "list" && output.Schema != nil {
			// For list types, convert based on schema element type
			// Try to extract type from schema (could be CozeOutputSchema or map)
			var schemaType string
			if schemaObj, ok := output.Schema.(CozeOutputSchema); ok {
				schemaType = schemaObj.Type
			} else if schemaMap, ok := output.Schema.(map[string]interface{}); ok {
				if typeVal, exists := schemaMap["type"]; exists {
					if typeStr, ok := typeVal.(string); ok {
						schemaType = typeStr
					}
				}
			}

			switch schemaType {
			case "string":
				outputType = models.DataTypeArrayString
			case "integer":
				outputType = models.DataTypeArrayString // iFlytek doesn't distinguish array element types
			case "float":
				outputType = models.DataTypeArrayString
			case "boolean":
				outputType = models.DataTypeArrayString
			default:
				outputType = models.DataTypeArrayString
			}
		} else {
			// For non-list types, use normal conversion
			outputType = p.convertDataType(output.Type)
		}

		modelOutput := models.Output{
			Name:        output.Name,
			Label:       output.Name,
			Type:        outputType,
			Description: "",
		}

		outputs = append(outputs, modelOutput)
	}

	return outputs
}

// convertDataType converts Coze data types to unified data types.
func (p *BaseNodeParser) convertDataType(cozeType string) models.UnifiedDataType {
	switch cozeType {
	case "string":
		return models.DataTypeString
	case "integer":
		return models.DataTypeInteger
	case "float":
		return models.DataTypeFloat
	case "boolean":
		return models.DataTypeBoolean
	case "array":
		return models.DataTypeArrayString
	case "object":
		return models.DataTypeObject
	default:
		return models.DataTypeString // Default to string type
	}
}
