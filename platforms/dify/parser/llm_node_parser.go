package parser

import (
	"ai-agents-transformer/internal/models"
	"fmt"
)

// LLMNodeParser parses Dify LLM nodes.
type LLMNodeParser struct {
	*BaseNodeParser
}

func NewLLMNodeParser(vrs *models.VariableReferenceSystem) NodeParser {
	return &LLMNodeParser{
		BaseNodeParser: NewBaseNodeParser("llm", vrs),
	}
}

// GetSupportedType returns supported node type.
func (p *LLMNodeParser) GetSupportedType() string {
	return "llm"
}

// ParseNode parses Dify LLM node.
func (p *LLMNodeParser) ParseNode(difyNode DifyNode) (*models.Node, error) {
	if err := p.ValidateNode(difyNode); err != nil {
		return nil, err
	}

	node := p.parseBasicNodeInfo(difyNode)
	node.Type = models.NodeTypeLLM

	config := p.buildLLMConfig(difyNode)
	node.Config = config

	node.Inputs = p.extractInputsFromTemplate(difyNode.Data.PromptTemplate, difyNode.Data.Context)
	node.Outputs = p.createLLMOutputs()

	return node, nil
}

// buildLLMConfig builds LLM configuration from Dify node
func (p *LLMNodeParser) buildLLMConfig(difyNode DifyNode) models.LLMConfig {
	config := models.LLMConfig{}

	if difyNode.Data.Model != nil {
		config.Model = p.parseModelConfig(difyNode.Data.Model)
		config.Parameters = p.parseModelParameters(difyNode.Data.Model)
	}

	config.Prompt = p.parsePromptConfig(difyNode.Data.PromptTemplate)
	config.Context = p.parseContextConfig(difyNode.Data.Context)
	config.Vision = p.parseVisionConfig(difyNode.Data.Vision)

	return config
}

// parseModelConfig parses model configuration
func (p *LLMNodeParser) parseModelConfig(model *DifyModel) models.ModelConfig {
	return models.ModelConfig{
		Provider: p.extractProvider(model.Provider),
		Name:     model.Name,
		Mode:     model.Mode,
	}
}

// parseModelParameters parses model parameters
func (p *LLMNodeParser) parseModelParameters(model *DifyModel) models.ModelParameters {
	if model.CompletionParams == nil {
		return models.ModelParameters{}
	}

	return models.ModelParameters{
		Temperature: p.getFloatFromParams(model.CompletionParams, "temperature", 0.7),
		MaxTokens:   p.getIntFromParams(model.CompletionParams, "max_tokens", 8192),
		TopK:        p.getIntFromParams(model.CompletionParams, "top_k", 4),
		TopP:        p.getFloatFromParams(model.CompletionParams, "top_p", 0.7),
	}
}

// parsePromptConfig parses prompt configuration
func (p *LLMNodeParser) parsePromptConfig(promptTemplate []DifyPrompt) models.PromptConfig {
	if len(promptTemplate) == 0 {
		return models.PromptConfig{}
	}

	config := models.PromptConfig{
		Messages: make([]models.Message, 0, len(promptTemplate)),
	}

	for _, template := range promptTemplate {
		p.processPromptTemplate(&config, template)
	}

	return config
}

// processPromptTemplate processes individual prompt template
func (p *LLMNodeParser) processPromptTemplate(config *models.PromptConfig, template DifyPrompt) {
	unifiedText := p.convertDifyTemplateToUnified(template.Text)

	if template.Role == "system" {
		config.SystemTemplate = unifiedText
	} else if template.Role == "user" {
		config.UserTemplate = unifiedText
	}

	message := models.Message{
		Role:    template.Role,
		Content: unifiedText,
	}
	config.Messages = append(config.Messages, message)
}

// parseContextConfig parses context configuration
func (p *LLMNodeParser) parseContextConfig(context *DifyContext) *models.ContextConfig {
	if context == nil {
		return nil
	}

	config := &models.ContextConfig{
		Enabled: context.Enabled,
	}

	if len(context.VariableSelector) >= 2 {
		config.VariableSelector = []string{
			context.VariableSelector[0], // nodeId
			context.VariableSelector[1], // outputName
		}
	}

	return config
}

// parseVisionConfig parses vision configuration
func (p *LLMNodeParser) parseVisionConfig(vision *DifyVision) *models.VisionConfig {
	if vision == nil {
		return nil
	}

	return &models.VisionConfig{
		Enabled: vision.Enabled,
	}
}

// createLLMOutputs creates default LLM outputs
func (p *LLMNodeParser) createLLMOutputs() []models.Output {
	return []models.Output{
		{
			Name:        "output",
			Type:        models.DataTypeString,
			Description: "LLM output result",
		},
	}
}

// ValidateNode validates Dify LLM node.
func (p *LLMNodeParser) ValidateNode(difyNode DifyNode) error {
	// Use base validation
	if err := p.BaseNodeParser.ValidateNode(difyNode); err != nil {
		return err
	}

	// LLM node specific validation
	if difyNode.Data.Type != "llm" {
		return fmt.Errorf("node type must be 'llm', got '%s'", difyNode.Data.Type)
	}

	// Validate required model configuration
	if difyNode.Data.Model == nil {
		return fmt.Errorf("LLM node must have model configuration")
	}

	return nil
}

// extractProvider extracts provider information.
func (p *LLMNodeParser) extractProvider(provider string) string {
	// Extract real provider name from complete provider path
	if provider == "langgenius/openai_api_compatible/openai_api_compatible" {
		return "openai_compatible"
	}
	return provider
}

// convertDifyTemplateToUnified converts Dify template format to unified format.
func (p *LLMNodeParser) convertDifyTemplateToUnified(difyTemplate string) string {
	// Dify format: {{#nodeId.outputName#}}
	// Unified format: {{$nodes.nodeId.outputName}}
	// Simple format conversion, complex parsing can be implemented as needed
	return difyTemplate // Return original format, variable reference conversion handled elsewhere
}

// extractInputsFromTemplate extracts input variables from templates.
func (p *LLMNodeParser) extractInputsFromTemplate(templates []DifyPrompt, context *DifyContext) []models.Input {
	inputs := make([]models.Input, 0)

	// Extract inputs from context
	if context != nil && context.Enabled && len(context.VariableSelector) >= 2 {
		input := models.Input{
			Name: context.VariableSelector[1], // Output name as input name
			Type: models.DataTypeString,       // Default to string
			Reference: &models.VariableReference{
				Type:       models.ReferenceTypeNodeOutput,
				NodeID:     context.VariableSelector[0],
				OutputName: context.VariableSelector[1],
				DataType:   models.DataTypeString,
			},
		}
		inputs = append(inputs, input)
	}

	// Variable reference parsing is already implemented through VariableSelectorConverter
	// Supports parsing {{#nodeId.outputName#}} format in templates

	return inputs
}

// getFloatFromParams gets float value from parameter map.
func (p *LLMNodeParser) getFloatFromParams(params map[string]interface{}, key string, defaultValue float64) float64 {
	if value, exists := params[key]; exists {
		if floatVal, ok := value.(float64); ok {
			return floatVal
		}
		if intVal, ok := value.(int); ok {
			return float64(intVal)
		}
	}
	return defaultValue
}

// getIntFromParams gets integer value from parameter map.
func (p *LLMNodeParser) getIntFromParams(params map[string]interface{}, key string, defaultValue int) int {
	if value, exists := params[key]; exists {
		if intVal, ok := value.(int); ok {
			return intVal
		}
		if floatVal, ok := value.(float64); ok {
			return int(floatVal)
		}
	}
	return defaultValue
}
