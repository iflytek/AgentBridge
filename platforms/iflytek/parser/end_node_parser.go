package parser

import (
	"github.com/iflytek/agentbridge/internal/models"
	"fmt"
)

// EndNodeParser parses end nodes.
type EndNodeParser struct {
	*BaseNodeParser
}

func NewEndNodeParser(variableRefSystem *models.VariableReferenceSystem) *EndNodeParser {
	return &EndNodeParser{
		BaseNodeParser: NewBaseNodeParser(variableRefSystem),
	}
}

// GetSupportedType returns the supported node type.
func (p *EndNodeParser) GetSupportedType() string {
	return "结束节点"
}

// ValidateNode validates node data.
func (p *EndNodeParser) ValidateNode(iflytekNode IFlytekNode) error {
	if iflytekNode.ID == "" {
		return fmt.Errorf("node ID is empty")
	}

	if iflytekNode.Type != p.GetSupportedType() {
		return fmt.Errorf("invalid node type: expected %s, got %s", p.GetSupportedType(), iflytekNode.Type)
	}

	return nil
}

// ParseNode parses a node.
func (p *EndNodeParser) ParseNode(iflytekNode IFlytekNode) (*models.Node, error) {
	if err := p.ValidateNode(iflytekNode); err != nil {
		return nil, err
	}

	// Parse basic information
	node := p.ParseBasicNodeInfo(iflytekNode, models.NodeTypeEnd)

	// Parse inputs and configuration
	if inputs, ok := iflytekNode.Data["inputs"].([]interface{}); ok {
		nodeInputs, err := p.ParseNodeInputs(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end node inputs: %w", err)
		}
		node.Inputs = nodeInputs

		// Generate configuration based on parsed inputs
		config, err := p.parseEndConfig(iflytekNode.Data, nodeInputs)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end config: %w", err)
		}
		node.Config = config
	} else {
		// Create default configuration if no inputs
		node.Config = models.EndConfig{
			OutputMode:   "template",
			StreamOutput: true,
			Outputs:      make([]models.EndOutput, 0),
		}
	}

	// Save platform-specific configuration
	p.SavePlatformConfig(node, iflytekNode)

	return node, nil
}

// parseEndConfig parses end node configuration.
func (p *EndNodeParser) parseEndConfig(data map[string]interface{}, inputs []models.Input) (models.EndConfig, error) {
	config := models.EndConfig{
		OutputMode:   "template",
		StreamOutput: true,
		Outputs:      make([]models.EndOutput, 0),
	}

	// Parse configuration from nodeParam
	if nodeParam, ok := data["nodeParam"].(map[string]interface{}); ok {
		if template, ok := nodeParam["template"].(string); ok {
			config.Template = template
		}
		if streamOutput, ok := nodeParam["streamOutput"].(bool); ok {
			config.StreamOutput = streamOutput
		}
		if outputMode, ok := nodeParam["outputMode"].(float64); ok {
			if outputMode == 1 {
				config.OutputMode = "template"
			} else {
				config.OutputMode = "variables"
			}
		}
	}

	// Build output configuration from inputs
	for _, input := range inputs {
		endOutput := models.EndOutput{
			Variable:  input.Name,
			ValueType: input.Type,
		}

		if input.Reference != nil {
			endOutput.ValueSelector = []string{input.Reference.NodeID, input.Reference.OutputName}
			endOutput.Reference = input.Reference
		}

		config.Outputs = append(config.Outputs, endOutput)
	}

	return config, nil
}
