package generator

import (
	"ai-agents-transformer/internal/models"
	"ai-agents-transformer/platforms/common"
	"fmt"
	"regexp"
	"strings"
)

// ClassifierNodeGenerator generates classifier decision nodes
type ClassifierNodeGenerator struct {
	*BaseNodeGenerator
	variableSelectorConverter *VariableSelectorConverter
}

func NewClassifierNodeGenerator() *ClassifierNodeGenerator {
	return &ClassifierNodeGenerator{
		BaseNodeGenerator:         NewBaseNodeGenerator(models.NodeTypeClassifier),
		variableSelectorConverter: NewVariableSelectorConverter(),
	}
}

// GenerateNode generates a classifier decision node
func (g *ClassifierNodeGenerator) GenerateNode(node models.Node) (DifyNode, error) {
	if node.Type != models.NodeTypeClassifier {
		return DifyNode{}, fmt.Errorf("unsupported node type: %s, expected: %s", node.Type, models.NodeTypeClassifier)
	}

	// Generate base node structure
	difyNode := g.generateBaseNode(node)

	// Set classifier node specific data - directly set to data field, don't use config wrapper
	g.setClassifierDataFields(&difyNode.Data, node)

	// Restore Dify-specific fields from platform configuration
	if difyConfig := node.PlatformConfig.Dify; difyConfig != nil {
		g.restoreDifyPlatformConfig(difyConfig, &difyNode)
	}

	return difyNode, nil
}

// SetNodeMapping sets node mapping for variable selector converter
func (g *ClassifierNodeGenerator) SetNodeMapping(nodes []models.Node) {
	g.variableSelectorConverter.SetNodeMapping(nodes)
}

// setClassifierDataFields directly sets classifier node data fields, consistent with Dify standard format
func (g *ClassifierNodeGenerator) setClassifierDataFields(data *DifyNodeData, node models.Node) {
	// Extract classifier configuration from unified DSL config
	var classifierConfig *models.ClassifierConfig
	if config, ok := common.AsClassifierConfig(node.Config); ok && config != nil {
		classifierConfig = config
	} else {
		// If config type doesn't match, use default config
		defaultConfig := models.ClassifierConfig{}
		classifierConfig = &defaultConfig
	}

	// Set classification classes
	data.Classes = g.generateClassesFromConfig(*classifierConfig)

	// Need to convert variable reference format: {{Query}} -> {{#nodeId.variableName#}}
	instruction := classifierConfig.Instructions
	if instruction != "" && len(node.Inputs) > 0 {
		// Get variable reference information from first input
		firstInput := node.Inputs[0]
		if firstInput.Reference != nil && firstInput.Reference.NodeID != "" {
			// Construct Dify format variable reference (using temporary nodeID, will be updated later by updateVariableSelectorsWithNewIDs)
			difyReference := fmt.Sprintf("{{#%s.%s#}}", firstInput.Reference.NodeID, firstInput.Reference.OutputName)
			// Find all {{variableName}} format references and replace them
			if strings.Contains(instruction, "{{") {
				// Use more generic replacement logic, find query variable name
				queryVarName := classifierConfig.QueryVariable
				if queryVarName != "" {
					oldPattern := fmt.Sprintf("{{%s}}", queryVarName)
					instruction = strings.ReplaceAll(instruction, oldPattern, difyReference)
				}
				// Compatibility handling: if specific variable name not found, try common patterns
				if strings.Contains(instruction, "{{Query}}") {
					instruction = strings.ReplaceAll(instruction, "{{Query}}", difyReference)
				}
			}
		}
	}

	// Use correct field name
	data.Instruction = instruction
	data.Instructions = "" // Keep empty string, consistent with Dify instance

	// Set model configuration - support complete parameters
	data.Model = g.generateModelFromClassifierConfig(*classifierConfig)

	// Set query variable selector
	data.QueryVariableSelector = g.generateQueryVariableSelector(node)

	// Set vision configuration - dynamically get from config or use reasonable defaults
	data.Vision = g.generateVisionConfig(*classifierConfig)

	// Set topics (ensure field exists, even if empty array)
	data.Topics = make([]string, 0)
}

