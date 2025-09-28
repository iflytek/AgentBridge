package parser

import (
	"github.com/iflytek/agentbridge/internal/models"
	"fmt"
)

// CodeNodeParser parses code nodes.
type CodeNodeParser struct {
	*BaseNodeParser
}

func NewCodeNodeParser(variableRefSystem *models.VariableReferenceSystem) *CodeNodeParser {
	return &CodeNodeParser{
		BaseNodeParser: NewBaseNodeParser(variableRefSystem),
	}
}

// GetSupportedType returns the supported node type.
func (p *CodeNodeParser) GetSupportedType() string {
	return "代码"
}

// ValidateNode validates node data.
func (p *CodeNodeParser) ValidateNode(iflytekNode IFlytekNode) error {
	if iflytekNode.ID == "" {
		return fmt.Errorf("node ID is empty")
	}

	if iflytekNode.Type != p.GetSupportedType() {
		return fmt.Errorf("invalid node type: expected %s, got %s", p.GetSupportedType(), iflytekNode.Type)
	}

	return nil
}

// ParseNode parses a node.
func (p *CodeNodeParser) ParseNode(iflytekNode IFlytekNode) (*models.Node, error) {
	if err := p.ValidateNode(iflytekNode); err != nil {
		return nil, err
	}

	// Parse basic information
	node := p.ParseBasicNodeInfo(iflytekNode, models.NodeTypeCode)

	// Parse inputs and outputs
	if err := p.parseInputsOutputs(node, iflytekNode.Data); err != nil {
		return nil, err
	}

	// Parse configuration
	config, err := p.parseCodeConfig(iflytekNode.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse code config: %w", err)
	}
	node.Config = config

	// Save platform-specific configuration
	p.SavePlatformConfig(node, iflytekNode)

	return node, nil
}

// parseInputsOutputs parses inputs and outputs.
func (p *CodeNodeParser) parseInputsOutputs(node *models.Node, data map[string]interface{}) error {
	// Parse inputs
	if inputs, ok := data["inputs"].([]interface{}); ok {
		nodeInputs, err := p.ParseNodeInputs(inputs)
		if err != nil {
			return fmt.Errorf("failed to parse code node inputs: %w", err)
		}
		node.Inputs = nodeInputs
	}

	// Parse outputs
	if outputs, ok := data["outputs"].([]interface{}); ok {
		nodeOutputs, err := p.ParseNodeOutputs(outputs)
		if err != nil {
			return fmt.Errorf("failed to parse code node outputs: %w", err)
		}
		node.Outputs = nodeOutputs
	}

	return nil
}

// parseCodeConfig parses code configuration.
func (p *CodeNodeParser) parseCodeConfig(data map[string]interface{}) (models.CodeConfig, error) {
	config := models.CodeConfig{
		Language: "python3",
	}

	// Parse configuration from nodeParam
	if nodeParam, ok := data["nodeParam"].(map[string]interface{}); ok {
		if code, ok := nodeParam["code"].(string); ok {
			config.Code = code
		}

		if dependencies, ok := nodeParam["dependencies"].([]interface{}); ok {
			config.Dependencies = p.parseDependencies(dependencies)
		}

		// Note: timeout and memoryLimit will be saved in platform-specific configuration
	}

	return config, nil
}

// parseDependencies parses dependency list.
func (p *CodeNodeParser) parseDependencies(dependencies []interface{}) []string {
	result := make([]string, 0)
	for _, dep := range dependencies {
		if depStr, ok := dep.(string); ok {
			result = append(result, depStr)
		}
	}
	return result
}
