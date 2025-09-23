package generator

import (
	"agentbridge/internal/models"
	"agentbridge/platforms/common"
	"fmt"
)

// LLMNodeGenerator generates Coze LLM nodes
type LLMNodeGenerator struct {
	idGenerator   *CozeIDGenerator
	isInIteration bool // Track if generating for iteration blocks
}

// NewLLMNodeGenerator creates an LLM node generator
func NewLLMNodeGenerator() *LLMNodeGenerator {
	return &LLMNodeGenerator{
		idGenerator: nil, // Set by the main generator
	}
}

// SetIDGenerator sets the shared ID generator
func (g *LLMNodeGenerator) SetIDGenerator(idGenerator *CozeIDGenerator) {
	g.idGenerator = idGenerator
}

// SetIterationContext sets the iteration context for generating blocks format
func (g *LLMNodeGenerator) SetIterationContext(isInIteration bool) {
	g.isInIteration = isInIteration
}

// GenerateNode generates a Coze workflow LLM node
func (g *LLMNodeGenerator) GenerateNode(unifiedNode *models.Node) (*CozeNode, error) {
	if err := g.ValidateNode(unifiedNode); err != nil {
		return nil, err
	}

	cozeNodeID := g.idGenerator.MapToCozeNodeID(unifiedNode.ID)

	// Generate input parameters from unified node inputs
	inputParams := g.generateInputParameters(unifiedNode)

	// Generate LLM parameters from unified node config
	llmParams, err := g.generateLLMParameters(unifiedNode)
	if err != nil {
		return nil, fmt.Errorf("failed to generate LLM parameters: %w", err)
	}

	// Generate error handling settings
	errorSettings := g.generateErrorSettings(unifiedNode)

	// Creates LLM node input structure with only essential fields (matching Coze iteration block format)
	llmInputs := map[string]interface{}{
		"inputParameters": inputParams,   // ✅ Essential: input parameters with correct camelCase
		"llmParam":        llmParams,     // ✅ Essential: LLM configuration parameters with correct camelCase
		"settingOnError":  errorSettings, // ✅ Essential: error handling settings with correct camelCase
	}

	return &CozeNode{
		ID:   cozeNodeID,
		Type: "3", // Coze LLM node type
		Meta: &CozeNodeMeta{
			Position: &CozePosition{
				X: unifiedNode.Position.X,
				Y: unifiedNode.Position.Y,
			},
		},
		Data: &CozeNodeData{
			Meta: &CozeNodeMetaInfo{
				Title:       g.getNodeTitle(unifiedNode),
				Description: g.getNodeDescription(unifiedNode),
				Icon:        g.getNodeIcon(unifiedNode),
				Subtitle:    "大模型",
				MainColor:   "#5C62FF",
			},
			Outputs: g.generateOutputs(unifiedNode),
			Inputs:  llmInputs,
		},
		Version: "3",
		Blocks:  []interface{}{},
		Edges:   []interface{}{},
	}, nil
}