// generateClassesFromConfig generates classification classes from unified DSL config
func (g *ClassifierNodeGenerator) generateClassesFromConfig(config models.ClassifierConfig) []map[string]interface{} {
	classes := make([]map[string]interface{}, 0)

	// Extract class information from config
	for i, class := range config.Classes {
		// Generate semantic ID based on class characteristics
		classID := g.generateSemanticClassID(class, i+1)

		// Process class name with better default intent handling
		className := g.processClassNameForDify(class)

		difyClass := map[string]interface{}{
			"id":   classID,
			"name": className,
		}
		classes = append(classes, difyClass)
	}

	return classes
}

// generateSemanticClassID generates semantic class ID to preserve meaning
func (g *ClassifierNodeGenerator) generateSemanticClassID(class models.ClassifierClass, fallbackIndex int) string {
	if class.IsDefault {
		return "default"
	}

	// Generate semantic ID based on class name, ensuring uniqueness
	if class.Name != "" {
		// Convert to safe ID format: remove spaces, convert to lowercase
		safeID := strings.ToLower(strings.ReplaceAll(class.Name, " ", "_"))
		// Remove special characters and keep only alphanumeric and underscore
		safeID = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(safeID, "")

		if safeID != "" {
			return safeID
		}
	}

	// Fallback to numeric ID if name processing fails
	return fmt.Sprintf("%d", fallbackIndex)
}

// processClassNameForDify processes class name for Dify with better default intent handling
func (g *ClassifierNodeGenerator) processClassNameForDify(class models.ClassifierClass) string {
	if class.IsDefault {
		// Handle default intent names with Chinese-friendly approach
		switch class.Name {
		case "default":
			return "其他分类" // Chinese-friendly default name
		case "默认意图":
			return "其他分类" // Keep consistent naming
		default:
			// For custom default intent names, preserve but add indicator
			if class.Name != "" {
				return class.Name + "(默认)"
			}
			return "其他分类"
		}
	}

	// For non-default intents, use original name
	return class.Name
}

// generateQueryVariableSelector generates query variable selector
func (g *ClassifierNodeGenerator) generateQueryVariableSelector(node models.Node) []string {
	// Get query variable selector from first input
	if len(node.Inputs) > 0 && node.Inputs[0].Reference != nil {
		// Use variable selector converter to handle field mapping
		valueSelector, err := g.variableSelectorConverter.ConvertVariableReference(node.Inputs[0].Reference)
		if err != nil {
			// Fallback to original logic if conversion fails
			return []string{
				node.Inputs[0].Reference.NodeID,
				node.Inputs[0].Reference.OutputName,
			}
		}
		return valueSelector
	}

	// If no input reference, return empty slice, avoid hardcoded placeholders
	return []string{}
}

// generateModelFromClassifierConfig generates model configuration based on classifier config
func (g *ClassifierNodeGenerator) generateModelFromClassifierConfig(classifierConfig models.ClassifierConfig) map[string]interface{} {
	modelConfig := classifierConfig.Model

	// Build parameters - dynamically get from Parameters struct
	params := map[string]interface{}{}

	// Map iFlytek SparkAgent model parameters to Dify format
	if classifierConfig.Parameters.Temperature > 0 {
		params["temperature"] = classifierConfig.Parameters.Temperature
	}
	if classifierConfig.Parameters.MaxTokens > 0 {
		params["max_tokens"] = classifierConfig.Parameters.MaxTokens
	}
	if classifierConfig.Parameters.TopK > 0 {
		params["top_k"] = classifierConfig.Parameters.TopK
	}

	model := map[string]interface{}{
		"completion_params": params,
		"mode":              "chat", // Dify classifier fixed to use chat mode
	}

	// Use provider information from config
	if modelConfig.Provider != "" {
		// Map iFlytek SparkAgent provider to Dify format
		provider := g.mapProviderToDify(modelConfig.Provider)
		model["provider"] = provider
	} else {
		// If no config, use default provider
		model["provider"] = "langgenius/openai_api_compatible/openai_api_compatible"
	}

	// Use model name from config
	if modelConfig.Name != "" {
		model["name"] = modelConfig.Name
	} else {
		// If no config, use default model
		model["name"] = "xdeepseekv32"
	}

	// Use mode from config
	if modelConfig.Mode != "" {
		model["mode"] = modelConfig.Mode
	}

	return model
}

