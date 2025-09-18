package generator

import (
	"ai-agents-transformer/internal/models"
	"fmt"
	"strings"
)

// ClassifierNodeGenerator generates Coze Intent Recognition nodes from unified DSL classifier configurations.
type ClassifierNodeGenerator struct {
	idGenerator *CozeIDGenerator
}

// NewClassifierNodeGenerator creates a classifier node generator for Coze platform.
func NewClassifierNodeGenerator() *ClassifierNodeGenerator {
	return &ClassifierNodeGenerator{
		idGenerator: nil, // Set by the main generator
	}
}

// SetIDGenerator configures the shared ID generator instance.
func (g *ClassifierNodeGenerator) SetIDGenerator(idGenerator *CozeIDGenerator) {
	g.idGenerator = idGenerator
}

// GetSupportedNodeType returns the supported node type.
func (g *ClassifierNodeGenerator) GetSupportedNodeType() models.NodeType {
	return models.NodeTypeClassifier
}

// GetNodeType returns the node type for interface compatibility.
func (g *ClassifierNodeGenerator) GetNodeType() models.NodeType {
	return models.NodeTypeClassifier
}

// ValidateNode validates the classifier node configuration.
func (g *ClassifierNodeGenerator) ValidateNode(node *models.Node) error {
	if node.Type != models.NodeTypeClassifier {
		return fmt.Errorf("expected classifier node, got %s", node.Type)
	}
	if _, ok := node.Config.(*models.ClassifierConfig); !ok {
		return fmt.Errorf("invalid classifier config type for node %s, got %T, expected *models.ClassifierConfig", node.ID, node.Config)
	}
	return nil
}

// GenerateNode generates Coze Intent Recognition node from unified classifier configuration.
func (g *ClassifierNodeGenerator) GenerateNode(unifiedNode *models.Node) (*CozeNode, error) {
	if err := g.ValidateNode(unifiedNode); err != nil {
		return nil, fmt.Errorf("validation failed for classifier node %s: %w", unifiedNode.ID, err)
	}

	// Parse classifier configuration
	classifierConfig, ok := unifiedNode.Config.(*models.ClassifierConfig)
	if !ok {
		return nil, fmt.Errorf("invalid classifier config type for node %s, got %T", unifiedNode.ID, unifiedNode.Config)
	}

	// Check if ID generator is set
	if g.idGenerator == nil {
		return nil, fmt.Errorf("ID generator not set for ClassifierNodeGenerator")
	}

	cozeNodeID := g.idGenerator.MapToCozeNodeID(unifiedNode.ID)

	// Generate input parameters from unified node inputs
	inputParams := g.generateInputParameters(unifiedNode)
	
	// Generate intents from classifier classes
	intents := g.generateIntents(classifierConfig)
	
	// Generate LLM parameters for intent recognition
	llmParam := g.generateLLMParam(classifierConfig)
	
	// Generate chat history setting
	chatHistorySetting := g.generateChatHistorySetting(classifierConfig)
	
	// Generate error handling settings
	errorSettings := g.generateErrorSettings()

	// Create intent recognition inputs structure
	intentInputs := map[string]interface{}{
		"chatHistorySetting": chatHistorySetting,
		"inputParameters":    inputParams,
		"intents":           intents,
		"llmParam":          llmParam,
		"mode":              "all", // Default mode for intent recognition
		"settingOnError":    errorSettings,
	}

	// Generate outputs based on unified DSL definition - NO hardcoded defaults
	outputs := g.generateOutputs(unifiedNode)

	// Create the complete Coze node
	cozeNode := &CozeNode{
		ID:   cozeNodeID,
		Type: "22", // Coze Intent Recognition node type
		Meta: &CozeNodeMeta{
			Position: &CozePosition{
				X: unifiedNode.Position.X,
				Y: unifiedNode.Position.Y,
			},
		},
		Data: &CozeNodeData{
			Meta: &CozeNodeMetaInfo{
				Title:       unifiedNode.Title,
				Description: "用于用户输入的意图识别，并将其与预设意图选项进行匹配。",
				Icon:        "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-Intent-v2.jpg",
				SubTitle:    "意图识别",
				MainColor:   "#00B2B2",
			},
			Outputs: outputs,
			Inputs:  intentInputs,
		},
		Blocks:  []interface{}{},
		Edges:   []interface{}{},
		Version: "",
		Size:    nil,
	}

	return cozeNode, nil
}