// GenerateSchemaNode generates schema node for LLM
func (g *LLMNodeGenerator) GenerateSchemaNode(unifiedNode *models.Node) (*CozeSchemaNode, error) {
	if err := g.ValidateNode(unifiedNode); err != nil {
		return nil, err
	}

	cozeNodeID := g.idGenerator.MapToCozeNodeID(unifiedNode.ID)

	// Generate LLM parameters from unified node config - CRITICAL for Coze import
	llmParams, err := g.generateLLMParameters(unifiedNode)
	if err != nil {
		return nil, fmt.Errorf("failed to generate LLM parameters for schema: %w", err)
	}

	// Generate error handling settings - CRITICAL for Coze import
	errorSettings := g.generateSchemaErrorSettings(unifiedNode)

	// Creates simplified input parameter structure for schema section
	schemaInputParams := make([]CozeInputParameter, 0)

	for _, input := range unifiedNode.Inputs {
		cozeType := g.mapDataTypeToCozeType(input.Type)

		if input.Reference != nil && input.Reference.Type == models.ReferenceTypeNodeOutput {
			// Variable reference input for schema - schema uses rawMeta with uppercase M
			schemaInput := CozeInputParameter{
				Name: input.Name,
				Input: &CozeInputValue{
					Type: cozeType,
					Value: &CozeInputRef{
						Type: "ref",
						Content: &CozeRefContent{
							BlockID: g.idGenerator.MapToCozeNodeID(input.Reference.NodeID),
							Name:    input.Reference.OutputName,
							Source:  "block-output",
						},
						RawMeta: &CozeRawMeta{ // Schema section uses RawMeta with uppercase M
							Type: g.mapDataTypeToRawMetaType(input.Type),
						},
					},
				},
			}
			schemaInputParams = append(schemaInputParams, schemaInput)
		}
	}

	// Creates complete input structure for schema node with all required configurations
	schemaInputs := &CozeSchemaNodeInputs{
		InputParameters: schemaInputParams,
		LLMParam:        llmParams,     // Adds essential LLM parameter configuration for schema
		SettingOnError:  errorSettings, // Adds error handling configuration for schema
	}

	return &CozeSchemaNode{
		Data: &CozeSchemaNodeData{
			NodeMeta: &CozeNodeMetaInfo{
				Title:       g.getNodeTitle(unifiedNode),
				Description: g.getNodeDescription(unifiedNode),
				Icon:        g.getNodeIcon(unifiedNode),
				SubTitle:    "大模型", // Uses correct field name for subtitle
				MainColor:   "#5C62FF",
			},
			Inputs:  schemaInputs,
			Outputs: g.generateSchemaOutputs(unifiedNode),
			Version: "3", // Specifies version field for compatibility
		},
		ID:   cozeNodeID,
		Type: "3",
		Meta: &CozeNodeMeta{
			Position: &CozePosition{
				X: unifiedNode.Position.X,
				Y: unifiedNode.Position.Y,
			},
		},
	}, nil
}

// generateInputParameters generates input parameters from unified node inputs
// Creates properly formatted input parameters for nodes section with required fields
func (g *LLMNodeGenerator) generateInputParameters(unifiedNode *models.Node) []interface{} {
	var inputParams []interface{}

	for _, input := range unifiedNode.Inputs {
		cozeType := g.mapDataTypeToCozeType(input.Type)

		if input.Reference != nil && input.Reference.Type == models.ReferenceTypeNodeOutput {
			// Creates complete variable reference structure for nodes section
			inputParam := map[string]interface{}{
				"name": input.Name,
				"input": map[string]interface{}{
					"type": cozeType,
					"value": map[string]interface{}{
						"type": "ref",
						"content": map[string]interface{}{
							"blockID": g.idGenerator.MapToCozeNodeID(input.Reference.NodeID),
							"name":    g.mapOutputFieldNameForCoze(input.Reference.NodeID, input.Reference.OutputName),
							"source":  "block-output",
						},
						"rawMeta": map[string]interface{}{ // Uses camelCase rawMeta for nodes section
							"type": g.mapDataTypeToRawMetaType(input.Type),
						},
					},
				},
				"left":      nil,
				"right":     nil,
				"variables": []interface{}{},
			}
			inputParams = append(inputParams, inputParam)
		} else {
			// Handles literal value inputs when available
			inputParam := map[string]interface{}{
				"name": input.Name,
				"input": map[string]interface{}{
					"Type":  cozeType,
					"Value": input.Default, // Uses default value from unified DSL
				},
				"left":      nil,
				"right":     nil,
				"variables": []interface{}{},
			}
			inputParams = append(inputParams, inputParam)
		}
	}

	return inputParams
}

