package generator

import (
	"ai-agents-transformer/internal/models"
	"ai-agents-transformer/platforms/common"
	"fmt"
	"regexp"
	"strings"
)

// LLMNodeGenerator LLM node generator
type LLMNodeGenerator struct {
	*BaseNodeGenerator
	idMapping        map[string]string // Dify ID -> iFlytek SparkAgent ID mapping
	nodeTitleMapping map[string]string // iFlytek SparkAgent ID -> node title mapping
}

func NewLLMNodeGenerator() *LLMNodeGenerator {
	return &LLMNodeGenerator{
		BaseNodeGenerator: NewBaseNodeGenerator(models.NodeTypeLLM),
		idMapping:         make(map[string]string),
		nodeTitleMapping:  make(map[string]string),
	}
}

// SetIDMapping sets ID mapping
func (g *LLMNodeGenerator) SetIDMapping(idMapping map[string]string) {
	g.idMapping = idMapping
}

// SetNodeTitleMapping sets node title mapping
func (g *LLMNodeGenerator) SetNodeTitleMapping(nodeTitleMapping map[string]string) {
	g.nodeTitleMapping = nodeTitleMapping
}

// GetSupportedNodeType gets supported node type
func (g *LLMNodeGenerator) GetSupportedNodeType() models.NodeType {
	return models.NodeTypeLLM
}

// GenerateNode generates iFlytek SparkAgent LLM node
func (g *LLMNodeGenerator) GenerateNode(node models.Node) (IFlytekNode, error) {
	// Validate node
	if err := g.ValidateNode(node); err != nil {
		return IFlytekNode{}, err
	}

	// Generate basic information
	iflytekNode := g.generateBasicNodeInfo(node)
	iflytekNode.Type = "大模型"
	iflytekNode.Data.Label = node.Title
	iflytekNode.Data.Description = "根据输入的提示词，调用选定的大模型，对提示词作出回答"
	iflytekNode.Data.NodeMeta.AliasName = "大模型"
	iflytekNode.Data.NodeMeta.NodeType = "基础节点"

	// LLM node special configuration
	iflytekNode.Data.AllowInputReference = true
	iflytekNode.Data.AllowOutputReference = true
	iflytekNode.Data.Icon = g.getNodeIcon(models.NodeTypeLLM)

	// Parse LLM configuration
	if llmConfig, ok := common.AsLLMConfig(node.Config); ok && llmConfig != nil {
		// Generate nodeParam configuration
		nodeParam := g.generateNodeParam(*llmConfig)
		iflytekNode.Data.NodeParam = nodeParam
	}

	// Generate inputs (LLM node receives variable references through inputs)
	iflytekNode.Data.Inputs = g.generateInputsWithMapping(node.Inputs)

	// Generate outputs (LLM node has default output)
	iflytekNode.Data.Outputs = g.generateOutputs(node.Outputs)

	// Generate variable reference information
	iflytekNode.Data.References = g.generateReferences(node.Inputs)

	return iflytekNode, nil
}

// ValidateNode validates unified DSL LLM node
func (g *LLMNodeGenerator) ValidateNode(node models.Node) error {
	if node.Type != models.NodeTypeLLM {
		return fmt.Errorf("node type must be 'llm', got '%s'", node.Type)
	}

	// Validate LLM configuration
	if cfg, ok := common.AsLLMConfig(node.Config); !ok || cfg == nil {
		return fmt.Errorf("LLM node must have LLMConfig")
	}

	return nil
}

// generateNodeParam generates iFlytek SparkAgent LLM node parameters
func (g *LLMNodeGenerator) generateNodeParam(config models.LLMConfig) map[string]interface{} {
	nodeParam := make(map[string]interface{})

	// Basic configuration
	nodeParam["topK"] = config.Parameters.TopK

	// User template (prompt) configuration
	if config.Prompt.UserTemplate != "" {
		nodeParam["template"] = g.convertTemplateFormat(config.Prompt.UserTemplate)
	} else {
		nodeParam["template"] = "无" // Default fallback
	}

	nodeParam["templateErrMsg"] = ""
	nodeParam["temperature"] = config.Parameters.Temperature
	nodeParam["maxTokens"] = config.Parameters.MaxTokens

	// Model configuration
	nodeParam["model"] = g.convertModelProvider(config.Model.Provider)
	nodeParam["domain"] = g.convertModelToDomain(config.Model.Name)
	nodeParam["serviceId"] = g.convertModelToServiceId(config.Model.Name)

	// iFlytek SparkAgent specific configuration (using reasonable defaults)
	nodeParam["modelId"] = g.getModelId(config.Model.Name)
	nodeParam["llmId"] = g.getModelId(config.Model.Name)
	nodeParam["llmIdErrMsg"] = ""
	nodeParam["url"] = g.getModelUrl(config.Model.Name)
	nodeParam["auditing"] = "default"
	nodeParam["multiMode"] = false
	nodeParam["uid"] = "20718349453" // Example value, should be obtained from configuration
	nodeParam["patchId"] = "0"
	nodeParam["appId"] = "12a0a7e2" // Example value, should be obtained from configuration
	nodeParam["isThink"] = false
	nodeParam["searchDisable"] = true

	// Response format: convert from Coze format (0=text, 2=json) to iFlytek format
	nodeParam["respFormat"] = g.convertResponseFormat(config.Parameters.ResponseFormat)

	// Chat history configuration
	nodeParam["chatHistory"] = map[string]interface{}{
		"isEnabled": false,
		"rounds":    1,
	}

	// System template
	if config.Prompt.SystemTemplate != "" {
		nodeParam["systemTemplate"] = g.convertTemplateFormat(config.Prompt.SystemTemplate)
	} else {
		nodeParam["systemTemplate"] = "无" // Default fallback for iFlytek requirement
	}

	return nodeParam
}