// mapProviderToDify maps iFlytek SparkAgent provider to Dify format
func (g *ClassifierNodeGenerator) mapProviderToDify(iflytekProvider string) string {
	// Map common providers
	providerMap := map[string]string{
		"xdeepseekv3":  "langgenius/openai_api_compatible/openai_api_compatible",
		"xdeepseekv32": "langgenius/openai_api_compatible/openai_api_compatible",
		"spark":        "langgenius/openai_api_compatible/openai_api_compatible",
		"openai":       "openai",
		"azure_openai": "azure_openai",
	}

	if difyProvider, exists := providerMap[iflytekProvider]; exists {
		return difyProvider
	}

	// If no mapping found, return default provider
	return "langgenius/openai_api_compatible/openai_api_compatible"
}


// generateVisionConfig generates vision configuration
func (g *ClassifierNodeGenerator) generateVisionConfig(classifierConfig models.ClassifierConfig) map[string]interface{} {
	// Get vision configuration from unified DSL config, if not available use reasonable defaults
	visionConfig := map[string]interface{}{
		"enabled": false, // Default to not enabling vision functionality
	}

	return visionConfig
}


// restoreDifyPlatformConfig restores Dify platform-specific configuration
func (g *ClassifierNodeGenerator) restoreDifyPlatformConfig(config map[string]interface{}, node *DifyNode) {
	g.ensureNodeConfig(node)
	
	g.restoreDirectConfigs(config, node)
	g.restoreStringArrayConfigs(config, node)
	g.restoreNodeMetadata(config, node)
}

func (g *ClassifierNodeGenerator) ensureNodeConfig(node *DifyNode) {
	if node.Data.Config == nil {
		node.Data.Config = make(map[string]interface{})
	}
}

func (g *ClassifierNodeGenerator) restoreDirectConfigs(config map[string]interface{}, node *DifyNode) {
	directConfigs := []string{"classes", "instructions", "model", "vision"}
	
	for _, key := range directConfigs {
		if value, exists := config[key]; exists {
			node.Data.Config[key] = value
		}
	}
}

func (g *ClassifierNodeGenerator) restoreStringArrayConfigs(config map[string]interface{}, node *DifyNode) {
	g.restoreQueryVariableSelector(config, node)
	g.restoreTopicsConfig(config, node)
}

func (g *ClassifierNodeGenerator) restoreQueryVariableSelector(config map[string]interface{}, node *DifyNode) {
	if queryVariableSelector, exists := config["query_variable_selector"].([]interface{}); exists {
		selector := g.convertToStringArray(queryVariableSelector)
		node.Data.Config["query_variable_selector"] = selector
	}
}

func (g *ClassifierNodeGenerator) restoreTopicsConfig(config map[string]interface{}, node *DifyNode) {
	if topics, exists := config["topics"].([]interface{}); exists {
		topicsList := g.convertToStringArray(topics)
		node.Data.Config["topics"] = topicsList
	}
}

func (g *ClassifierNodeGenerator) convertToStringArray(items []interface{}) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if str, ok := item.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

func (g *ClassifierNodeGenerator) restoreNodeMetadata(config map[string]interface{}, node *DifyNode) {
	if desc, ok := config["desc"].(string); ok {
		node.Data.Desc = desc
	}
	if title, ok := config["title"].(string); ok {
		node.Data.Title = title
	}
}