// generateLLMParameters generates LLM-specific parameters from unified node config
func (g *LLMNodeGenerator) generateLLMParameters(unifiedNode *models.Node) ([]map[string]interface{}, error) {
	llmConfig, ok := common.AsLLMConfig(unifiedNode.Config)
	if !ok || llmConfig == nil {
		return nil, fmt.Errorf("invalid LLM config type for node %s", unifiedNode.ID)
	}

	var llmParams []map[string]interface{}

	// Model Type uses fixed Doubao model configuration
	llmParams = append(llmParams, map[string]interface{}{
		"name": "modelType",
		"input": map[string]interface{}{
			"type": "integer",
			"value": map[string]interface{}{
				"content": "61010",
				"rawMeta": map[string]interface{}{
					"type": 2,
				},
				"type": "literal",
			},
		},
	})

	// Model Name maps Spark domain to Coze model names
	modelName := g.mapSparkDomainToCozeModel(*llmConfig)
	llmParams = append(llmParams, map[string]interface{}{
		"name": "modleName",
		"input": map[string]interface{}{
			"type": "string",
			"value": map[string]interface{}{
				"content": modelName,
				"rawMeta": map[string]interface{}{
					"type": 1,
				},
				"type": "literal",
			},
		},
	})

	// Generation Diversity
	llmParams = append(llmParams, map[string]interface{}{
		"name": "generationDiversity",
		"input": map[string]interface{}{
			"type": "string",
			"value": map[string]interface{}{
				"content": "balance",
				"rawMeta": map[string]interface{}{
					"type": 1,
				},
				"type": "literal",
			},
		},
	})

	// Temperature
	temperature := llmConfig.Parameters.Temperature
	if temperature == 0 {
		temperature = 0.8 // Default temperature value
	}
	llmParams = append(llmParams, map[string]interface{}{
		"name": "temperature",
		"input": map[string]interface{}{
			"type": "float",
			"value": map[string]interface{}{
				"content": fmt.Sprintf("%.1f", temperature),
				"rawMeta": map[string]interface{}{
					"type": 4,
				},
				"type": "literal",
			},
		},
	})

	// Max Tokens
	maxTokens := llmConfig.Parameters.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096 // Default max tokens value
	}
	llmParams = append(llmParams, map[string]interface{}{
		"name": "maxTokens",
		"input": map[string]interface{}{
			"type": "integer",
			"value": map[string]interface{}{
				"content": fmt.Sprintf("%d", maxTokens),
				"rawMeta": map[string]interface{}{
					"type": 2,
				},
				"type": "literal",
			},
		},
	})

	// Top P
	topP := llmConfig.Parameters.TopP
	if topP == 0 {
		topP = 0.7 // Default top P value
	}
	llmParams = append(llmParams, map[string]interface{}{
		"name": "topP",
		"input": map[string]interface{}{
			"type": "float",
			"value": map[string]interface{}{
				"content": fmt.Sprintf("%.1f", topP),
				"rawMeta": map[string]interface{}{
					"type": 4,
				},
				"type": "literal",
			},
		},
	})

	// Response Format
	llmParams = append(llmParams, map[string]interface{}{
		"name": "responseFormat",
		"input": map[string]interface{}{
			"type": "integer",
			"value": map[string]interface{}{
				"content": "2",
				"rawMeta": map[string]interface{}{
					"type": 2,
				},
				"type": "literal",
			},
		},
	})

	// Prompt corresponds to UserTemplate in iFlytek DSL
	// Empty UserTemplate indicates iFlytek had placeholder value, map to corresponding Coze format
	prompt := llmConfig.Prompt.UserTemplate
	if prompt == "" {
		prompt = "无" // Map empty UserTemplate to placeholder value for Coze compatibility
	}
	// Converts English commas to Chinese commas for Coze format compatibility
	prompt = g.convertCommasToChineseFormat(prompt)

	llmParams = append(llmParams, map[string]interface{}{
		"name": "prompt",
		"input": map[string]interface{}{
			"type": "string",
			"value": map[string]interface{}{
				"content": prompt,
				"rawMeta": map[string]interface{}{
					"type": 1,
				},
				"type": "literal",
			},
		},
	})

	// Enable Chat History
	llmParams = append(llmParams, map[string]interface{}{
		"name": "enableChatHistory",
		"input": map[string]interface{}{
			"type": "boolean",
			"value": map[string]interface{}{
				"content": false,
				"rawMeta": map[string]interface{}{
					"type": 3,
				},
				"type": "literal",
			},
		},
	})

	// Chat History Round
	llmParams = append(llmParams, map[string]interface{}{
		"name": "chatHistoryRound",
		"input": map[string]interface{}{
			"type": "integer",
			"value": map[string]interface{}{
				"content": "3",
				"rawMeta": map[string]interface{}{
					"type": 2,
				},
				"type": "literal",
			},
		},
	})

	// System Prompt gets template from SystemTemplate
	systemPrompt := llmConfig.Prompt.SystemTemplate
	if systemPrompt == "" {
		systemPrompt = "你是一个有用的AI助手"
	}
	llmParams = append(llmParams, map[string]interface{}{
		"name": "systemPrompt",
		"input": map[string]interface{}{
			"type": "string",
			"value": map[string]interface{}{
				"content": systemPrompt,
				"rawMeta": map[string]interface{}{
					"type": 1,
				},
				"type": "literal",
			},
		},
	})

	return llmParams, nil
}

