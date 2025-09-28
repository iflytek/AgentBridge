package parser

import (
	"github.com/iflytek/agentbridge/internal/models"
	"fmt"
)

// ClassifierNodeParser parses classification decision nodes.
type ClassifierNodeParser struct {
	*BaseNodeParser
}

func NewClassifierNodeParser(variableRefSystem *models.VariableReferenceSystem) *ClassifierNodeParser {
	return &ClassifierNodeParser{
		BaseNodeParser: NewBaseNodeParser(variableRefSystem),
	}
}

// GetSupportedType returns the supported node type.
func (p *ClassifierNodeParser) GetSupportedType() string {
	return "决策"
}

// ValidateNode validates node data.
func (p *ClassifierNodeParser) ValidateNode(iflytekNode IFlytekNode) error {
	if iflytekNode.ID == "" {
		return fmt.Errorf("node ID is empty")
	}

	if iflytekNode.Type != p.GetSupportedType() {
		return fmt.Errorf("invalid node type: expected %s, got %s", p.GetSupportedType(), iflytekNode.Type)
	}

	return nil
}

// ParseNode parses a node.
func (p *ClassifierNodeParser) ParseNode(iflytekNode IFlytekNode) (*models.Node, error) {
	if err := p.ValidateNode(iflytekNode); err != nil {
		return nil, err
	}

	// Parse basic information
	node := p.ParseBasicNodeInfo(iflytekNode, models.NodeTypeClassifier)

	// Parse inputs and outputs
	if err := p.parseInputsOutputs(node, iflytekNode.Data); err != nil {
		return nil, err
	}

	// Parse configuration
	config, err := p.parseClassifierConfig(iflytekNode.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse classifier config: %w", err)
	}
	node.Config = config

	// Save platform-specific configuration
	p.SavePlatformConfig(node, iflytekNode)

	return node, nil
}

// parseInputsOutputs parses inputs and outputs.
func (p *ClassifierNodeParser) parseInputsOutputs(node *models.Node, data map[string]interface{}) error {
	// Parse inputs
	if inputs, ok := data["inputs"].([]interface{}); ok {
		nodeInputs, err := p.ParseNodeInputs(inputs)
		if err != nil {
			return fmt.Errorf("failed to parse classifier node inputs: %w", err)
		}
		node.Inputs = nodeInputs
	}

	// Parse outputs
	if outputs, ok := data["outputs"].([]interface{}); ok {
		nodeOutputs, err := p.ParseNodeOutputs(outputs)
		if err != nil {
			return fmt.Errorf("failed to parse classifier node outputs: %w", err)
		}
		node.Outputs = nodeOutputs
	}

	return nil
}

// parseClassifierConfig parses classification decision configuration.
func (p *ClassifierNodeParser) parseClassifierConfig(data map[string]interface{}) (*models.ClassifierConfig, error) {
	config := &models.ClassifierConfig{
		Classes: make([]models.ClassifierClass, 0),
	}

	// Get configuration from nodeParam
	nodeParam, ok := data["nodeParam"].(map[string]interface{})
	if !ok {
		return config, nil
	}

	// Parse all configuration sections
	if err := p.parseClassifierModelConfig(nodeParam, config); err != nil {
		return nil, err
	}

	if err := p.parseClassifierClasses(nodeParam, config); err != nil {
		return nil, err
	}

	p.parseClassifierQueryVariable(data, config)
	p.parseClassifierInstructions(nodeParam, config)

	return config, nil
}

// parseClassifierModelConfig parses model configuration for classifier
func (p *ClassifierNodeParser) parseClassifierModelConfig(nodeParam map[string]interface{}, config *models.ClassifierConfig) error {
	modelConfig, err := p.parseModelConfig(nodeParam)
	if err != nil {
		return fmt.Errorf("failed to parse model config: %w", err)
	}
	config.Model = *modelConfig
	config.Parameters = p.parseModelParameters(nodeParam)
	return nil
}