// convertModelProvider converts model provider
func (g *LLMNodeGenerator) convertModelProvider(provider string) string {
	switch provider {
	case "openai_compatible":
		return "spark"
	default:
		return "spark" // Default to spark
	}
}

// convertModelToDomain converts model name to domain
func (g *LLMNodeGenerator) convertModelToDomain(modelName string) string {
	switch {
	case strings.Contains(modelName, "deepseek"):
		return "xdeepseekv3"
	default:
		return "xdeepseekv3" // Default value
	}
}

// convertModelToServiceId converts model name to serviceId
func (g *LLMNodeGenerator) convertModelToServiceId(modelName string) string {
	return g.convertModelToDomain(modelName) // Usually the same as domain
}

// getModelId gets model ID
func (g *LLMNodeGenerator) getModelId(modelName string) int {
	switch {
	case strings.Contains(modelName, "deepseek"):
		return 141
	default:
		return 141 // Default value
	}
}

// getModelUrl gets model URL
func (g *LLMNodeGenerator) getModelUrl(modelName string) string {
	return "wss://maas-api.cn-huabei-1.xf-yun.com/v1.1/chat"
}

// convertTemplateFormat converts template format
func (g *LLMNodeGenerator) convertTemplateFormat(template string) string {
	// Convert Dify format to iFlytek SparkAgent format
	result := template

	// Replace {{#nodeId.outputName#}} with {{outputName}}
	result = strings.ReplaceAll(result, "{{#", "{{")
	result = strings.ReplaceAll(result, "#}}", "}}")

	// Use regex to remove nodeId prefix, only keep variable name
	// For example: {{1754269219469.input_01}} -> {{input_01}}

	// Match {{nodeId.variableName}} pattern
	re := regexp.MustCompile(`\{\{[^.]+\.([^}]+)\}\}`)
	result = re.ReplaceAllString(result, "{{$1}}")

	return result
}

// convertResponseFormat converts response format from Coze to iFlytek
func (g *LLMNodeGenerator) convertResponseFormat(cozeFormat int) int {
	switch cozeFormat {
	case 0: // Coze text format
		return 0 // iFlytek text format
	case 2: // Coze JSON format
		return 2 // iFlytek JSON format
	case 1: // Coze markdown format (ignored as per requirements)
		return 0 // Default to text format
	default:
		return 0 // Default to text format
	}
}

// generateInputsWithMapping generates inputs with ID mapping
func (g *LLMNodeGenerator) generateInputsWithMapping(inputs []models.Input) []IFlytekInput {
	iflytekInputs := make([]IFlytekInput, 0, len(inputs))

	for _, input := range inputs {
		iflytekInput := IFlytekInput{
			ID:         g.generateRefID(),
			Name:       input.Name,
			NameErrMsg: "",
			Schema: IFlytekSchema{
				Type:    g.convertDataType(input.Type),
				Default: input.Default,
			},
			FileType: "",
		}

		// Handle variable references, use mapped ID
		if input.Reference != nil {
			// Get mapped node ID
			mappedNodeID := input.Reference.NodeID
			if g.idMapping != nil {
				if mapped, exists := g.idMapping[input.Reference.NodeID]; exists {
					mappedNodeID = mapped
				}
			}

			iflytekInput.Schema.Value = &IFlytekSchemaValue{
				Type: "ref",
				Content: &IFlytekRefContent{
					Name:   input.Reference.OutputName,
					ID:     g.generateRefID(),
					NodeID: mappedNodeID,
				},
				ContentErrMsg: "",
			}
		}

		iflytekInputs = append(iflytekInputs, iflytekInput)
	}

	return iflytekInputs
}

// generateReferences generates variable reference information
func (g *LLMNodeGenerator) generateReferences(inputs []models.Input) []IFlytekReference {
	// Group inputs by node ID
	nodeGroups := make(map[string][]models.Input)

	for _, input := range inputs {
		if input.Reference == nil || input.Reference.NodeID == "" {
			continue
		}

		// Get mapped node ID
		mappedNodeID := input.Reference.NodeID
		if g.idMapping != nil {
			if mapped, exists := g.idMapping[input.Reference.NodeID]; exists {
				mappedNodeID = mapped
			}
		}

		nodeGroups[mappedNodeID] = append(nodeGroups[mappedNodeID], input)
	}

	// Create a parent reference for each node
	references := make([]IFlytekReference, 0, len(nodeGroups))

	for nodeID, nodeInputs := range nodeGroups {
		// Create all output references for this node
		refDetails := make([]IFlytekRefDetail, 0, len(nodeInputs))

		for _, input := range nodeInputs {
			refDetail := IFlytekRefDetail{
				OriginID: nodeID,
				ID:       g.generateRefID(),
				Label:    input.Reference.OutputName,
				Type:     g.convertDataType(input.Type),
				Value:    input.Reference.OutputName,
				FileType: "",
			}
			refDetails = append(refDetails, refDetail)
		}

		// Create parent reference
		ref := IFlytekReference{
			Label:      g.determineLabelByID(nodeID, g.nodeTitleMapping),
			ParentNode: true,
			Value:      nodeID,
			Children: []IFlytekReference{
				{
					Label:      "",
					Value:      "",
					References: refDetails,
				},
			},
		}

		references = append(references, ref)
	}

	return references
}