// mapOutputFieldNameForCoze maps output field names from unified DSL to Coze platform format
func (g *LLMNodeGenerator) mapOutputFieldNameForCoze(nodeID, outputName string) string {

	// Map classifier output names from iFlytek format to Coze format
	if outputName == "class_name" {
		// iFlytek classifier outputs "class_name", but Coze uses "classificationId"
		return "classificationId"
	}

	// Default: return original name if no mapping needed
	return outputName
}

// generateErrorSettings generates error handling settings
func (g *LLMNodeGenerator) generateErrorSettings(unifiedNode *models.Node) map[string]interface{} {
	return map[string]interface{}{
		"processType": 1,
		"retryTimes":  0,
		"timeoutMs":   180000,
	}
}

// generateSchemaErrorSettings generates error handling settings for schema node
func (g *LLMNodeGenerator) generateSchemaErrorSettings(unifiedNode *models.Node) map[string]interface{} {
	return map[string]interface{}{
		"processType": 1,      // Schema section uses camelCase naming convention
		"retryTimes":  0,      // Schema section uses camelCase naming convention
		"timeoutMs":   180000, // Schema section uses camelCase naming convention
	}
}

// generateOutputs generates outputs for LLM node
func (g *LLMNodeGenerator) generateOutputs(unifiedNode *models.Node) []CozeNodeOutput {
	var outputs []CozeNodeOutput

	// Generate outputs completely based on unified DSL definition - NO hardcoded defaults
	for _, output := range unifiedNode.Outputs {
		outputs = append(outputs, CozeNodeOutput{
			Name:     output.Name,
			Type:     g.mapDataTypeToCozeType(output.Type),
			Required: output.Required,
		})
	}

	return outputs
}

// generateSchemaOutputs generates outputs for schema node
func (g *LLMNodeGenerator) generateSchemaOutputs(unifiedNode *models.Node) []CozeNodeOutput {
	return g.generateOutputs(unifiedNode)
}

// ValidateNode validates the unified node for LLM generation
func (g *LLMNodeGenerator) ValidateNode(unifiedNode *models.Node) error {
	if unifiedNode == nil {
		return fmt.Errorf("unified node is nil")
	}

	if unifiedNode.Type != models.NodeTypeLLM {
		return fmt.Errorf("invalid node type: expected %s, got %s", models.NodeTypeLLM, unifiedNode.Type)
	}

	if g.idGenerator == nil {
		return fmt.Errorf("ID generator not set")
	}

	// Validates configuration type
	if cfg, ok := common.AsLLMConfig(unifiedNode.Config); !ok || cfg == nil {
		return fmt.Errorf("invalid config type for LLM node")
	}

	return nil
}

