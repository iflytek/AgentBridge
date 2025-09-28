package parser

import (
	"fmt"
	"github.com/iflytek/agentbridge/internal/models"
	"strconv"
	"strings"
)

// ResponseFormat constants for Coze LLM nodes
const (
	ResponseFormatText     = "0"
	ResponseFormatJSON     = "2"
	ResponseFormatMarkdown = "1" // Ignored in conversion
)

// LLMNodeParser parses Coze LLM nodes.
type LLMNodeParser struct {
	*BaseNodeParser
}

func NewLLMNodeParser(variableRefSystem *models.VariableReferenceSystem) *LLMNodeParser {
	return &LLMNodeParser{
		BaseNodeParser: NewBaseNodeParser("3", variableRefSystem),
	}
}

// GetSupportedType returns the supported node type.
func (p *LLMNodeParser) GetSupportedType() string {
	return "3"
}

// ParseNode parses a Coze LLM node into unified DSL.
func (p *LLMNodeParser) ParseNode(cozeNode CozeNode) (*models.Node, error) {
	if err := p.ValidateNode(cozeNode); err != nil {
		return nil, err
	}

	node := p.parseBasicNodeInfo(cozeNode)
	node.Type = models.NodeTypeLLM

	// Parse LLM configuration
	config, err := p.parseLLMConfig(cozeNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM config: %w", err)
	}
	node.Config = config

	// Parse inputs from inputParameters
	node.Inputs = p.parseInputs(cozeNode)

	// Parse outputs (filtering out reasoning_content)
	node.Outputs = p.parseNodeOutputs(cozeNode)

	return node, nil
}

// parseLLMConfig extracts LLM configuration from Coze node
func (p *LLMNodeParser) parseLLMConfig(cozeNode CozeNode) (models.LLMConfig, error) {
	config := models.LLMConfig{}

	if cozeNode.Data.Inputs.LLMParam == nil {
		return config, fmt.Errorf("llmParam is required for LLM node")
	}

	llmParams := p.extractLLMParams(cozeNode.Data.Inputs.LLMParam)

	// Parse model configuration
	config.Model = models.ModelConfig{
		Provider: p.getStringParam(llmParams, "modleName", ""),
		Name:     p.getStringParam(llmParams, "modleName", ""),
		Mode:     "chat",
	}

	// Parse model parameters
	config.Parameters = models.ModelParameters{
		Temperature:    p.getFloatParam(llmParams, "temperature", 0.8),
		MaxTokens:      p.getIntParam(llmParams, "maxTokens", 4096),
		TopP:           p.getFloatParam(llmParams, "topP", 0.7),
		ResponseFormat: p.getIntParam(llmParams, "responseFormat", 0), // 0=text, 2=json
	}

	// Parse prompt configuration
	config.Prompt = models.PromptConfig{
		SystemTemplate: p.getStringParam(llmParams, "systemPrompt", ""),
		UserTemplate:   p.getStringParam(llmParams, "prompt", ""),
	}

	return config, nil
}

// extractLLMParams converts LLMParam to map for easier access
// Supports both array format (main layer) and object format (iteration subnodes)
func (p *LLMNodeParser) extractLLMParams(llmParam interface{}) map[string]interface{} {
	params := make(map[string]interface{})

	// Handle object format (already converted by iteration parser)
	if paramsMap, ok := llmParam.(map[string]interface{}); ok {
		return paramsMap
	}

	// Handle array format (main layer nodes)
	if llmParamArray, ok := llmParam.([]interface{}); ok {
		for _, param := range llmParamArray {
			if paramMap, ok := param.(map[string]interface{}); ok {
				if name, hasName := paramMap["name"].(string); hasName {
					if input, hasInput := paramMap["input"].(map[string]interface{}); hasInput {
						if value, hasValue := input["value"].(map[string]interface{}); hasValue {
							if content, hasContent := value["content"]; hasContent {
								params[name] = content
							}
						}
					}
				}
			}
		}
	}

	return params
}

// Helper methods to extract typed parameters
func (p *LLMNodeParser) getStringParam(params map[string]interface{}, key string, defaultValue string) string {
	if value, exists := params[key]; exists {
		if strVal, ok := value.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func (p *LLMNodeParser) getFloatParam(params map[string]interface{}, key string, defaultValue float64) float64 {
	if value, exists := params[key]; exists {
		switch v := value.(type) {
		case string:
			if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
				return floatVal
			}
		case float64:
			return v
		case int:
			return float64(v)
		}
	}
	return defaultValue
}

func (p *LLMNodeParser) getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if value, exists := params[key]; exists {
		switch v := value.(type) {
		case string:
			if intVal, err := strconv.Atoi(v); err == nil {
				return intVal
			}
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

// parseNodeOutputs processes node outputs, filtering out reasoning_content
func (p *LLMNodeParser) parseNodeOutputs(cozeNode CozeNode) []models.Output {
	var outputs []models.Output

	if cozeNode.Data.Outputs != nil {
		for _, output := range cozeNode.Data.Outputs {
			outputName := output.Name

			// Skip reasoning_content output as per requirements
			if strings.ToLower(outputName) == "reasoning_content" {
				continue
			}

			// Map output type
			outputType := p.mapOutputType(output.Type)

			outputs = append(outputs, models.Output{
				Name:        outputName,
				Type:        outputType,
				Description: "",
				Required:    true,
			})
		}
	}

	// Ensure at least one output exists (default to "output" if none found)
	if len(outputs) == 0 {
		outputs = append(outputs, models.Output{
			Name:        "output",
			Type:        models.DataTypeString,
			Description: "LLM generated output",
			Required:    true,
		})
	}

	return outputs
}

// mapOutputType maps Coze output types to unified DSL types
func (p *LLMNodeParser) mapOutputType(cozeType string) models.UnifiedDataType {
	switch strings.ToLower(cozeType) {
	case "string", "text":
		return models.DataTypeString
	case "integer", "int":
		return models.DataTypeInteger
	case "number", "float":
		return models.DataTypeFloat
	case "boolean", "bool":
		return models.DataTypeBoolean
	case "array", "list":
		return models.DataTypeArrayString // Default to array of strings
	case "object":
		return models.DataTypeObject
	default:
		return models.DataTypeString // Default to string if type is unknown
	}
}