// parseClassifierClasses parses classification categories
func (p *ClassifierNodeParser) parseClassifierClasses(nodeParam map[string]interface{}, config *models.ClassifierConfig) error {
	intentChains, ok := nodeParam["intentChains"].([]interface{})
	if !ok {
		return nil
	}

	for _, intentData := range intentChains {
		intentMap, ok := intentData.(map[string]interface{})
		if !ok {
			continue
		}

		classifierClass, err := p.parseClassifierClass(intentMap)
		if err != nil {
			return fmt.Errorf("failed to parse classifier class: %w", err)
		}
		config.Classes = append(config.Classes, *classifierClass)
	}
	return nil
}

// parseClassifierQueryVariable parses query variable from inputs
func (p *ClassifierNodeParser) parseClassifierQueryVariable(data map[string]interface{}, config *models.ClassifierConfig) {
	inputs, ok := data["inputs"].([]interface{})
	if !ok || len(inputs) == 0 {
		return
	}

	firstInput, ok := inputs[0].(map[string]interface{})
	if !ok {
		return
	}

	if name, ok := firstInput["name"].(string); ok {
		config.QueryVariable = name
	}
}

// parseClassifierInstructions parses prompt prefix as instructions
func (p *ClassifierNodeParser) parseClassifierInstructions(nodeParam map[string]interface{}, config *models.ClassifierConfig) {
	if promptPrefix, ok := nodeParam["promptPrefix"].(string); ok {
		config.Instructions = promptPrefix
	}
}

// parseModelConfig parses model configuration.
func (p *ClassifierNodeParser) parseModelConfig(nodeParam map[string]interface{}) (*models.ModelConfig, error) {
	modelConfig := &models.ModelConfig{
		Mode: "chat", // Default to chat mode
	}

	// Parse provider (inferred from serviceId or domain)
	if serviceId, ok := nodeParam["serviceId"].(string); ok {
		modelConfig.Provider = serviceId
	} else if domain, ok := nodeParam["domain"].(string); ok {
		modelConfig.Provider = domain
	}

	// Parse model name
	if model, ok := nodeParam["model"].(string); ok {
		modelConfig.Name = model
	}

	return modelConfig, nil
}

// parseClassifierClass parses classification category.
func (p *ClassifierNodeParser) parseClassifierClass(intentData map[string]interface{}) (*models.ClassifierClass, error) {
	classifierClass := &models.ClassifierClass{}

	// Parse ID
	if id, ok := intentData["id"].(string); ok {
		classifierClass.ID = id
	}

	// Parse name
	if name, ok := intentData["name"].(string); ok {
		classifierClass.Name = name
	}

	// Parse description
	if description, ok := intentData["description"].(string); ok {
		classifierClass.Description = description
	}

	// Parse intent type to determine if it's default intent
	if intentType, ok := intentData["intentType"].(float64); ok {
		// intentType: 1 means default intent, 2 means normal classification
		classifierClass.IsDefault = (int(intentType) == 1)
	} else if intentType, ok := intentData["intentType"].(int); ok {
		classifierClass.IsDefault = (intentType == 1)
	}

	return classifierClass, nil
}

// parseModelParameters parses model parameters - uses unified ModelParameters structure.
func (p *ClassifierNodeParser) parseModelParameters(nodeParam map[string]interface{}) models.ModelParameters {
	params := models.ModelParameters{
		Temperature: 0.7,  // Default value
		MaxTokens:   4096, // Default value
	}

	// Temperature parameter - supports float64 type
	if temperature, ok := nodeParam["temperature"].(float64); ok {
		params.Temperature = temperature
	}

	// MaxTokens parameter - supports int and float64 types
	if maxTokens, ok := nodeParam["maxTokens"].(int); ok {
		params.MaxTokens = maxTokens
	} else if maxTokens, ok := nodeParam["maxTokens"].(float64); ok {
		params.MaxTokens = int(maxTokens)
	}

	// TopK parameter - supports int and float64 types
	if topK, ok := nodeParam["topK"].(int); ok {
		params.TopK = topK
	} else if topK, ok := nodeParam["topK"].(float64); ok {
		params.TopK = int(topK)
	}

	return params
}