// GenerateSchemaNode generates schema node for Coze classifier workflow definition.
func (g *ClassifierNodeGenerator) GenerateSchemaNode(unifiedNode *models.Node) (*CozeSchemaNode, error) {
	if err := g.ValidateNode(unifiedNode); err != nil {
		return nil, fmt.Errorf("validation failed for classifier schema node %s: %w", unifiedNode.ID, err)
	}

	// Parse classifier configuration with detailed error reporting
	classifierConfig, ok := unifiedNode.Config.(*models.ClassifierConfig)
	if !ok {
		return nil, fmt.Errorf("invalid classifier config type for schema node %s, got %T, expected *models.ClassifierConfig", unifiedNode.ID, unifiedNode.Config)
	}

	cozeNodeID := g.idGenerator.MapToCozeNodeID(unifiedNode.ID)

	// Generate input parameters for schema (simplified structure)
	schemaInputParams := []map[string]interface{}{}
	
	for _, input := range unifiedNode.Inputs {
		if input.Reference != nil {
			param := map[string]interface{}{
				"name": "query", 
				"input": map[string]interface{}{
					"type": g.convertDataType(input.Type),
					"value": g.convertVariableReference(*input.Reference),
				},
			}
			schemaInputParams = append(schemaInputParams, param)
			break // Use only the first input as query parameter
		}
	}

	// Generate intents for schema
	schemaIntents := g.generateIntents(classifierConfig)
	
	// Generate LLM parameters for schema
	schemaLLMParam := g.generateLLMParam(classifierConfig)
	
	// Generate chat history setting for schema
	schemaChatHistorySetting := g.generateChatHistorySetting(classifierConfig)
	
	// Generate error settings for schema
	schemaErrorSettings := g.generateErrorSettings()

	// Create schema inputs structure
	schemaInputs := map[string]interface{}{
		"inputParameters":    schemaInputParams,
		"chatHistorySetting": schemaChatHistorySetting,
		"intents":           schemaIntents,
		"llmParam":          schemaLLMParam,
		"mode":              "all",
		"settingOnError":    schemaErrorSettings,
	}

	// Generate outputs for schema based on unified DSL definition - NO hardcoded defaults
	schemaOutputs := g.generateOutputs(unifiedNode)

	// Create schema node
	schemaNode := &CozeSchemaNode{
		ID:   cozeNodeID,
		Type: "22", // Intent Recognition node type
		Meta: &CozeNodeMeta{
			Position: &CozePosition{
				X: unifiedNode.Position.X,
				Y: unifiedNode.Position.Y,
			},
		},
		Data: &CozeSchemaNodeData{
			NodeMeta: &CozeNodeMetaInfo{
				Title:       unifiedNode.Title,
				Description: "用于用户输入的意图识别，并将其与预设意图选项进行匹配。",
				Icon:        "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-Intent-v2.jpg",
				SubTitle:    "意图识别",
				MainColor:   "#00B2B2",
			},
			Outputs: schemaOutputs,
			Inputs:  schemaInputs,
		},
	}

	return schemaNode, nil
}

// generateInputParameters converts unified node inputs to Coze input parameter format.
func (g *ClassifierNodeGenerator) generateInputParameters(unifiedNode *models.Node) []map[string]interface{} {
	params := []map[string]interface{}{}
	
	for _, input := range unifiedNode.Inputs {
		if input.Reference != nil {
			param := map[string]interface{}{
				"name": "query", // Use fixed query parameter name for Coze format compatibility
				"input": map[string]interface{}{
					"type": g.convertDataType(input.Type),
					"value": g.convertVariableReference(*input.Reference),
				},
			}
			params = append(params, param)
			break // Use only the first input as query parameter
		}
	}
	
	return params
}

// generateIntents converts classifier classes to Coze intent format.
func (g *ClassifierNodeGenerator) generateIntents(config *models.ClassifierConfig) []map[string]interface{} {
	intents := []map[string]interface{}{}
	
	for _, class := range config.Classes {

		if (class.Name == "default" && strings.Contains(class.Description, "默认")) ||
		   strings.Contains(class.Description, "默认意图") {
			continue // Skip the default intent
		}
		
		intent := map[string]interface{}{
			"name": class.Name,
		}
		intents = append(intents, intent)
	}
	
	return intents
}

// generateLLMParam converts classifier model configuration to Coze LLM parameters.
func (g *ClassifierNodeGenerator) generateLLMParam(config *models.ClassifierConfig) map[string]interface{} {
	return map[string]interface{}{
		"chatHistoryRound":    3,                              // Default value
		"enableChatHistory":   false,                          // Default value
		"generationDiversity": "balance",                      // Default value
		"maxTokens":          config.Parameters.MaxTokens,
		"modelName":          config.Model.Name,
		"modelType":          g.convertModelType(config.Model.Provider),
		"prompt": map[string]interface{}{
			"type": "string",
			"value": map[string]interface{}{
				"content": g.generatePromptContent(config),
				"type":    "literal",
			},
		},
		"responseFormat": 2, // Default response format
		"systemPrompt": map[string]interface{}{
			"type": "string",
			"value": map[string]interface{}{
				"content": g.normalizeSystemPromptForCoze(config.Instructions),
				"type":    "literal",
			},
		},
		"temperature": config.Parameters.Temperature,
		"topP":        config.Parameters.TopP,
	}
}

