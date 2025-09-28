package parser

import (
	"github.com/iflytek/agentbridge/internal/models"
	"fmt"
)

// BaseNodeParser provides common functionality for node parsing.
type BaseNodeParser struct {
	variableRefSystem *models.VariableReferenceSystem
}

func NewBaseNodeParser(variableRefSystem *models.VariableReferenceSystem) *BaseNodeParser {
	return &BaseNodeParser{
		variableRefSystem: variableRefSystem,
	}
}

// ParseBasicNodeInfo parses basic node information.
func (p *BaseNodeParser) ParseBasicNodeInfo(iflytekNode IFlytekNode, nodeType models.NodeType) *models.Node {
	node := models.NewNode(iflytekNode.ID, nodeType, "")

	// Set basic properties
	if label, ok := iflytekNode.Data["label"].(string); ok {
		node.Title = label
	}
	if desc, ok := iflytekNode.Data["description"].(string); ok {
		node.Description = desc
	}

	// Set position and size
	node.Position = models.Position{X: iflytekNode.Position.X, Y: iflytekNode.Position.Y}
	node.Size = models.Size{Width: iflytekNode.Width, Height: iflytekNode.Height}

	return node
}

// ParseNodeInputs parses node inputs.
func (p *BaseNodeParser) ParseNodeInputs(inputsData []interface{}) ([]models.Input, error) {
	inputs := make([]models.Input, 0)

	for _, inputData := range inputsData {
		if inputMap, ok := inputData.(map[string]interface{}); ok {
			input, err := p.parseInput(inputMap)
			if err != nil {
				return nil, fmt.Errorf("failed to parse input: %w", err)
			}
			if input != nil {
				inputs = append(inputs, *input)
			}
		}
	}

	return inputs, nil
}

// ParseNodeOutputs parses node outputs.
func (p *BaseNodeParser) ParseNodeOutputs(outputsData []interface{}) ([]models.Output, error) {
	outputs := make([]models.Output, 0)

	for _, outputData := range outputsData {
		if outputMap, ok := outputData.(map[string]interface{}); ok {
			output, err := p.parseOutput(outputMap)
			if err != nil {
				return nil, fmt.Errorf("failed to parse output: %w", err)
			}
			if output != nil {
				outputs = append(outputs, *output)
			}
		}
	}

	return outputs, nil
}

// SavePlatformConfig saves platform-specific configuration.
func (p *BaseNodeParser) SavePlatformConfig(node *models.Node, iflytekNode IFlytekNode) {
	node.PlatformConfig.IFlytek = map[string]interface{}{
		"allowInputReference":  iflytekNode.Data["allowInputReference"],
		"allowOutputReference": iflytekNode.Data["allowOutputReference"],
		"nodeMeta":             iflytekNode.Data["nodeMeta"],
		"nodeParam":            iflytekNode.Data["nodeParam"],
		"references":           iflytekNode.Data["references"],
		"icon":                 iflytekNode.Data["icon"],
		"updatable":            iflytekNode.Data["updatable"],
		"dragging":             iflytekNode.Dragging,
		"selected":             iflytekNode.Selected,
		"positionAbsolute":     iflytekNode.PositionAbsolute,
		// Preserve original inputs and outputs data for type mapping
		"outputs": iflytekNode.Data["outputs"],
		"inputs":  iflytekNode.Data["inputs"],
	}

	// Handle special properties for iteration nodes
	if iflytekNode.ParentID != "" {
		platformConfigMap := node.PlatformConfig.IFlytek
		platformConfigMap["parentId"] = iflytekNode.ParentID
		platformConfigMap["extent"] = iflytekNode.Extent
		platformConfigMap["zIndex"] = iflytekNode.ZIndex
		platformConfigMap["draggable"] = iflytekNode.Draggable
	}
}

// parseInput parses a single input.
func (p *BaseNodeParser) parseInput(inputData map[string]interface{}) (*models.Input, error) {
	name, _ := inputData["name"].(string)
	if name == "" {
		// Skip inputs with empty names, which may be invalid data from user interface
		// Return nil to skip this input, it won't be added to the inputs list
		return nil, nil
	}

	input := &models.Input{
		Name:     name,
		Required: true,
	}

	// Parse schema information
	if schema, ok := inputData["schema"].(map[string]interface{}); ok {
		if schemaType, ok := schema["type"].(string); ok {
			mapping := models.GetDefaultDataTypeMapping()
			input.Type = mapping.FromIFlytekType(schemaType)
		}

		// Parse variable references
		if value, ok := schema["value"].(map[string]interface{}); ok {
			if ref, err := p.variableRefSystem.ParseIFlytekReference(value); err == nil {
				input.Reference = ref
			}
		}

		if defaultValue, ok := schema["default"]; ok {
			input.Default = defaultValue
		}
	}

	return input, nil
}

// parseOutput parses a single output.
func (p *BaseNodeParser) parseOutput(outputData map[string]interface{}) (*models.Output, error) {
	name, _ := outputData["name"].(string)
	if name == "" {
		return nil, fmt.Errorf("output name is empty")
	}

	output := &models.Output{
		Name: name,
	}

	// Parse schema information
	if schema, ok := outputData["schema"].(map[string]interface{}); ok {
		if schemaType, ok := schema["type"].(string); ok {
			mapping := models.GetDefaultDataTypeMapping()
			output.Type = mapping.FromIFlytekType(schemaType)
		}

		if defaultValue, ok := schema["default"]; ok {
			output.Default = defaultValue
		}
	}

	// Parse required field
	if required, ok := outputData["required"].(bool); ok {
		output.Required = required
	}

	return output, nil
}

// IFlytekNode represents iFlytek SparkAgent node structure.
type IFlytekNode struct {
	ID               string                 `yaml:"id"`
	Type             string                 `yaml:"type"`
	Width            float64                `yaml:"width"`
	Height           float64                `yaml:"height"`
	Position         IFlytekPosition        `yaml:"position"`
	PositionAbsolute IFlytekPosition        `yaml:"positionAbsolute"`
	Dragging         bool                   `yaml:"dragging"`
	Selected         bool                   `yaml:"selected"`
	Data             map[string]interface{} `yaml:"data"`
	ParentID         string                 `yaml:"parentId,omitempty"`
	Extent           string                 `yaml:"extent,omitempty"`
	ZIndex           int                    `yaml:"zIndex,omitempty"`
	Draggable        bool                   `yaml:"draggable,omitempty"`
}

// IFlytekPosition contains position information.
type IFlytekPosition struct {
	X float64 `yaml:"x"`
	Y float64 `yaml:"y"`
}