// GetNodeType returns the node type for this generator
func (g *LLMNodeGenerator) GetNodeType() models.NodeType {
	return models.NodeTypeLLM
}

// Helper methods for node metadata

func (g *LLMNodeGenerator) getNodeTitle(unifiedNode *models.Node) string {
	if unifiedNode.Title != "" {
		return unifiedNode.Title
	}
	return "大模型"
}

func (g *LLMNodeGenerator) getNodeDescription(unifiedNode *models.Node) string {
	if unifiedNode.Description != "" {
		return unifiedNode.Description
	}
	return "调用大语言模型,使用变量和提示词生成回复"
}

func (g *LLMNodeGenerator) getNodeIcon(unifiedNode *models.Node) string {
	return "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-LLM-v2.jpg"
}

// Helper methods for data type mapping

func (g *LLMNodeGenerator) mapDataTypeToCozeType(dataType models.UnifiedDataType) string {
	switch dataType {
	case models.DataTypeString:
		return "string"
	case models.DataTypeInteger:
		return "integer"
	case models.DataTypeFloat:
		return "float"
	case models.DataTypeNumber:
		return "float" // Maps generic number to float in Coze
	case models.DataTypeBoolean:
		return "boolean"
	default:
		return "string"
	}
}

func (g *LLMNodeGenerator) mapDataTypeToRawMetaType(dataType models.UnifiedDataType) int {
	switch dataType {
	case models.DataTypeString:
		return 1
	case models.DataTypeInteger:
		return 2
	case models.DataTypeBoolean:
		return 3
	case models.DataTypeFloat, models.DataTypeNumber:
		return 4
	default:
		return 1
	}
}

// mapSparkDomainToCozeModel maps Spark domain to Coze model name
func (g *LLMNodeGenerator) mapSparkDomainToCozeModel(llmConfig models.LLMConfig) string {
	// Creates mapping table from Spark domain to Coze model names
	domainToModelMap := map[string]string{
		"4.0Ultra":    "Doubao-Seed-1.6",
		"generalv3.5": "Doubao-Seed-1.6",
		"generalv3":   "Doubao-Seed-1.6",
		"generalv2":   "Doubao-Seed-1.6",
		"general":     "Doubao-Seed-1.6",
		"spark":       "Doubao-Seed-1.6", // Default Spark model mapping
	}

	// Gets model information from LLM configuration
	modelKey := ""
	if llmConfig.Model.Name != "" {
		modelKey = llmConfig.Model.Name
	} else if llmConfig.Model.Provider != "" {
		modelKey = llmConfig.Model.Provider
	}

	// Returns mapped model name or default if mapping not found
	if modelName, exists := domainToModelMap[modelKey]; exists {
		return modelName
	}

	return "Doubao-Seed-1.6" // Default model fallback
}

// convertCommasToChineseFormat converts English commas to Chinese commas in template variables
// Transforms variable template comma format for Coze platform compatibility
func (g *LLMNodeGenerator) convertCommasToChineseFormat(text string) string {
	// Example: {{name}},{{birth_year}} → {{name}}，{{birth_year}}
	result := ""
	inBraces := 0

	for i, char := range text {
		if char == '{' {
			inBraces++
		} else if char == '}' {
			inBraces--
		} else if char == ',' && inBraces == 0 {
			// Converts English commas outside variables to Chinese commas
			result += "，"
			continue
		} else if char == ',' && inBraces > 0 && i+1 < len(text) && text[i+1] == '{' {
			// Converts English commas between variables to Chinese commas
			result += "，"
			continue
		}
		result += string(char)
	}

	return result
}