// generateChatHistorySetting creates chat history configuration for Coze classifier.
func (g *ClassifierNodeGenerator) generateChatHistorySetting(config *models.ClassifierConfig) map[string]interface{} {
	return map[string]interface{}{
		"enableChatHistory": false, // Default value
		"chatHistoryRound":  3,     // Default value
	}
}

// generateErrorSettings creates error handling configuration for Coze classifier nodes.
func (g *ClassifierNodeGenerator) generateErrorSettings() map[string]interface{} {
	return map[string]interface{}{
		"processType": 1,
		"retryTimes":  0,
		"timeoutMs":   60000,
	}
}

// generatePromptContent creates prompt template content using classifier query variable.
func (g *ClassifierNodeGenerator) generatePromptContent(config *models.ClassifierConfig) string {
	// Use fixed {{query}} template for Coze format consistency
	return "{{query}}"
}

// normalizeSystemPromptForCoze converts iFlytek systemPrompt format to Coze compatible format
// Main conversion: {{Query}} to {{query}}
func (g *ClassifierNodeGenerator) normalizeSystemPromptForCoze(instructions string) string {
	if instructions == "" {
		return ""
	}
	
	// Convert iFlytek {{Query}} (uppercase) to Coze standard {{query}} (lowercase)
	normalized := strings.ReplaceAll(instructions, "{{Query}}", "{{query}}")
	
	// Handle other possible variants
	normalized = strings.ReplaceAll(normalized, "{{QUERY}}", "{{query}}")
	
	return normalized
}

// convertModelType maps unified model provider to Coze model type identifier.
func (g *ClassifierNodeGenerator) convertModelType(provider string) int {
	// Map common model providers to Coze model types
	switch provider {
	case "doubao", "Doubao":
		return 61010
	case "gpt-4", "openai":
		return 61001
	case "claude":
		return 61002
	default:
		return 61010 // Default to Doubao
	}
}

// convertDataType maps unified data types to Coze data type strings.
func (g *ClassifierNodeGenerator) convertDataType(dataType models.UnifiedDataType) string {
	switch dataType {
	case models.DataTypeString:
		return "string"
	case models.DataTypeInteger:
		return "integer"
	case models.DataTypeFloat:
		return "float"
	case models.DataTypeBoolean:
		return "boolean"
	case models.DataTypeArrayString:
		return "array"
	case models.DataTypeArrayObject:
		return "array"
	case models.DataTypeObject:
		return "object"
	default:
		return "string" // Default to string
	}
}

// convertVariableReference transforms unified variable references to Coze input value format.
func (g *ClassifierNodeGenerator) convertVariableReference(ref models.VariableReference) map[string]interface{} {
	switch ref.Type {
	case models.ReferenceTypeNodeOutput:
		return map[string]interface{}{
			"content": map[string]interface{}{
				"blockID": g.idGenerator.MapToCozeNodeID(ref.NodeID),
				"name":    g.mapOutputFieldNameForCoze(ref.NodeID, ref.OutputName),
				"source":  "block-output",
			},
			"rawMeta": map[string]interface{}{
				"type": 1, // Reference type
			},
			"type": "ref",
		}
	case models.ReferenceTypeLiteral:
		return map[string]interface{}{
			"content": ref.Value,
			"rawMeta": map[string]interface{}{
				"type": g.getLiteralType(ref.Value),
			},
			"type": "literal",
		}
	case models.ReferenceTypeTemplate:
		return map[string]interface{}{
			"content": ref.Template,
			"rawMeta": map[string]interface{}{
				"type": 1, // String type for templates
			},
			"type": "literal",
		}
	default:
		return map[string]interface{}{
			"content": "",
			"type":    "literal",
		}
	}
}

// getLiteralType determines the Coze type identifier for literal values.
func (g *ClassifierNodeGenerator) getLiteralType(value interface{}) int {
	switch value.(type) {
	case string:
		return 1
	case int, int32, int64:
		return 2
	case bool:
		return 3
	case float32, float64:
		return 4
	default:
		return 1 // Default to string
	}
}

// generateOutputs generates standard Coze classifier outputs
func (g *ClassifierNodeGenerator) generateOutputs(unifiedNode *models.Node) []CozeNodeOutput {
	// Coze intent recognition always has these two standard outputs
	outputs := []CozeNodeOutput{
		{
			Name: "classificationId",
			Type: "integer",
		},
		{
			Name: "reason",
			Type: "string",
		},
	}
	return outputs
}

// mapOutputFieldNameForCoze maps output field names from unified DSL to Coze platform format
func (g *ClassifierNodeGenerator) mapOutputFieldNameForCoze(nodeID, outputName string) string {
	// Map classifier output names from iFlytek format to Coze format
	if outputName == "class_name" {
		// iFlytek classifier outputs "class_name", but Coze uses "classificationId"
		return "classificationId"
	}
	
	// Add more mappings as needed for other node types
	// Example: if outputName == "some_other_field" { return "mappedField" }
	
	// Default: return original name if no mapping needed
	return outputName
}