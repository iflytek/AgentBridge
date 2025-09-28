// Package parser provides functionality for parsing Coze DSL nodes
package parser

import (
	"github.com/iflytek/agentbridge/internal/models"
	"fmt"
)

// ClassifierNodeParser handles parsing of classifier (intent detection) nodes
type ClassifierNodeParser struct {
	*BaseNodeParser
}

// NewClassifierNodeParser creates a classifier node parser instance
func NewClassifierNodeParser(variableRefSystem *models.VariableReferenceSystem) *ClassifierNodeParser {
	return &ClassifierNodeParser{
		BaseNodeParser: NewBaseNodeParser("22", variableRefSystem),
	}
}

// GetSupportedType returns the supported node type
func (p *ClassifierNodeParser) GetSupportedType() string {
	return "22"
}

// ParseNode converts a Coze classifier node to unified DSL format
func (p *ClassifierNodeParser) ParseNode(cozeNode CozeNode) (*models.Node, error) {
	if err := p.ValidateNode(cozeNode); err != nil {
		return nil, err
	}

	node := p.parseBasicNodeInfo(cozeNode)
	node.Type = models.NodeTypeClassifier

	// Parse classifier-specific configuration
	config, err := p.parseClassifierConfig(cozeNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse classifier config: %w", err)
	}
	node.Config = config

	// Parse inputs and outputs
	node.Inputs = p.parseInputs(cozeNode)
	node.Outputs = p.parseClassifierOutputs(cozeNode)

	return node, nil
}

// parseClassifierConfig extracts classifier configuration from Coze node
func (p *ClassifierNodeParser) parseClassifierConfig(cozeNode CozeNode) (models.ClassifierConfig, error) {
	config := models.ClassifierConfig{
		Classes: []models.ClassifierClass{},
	}

	// Parse input parameters to get query variable - check both formats
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
			if param.Name == "query" {
				if param.Input.Value.Type == "ref" {
					// Extract variable reference as string
					config.QueryVariable = p.extractVariableReference(param.Input)
				}
				break
			}
		}
	}

	// Parse LLM parameters
	if cozeNode.Data.Inputs.LLMParam != nil {
		err := p.parseLLMConfigForClassifier(cozeNode.Data.Inputs.LLMParam, &config)
		if err != nil {
			return config, fmt.Errorf("failed to parse LLM config: %w", err)
		}
	}

	// Parse intent detector configuration
	if cozeNode.Data.Inputs.IntentDetector != nil {
		err := p.parseIntentDetectorConfig(&config, cozeNode.Data.Inputs.IntentDetector)
		if err != nil {
			return config, fmt.Errorf("failed to parse intent detector config: %w", err)
		}
	}

	return config, nil
}

// parseClassifierOutputs generates iFlytek-compatible classifier outputs
func (p *ClassifierNodeParser) parseClassifierOutputs(cozeNode CozeNode) []models.Output {
	// iFlytek classifier node only supports one string output parameter
	outputs := []models.Output{
		{
			Name:        "classificationId",
			Label:       "classificationId",
			Type:        models.DataTypeString, // Force to string type for iFlytek compatibility
			Description: "Classification result ID",
		},
	}

	return outputs
}

// extractVariableReference extracts variable reference string from node input
func (p *ClassifierNodeParser) extractVariableReference(input CozeNodeInput) string {
	if input.Value.Content.BlockID == "" || input.Value.Content.Name == "" {
		return ""
	}
	return fmt.Sprintf("%s.%s", input.Value.Content.BlockID, input.Value.Content.Name)
}

// parseLLMConfigForClassifier processes LLM parameters for classifier
func (p *ClassifierNodeParser) parseLLMConfigForClassifier(llmParam interface{}, config *models.ClassifierConfig) error {
	// Initialize model configuration
	config.Model = models.ModelConfig{
		Provider: "coze",
		Name:     "default",
		Mode:     "chat",
	}

	// Initialize parameters with default values
	config.Parameters = models.ModelParameters{
		Temperature: 0.5,
		MaxTokens:   2048,
	}

	// Handle different formats of LLM parameters
	switch param := llmParam.(type) {
	case []interface{}:
		// Parse array format (similar to LLM node parser)
		for _, paramItem := range param {
			if paramMap, ok := paramItem.(map[string]interface{}); ok {
				if nameInterface, exists := paramMap["name"]; exists {
					if name, ok := nameInterface.(string); ok {
						if inputInterface, exists := paramMap["input"]; exists {
							p.processLLMParam(name, inputInterface, config)
						}
					}
				}
			}
		}
	case map[string]interface{}:
		// Handle map format if needed - process directly as LLM config
		for key, value := range param {
			p.processLLMParamDirect(key, value, config)
		}
	}

	return nil
}

