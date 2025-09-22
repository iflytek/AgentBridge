package parser

import (
	"ai-agents-transformer/internal/models"
	"fmt"
)

// LLMNodeParser parses LLM nodes.
type LLMNodeParser struct {
	*BaseNodeParser
}

func NewLLMNodeParser(variableRefSystem *models.VariableReferenceSystem) *LLMNodeParser {
	return &LLMNodeParser{
		BaseNodeParser: NewBaseNodeParser(variableRefSystem),
	}
}

// GetSupportedType returns the supported node type.
func (p *LLMNodeParser) GetSupportedType() string {
	return "大模型"
}

// ValidateNode validates node data.
func (p *LLMNodeParser) ValidateNode(iflytekNode IFlytekNode) error {
	if iflytekNode.ID == "" {
		return fmt.Errorf("node ID is empty")
	}

	if iflytekNode.Type != p.GetSupportedType() {
		return fmt.Errorf("invalid node type: expected %s, got %s", p.GetSupportedType(), iflytekNode.Type)
	}

	return nil
}

// ParseNode parses a node.
func (p *LLMNodeParser) ParseNode(iflytekNode IFlytekNode) (*models.Node, error) {
	if err := p.ValidateNode(iflytekNode); err != nil {
		return nil, err
	}

	// Parse basic information
	node := p.ParseBasicNodeInfo(iflytekNode, models.NodeTypeLLM)

	// Parse inputs
	if inputs, ok := iflytekNode.Data["inputs"].([]interface{}); ok {
		nodeInputs, err := p.ParseNodeInputs(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to parse LLM node inputs: %w", err)
		}
		node.Inputs = nodeInputs
	}

	// Parse outputs
	if outputs, ok := iflytekNode.Data["outputs"].([]interface{}); ok {
		nodeOutputs, err := p.ParseNodeOutputs(outputs)
		if err != nil {
			return nil, fmt.Errorf("failed to parse LLM node outputs: %w", err)
		}
		node.Outputs = nodeOutputs
	}

	// Parse configuration
	config, err := p.parseLLMConfig(iflytekNode.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM config: %w", err)
	}
	node.Config = config

	// Save platform-specific configuration
	p.SavePlatformConfig(node, iflytekNode)

	return node, nil
}

// parseLLMConfig parses LLM configuration.
func (p *LLMNodeParser) parseLLMConfig(data map[string]interface{}) (models.LLMConfig, error) {
	config := models.LLMConfig{
		Model: models.ModelConfig{
			Provider: "iflytek",
			Mode:     "chat",
		},
		Parameters: models.ModelParameters{
			Temperature: 0.7,
			MaxTokens:   4096,
		},
		Prompt: models.PromptConfig{},
	}

	// Parse detailed configuration from nodeParam
	if nodeParam, ok := data["nodeParam"].(map[string]interface{}); ok {
		p.parseModelConfig(&config, nodeParam)
		p.parseModelParameters(&config, nodeParam)
		p.parsePromptConfig(&config, nodeParam)
		p.parseContextConfig(&config, nodeParam)
		p.parseVisionConfig(&config, nodeParam)
	}

	return config, nil
}

// parseModelConfig parses model configuration.
func (p *LLMNodeParser) parseModelConfig(config *models.LLMConfig, nodeParam map[string]interface{}) {
	if domain, ok := nodeParam["domain"].(string); ok {
		config.Model.Name = domain
	}
	if serviceId, ok := nodeParam["serviceId"].(string); ok {
		config.Model.Provider = fmt.Sprintf("iflytek/%s", serviceId)
	}
}

// parseModelParameters parses model parameters.
func (p *LLMNodeParser) parseModelParameters(config *models.LLMConfig, nodeParam map[string]interface{}) {
	// Temperature parameter - supports float64 type
	if temperature, ok := nodeParam["temperature"].(float64); ok {
		config.Parameters.Temperature = temperature
	}

	// MaxTokens parameter - supports int and float64 types
	if maxTokens, ok := nodeParam["maxTokens"].(int); ok {
		config.Parameters.MaxTokens = maxTokens
	} else if maxTokens, ok := nodeParam["maxTokens"].(float64); ok {
		config.Parameters.MaxTokens = int(maxTokens)
	}

	// TopK parameter - supports int and float64 types
	if topK, ok := nodeParam["topK"].(int); ok {
		config.Parameters.TopK = topK
	} else if topK, ok := nodeParam["topK"].(float64); ok {
		config.Parameters.TopK = int(topK)
	}
}

// parsePromptConfig parses prompt configuration.
func (p *LLMNodeParser) parsePromptConfig(config *models.LLMConfig, nodeParam map[string]interface{}) {
	if systemTemplate, ok := nodeParam["systemTemplate"].(string); ok {
		config.Prompt.SystemTemplate = systemTemplate
	}
	if userTemplate, ok := nodeParam["template"].(string); ok && userTemplate != "无" {
		config.Prompt.UserTemplate = userTemplate
	}
}

// parseContextConfig parses context configuration.
func (p *LLMNodeParser) parseContextConfig(config *models.LLMConfig, nodeParam map[string]interface{}) {
	if searchDisable, ok := nodeParam["searchDisable"].(bool); ok {
		config.Context = &models.ContextConfig{
			Enabled: !searchDisable,
		}
	}
}

// parseVisionConfig parses vision configuration.
func (p *LLMNodeParser) parseVisionConfig(config *models.LLMConfig, nodeParam map[string]interface{}) {
	if multiMode, ok := nodeParam["multiMode"].(bool); ok {
		config.Vision = &models.VisionConfig{
			Enabled: multiMode,
		}
	}
}