// processLLMParam processes individual LLM parameter
func (p *ClassifierNodeParser) processLLMParam(name string, input interface{}, config *models.ClassifierConfig) {
	if inputMap, ok := input.(map[string]interface{}); ok {
		if value, hasValue := inputMap["value"].(map[string]interface{}); hasValue {
			if content, hasContent := value["content"]; hasContent {
				switch name {
				case "modelType":
					config.Model.Name = fmt.Sprintf("%v", content)
				case "temperature":
					if temp, ok := content.(float64); ok {
						config.Parameters.Temperature = temp
					}
				case "maxTokens":
					if tokens, ok := content.(float64); ok {
						config.Parameters.MaxTokens = int(tokens)
					} else if tokens, ok := content.(int); ok {
						config.Parameters.MaxTokens = tokens
					}
				case "systemPrompt":
					config.Instructions = p.extractNestedContent(content)
				}
			}
		}
	}
}

// processLLMParamDirect processes LLM parameter in direct map format
func (p *ClassifierNodeParser) processLLMParamDirect(key string, value interface{}, config *models.ClassifierConfig) {
	switch key {
	case "modelType":
		config.Model.Name = fmt.Sprintf("%v", value)
	case "temperature":
		if temp, ok := value.(float64); ok {
			config.Parameters.Temperature = temp
		}
	case "maxTokens":
		if tokens, ok := value.(float64); ok {
			config.Parameters.MaxTokens = int(tokens)
		} else if tokens, ok := value.(int); ok {
			config.Parameters.MaxTokens = tokens
		}
	case "systemPrompt":
		config.Instructions = p.extractNestedContent(value)
	}
}

// parseIntentDetectorConfig processes the intentdetector configuration
func (p *ClassifierNodeParser) parseIntentDetectorConfig(config *models.ClassifierConfig, intentDetector interface{}) error {
	// Convert interface{} to map
	intentMap, ok := intentDetector.(map[string]interface{})
	if !ok {
		return fmt.Errorf("intentdetector is not a map")
	}

	// Parse intents list
	if intentsInterface, exists := intentMap["intents"]; exists {
		intentsList, ok := intentsInterface.([]interface{})
		if !ok {
			return fmt.Errorf("intents is not a list")
		}

		for i, intentInterface := range intentsList {
			intentItem, ok := intentInterface.(map[string]interface{})
			if !ok {
				continue
			}

			// Extract intent name
			var intentName string
			if nameInterface, exists := intentItem["name"]; exists {
				if name, ok := nameInterface.(string); ok {
					intentName = name
				}
			}

			if intentName != "" {
				classifierClass := models.ClassifierClass{
					Name:        intentName,
					Description: intentName, // Use name as description
					ID:          fmt.Sprintf("class_%d", i),
				}
				config.Classes = append(config.Classes, classifierClass)
			}
		}
	}

	// Parse mode (if available) - mode information is preserved in the classifier config
	// but not appended to instructions to keep original prompt text clean
	if modeInterface, exists := intentMap["mode"]; exists {
		if mode, ok := modeInterface.(string); ok {
			// Mode information is available but not appended to preserve original prompt purity
			_ = mode // Keep the mode parsing logic but don't use it for instructions
		}
	}

	return nil
}

// extractNestedContent extracts the actual content from nested map structures
func (p *ClassifierNodeParser) extractNestedContent(content interface{}) string {
	// First try direct string conversion
	if strContent, ok := content.(string); ok {
		return strContent
	}

	// Handle nested map structure: map[type:string value:map[content:actual_text type:literal]]
	if contentMap, ok := content.(map[string]interface{}); ok {
		// Check if there's a "value" field
		if valueInterface, hasValue := contentMap["value"]; hasValue {
			if valueMap, ok := valueInterface.(map[string]interface{}); ok {
				// Check if the value map contains "content"
				if actualContent, hasContent := valueMap["content"]; hasContent {
					if actualStr, ok := actualContent.(string); ok {
						return actualStr
					}
				}
			}
		}

		// Check if there's a direct "content" field
		if directContent, hasContent := contentMap["content"]; hasContent {
			if contentStr, ok := directContent.(string); ok {
				return contentStr
			}
		}
	}

	// Fallback to string representation if no nested content found
	return fmt.Sprintf("%v", content)
}
